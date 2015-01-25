package main

import (
	"flag"
	"time"
)

var forceTime = flag.String("time", "", "set the clock to this time")

func timeNow() time.Time {
	if *forceTime != "" {
		if t, err := time.Parse(time.RFC3339, *forceTime); err != nil {
			panic(err)
		} else {
			return t
		}
	} else {
		return time.Now()
	}
}

func parseDateStr(s string) (time.Time, error) {
	// Chop off the last token (I think it's day-of-week, not sure)
	s = s[:len(s)-2]
	return time.ParseInLocation("15:4:5:1:2:2006", s, time.UTC)
}
