package model

import "math"

type Amount int64

func FromFloat64(f float64) Amount {
	return Amount(math.Round(f * 100))
}

func (a Amount) ToFloat64() float64 {
	return float64(a) / 100.0
}

func FromInt64(i int64) Amount {
	return Amount(i)
}

func (a Amount) Int64() int64 {
	return int64(a)
}
