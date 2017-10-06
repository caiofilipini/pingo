package pinger

import (
	"time"

	"github.com/caiofilipini/go-ping/math"
)

// Stats stores the packet statistics.
type Stats struct {
	totalCount   int
	successCount int
	rtts         []time.Duration
}

// Transmitted returns the total number of packets transmitted.
func (s *Stats) Transmitted() int {
	return s.totalCount
}

// Received returns the total number of packets successfully received back.
func (s *Stats) Received() int {
	return s.successCount
}

// PacketLoss calculates and returns the percentage of packets that have been
// lost (i.e. a packet was sent, but a reply was not received due to a timeout).
func (s *Stats) PacketLoss() float64 {
	return (1 - float64(s.successCount)/float64(s.totalCount)) * 100
}

// RTTStats calculates and returns, respectively, the min, average, max and
// standard deviation for round-trip latencies.
func (s *Stats) RTTStats() (float64, float64, float64, float64) {
	rttsInMillis := make([]float64, len(s.rtts))
	for i, rtt := range s.rtts {
		rttsInMillis[i] = math.TimeInMillis(rtt)
	}

	return math.Min(rttsInMillis),
		math.Mean(rttsInMillis),
		math.Max(rttsInMillis),
		math.StdDev(rttsInMillis)
}

// incSuccess increments both the totalCount and the successCount,
// as well as appends the given rtt to the list of rtts.
func (s *Stats) incSuccess(rtt time.Duration) {
	s.totalCount++
	s.successCount++
	s.rtts = append(s.rtts, rtt)
}

// incTimeout increments only the totalCount.
func (s *Stats) incTimeout() {
	s.totalCount++
}
