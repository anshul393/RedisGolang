package main

import (
	"errors"
	"math"
	"time"
)

func Incr(num int) (int, error) {
	if num > math.MaxInt-1 {
		return 0, errors.New("out of range")
	}

	return num + 1, nil
}

func Decr(num int) (int, error) {
	if num < math.MinInt+1 {
		return 0, errors.New("out of range")
	}

	return num - 1, nil
}

// isExpired , // can be returned or not
func Validate(v *StoreData) (bool, bool) {

	if v.Expiry.IsZero() {
		return false, true
	}
	currentTime := time.Now()
	isExpired := false
	isValid := false

	if v.Expiry.Compare(currentTime) <= 0 {
		isExpired = true
	}

	if v.Expiry.Compare(currentTime) == 0 {
		isValid = true
	}

	isExpired = true
	isValid = true

	return isExpired, isValid
}
