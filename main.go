package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var baseURL = "https://russia.tmembassy.gov.tm/ru/appointment/available"

func main() {
	client := &http.Client{}

	slotsMap := buildMap()
	formData := buildData(slotsMap)
	req, err := buildRequest(formData)
	if err != nil {
		log.Fatal(err)
	}
	freeSlotsArray := doRequest(req, client)

	var isSuccess bool
	for index, isFreeSlot := range freeSlotsArray {
		if isFreeSlot {
			fmt.Printf("Date: %s is free\n", slotsMap[index]["datetime"])
			isSuccess = true
		}
	}

	if isSuccess {
		os.Exit(1)
	}
}

func buildData(dataMap map[int]map[string]string) string {
	form := url.Values{}

	for k, v := range dataMap {
		form.Add(fmt.Sprintf("data[%d][service]", k), v["service"])
		form.Add(fmt.Sprintf("data[%d][datetime]", k), v["datetime"])
	}

	return form.Encode()
}

func buildMap() map[int]map[string]string {
	var data = make(map[int]map[string]string)

	index := 0

	now := time.Now()
	for i := 0; i < 15; i++ {
		dateStart := time.Date(now.Year(), now.Month(), now.Day(), 8, 45, 0, 0, time.Local)
		dateStart = dateStart.AddDate(0, 0, i)
		for j := 0; j < 8; j++ {
			dateStart = dateStart.Add(time.Duration(15) * time.Minute)

			data[index] = map[string]string{
				"service":  "requesting_information",
				"datetime": dateStart.Format(time.DateTime),
			}
			index++
		}
	}
	return data
}

// go
func buildRequest(data string) (*http.Request, error) {
	req, err := http.NewRequest("POST", baseURL, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	// Устанавливаем Content\-Type всегда
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	return req, nil
}

func doRequest(req *http.Request, client *http.Client) []bool {
	resp, err := client.Do(req)
	if err != nil {
		return []bool{}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return []bool{}
	}

	var result []bool
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return []bool{}
	}
	return result
}
