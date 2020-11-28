package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
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
	Observed    string
	Link        string
	Short       string
	Condition   string
	Temperature string
	Visibility  string
	Pressure    string
	Humidity    string
	Dewpoint    string
	Wind        string
	AirQuality  string
}

func main() {
	resp, err := http.Get("https://weather.gc.ca/rss/city/qc-147_e.xml")
	if err != nil {
		fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Errorf("Status error: %v", resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Errorf("Read body: %v", err)
	}

	var weatherRes weatherXML
	xml.Unmarshal(data, &weatherRes)

	weeklyForecast := make([]Forecast, 0)

	for _, entry := range weatherRes.Entry {
		switch entry.Category.Term {
		case "Weather Forecasts":
			weeklyForecast = append(weeklyForecast, Forecast{Short: entry.Title, Long: entry.Summary.Text, Link: entry.Link.Href, Updated: entry.Updated})
		case "Current Conditions":
			fmt.Printf("%+v\n", entry)
		case "Warnings and Watches":
			fmt.Printf("%+v\n", entry)
		}
	}

	fmt.Println("===============")

	for _, day := range weeklyForecast {
		fmt.Println(day.Long)
	}
}
