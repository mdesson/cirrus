package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

var config MapleConfig

func getConfig(filename string) (MapleConfig, error) {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		return MapleConfig{}, fmt.Errorf("error opening file: %v", err)
	}
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return MapleConfig{}, fmt.Errorf("error reading file: %v", err)
	}

	json.Unmarshal(bytes, &config)

	return config, nil
}

func getCurrentConditions() (WeatherData, error) {
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/onecall?lat=%v&lon=%v&appid=%v&units=%v",
		config.Latitude, config.Longitude, config.OpenWeatherApiKey, config.Unit)

	resp, err := http.Get(url)
	if err != nil {
		return WeatherData{}, fmt.Errorf("error in http request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return WeatherData{}, fmt.Errorf("error reading file: %v", err)
	}

	// One-off struct to hold openeweathermap's API response
	weatherJson := struct {
		Current struct {
			Dt        int     `json:"dt"`
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			Pressure  int     `json:"pressure"`
			Humidity  int     `json:"humidity"`
			WindSpeed float64 `json:"wind_speed"`
			WindDeg   int     `json:"wind_deg"`
			Weather   []struct {
				Description string `json:"description"`
			} `json:"weather"`
		} `json:"current"`
		Minutely []PrecipitationChance `json:"minutely"`
		Hourly   []struct {
			Dt        int     `json:"dt"`
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			WindSpeed float64 `json:"wind_speed"`
			WindGust  float64 `json:"wind_gust"`
			Weather   []struct {
				Description string `json:"description"`
			} `json:"weather"`
			Pop int `json:"pop"`
		} `json:"hourly"`
		Daily []struct {
			Dt        int               `json:"dt"`
			Temp      TempForecast      `json:"temp"`
			FeelsLike FeelsLikeForecast `json:"feels_like"`
			Humidity  int               `json:"humidity"`
			WindSpeed float64           `json:"wind_speed"`
			WindGust  float64           `json:"wind_gust"`
			Weather   []struct {
				Description string `json:"description"`
			} `json:"weather"`
			Pop  int     `json:"pop"`
			Rain float64 `json:"rain,omitempty"`
		} `json:"daily"`
	}{}

	json.Unmarshal(body, &weatherJson)

	data := WeatherData{}

	data.Current = CurrentConditions{
		Timestamp:     weatherJson.Current.Dt,
		Temp:          weatherJson.Current.Temp,
		FeelsLike:     weatherJson.Current.FeelsLike,
		Pressure:      weatherJson.Current.Pressure,
		Humidity:      weatherJson.Current.Humidity,
		WindSpeed:     weatherJson.Current.WindSpeed,
		WindDirection: weatherJson.Current.WindDeg,
		Description:   weatherJson.Current.Weather[0].Description,
	}

	data.Minutely = weatherJson.Minutely

	data.Hourly = make([]HourlyForecast, 48)
	for i, forecast := range weatherJson.Hourly {
		data.Hourly[i] = HourlyForecast{
			Timestamp:   forecast.Dt,
			Temp:        forecast.Temp,
			FeelsLike:   forecast.FeelsLike,
			WindSpeed:   forecast.WindSpeed,
			WindGust:    forecast.WindGust,
			Description: forecast.Weather[0].Description,
			Pop:         forecast.Pop,
		}
	}

	data.Daily = make([]DailyForecast, 7)
	for i, forecast := range weatherJson.Daily[1:] {
		data.Daily[i] = DailyForecast{
			Timestamp:   forecast.Dt,
			Humidity:    forecast.Humidity,
			WindSpeed:   forecast.WindSpeed,
			WindGust:    forecast.WindGust,
			Description: forecast.Weather[0].Description,
			Temp:        forecast.Temp,
			FeelsLike:   forecast.FeelsLike,
		}
	}

	return data, nil
}

func main() {
	config, err := getConfig("config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(config)
	data, err := getCurrentConditions()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", data)
}
