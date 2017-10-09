package math

import (
	"testing"
)

func TestMin(t *testing.T) {
	tests := []struct {
		desc       string
		population []float64
		expected   float64
	}{
		{
			desc:       "returns zero for an empty population",
			population: []float64{},
			expected:   0,
		},
		{
			desc:       "returns the single value",
			population: []float64{4.2},
			expected:   4.2,
		},
		{
			desc:       "returns the negative number",
			population: []float64{-3.14, 0, 23.34},
			expected:   -3.14,
		},
		{
			desc:       "returns 0",
			population: []float64{0, 4.2, 4.21, 4.22},
			expected:   0,
		},
		{
			desc:       "returns the lowest positive number",
			population: []float64{4.2, 4.21, 4.22},
			expected:   4.2,
		},
		{
			desc:       "returns 42.42 when they're all the same",
			population: []float64{42.42, 42.42, 42.42},
			expected:   42.42,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			min := Min(tc.population)
			if min != tc.expected {
				t.Errorf("wanted %f, got %f", tc.expected, min)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		desc       string
		population []float64
		expected   float64
	}{
		{
			desc:       "returns zero for an empty population",
			population: []float64{},
			expected:   0,
		},
		{
			desc:       "returns the single value",
			population: []float64{4.2},
			expected:   4.2,
		},
		{
			desc:       "returns the highest negative number",
			population: []float64{-3.14, -2.23, -1.42},
			expected:   -1.42,
		},
		{
			desc:       "returns the highest positive number",
			population: []float64{4.2, 4.21, 4.22},
			expected:   4.22,
		},
		{
			desc:       "returns 42.42 when they're all the same",
			population: []float64{42.42, 42.42, 42.42},
			expected:   42.42,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			max := Max(tc.population)
			if max != tc.expected {
				t.Errorf("wanted %f, got %f", tc.expected, max)
			}
		})
	}
}

func TestMean(t *testing.T) {
	tests := []struct {
		desc       string
		population []float64
		expected   float64
	}{
		{
			desc:       "returns zero for an empty population",
			population: []float64{},
			expected:   0,
		},
		{
			desc:       "returns the single value in the population",
			population: []float64{4.22},
			expected:   4.22,
		},
		{
			desc:       "returns the average of the population",
			population: []float64{-1.1, 5.8, 9.9, 1.4},
			expected:   4.0, // -1.1 + 5.8 + 9.9 + 1.4 = 16.0 => 16.0 / 4 = 4.0
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			mean := Mean(tc.population)
			if mean != tc.expected {
				t.Errorf("wanted %f, got %f", tc.expected, mean)
			}
		})
	}
}

func TestStdDev(t *testing.T) {
	tests := []struct {
		desc       string
		population []float64
		expected   float64
	}{
		{
			desc:       "returns zero for an empty population",
			population: []float64{},
			expected:   0,
		},
		{
			desc:       "returns zero for a single value",
			population: []float64{4.22},
			expected:   0,
		},
		{
			desc:       "returns the standard deviation of the population",
			population: []float64{3.11, 4.22, 5.33, 6.44},
			expected:   1.24,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			stddev := round(StdDev(tc.population))
			if stddev != tc.expected {
				t.Errorf("wanted %f, got %f", tc.expected, stddev)
			}
		})
	}
}

// round truncates the given float64 to 2 decimal places.
func round(n float64) float64 {
	return float64(int(n*100)) / 100
}
