package domain

import "math"

type Money struct {
	Cents int64 `json:"cents"`
}

func NewMoneyFromFloat(value float64) Money {
	return Money{Cents: int64(math.Round(value * 100))}
}

func (m Money) Add(other Money) Money {
	return Money{Cents: m.Cents + other.Cents}
}

func (m Money) Sub(other Money) Money {
	return Money{Cents: m.Cents - other.Cents}
}

func (m Money) Float64() float64 {
	return float64(m.Cents) / 100
}

func (m Money) DiffAbs(other Money) Money {
	diff := m.Cents - other.Cents
	if diff < 0 {
		diff = -diff
	}
	return Money{Cents: diff}
}

func (m Money) WithinCents(other Money, tolerance int64) bool {
	return m.DiffAbs(other).Cents <= tolerance
}
