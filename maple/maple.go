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

	// config := MapleConfig{}
	json.Unmarshal(bytes, &config)

	return config, nil
}

func getCurrentConditions() (CurrentConditions, error) {
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?lat=%v&lon=%v&appid=%v&units=%v", config.Latitude, config.Longitude, config.OpenWeatherApiKey, config.Unit)
	resp, err := http.Get(url)
	if err != nil {
		return CurrentConditions{}, fmt.Errorf("error in http request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CurrentConditions{}, fmt.Errorf("error reading file: %v", err)
	}
	weatherJson := struct {
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`

		Main struct {
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			Pressure  int     `json:"pressure"`
			Humidity  int     `json:"humidity"`
		} `json:"main"`
		Wind struct {
			Speed float64 `json:"speed"`
			Deg   int     `json:"deg"`
		} `json:"wind"`
		Dt int `json:"dt"`
	}{}

	json.Unmarshal(body, &weatherJson)

	currentConditions := CurrentConditions{
		Temperature:   weatherJson.Main.Temp,
		FeelsLike:     weatherJson.Main.FeelsLike,
		Humidity:      weatherJson.Main.Humidity,
		WindSpeed:     weatherJson.Wind.Speed,
		WindDirection: weatherJson.Wind.Deg,
		Pressure:      weatherJson.Main.Pressure,
		Condition:     weatherJson.Weather[0].Description,
		Timestamp:     weatherJson.Dt,
	}

	return currentConditions, nil
}

func main() {
	config, err := getConfig("config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(config)
	current, err := getCurrentConditions()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", current)
}
