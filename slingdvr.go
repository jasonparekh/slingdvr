package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
)

var async = flag.Bool("async", true, "Records DVR programs at a later time")

func main() {
	flag.Parse()

	if err := ReadConfig(); err != nil {
		panic(err)
	}

	if *dumpSchedule {
		if showings, err := FetchSchedule(); err == nil {
			fmt.Printf("URL: %s\n", scheduleUrl())
			spew.Printf("%#v\n", showings)
			return
		} else {
			panic(err)
		}
	} else if *dumpRecs {
		if showings, err := FetchRecs(); err == nil {
			fmt.Printf("URL: %s\n", recUrl())
			spew.Printf("%#v\n", showings)
			return
		} else {
			panic(err)
		}
	}

	if *forceTime != "" {
		// TODO hack, isn't concurrency-safe
		go func() {
			const timeMultiple = 1
			for {
				time.Sleep(timeMultiple * time.Second)
				*forceTime = timeNow().Add(timeMultiple * time.Second).Format(time.RFC3339)
			}
		}()
	}

	if !*async {
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

		for {
			time.Sleep(15 * 60 * time.Second)
			refreshScheduleC <- struct{}{}
		}

	} else {
		allRecs := make(chan Showing)
		recReqs := make(chan RecRequest)

		go func() {
			if err := SendNewRecs(allRecs); err != nil {
				panic(err)
			}
		}()

		go func() {
			if err := SendRecReqs(recReqs, allRecs); err != nil {
				panic(err)
			}
		}()

		go func() {
			if err := AsyncRecorder(recReqs); err != nil {
				panic(err)
			}
		}()

		// Block forever
		<-make(chan struct{})
	}
}
