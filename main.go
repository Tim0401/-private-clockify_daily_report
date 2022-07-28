package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

const (
	// Endpoint from https://clockify.me/developers-api

	// use for getting report
	reportsEndpoint = "https://reports.api.clockify.me/v1"
	//  /workspaces/{workspaceId}/reports/detailed
	detailedReportPath = "/workspaces/%s/reports/detailed"
)

// Type Definition from https://clockify.me/developers-api
type DetailedReportReq struct {
	DateRangeStart string         `json:"dateRangeStart"`
	DateRangeEnd   string         `json:"dateRangeEnd"`
	DetailedFilter DetailedFilter `json:"detailedFilter"`
	SortOrder      string         `json:"sortOrder"`
}

type DetailedFilter struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	SortColumn string `json:"sortColumn"`
}

type DetailedReport struct {
	TimeEntries []TimeEntry `json:"timeentries"`
}

type TimeEntry struct {
	Description  string       `json:"description"`
	Tags         []Tag        `json:"tags"`
	TimeInterval TimeInterval `json:"timeInterval"`
}

type TimeInterval struct {
	Start    time.Time `json:"start"`
	Duration int       `json:"duration"`
}

type Tag struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}

func main() {
	// APIを叩くのに必要な情報を取得
	accessToken := os.Getenv("CLOCKIFY_ACCESS_TOKEN")
	workspaceID := os.Getenv("CLOCKIFY_WORKSPACE_ID")
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	// Get Today Detail
	today := time.Now().Format("2006-01-02")
	reqBody := DetailedReportReq{
		DateRangeStart: today + "T00:00:00.000",
		DateRangeEnd:   today + "T23:59:59.999",
		DetailedFilter: DetailedFilter{
			Page:       1,
			PageSize:   100,
			SortColumn: "Date",
		},
		SortOrder: "ASCENDING",
	}
	rawBody, _ := json.Marshal(reqBody)

	// Get Detailed Report
	req, err := http.NewRequest("POST", reportsEndpoint+fmt.Sprintf(detailedReportPath, workspaceID), bytes.NewBuffer(rawBody))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("X-Api-Key", accessToken)
	req.Header.Add("content-type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatal(res.StatusCode)
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var report DetailedReport
	if err := json.Unmarshal(data, &report); err != nil {
		log.Fatal(err)
	}

	// format report

	// group by tag
	tagTimeEntryMap := make(map[string][]TimeEntry)
	for _, entry := range report.TimeEntries {
		// skip if no tag
		if len(entry.Tags) == 0 {
			continue
		}
		tagTimeEntryMap[entry.Tags[0].Name] = append(tagTimeEntryMap[entry.Tags[0].Name], entry)
	}

	// entry and tag at least 1
	// sort by tag
	sortedTagEntries := make([][]TimeEntry, 0, len(tagTimeEntryMap))
	for key := range tagTimeEntryMap {
		sortedTagEntries = append(sortedTagEntries, tagTimeEntryMap[key])
	}
	sort.Slice(sortedTagEntries, func(i, j int) bool {
		return sortedTagEntries[i][0].TimeInterval.Start.Before(sortedTagEntries[j][0].TimeInterval.Start)
	})

	// print
	for _, entries := range sortedTagEntries {
		fmt.Printf("【%s】\n", entries[0].Tags[0].Name)
		// unique entries by description
		// key description, value TimeEntry
		uniqEntryMap := make(map[string]TimeEntry, 0)
		for i := range entries {
			if _, ok := uniqEntryMap[entries[i].Description]; !ok {
				uniqEntryMap[entries[i].Description] = entries[i]
			} else {
				orig := uniqEntryMap[entries[i].Description]
				orig.TimeInterval.Duration += entries[i].TimeInterval.Duration
				uniqEntryMap[entries[i].Description] = orig
			}
		}

		// sort by start time
		sortedEntries := make([]TimeEntry, 0, len(uniqEntryMap))
		for key := range uniqEntryMap {
			sortedEntries = append(sortedEntries, uniqEntryMap[key])
		}
		sort.Slice(sortedEntries, func(i, j int) bool {
			return sortedEntries[i].TimeInterval.Start.Before(sortedEntries[j].TimeInterval.Start)
		})

		// print
		for _, e := range sortedEntries {
			fmt.Printf(" └%.1f h %s \n", float64(e.TimeInterval.Duration)/3600, e.Description)
		}
	}
}
