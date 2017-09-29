package pinger

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	// DefaultTimeout is the default timeout to be used to a ping request.
	DefaultTimeout = time.Second

	// DefaultPacketSize is the default packet size for ping requests.
	DefaultPacketSize = uint(56)
)

// Pinger defines the operations of a pinger.
type Pinger interface {
	// Ping accepts a string containing a host and sends ICMP ping packets
	// to that host. It returns a non-nil error in case it fails to parse
	// the given host, or it fails to send the first packet, or the connection
	// to the host is interrupted.
	Ping(host string) error
}

// Options defines the options for a Pinger.
type Options struct {
	// Timeout sets the timeout for each ping request.
	// The default timeout is 1 second.
	Timeout time.Duration

	// Count sets the number of packets to be sent/received.
	// The default count is 0, which means ping requests will be sent
	// indefinitely.
	Count uint

	// PacketSize sets the size of packets to be sent/received.
	// The default packet size is 56 bytes.
	PacketSize uint
}

// setDefaults sets each option to its default value in case one
// hasn't been provided.
func (o *Options) setDefaults() {
	if o.Timeout <= 0 {
		o.Timeout = DefaultTimeout
	}
	if o.Count < 0 {
		o.Count = 0
	}
	if o.PacketSize <= 0 {
		o.PacketSize = DefaultPacketSize
	}
}

// NewPinger accepts an Options object and returns a new Pinger
// configured with the given options.
func NewPinger(opts *Options) Pinger {
	opts.setDefaults()
	return &pinger{
		opts: opts,
	}
}

// pinger is the default implementation for Pinger.
type pinger struct {
	opts *Options
}

// Ping uses Go's x/net/icmp package to send ping packets to the given host.
func (p *pinger) Ping(host string) error {
	addr, err := net.ResolveIPAddr("ip4:icmp", host)
	if err != nil {
		return fmt.Errorf("cannot resolve host %s: %v", host, err)
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "")
	if err != nil {
		return fmt.Errorf("cannot connect to host %s: %v", host, err)
	}
	defer conn.Close()

	fmt.Printf("PING %s: %d data bytes\n", addr, p.opts.PacketSize)

	seq := 0
	stop := false

	for !stop {
		pktBytes, err := createPacket(seq, int(p.opts.PacketSize))
		if err != nil {
			return fmt.Errorf("cannot encode packet: %v", err)
		}

		// Send ping
		start := time.Now()
		if _, err := conn.WriteTo(pktBytes, addr); err != nil {
			return fmt.Errorf("cannot send ping packet for icmp_seq %d: %v", seq, err)
		}

		conn.SetReadDeadline(time.Now().Add(p.opts.Timeout))

		resBytes := make([]byte, len(pktBytes))
		n, ra, err := conn.ReadFrom(resBytes)
		if err != nil {
			if _, ok := err.(*net.OpError); ok {
				fmt.Printf("Request timeout for icmp_seq %d\n", seq)
			} else {
				return fmt.Errorf("cannot read packet for icmp_seq %d: %v", seq, err)
			}
		} else {
			fmt.Printf("%d bytes from %v: icmp_seq=%d time=%.3f\n", n, ra, seq, timeElapsedInMillis(start))
		}

		seq++

		if p.opts.Count != 0 {
			stop = int(p.opts.Count) == seq
		}

		if !stop {
			time.Sleep(time.Second)
		}
	}

	return nil
}

func createPacket(seq int, size int) ([]byte, error) {
	pkt := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   seq,
			Seq:  seq,
			Data: bytes.Repeat([]byte("a"), size),
		},
	}
	return pkt.Marshal(nil)
}

// timeElapsedInMillis returns the amount of milliseconds since start as a float64.
func timeElapsedInMillis(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds()) / (float64(time.Millisecond) / float64(time.Nanosecond))
}
