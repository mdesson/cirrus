package main

import "time"

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
