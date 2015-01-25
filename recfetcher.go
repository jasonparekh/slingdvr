package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
)

type RawRecContainer struct {
	ReqPack struct {
		XMLFile struct {
			PVRList struct {
				PVRRecords []RawRec `json:"pvr_record"`
			} `json:"pvr_list"`
		} `json:"xml_file"`
	} `json:"req_pack"`
}

type RawRec struct {
	EventName string `json:"event_name"`
	Id        int    `json:"pgm_id"`
	PVRAttrib int    `json:"pvr_attrib"`
	Duration  int64  `json:"duration"` // In minutes
	MediaView []struct {
		Title     string `json:"title"`
		ShortDesc string `json:"short_description"`
	} `json:"mediaview"`
	RecTime  string `json:"rec_time"` // GMT formatted like "2:30:0:9:28:2014:0"
	ChanName string `json:"svc_name"`
}

var dumpRecs = flag.Bool("dumpRecs", false, "Dumps recordings on DVR")

const recfetcherPollingPeriod = 1 * time.Minute

func FetchRecs() (showings []Showing, err error) {
	url := recUrl()
	var resp *http.Response
	if resp, err = http.Get(url); err != nil {
		return
	}

	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var container RawRecContainer
	if err = json.Unmarshal(res, &container); err != nil {
		return
	}

	for _, rec := range container.ReqPack.XMLFile.PVRList.PVRRecords {
		var startTime time.Time
		startTime, err = parseDateStr(rec.RecTime)
		if err != nil {
			return
		}

		var shortDesc string
		if len(rec.MediaView) > 0 {
			shortDesc = rec.MediaView[0].ShortDesc
		}
		showing := Showing{
			rec.EventName,
			shortDesc,
			strconv.Itoa(rec.Id),
			startTime,
			startTime.Add(time.Duration(rec.Duration) * time.Minute),
			strconv.Itoa(rec.PVRAttrib),
		}
		showings = append(showings, showing)
	}

	return
}

func recUrl() string {
	return fmt.Sprintf("http://www.dishanywhere.com/sge/dvr/rest/v1/receiverid/%s?correlationid=%s&pvrattribute=0&programid=0&type=raw&search=all&limitstart=1&limitoffset=999&sortby=date&sortorder=asc&cache=false", config.ReceiverId, config.CorrelationId)
}

func SendNewRecs(c chan<- Showing) error {
	for {
		var recs []Showing
		fetchFunc := func() (err error) {
			recs, err = FetchRecs()
			return
		}
		if err := backoff.Retry(fetchFunc, backoff.NewExponentialBackOff()); err != nil {
			return err
		}

		start := timeNow()
		for _, rec := range recs {
			c <- rec
		}

		if elapsed := timeNow().Sub(start); elapsed < recfetcherPollingPeriod {
			time.Sleep(recfetcherPollingPeriod - elapsed)
		}
	}
}
