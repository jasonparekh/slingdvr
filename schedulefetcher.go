package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var dumpSchedule = flag.Bool("dumpSchedule", false, "Dumps the schedule and exits")

type RawScheduleContainer struct {
	ReqPack struct {
		XmlFile struct {
			TimerList struct {
				DTimer []RawShowing `json:"d_timer"` // Daily schedule
				GTimer []RawShowing `json:"g_timer"` // My timers
			} `json:"timer_list"`
		} `json:"xml_file"`
	} `json:"req_pack"`
}

type RawShowing struct {
	EventName      string `json:"event_name"`
	Id             string `json:"tms_id"`
	StartTimestamp int64  `json:"startTimestamp"` // Not sure what timezone this is in, it is not GMT. Is in seconds
	EndTimestamp   int64  `json:"endTimestamp"`
	MediaView      []struct {
		Title string `json:"title"`
	} `json:"mediaview"`
	TmPeriod struct {
		StartTime string `json:"start_time"` // GMT formatted like "2:30:0:9:28:2014:0"
	} `json:"tm_period"`
}

func FetchSchedule() (showings []Showing, err error) {
	url := scheduleUrl()
	var resp *http.Response
	if resp, err = http.Get(url); err != nil {
		return
	}

	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var container RawScheduleContainer
	if err = json.Unmarshal(res, &container); err != nil {
		return
	}

	for _, rawShowing := range container.ReqPack.XmlFile.TimerList.DTimer {
		var startTime time.Time
		startTime, err = parseDateStr(rawShowing.TmPeriod.StartTime)
		if err != nil {
			return
		}

		// Includes any padding specified by DVR timer
		duration := rawShowing.EndTimestamp - rawShowing.StartTimestamp
		endTime := startTime.Add(time.Duration(duration) * time.Second)

		showing := Showing{
			rawShowing.MediaView[0].Title,
			rawShowing.EventName,
			rawShowing.Id,
			startTime,
			endTime,
			"",
		}
		showings = append(showings, showing)
	}

	return
}

func scheduleUrl() string {
	return fmt.Sprintf("http://www.dishanywhere.com/sge/timer/rest/v1/receiverid/%s?correlationid=%s&type=raw&ptat_filter=&limitstart=1&limitoffset=999&sortby=name&sortorder=asc", config.ReceiverId, config.CorrelationId)
}
