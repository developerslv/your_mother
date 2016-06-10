package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
)

type Weather struct{}

type weatherStation struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Temperature string `json:"temperature"`
}

type weatherResponse struct {
	Date     string           `json:"date"`
	Time     string           `json:"time"`
	Stations []weatherStation `json:"stations"`
}

func NewWeather() *Weather {
	return &Weather{}
}

func (w *Weather) ForCity(city string) (string, error) {
	if strings.EqualFold("Rīga", city) {
		city = "Rīga - LU"
	}

	meteoResp, err := w.getWeatherResponse()

	if err != nil {
		return "", err
	}

	for _, station := range meteoResp.Stations {
		if strings.EqualFold(station.Name, city) {
			return city + " : " + station.Temperature, nil
		}
	}

	return "", errors.New("Could not find requested weather station")
}

func (w *Weather) ListCities() ([]string, error) {
	meteoResp, err := w.getWeatherResponse()

	cities := make([]string, 0, 10)

	if err != nil {
		return nil, err
	}

	for _, station := range meteoResp.Stations {
		cities = append(cities, station.Name)
	}

	return cities, nil
}

func (w *Weather) getWeatherResponse() (*weatherResponse, error) {
	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar: jar,
	}

	res, err := client.Get("http://www.meteo.lv/meteorologijas-operativie-dati/?date=&time=&parameterId=122&fullMap=0&rnd=" + strconv.FormatUint(uint64(time.Now().Unix()), 10))

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	forecast, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	var meteoResp weatherResponse

	if err = json.Unmarshal(forecast, &meteoResp); err != nil {
		return nil, err
	}

	return &meteoResp, nil
}
