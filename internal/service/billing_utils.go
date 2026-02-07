package service

import "time"

/*
weekSince internal helper method to calculte week duration between 2 different time
*/
func weekSince(start, now time.Time) int {
	if now.Before(start) {
		return 0
	}

	duration := now.Sub(start)
	weeks := int(duration.Hours() / (24 * 7))

	return weeks + 1
}
