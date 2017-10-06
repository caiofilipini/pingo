package pinger

import (
	"math"
	"time"
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
	var min, max, sum float64

	rttsInMillis := make([]float64, len(s.rtts))
	for i, rtt := range s.rtts {
		rttInMillis := timeInMillis(rtt)
		rttsInMillis[i] = rttInMillis

		if min == float64(0) || rttInMillis < min {
			min = rttInMillis
		}
		if max == float64(0) || rttInMillis > max {
			max = rttInMillis
		}

		sum += rttInMillis
	}

	avg := sum / float64(len(s.rtts))
	stddev := calcStdDev(avg, rttsInMillis)

	return min, avg, max, stddev
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

// timeInMillis returns the amount of milliseconds in d as a float64.
// TODO: figure out a way to reuse this func in main
func timeInMillis(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / (float64(time.Millisecond) / float64(time.Nanosecond))
}

// calcStdDev calculates the standard deviation for rtts based on the
// given mean.
func calcStdDev(mean float64, rtts []float64) float64 {
	var sumDist float64
	for _, rtt := range rtts {
		sumDist += math.Pow(math.Abs(rtt-mean), 2)
	}
	return math.Sqrt(sumDist / float64(len(rtts)))
}
