package tcpdump_demo

import (
	"testing"

	"fmt"
	"math/big"
	"net"
	"syscall"

	"github.com/google/gopacket/afpacket"
	"golang.org/x/net/bpf"
)

// 1. python -m SimpleHTTPServer 9007 start the http server
// 2. start this test program
// 3. then use `curl http://127.0.0.1:9007` to start to debug this program and watch the output
func TestBpfGenProg(t *testing.T) {
	bpfIns, err := generateBpfProgram()
	if err != nil {
		t.Error(err)
		return
	}
	for _, item := range bpfIns {
		fmt.Println(item.Op, item.Jt, item.Jf, item.K)
	}

	rawSocket, err := afpacket.NewTPacket(
		// afpacket.OptPollTimeout(10*time.Second),
		// This setup will require ~4Mb that is mmap'd into the process virtual space
		// More information here: https://www.kernel.org/doc/Documentation/networking/packet_mmap.txt
		afpacket.OptFrameSize(4096),
		afpacket.OptBlockSize(4096*128),
		afpacket.OptNumBlocks(8),
		afpacket.OptInterface("lo"),
	)
	if err != nil {
		t.Error(err)
		return
	}
	err = rawSocket.SetBPF(bpfIns)
	for {

		data, stats, err2 := rawSocket.ZeroCopyReadPacketData()

		// Immediately retry for EAGAIN
		if err2 == syscall.EAGAIN {
			continue
		}

		if err2 == afpacket.ErrTimeout {
			return
		}

		if err2 != nil {
			t.Error(err)
			return
		}
		fmt.Println(stats.Length, stats.Timestamp, stats.InterfaceIndex)
		EthHlen := 14

		totalLength := int(data[EthHlen+2])              // load MSB
		totalLength = totalLength << 8                   // shift MSB
		totalLength = totalLength + int(data[EthHlen+3]) // add LSB

		ipHeaderLength := int(data[EthHlen])   // load Byte
		ipHeaderLength = ipHeaderLength & 0x0F // mask bits 0..3
		ipHeaderLength = ipHeaderLength << 2   // shift to obtain length

		ipSrcStr := data[EthHlen+12 : EthHlen+16] // ip source offset 12..15
		ipDstStr := data[EthHlen+16 : EthHlen+20] // ip dest   offset 16..19

		tcpHeaderLength := int(data[EthHlen+ipHeaderLength+12]) // load Byte
		tcpHeaderLength = tcpHeaderLength & 0xF0                // mask bit 4..7
		tcpHeaderLength = tcpHeaderLength >> 2                  // SHR 4 ; SHL 2 -> SHR 2

		portSrcStr := data[EthHlen+ipHeaderLength : EthHlen+ipHeaderLength+2]
		portDstStr := data[EthHlen+ipHeaderLength+2 : EthHlen+ipHeaderLength+4]
		srcAddr := &net.TCPAddr{IP: ipSrcStr, Port: int(big.NewInt(0).SetBytes(portSrcStr).Uint64())}
		dstAddr := &net.TCPAddr{IP: ipDstStr, Port: int(big.NewInt(0).SetBytes(portDstStr).Uint64())}

		// payloadOffset := EthHlen + ipHeaderLength + tcpHeaderLength
		// payloadString := data[(payloadOffset):stats.Length]
		fmt.Println(srcAddr.String(), dstAddr.String())
	}
}

// tcpdump -i lo dst host 127.0.0.1 and dst port 9007 -d
func generateBpfProgram() ([]bpf.RawInstruction, error) {
	return bpf.Assemble([]bpf.Instruction{
		// https://www.networxsecurity.org/members-area/glossary/e/ethertype.html
		// https://en.wikipedia.org/wiki/Ethernet_frame
		// 6 src mac, 6 dest mac, 2 ethernet type
		// 0x0800 IPv4
		// 0x86dd IPv6
		//(000) ldh      [12] -- load Ethertype
		bpf.LoadAbsolute{Size: 2, Off: 12},
		//(001) jeq      #0x800           jt 2    jf 14 -- if ethernet packet, goto 2, else 14
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x800, SkipTrue: 0, SkipFalse: 12},
		// https://en.wikipedia.org/wiki/Transmission_Control_Protocol
		//(002) ld      [30] -- load dst addr
		bpf.LoadAbsolute{Size: 4, Off: 30},
		//(003) jeq      #0x7f000001      jt 4    jf 14 -- if address is correct, then 4, else 14
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x7f000001, SkipTrue: 0, SkipFalse: 10},
		// https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml
		// as protocol numbers above,
		// 132 == 0x84 -> sctp
		// 6   == 0x6  -> tcp
		// 17  == 0x11 -> udp
		//(004) ldb      [23] -- load ip Next Header
		bpf.LoadAbsolute{Size: 1, Off: 23},
		//(005) jeq      #0x84            jt 8   jf 6 -- if sctp, goto 8
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x84, SkipTrue: 2, SkipFalse: 0},
		//(006) jeq      #0x6             jt 8   jf 7  -- if TCP, goto 8
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x6, SkipTrue: 1, SkipFalse: 0},
		//(007) jeq      #0x11            jt 8   jf 14 -- if UDP, goto 8, else drop
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x11, SkipTrue: 0, SkipFalse: 6},
		// https://en.wikipedia.org/wiki/Internet_Protocol_version_4#Fragment_offset
		//(008) ldh      [20] -- load Fragment Offset
		bpf.LoadAbsolute{Size: 2, Off: 20},
		//(009) jset     #0x1fff          jt 14   jf 10  -- use 0x1fff as mask for fragment offset, if != 0, drop
		bpf.JumpIf{Cond: bpf.JumpBitsSet, Val: 0x1fff, SkipTrue: 4, SkipFalse: 0},
		//(010) ldxb     4*([14]&0xf) -- x = IP header length // remove 14 eth frame length and load one byte and stipe high 4 bits
		bpf.LoadMemShift{Off: 14},
		//(011) ldh      [x + 16]   -- load dst port
		// 14 + x + 2 // 2 to skip src addr
		bpf.LoadIndirect{Size: 2, Off: 16},
		//(012) jeq      #0x232f            jt 13   jf 14 -- if 9007 port, capture
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x232f, SkipTrue: 0, SkipFalse: 1},
		//power(2, 18) --> 256K
		//https://github.com/the-tcpdump-group/tcpdump/blob/tcpdump-4.9/netdissect.h#L263
		//(013) ret      #262144 -- capture
		bpf.RetConstant{Val: 262144},
		//(014) ret      #0 -- drop
		bpf.RetConstant{Val: 0},
	})
}
