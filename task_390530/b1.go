// weatherforecast.go

package main

import (
	"fmt"
	"net/http"
)

// Forecast represents the weather forecast for a given location.
type Forecast struct {
	Location string
	Temperature float64
	Description string
}

func main() {
	// Forecast data for New York
	nyForecast := Forecast{
		Location:    "New York",
		Temperature: 25.3,
		Description: "Partly cloudy",
	}

	// Forecast data for London
	londonForecast := Forecast{
		Location:    "London",
		Temperature: 18.0,
		Description: "Rainy",
	}

	fmt.Println("Weather Forecast:")
	fmt.Println(nyForecast.Location, ":", nyForecast.Temperature, "°C, ", nyForecast.Description)
	fmt.Println(londonForecast.Location, ":", londonForecast.Temperature, "°C, ", londonForecast.Description)

	// Fetch weather data from an external API (simulated here)
	// ...
}