package main

import (
	"time"
	"flag"
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
