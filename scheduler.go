package main

import (
	"strings"
	"time"
	"fmt"
	"flag"
)

var recordAll = flag.Bool("recordAll", false, "records everything")

func Scheduler(recordC chan<- Showing, refreshC <-chan struct{}) (err error) {
	var recordTitles []string
	for {
		fmt.Println("Schedule: ")
		recordTitles, err = readRecordTitles()
		if err != nil {
			return
		}

		abortC := make(chan struct{})
		if err = runSchedule(recordTitles, recordC, abortC); err != nil {
			fmt.Println(err.Error())
			time.Sleep(5 * time.Second)
		}

		<- refreshC
		close(abortC)

		fmt.Println()
	}

	return
}

func runSchedule(recordTitles []string, recordC chan<- Showing, abortC <-chan struct{}) (error) {
	recordLookup := genRecordTitlesMap(recordTitles)

	showings, err := FetchSchedule()
	if err != nil {
		return err
	}

	for _, showing := range showings {
		if (!recordLookup[strings.ToLower(showing.Title)] && !*recordAll) || timeNow().After(showing.End) {
			continue
		}

		go func(showing Showing) {
			fmt.Printf("%s will record at %s\n", showing.Title, showing.Start.Local())

			select {
			case <-time.After(showing.Start.Sub(timeNow())):
				// Note this will also run if we are in the middle of the recording
				recordC <- showing
			case <-abortC:
			}
		}(showing)
	}

	return nil
}

func genRecordTitlesMap(recordTitles []string) map[string]bool {
	m := make(map[string]bool)
	for _, title := range recordTitles {
		m[strings.ToLower(title)] = true
	}

	return m
}

func readRecordTitles() ([]string, error) {
	if err := ReadConfig(); err != nil {
		return nil, err
	}

	return config.Titles, nil
}
