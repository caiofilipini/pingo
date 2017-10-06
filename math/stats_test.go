package math

import "testing"

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
