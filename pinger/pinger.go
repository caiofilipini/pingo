package pinger

import (
	"fmt"
	"math/rand"
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

	// maxID is the maximum value for a packet identifier
	// (i.e. max 16 bits integer = 65536).
	maxID = 0xffff

	// ipv4Proto is the type used for parsing the echo response.
	ipv4Proto = 1

	// timeByteSize is the number of bytes used to represent the timestamp
	// in the payload.
	timeByteSize = 8
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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
		id:   rand.Intn(maxID),
		opts: opts,
	}
}

// pinger is the default implementation for Pinger.
type pinger struct {
	id   int
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

	seq := 0
	stop := false

	fmt.Printf("PING %s: %d data bytes\n", addr, p.opts.PacketSize)

	for !stop {
		pktSize, err := p.send(conn, addr, seq)
		if err != nil {
			return fmt.Errorf("cannot send ping packet for icmp_seq %d: %v", seq, err)
		}

		if err := p.recv(conn, seq, pktSize); err != nil {
			return err
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

func (p *pinger) send(conn *icmp.PacketConn, addr net.Addr, seq int) (int, error) {
	pktBytes, err := createPacket(p.id, seq, int(p.opts.PacketSize))
	if err != nil {
		return 0, fmt.Errorf("cannot encode packet: %v", err)
	}

	if _, err := conn.WriteTo(pktBytes, addr); err != nil {
		return 0, fmt.Errorf("cannot send ping packet for icmp_seq %d: %v", seq, err)
	}

	return len(pktBytes), nil
}

func (p *pinger) recv(conn *icmp.PacketConn, seq int, pktSize int) error {
	conn.SetReadDeadline(time.Now().Add(p.opts.Timeout))

	resBytes := make([]byte, pktSize)
	n, ra, err := conn.ReadFrom(resBytes)
	if err != nil {
		if neterr, ok := err.(*net.OpError); ok && neterr.Timeout() {
			fmt.Printf("Request timeout for icmp_seq %d\n", seq)
			return nil
		} else {
			return fmt.Errorf("cannot read packet for icmp_seq %d: %v", seq, err)
		}
	}

	res, err := p.parse(seq, resBytes)
	if err != nil {
		return err
	}

	fmt.Printf("%d bytes from %v: icmp_seq=%d time=%.3f ms\n", n, ra, seq, timeElapsedInMillis(bytesToTime(res.Data[:timeByteSize])))

	return nil
}

func (p *pinger) parse(seq int, resBytes []byte) (*icmp.Echo, error) {
	res, err := icmp.ParseMessage(ipv4Proto, resBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse response for icmp_seq %d: %v", seq, err)
	}

	if res.Type != ipv4.ICMPTypeEchoReply {
		return nil, fmt.Errorf("cannot parse response for icmp_seq %d: %T", seq, res.Body)
	}
	pkt, ok := res.Body.(*icmp.Echo)
	if !ok {
		return nil, fmt.Errorf("unexpected response type for icmp_seq %d: %T", seq, res.Body)
	}

	if pkt.ID != p.id || pkt.Seq != seq {
		return nil, fmt.Errorf("unexpected response for icmp_seq %d: %v", seq, pkt)
	}

	return pkt, nil
}

func createPacket(id int, seq int, size int) ([]byte, error) {
	payload := timeToBytes(time.Now())

	remaining := size - len(payload)
	if remaining > 0 {
		trail := make([]byte, remaining)
		for i := 0; i < len(trail); i++ {
			trail[i] = 1
		}
		payload = append(payload, trail...)
	}

	pkt := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  seq,
			Data: payload,
		},
	}
	return pkt.Marshal(nil)
}

// This function was copied from https://github.com/tatsushid/go-fastping and adapted.
func timeToBytes(t time.Time) []byte {
	nsec := t.UnixNano()
	b := make([]byte, timeByteSize)
	for i := uint8(0); i < timeByteSize; i++ {
		b[i] = byte((nsec >> ((7 - i) * timeByteSize)) & 0xff)
	}
	return b
}

// This function was copied from https://github.com/tatsushid/go-fastping and adapted.
func bytesToTime(b []byte) time.Time {
	var nsec int64
	for i := uint8(0); i < timeByteSize; i++ {
		nsec += int64(b[i]) << ((7 - i) * timeByteSize)
	}
	return time.Unix(nsec/1000000000, nsec%1000000000)
}

// timeElapsedInMillis returns the amount of milliseconds since start as a float64.
func timeElapsedInMillis(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds()) / (float64(time.Millisecond) / float64(time.Nanosecond))
}
