package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const ua = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, " +
	"like Gecko) Chrome/133.0.0.0 Safari/537.36"

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

func doRequest(req *http.Request, client *http.Client) []bool {
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7,zh-CN;q=0.6,zh-TW;q=0.5,zh;q=0.4")
	req.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("origin", "https://russia.tmembassy.gov.tm")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://russia.tmembassy.gov.tm/ru/appointment")
	req.Header.Set("sec-ch-ua", `"Not(A:Brand";v="99", "Google Chrome";v="133", "Chromium";v="133"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", ua)
	resp, err := client.Do(req)
	bClose := func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}
	if err != nil {
		log.Println(err)
		bClose(resp.Body)
		log.Fatal(err)
	}

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		bClose(resp.Body)
		log.Fatal(err)
	}
	bClose(resp.Body)
	var d []bool

	err = json.Unmarshal(bodyText, &d)
	if err != nil {
		log.Fatal(err)
	}

	return d
}
func buildRequest(data string) (*http.Request, error) {
	return http.NewRequest("POST", "https://russia.tmembassy.gov.tm/ru/appointment/available", strings.NewReader(data))
}
