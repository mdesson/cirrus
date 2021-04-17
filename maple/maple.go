package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type weatherXML struct {
	Entry []struct {
		Title   string `xml:"title"`
		Updated string `xml:"updated"`
		Type    string `xml:"type,attr"`
		Link    struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
		Summary struct {
			Text string `xml:",chardata"`
		} `xml:"summary"`
		Category struct {
			Term string `xml:"term,attr"`
		} `xml:"category"`
	} `xml:"entry"`
}

type Forecast struct {
	Short   string
	Long    string
	Link    string
	Updated string
}

type Warning struct {
	Short   string
	Long    string
	Link    string
	Updated string
}

type CurrentCondition struct {
	Temperature float64
	Humidity    float64
	Wind        float64
	Pressure    float64
	Condition   string
	Time        time.Time
}

func regexFind(pattern, text string) ([]string, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	// Get current conditions string and the temperature
	match := r.FindStringSubmatch(text)
	return match, nil
}

func main() {
	// Get weather from Environment Canada
	resp, err := http.Get("https://weather.gc.ca/rss/city/qc-147_e.xml")
	if err != nil {
		fmt.Printf("GET error: %v\n", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Close resp body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Status error: %v\n", resp.StatusCode)
	}

	// Read data from response
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Read body: %v\n", err)
	}

	// Unmarshal data
	var weatherRes weatherXML
	if err := xml.Unmarshal(data, &weatherRes); err != nil {
		fmt.Printf("xml unmarshal: %v\n", err)
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
				fmt.Printf("Regex Error Getting Current Conditions: %v\n", err)
			}

			currentConditions.Condition = match[1]
			temp, err := strconv.ParseFloat(match[2], 64)
			if err != nil {
				fmt.Printf("Error getting temperature float: %v", err)
			}
			currentConditions.Temperature = temp

			// Get remaining current conditions information
			conditions := strings.Split(entry.Summary.Text, "\n")

			// Time
			match, err = regexFind("Airport (.+) <br\\/>", entry.Summary.Text)
			if err != nil {
				fmt.Printf("Regex Error Getting Current Conditions: %v\n", err)
			}
			t, err := time.Parse("3:04 PM MST Monday 2 January 2006", match[1])
			if err != nil {
				fmt.Printf("Error getting date: %v\n", err)
			}
			currentConditions.Time = t

			// Pressure
			match, err = regexFind("Pressure \\/ Tendency:<\\/b> (.+) kPa", conditions[3])
			if err != nil {
				fmt.Printf("Regex Error Getting Pressure: %v\n", err)
			}
			pressure, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				fmt.Printf("Error compiling getting pressure float: %v", err)
			}
			currentConditions.Pressure = pressure

			// Humidity[5]
			match, err = regexFind("</b> (.+) %", conditions[5])
			if err != nil {
				fmt.Printf("Regex Error Getting Humidity: %v\n", err)
			}
			humidity, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				fmt.Printf("Error compiling getting pressure float: %v", err)
			}
			currentConditions.Humidity = humidity
			// Wind[7]
			match, err = regexFind("(\\w+) km", conditions[7])
			if err != nil {
				fmt.Printf("Regex Error Getting wind: %v\n", err)
			}
			wind, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				fmt.Printf("Error compiling getting wind float: %v", err)
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

	fmt.Println(currentConditions)

	fmt.Println("===============")

	for _, day := range weeklyForecast {
		fmt.Println(day.Long)
	}
}
