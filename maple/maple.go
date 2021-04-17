package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func regexFind(pattern, text string) ([]string, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	// Get current conditions string and the temperature
	match := r.FindStringSubmatch(text)
	return match, nil
}

func fetch() (CurrentCondition, []Warning, []Forecast, error) {
	// Get weather from Environment Canada
	resp, err := http.Get("https://weather.gc.ca/rss/city/qc-147_e.xml")
	if err != nil {
		return CurrentCondition{}, nil, nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error getting weather from server: ", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Status error: %v", resp.StatusCode)
		return CurrentCondition{}, nil, nil, errors.New(msg)

	}

	// Read data from response
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return CurrentCondition{}, nil, nil, err
	}

	// Unmarshal data
	var weatherRes weatherXML
	if err := xml.Unmarshal(data, &weatherRes); err != nil {
		return CurrentCondition{}, nil, nil, err
	}

	weeklyForecast := make([]Forecast, 0)
	currentConditions := CurrentCondition{}
	watchesAndWarnings := make([]Warning, 0)

	for _, entry := range weatherRes.Entry {
		switch entry.Category.Term {
		case "Weather Forecasts":
			weeklyForecast = append(weeklyForecast, Forecast{Short: entry.Title, Long: entry.Summary.Text, Link: entry.Link.Href, Updated: entry.Updated})
		case "Current Conditions":
			// Get the temperature and current conditions string
			match, err := regexFind("^Current Conditions: (.*), ((?:\\d+|\\d+\\.\\d+))°", entry.Title)
			if err != nil {
				return CurrentCondition{}, nil, nil, err
			}

			currentConditions.Condition = match[1]
			temp, err := strconv.ParseFloat(match[2], 64)
			if err != nil {
				return CurrentCondition{}, nil, nil, err
			}
			currentConditions.Temperature = temp

			// Get remaining current conditions information
			conditions := strings.Split(entry.Summary.Text, "\n")

			// Time
			match, err = regexFind("Airport (.+) <br\\/>", entry.Summary.Text)
			if err != nil {
				return CurrentCondition{}, nil, nil, err
			}
			t, err := time.Parse("3:04 PM MST Monday 2 January 2006", match[1])
			if err != nil {
				return CurrentCondition{}, nil, nil, err
			}
			currentConditions.Time = t

			// Pressure
			match, err = regexFind("Pressure \\/ Tendency:<\\/b> (.+) kPa", conditions[3])
			if err != nil {
				return CurrentCondition{}, nil, nil, err
			}
			pressure, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				return CurrentCondition{}, nil, nil, err
			}
			currentConditions.Pressure = pressure

			// Humidity[5]
			match, err = regexFind("</b> (.+) %", conditions[5])
			if err != nil {
				return CurrentCondition{}, nil, nil, err
			}
			humidity, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				return CurrentCondition{}, nil, nil, err
			}
			currentConditions.Humidity = humidity
			// Wind[7]
			match, err = regexFind("(\\w+) km", conditions[7])
			if err != nil {
				return CurrentCondition{}, nil, nil, err
			}
			wind, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				return CurrentCondition{}, nil, nil, err
			}
			currentConditions.Wind = wind

		case "Warnings and Watches":
			watchesAndWarnings = append(watchesAndWarnings, Warning{
				Short:   entry.Title,
				Long:    entry.Summary.Text,
				Link:    entry.Link.Href,
				Updated: entry.Updated,
			})
		}
	}

	return currentConditions, watchesAndWarnings, weeklyForecast, nil
}

func main() {
	current, warnings, weekly, err := fetch()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("CURRENT:\n%v\n", current)
	fmt.Println("WARNINGS:")
	for _, w := range warnings {
		fmt.Println(w)
	}
	fmt.Println("WEEKLY:")
	for _, w := range weekly {
		fmt.Println(w)
	}
}
