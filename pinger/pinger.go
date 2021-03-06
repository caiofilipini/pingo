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
	// to that host. Ping is a blocking operation.
	Ping(addr net.Addr)

	// Stop signals the Pinger to stop sending ping requests to the host.
	// After a call to Stop(), Ping() is expected to return.
	Stop()

	// Report returns the pair of channels where results will be reported to:
	// 1) a channel of type Ping for successful requests (including temporary errors, e.g. timeouts)
	// 2) a channel of type error for unrecoverable errors
	Report() (<-chan Ping, <-chan error)

	// Stats returns the packet statistics accumulated for the host being
	// pinged.
	Stats() Stats
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
		opts:       opts,
		reportChan: make(chan Ping), // TODO: use buffer?
		errChan:    make(chan error, 1),
		stop:       make(chan struct{}, 1),
		stats:      &Stats{},
		clock:      defaultClock{},
	}
}

// pinger is the default implementation for Pinger.
type pinger struct {
	id         int
	opts       *Options
	reportChan chan Ping
	errChan    chan error
	stats      *Stats
	stop       chan struct{}
	clock      clock
}

// Report returns the pair of channels used for reporting.
func (p *pinger) Report() (<-chan Ping, <-chan error) {
	return p.reportChan, p.errChan
}

// Stats returns the stats for the pinger.
func (p *pinger) Stats() Stats {
	return *p.stats
}

// Ping uses Go's x/net/icmp package to send ping packets to the given addr.
// Ping is a blocking operation.
func (p *pinger) Ping(addr net.Addr) {
	defer close(p.reportChan)
	defer close(p.errChan)

	conn, err := icmp.ListenPacket("ip4:icmp", "")
	if err != nil {
		p.errChan <- fmt.Errorf("cannot connect to addr %s: %v", addr, err)
		return
	}
	defer conn.Close()

	seq := 0
	for {
		select {
		case <-p.stop:
			return
		default:
			ping, err := p.ping(conn, addr, seq)
			if err != nil {
				p.errChan <- err
				return
			}

			p.reportChan <- ping
			seq++

			if p.opts.Count != 0 && int(p.opts.Count) == seq {
				p.Stop()
			} else {
				time.Sleep(time.Second)
			}
		}
	}
}

// Stop signals the Pinger to stop sending ping requests to the host.
func (p *pinger) Stop() {
	p.stop <- struct{}{}
}

func (p *pinger) ping(conn net.PacketConn, addr net.Addr, seq int) (Ping, error) {
	pktSize, err := p.send(conn, addr, seq)
	if err != nil {
		return Ping{}, fmt.Errorf("cannot send ping packet for icmp_seq %d: %v", seq, err)
	}

	return p.recv(conn, seq, pktSize)
}

func (p *pinger) send(conn net.PacketConn, addr net.Addr, seq int) (int, error) {
	pktBytes, err := createPacket(p.id, seq, int(p.opts.PacketSize), p.clock.Now())
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
			p.stats.incTimeout()
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

	rtt := p.clock.Now().Sub(bytesToTime(res.Data[:timeByteSize]))
	p.stats.incSuccess(rtt)

	return Ping{
		Seq:  seq,
		Size: n,
		RTT:  rtt,
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

func createPacket(id int, seq int, size int, now time.Time) ([]byte, error) {
	payload := timeToBytes(now)

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
