package suppliers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
)

type Weather struct{}

type weatherStation struct {
	Id          string             `json:"id"`
	Name        string             `json:"name"`
	Temperature string             `json:"temperature"`
	Parameters  []weatherParameter `json:"parameters"`
	IconTitle   string             `json:"iconTitle"`
}

type weatherParameter struct {
	Name  string `json:"name"`
	Id    string `json:"parameterId"`
	Value string `json:"value"`
}

type weatherResponse struct {
	Date     string           `json:"date"`
	Time     string           `json:"time"`
	Stations []weatherStation `json:"stations"`
}

func NewWeather() *Weather {
	return &Weather{}
}

func (s *weatherStation) String() string {
	res := s.Name + " : " + s.Temperature

	if s.IconTitle != "" {
		res += fmt.Sprintf("(%s)", s.IconTitle)
	}

	for _, p := range s.Parameters {
		if p.Id == "121" { //temperature
			continue
		}

		if p.Value != "" {
			if p.Id == "113" { //wind speed is in format normal/gust
				parts := strings.Split(p.Value, "/")
				p.Value = fmt.Sprintf("%s (brāzmās %s)", parts[0], parts[1])
			}

			res = res + ", " + p.Name + " : " + p.Value
		}
	}

	return res
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
			return station.String(), nil
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

	u := "http://www.meteo.lv/meteorologijas-operativie-dati/?date=&time=&parameterId=&fullMap=0&rnd=" + strconv.FormatUint(uint64(time.Now().Unix()), 10)

	req, _ := http.NewRequest("GET", u, nil)

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/49.0.2623.112 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", "http://www.meteo.lv/meteorologijas-operativa-informacija/?nid=459&pid=122")

	res, err := client.Do(req)

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
