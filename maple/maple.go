package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// globals
var config MapleConfig
var logInfo *log.Logger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
var logWarn *log.Logger = log.New(os.Stdout, "[WARNING] ", log.Ldate|log.Ltime|log.Lshortfile)
var logError *log.Logger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

// config fetcher
func getConfig(filename string) error {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	json.Unmarshal(bytes, &config)

	return nil
}

// database

// controllers
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

// handlers
func handleNotFound(w http.ResponseWriter, r *http.Request) {
	logWarn.Printf("404 on %v", r.RequestURI)
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func currentHandler(w http.ResponseWriter, r *http.Request) {
	logInfo.Printf("Hit on %v", r.RequestURI)

	data, err := getCurrentConditions()
	if err != nil {
		logError.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	err := getConfig("config.json")
	if err != nil {
		logError.Fatal("error getting config file:", err)
	}
	logInfo.Println("loaded config")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	// NOTE: I never release this connection, may cause issuse later
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logError.Fatal("error making database connection: ", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logError.Fatal("error pinging database: ", err)
	}
	logInfo.Println("connected to database")

	// sqlStatement1 := `CREATE TABLE temp (name TEXT);`
	// sqlStatement2 := `INSERT INTO temp (name) VALUES ('bob');`
	// _, err = db.Exec(sqlStatement1)
	// if err != nil {
	// 	logError.Fatal(err)
	// }

	// _, err = db.Exec(sqlStatement2)
	// if err != nil {
	// 	logError.Fatal(err)
	// }

	res, err := db.Query(`SELECT * FROM temp;`)
	if err != nil {
		logError.Fatal(err)
	}

	for res.Next() {
		var name string
		if err = res.Scan(&name); err != nil {
			logError.Fatal(err)
		}
		logInfo.Println(name)
	}

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(handleNotFound)
	r.HandleFunc("/current", currentHandler)

	address := fmt.Sprintf(":%v", config.Port)
	logInfo.Printf("listening on port %v", address)
	log.Fatal(http.ListenAndServe(address, r))

}
