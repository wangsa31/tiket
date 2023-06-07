package utils

import (
	"strconv"
	"strings"
)

/*
for convert time example time 1:00:00 to string become 1 j , time 1:30:00 become 1 j 30 m
if time under 60 minutes will become xx minute (example 8 minut)
*/
func Format_time(time string) string {
	parts := strings.Split(time, ":")
	hoursStr := parts[0]
	minutesStr := parts[1]

	hours, _ := strconv.Atoi(hoursStr)
	minutes, _ := strconv.Atoi(minutesStr)

	result := ""
	if hours > 0 {
		result += strconv.Itoa(hours) + " j "
	}

	result += strconv.Itoa(minutes) + " m"

	return strings.TrimSpace(result)
}
