package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"log"
	"net/netip"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/rlimit"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang tcpconn ./kern/tcp_connect.c -- -D__TARGET_ARCH_x86 -I../ebpf_headers -Wall

var obj tcpconnObjects

func main() {
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatalf("Failed to remove rlimit memlock: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := loadTcpconnObjects(&obj, nil); err != nil {
		log.Fatalf("Failed to load bpf obj: %v", err)
	}
	defer obj.Close()

	// prepare programs for bpf_tail_call()
	prog := obj.tcpconnPrograms.HandleNewConnection
	key := uint32(0)
	if err := obj.tcpconnMaps.Progs.Update(key, prog, ebpf.UpdateAny); err != nil {
		log.Printf("Failed to prepare prog(handle_new_connection): %v", err)
		return
	}

	if kp, err := link.Kprobe("tcp_connect", obj.K_tcpConnect, nil); err != nil {
		log.Printf("Failed to attach kprobe(tcp_connect): %v", err)
		return
	} else {
		defer kp.Close()
		log.Printf("Attached kprobe(tcp_connect)")
	}

	if kp, err := link.Kprobe("inet_csk_complete_hashdance", obj.K_icskCompleteHashdance, nil); err != nil {
		log.Printf("Failed to attach kprobe(inet_csk_complete_hashdance): %v", err)
		return
	} else {
		defer kp.Close()
		log.Printf("Attached kprobe(inet_csk_complete_hashdance)")
	}

	go handlePerfEvent(ctx, obj.Events, obj.Logs)

	<-ctx.Done()
}

func updateTailCall(prog *ebpf.Program) {
	key := uint32(0)
	if err := obj.tcpconnMaps.Progs.Update(key, prog, ebpf.UpdateAny); err != nil {
		log.Printf("Failed to prepare prog(handle_new_connection): %v", err)
		return
	}
}

func handlePerfEvent(ctx context.Context, events *ebpf.Map, logs *ebpf.Map) {
	eventReader, err := perf.NewReader(events, 4096)
	if err != nil {
		log.Printf("Failed to create perf-event reader: %v", err)
		return
	}
	logReader, err := perf.NewReader(logs, 4096)
	if err != nil {
		log.Printf("Failed to create perf-event reader: %v", err)
		return
	}

	log.Printf("Listening events...")

	go func() {
		<-ctx.Done()
		eventReader.Close()
	}()

	var ev struct {
		Saddr, Daddr [4]byte
		Sport, Dport uint16
	}
	var str [4]byte
	var counter int32 = 0
	go func() {
		for {
			logEntry, err := logReader.Read()
			if err != nil {
				if errors.Is(err, perf.ErrClosed) {
					return
				}
				log.Printf("Reading perf-event: %v", err)
			}

			if logEntry.LostSamples != 0 {
				log.Printf("Lost %d events", logEntry.LostSamples)
			}
			if counterTmp := atomic.LoadInt32(&counter); counterTmp%100 == 99 {
				updateTailCall(obj.HandleNewConnection)
			}

			binary.Read(bytes.NewBuffer(logEntry.RawSample), binary.LittleEndian, &str)
			log.Printf("received from :%s", string(str[:]))

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()
	for {
		atomic.AddInt32(&counter, 1)
		event, err := eventReader.Read()
		if err != nil {
			if errors.Is(err, perf.ErrClosed) {
				return
			}

			log.Printf("Reading perf-event: %v", err)
		}

		if event.LostSamples != 0 {
			log.Printf("Lost %d events", event.LostSamples)
		}

		binary.Read(bytes.NewBuffer(event.RawSample), binary.LittleEndian, &ev)

		if counterTmp := atomic.LoadInt32(&counter); counterTmp%100 == 50 {
			updateTailCall(obj.FakeNewConnection)
		}

		log.Printf("new tcp connection: %s:%d -> %s:%d",
			netip.AddrFrom4(ev.Saddr), ev.Sport,
			netip.AddrFrom4(ev.Daddr), ev.Dport)

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}
