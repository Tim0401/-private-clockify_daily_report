package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	baseEndpoint = "https://api.clockify.me/api/v1/"
	userPath     = "user"

	reportsEndpoint = "https://reports.api.clockify.me/v1"
	//  /workspaces/{workspaceId}/reports/detailed
	detailedReportPath = "/workspaces/%s/reports/detailed"
)

type User struct {
	ID string `json:"id"`
}

type DetailedReportReq struct {
	DateRangeStart string         `json:"dateRangeStart"`
	DateRangeEnd   string         `json:"dateRangeEnd"`
	DetailedFilter DetailedFilter `json:"detailedFilter"`
}

type DetailedFilter struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
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
	Duration int `json:"duration"`
}

type Tag struct {
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
			Page:     1,
			PageSize: 100,
		},
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
	fmt.Printf("%+v", report)

	// format report
}
