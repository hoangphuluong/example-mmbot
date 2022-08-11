package service

import "math"

func GetFormattedPrice(origPrice float64, priceDecimals int) float64 {
	if priceDecimals > 0 {
		return origPrice / math.Pow(10, float64(priceDecimals))
	} else {
		return origPrice
	}
}
