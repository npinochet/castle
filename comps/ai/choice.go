package ai

import "math/rand"

type Choice []struct {
	Weight float64
	Act    func()
}

func (c Choice) Play() {
	totalWeight := 0.0
	for _, c := range c {
		totalWeight += c.Weight
	}

	r := rand.Float64() * totalWeight
	for _, c := range c {
		if r -= c.Weight; r <= 0 {
			c.Act()

			return
		}
	}
}
