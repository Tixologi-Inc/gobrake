package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Tixologi-Inc/gobrake"
	httpbrake "github.com/Tixologi-Inc/gobrake/http"
)

// Define models

type Date struct {
	Date int64 `json:"date,omitempty"`
}

type WeatherInfo struct {
	Lat            float32       `json:"lat,omitempty"`
	Lon            float32       `json:"lon,omitempty"`
	Timezone       string        `json:"timezone,omitempty"`
	TimezoneOffset float32       `json:"timezone_offset,omitempty"`
	Current        interface{}   `json:"current,omitempty"`
	Minutely       []interface{} `json:"minutely,omitempty"`
	Hourly         []interface{} `json:"hourly,omitempty"`
	Daily          []interface{} `json:"daily,omitempty"`
}

var ProjectId int64 = 999999                               // Insert your Project Id here
var ProjectKey string = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" // Insert your Project Key here

var notifier = gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
	ProjectId:   ProjectId,
	ProjectKey:  ProjectKey,
	Environment: "production",
})

func main() {
	mux := http.NewServeMux()
	fmt.Println("Server listening at http://localhost:3000/")
	airbrakeRouteNotifier := httpbrake.New(notifier)
	d := Date{Date: time.Now().Unix()}
	// Example on how to use Handle method with airbrake notifier
	mux.Handle("/date", airbrakeRouteNotifier.Handle(d))
	// Example on how to use HandleFunc method with airbrake notifier
	mux.HandleFunc("/locations", airbrakeRouteNotifier.HandleFunc(getLocations))
	mux.HandleFunc("/weather/", airbrakeRouteNotifier.HandleFunc(getWeather))
	http.ListenAndServe(":3000", mux)
}

func (d Date) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}

func getLocations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	availableLocations, err := locations()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availableLocations)
	if err != nil {
		notifier.Notify(err, nil) // send error to airbrake
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var weatherInfo WeatherInfo
	w.Header().Set("Content-Type", "application/json")
	location := strings.TrimPrefix(r.URL.Path, "/weather/")
	location = strings.TrimSpace(location) // It lets you allow whitespaces but case-sensitive
	weatherResp, err := checkWeather(location)
	if err != nil {
		notifier.Notify(err, nil) // send error to airbrake
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(weatherInfo)
		return
	}
	json.Unmarshal(weatherResp, &weatherInfo)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(weatherInfo)
}

func locations() ([]string, error) {
	var locations []string
	weatherInfoFile := "https://airbrake.github.io/weatherapi/locations"
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, weatherInfoFile, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("locations not found")
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(bodyBytes, &locations)
	return locations, nil
}

func checkWeather(location string) ([]byte, error) {
	weatherInfoFile := "https://airbrake.github.io/weatherapi/weather/" + location
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, weatherInfoFile, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		errMsg := fmt.Sprintf("%s weather data not found", location)
		return nil, errors.New(errMsg)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}
