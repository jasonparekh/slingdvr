package main

import (
	"flag"
	"time"
)

func main() {
	flag.Parse()

	if err := ReadConfig(); err != nil {
		panic(err)
	}

	recordC := make(chan Showing)
	refreshScheduleC := make(chan struct{})

	go func() {
		if err := Scheduler(recordC, refreshScheduleC); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := Recorder(recordC); err != nil {
			panic(err)
		}
	}()

	if *forceTime != "" {
		for {
			time.Sleep(5 * time.Second)
			*forceTime = timeNow().Add(30 * 60 * time.Second).Format(time.RFC3339)

			refreshScheduleC <- struct{}{}
		}
	} else {
		for {
			time.Sleep(15 * 60 * time.Second)
			refreshScheduleC <- struct{}{}
		}
	}
}
