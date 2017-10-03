package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caiofilipini/go-ping/pinger"
)

func main() {
	bin := os.Args[0]
	count := flag.Uint("c", 0, fmt.Sprintf("number of packets to be sent and received; if not specified, %s will send requests until interrupted", bin))
	packetSize := flag.Uint("s", pinger.DefaultPacketSize, "number of data bytes to be sent in each request")
	timeout := flag.Uint("t", uint(pinger.DefaultTimeout.Seconds()), "timeout in seconds for each request")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s host\n", bin)
		flag.PrintDefaults()
		os.Exit(2)
	}

	host := flag.Arg(0)
	addr, err := pinger.Resolve(host)
	if err != nil {
		fmt.Printf("failed to resolve host %s: %v", host, err)
		os.Exit(2)
	}

	pinger := pinger.NewPinger(&pinger.Options{
		Count:      *count,
		PacketSize: *packetSize,
		Timeout:    time.Duration(*timeout) * time.Second,
	})

	if err := pinger.Ping(addr); err != nil {
		fmt.Printf("failed to ping %s: %v\n", addr, err)
		os.Exit(2)
	}
}
