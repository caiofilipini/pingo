package math

import "math"

// Min returns the minimum value in the given population.
func Min(population []float64) float64 {
	return reduce(population, math.MaxFloat64, func(v float64, acc float64) float64 {
		return math.Min(acc, v)
	})
}

// Max returns the maximum value in the given population.
func Max(population []float64) float64 {
	return reduce(population, -math.MaxFloat64, func(v float64, acc float64) float64 {
		return math.Max(acc, v)
	})
}

// Mean calculates the mean value for the given population.
func Mean(population []float64) float64 {
	if len(population) == 0 {
		return 0
	}
	sum := reduce(population, 0, func(v float64, acc float64) float64 {
		return acc + v
	})
	return sum / float64(len(population))
}

// StdDev calculates the standard deviation for the given population.
func StdDev(population []float64) float64 {
	mean := Mean(population)
	if mean == 0 {
		return 0
	}

	sumDist := reduce(population, 0, func(v float64, acc float64) float64 {
		return acc + math.Pow(math.Abs(v-mean), 2)
	})
	return math.Sqrt(sumDist / float64(len(population)))
}

type reducer func(v float64, acc float64) float64

func reduce(population []float64, acc float64, fn reducer) float64 {
	if len(population) == 0 {
		return 0
	}

	for _, v := range population {
		acc = fn(v, acc)
	}

	return acc
}
