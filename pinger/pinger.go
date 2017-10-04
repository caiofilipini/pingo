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
	// Ping accepts a net.Addr representing a host and sends ICMP ping packets
	// to that host.
	Ping(addr net.Addr)

	// Report returns the pair of channels where results will be reported to:
	// 1) a channel of type Ping for successful requests (including temporary errors, e.g. timeouts)
	// 2) a channel of type error for unrecoverable errors
	Report() (<-chan Ping, <-chan error)
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

// Resolve resolves the given host to a net.Addr.
func Resolve(host string) (net.Addr, error) {
	return net.ResolveIPAddr("ip4:icmp", host)
}

// Ping represents a ping request/response.
type Ping struct {
	// Seq is the sequence number.
	Seq int

	// Size is the number of bytes in the response.
	Size int

	// RTT is the duration for the round trip.
	RTT time.Duration

	// Timeout is whether or not the request timed out.
	Timeout bool
}

// NewPinger accepts an Options object and returns a new Pinger
// configured with the given options.
func NewPinger(opts *Options) Pinger {
	opts.setDefaults()
	return &pinger{
		id:         rand.Intn(maxID),
		reportChan: make(chan Ping), // TODO: use buffer?
		errChan:    make(chan error, 1),
		opts:       opts,
	}
}

// pinger is the default implementation for Pinger.
type pinger struct {
	id         int
	reportChan chan Ping
	errChan    chan error
	opts       *Options
}

// Report returns the pair of channels used for reporting.
func (p *pinger) Report() (<-chan Ping, <-chan error) {
	return p.reportChan, p.errChan
}

// Ping uses Go's x/net/icmp package to send ping packets to the given addr.
func (p *pinger) Ping(addr net.Addr) {
	defer close(p.reportChan)

	conn, err := icmp.ListenPacket("ip4:icmp", "")
	if err != nil {
		p.errChan <- fmt.Errorf("cannot connect to addr %s: %v", addr, err)
		return
	}
	defer conn.Close()

	seq := 0
	for {
		pktSize, err := p.send(conn, addr, seq)
		if err != nil {
			p.errChan <- fmt.Errorf("cannot send ping packet for icmp_seq %d: %v", seq, err)
			return
		}

		ping, err := p.recv(conn, seq, pktSize)
		if err != nil {
			p.errChan <- err
			return
		}

		p.reportChan <- ping
		seq++

		if p.opts.Count != 0 && int(p.opts.Count) == seq {
			break
		}

		time.Sleep(time.Second)
	}
}

func (p *pinger) send(conn net.PacketConn, addr net.Addr, seq int) (int, error) {
	pktBytes, err := createPacket(p.id, seq, int(p.opts.PacketSize))
	if err != nil {
		return 0, fmt.Errorf("cannot encode packet: %v", err)
	}

	if _, err := conn.WriteTo(pktBytes, addr); err != nil {
		return 0, fmt.Errorf("cannot send ping packet for icmp_seq %d: %v", seq, err)
	}

	return len(pktBytes), nil
}

func (p *pinger) recv(conn net.PacketConn, seq int, pktSize int) (Ping, error) {
	conn.SetReadDeadline(time.Now().Add(p.opts.Timeout))
	resBytes := make([]byte, pktSize)
	n, _, err := conn.ReadFrom(resBytes)
	if err != nil {
		if neterr, ok := err.(*net.OpError); ok && neterr.Timeout() {
			return Ping{
				Seq:     seq,
				Timeout: true,
			}, nil
		} else {
			return Ping{}, fmt.Errorf("cannot read packet for icmp_seq %d: %v", seq, err)
		}
	}

	res, err := p.parse(seq, resBytes)
	if err != nil {
		return Ping{}, err
	}

	return Ping{
		Seq:  seq,
		Size: n,
		RTT:  time.Since(bytesToTime(res.Data[:timeByteSize])),
	}, nil
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
