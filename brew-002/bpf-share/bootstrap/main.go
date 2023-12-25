package main

import (
	"fmt"
	"math/rand"
	"time"
)

// output target file path: /root/bpf-share/bootstrap/bootstrap
// docker run -v /home/martinho:/root --privileged=true -ti --network=host --pid=host --ipc=host quay.io/iovisor/bpftrace:latest bash
// bpftrace  -e 'uprobe:/root/bpf-share/bootstrap/bootstrap:main.sum {printf("pid:%d, arg0:%lld, arg1:%ld\n", pid, reg("ax"), arg1);}'
func main() {
	for {
		a := rand.Int()
		b := rand.Int()
		c := sum(a, b)
		fmt.Println("sum up is: ", c)
		// fmt.Println("a, ", a, "b,", b)
		time.Sleep(1 * time.Second)
	}
}

//go:noinline
func sum(a, b int) uint64 {
	c := uint64(a) + uint64(b)
	return c
}
