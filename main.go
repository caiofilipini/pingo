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
		fmt.Printf("failed to resolve host %s: %v\n", host, err)
		os.Exit(2)
	}

	pinger := pinger.NewPinger(&pinger.Options{
		Count:      *count,
		PacketSize: *packetSize,
		Timeout:    time.Duration(*timeout) * time.Second,
	})

	done := make(chan struct{})
	results, errors := pinger.Report()
	stop := false

	fmt.Printf("PING %s: %d data bytes\n", addr, *packetSize)

	go func(done chan struct{}) {
		pinger.Ping(addr)
		done <- struct{}{}
	}(done)

	for !stop {
		select {
		case <-done:
			stop = true
		case res, ok := <-results:
			if !ok {
				continue
			}

			if res.Timeout {
				fmt.Printf("Request timeout for icmp_seq %d\n", res.Seq)
			} else {
				fmt.Printf("%d bytes from %v: icmp_seq=%d time=%.3f ms\n", res.Size, addr, res.Seq, timeInMillis(res.RTT))
			}
		case err := <-errors:
			fmt.Printf("failed to ping %s: %v\n", host, err)
			os.Exit(2)
		}
	}

	printStats(host, pinger.Stats())
}

func printStats(host string, stats pinger.Stats) {
	fmt.Println()
	fmt.Printf("--- %s ping statistics ---\n", host)
	fmt.Printf(
		"%d packets transmitted, %d packets received, %.1f%% packet loss\n",
		stats.Transmitted(),
		stats.Received(),
		stats.PacketLoss(),
	)

	min, avg, max, stddev := stats.RTTStats()
	fmt.Printf("round-trip min/avg/max/stddev = %.3f/%.3f/%.3f/%.3f ms\n", min, avg, max, stddev)
}

// timeInMillis returns the amount of milliseconds in d as a float64.
func timeInMillis(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / (float64(time.Millisecond) / float64(time.Nanosecond))
}
