package main

type MapleConfig struct {
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	OpenWeatherApiKey string  `json:"open_weather_api_key"`
	Unit              string  `json:"unit"`
}

type CurrentConditions struct {
	Timestamp     int
	Temp          float64
	FeelsLike     float64
	Pressure      int
	Humidity      int
	WindSpeed     float64
	WindDirection int
	Description   string
}

type PrecipitationChance struct {
	Timestamp int `json:"dt"`
	Pop       int `json:"precipitation"`
}

type HourlyForecast struct {
	Timestamp   int
	Temp        float64
	FeelsLike   float64
	WindSpeed   float64
	WindGust    float64
	Description string
	Pop         int
}

type TempForecast struct {
	Day   float64 `json:"day"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Night float64 `json:"night"`
	Eve   float64 `json:"eve"`
	Morn  float64 `json:"morn"`
}
type FeelsLikeForecast struct {
	Day   float64 `json:"day"`
	Night float64 `json:"night"`
	Eve   float64 `json:"eve"`
	Morn  float64 `json:"morn"`
}

type DailyForecast struct {
	Timestamp   int
	Humidity    int
	WindSpeed   float64
	WindGust    float64
	Description string
	Pop         int
	Rain        float64
	Temp        TempForecast
	FeelsLike   FeelsLikeForecast
}

type WeatherData struct {
	Current  CurrentConditions
	Minutely []PrecipitationChance
	Hourly   []HourlyForecast
	Daily    []DailyForecast
}
