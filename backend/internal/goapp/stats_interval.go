package goapp

import "math"

type ProportionStat struct {
	Count    int64   `json:"count"`
	Total    int64   `json:"total"`
	Rate     float64 `json:"rate"`
	CI95Low  float64 `json:"ci95_low"`
	CI95High float64 `json:"ci95_high"`
}

func newProportionStat(count int64, total int64) *ProportionStat {
	stat := &ProportionStat{Count: count, Total: total}
	if total <= 0 || count < 0 {
		return stat
	}
	stat.Rate = rounded(float64(count)/float64(total)*100, 2)
	if count > total {
		return stat
	}

	p := float64(count) / float64(total)
	z := 1.959963984540054
	z2 := z * z
	n := float64(total)
	denom := 1 + z2/n
	center := (p + z2/(2*n)) / denom
	margin := (z * math.Sqrt((p*(1-p)+z2/(4*n))/n)) / denom
	low := math.Max(0, center-margin)
	high := math.Min(1, center+margin)

	stat.CI95Low = rounded(low*100, 2)
	stat.CI95High = rounded(high*100, 2)
	return stat
}
