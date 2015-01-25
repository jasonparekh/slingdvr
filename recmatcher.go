package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type RecRequest struct {
	Showing  Showing
	Finished chan<- struct{}
}

func SendRecReqs(c chan<- RecRequest, recs <-chan Showing) error {
	titles := getRecTitles()
	recorded, err := ReadRecorded()
	if err != nil {
		return err
	}

	reloadConfigTick := time.Tick(1 * time.Minute)
	for rec := range recs {
		sleepUntilRecStartTime()

		select {
		case <-reloadConfigTick:
			titles = getRecTitles()
		default:
		}

		if !doesMatch(titles, rec) || contains(rec.Id, recorded) {
			continue
		}

		// Signal to record this
		finished := make(chan struct{})
		c <- RecRequest{rec, finished}
		<-finished

		recorded = append(recorded, rec.Id)
		if err := WriteRecorded(recorded); err != nil {
			fmt.Fprintf(os.Stderr, "Could not write recorded: %s\n", err)
		}
	}

	return nil
}

func getRecTitles() map[string]bool {
	if err := ReadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not read config: %s\n", err)
		return nil
	}

	return genRecordTitlesMap(config.Titles)
}

func doesMatch(titles map[string]bool, rec Showing) bool {
	if afterEarliestShowingTime := !config.EarliestShowingTime.After(rec.Start); !afterEarliestShowingTime {
		return false
	}

	titleMatches := titles[strings.ToLower(rec.Title)]
	return titleMatches || *recordAll
}

func sleepUntilRecStartTime() {
	if config.RecStartTime.IsZero() {
		return
	}

	rsh, rsm, _ := config.RecStartTime.Clock()
	rsmins := rsh*60 + rsm
	reh, rem, _ := config.RecEndTime.Clock()
	remins := reh*60 + rem

	nh, nm, _ := timeNow().Clock()
	nmins := nh*60 + nm

	// Either in the range, or the range crosses midnight and now is past midnight
	if rsmins <= nmins && nmins < remins || remins < rsmins && nmins < remins {
		return
	}

	if rsmins < nmins {
		rsmins += 24 * 60
	}
	durMins := rsmins - nmins
	fmt.Printf("Sleeping for %d mins until rec start time\n", durMins)
	time.Sleep(time.Duration(durMins) * time.Minute)
}

func contains(s string, ss []string) bool {
	for _, c := range ss {
		if c == s {
			return true
		}
	}

	return false
}
