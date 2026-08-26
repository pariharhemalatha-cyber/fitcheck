package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const openMeteoURL = "https://api.open-meteo.com/v1/forecast"

// DailyForecast holds one day of weather for outfit planning.
type DailyForecast struct {
	Date              string  `json:"date"`
	TempHighC         float64 `json:"temp_high_c"`
	TempLowC          float64 `json:"temp_low_c"`
	PrecipProbability float64 `json:"precip_probability"`
	WindSpeedKmh      float64 `json:"wind_speed_kmh"`
	WeatherCode       int     `json:"weather_code"`
}

type openMeteoResponse struct {
	Daily struct {
		Time                    []string  `json:"time"`
		Temperature2mMax        []float64 `json:"temperature_2m_max"`
		Temperature2mMin        []float64 `json:"temperature_2m_min"`
		PrecipitationProbabilityMax []float64 `json:"precipitation_probability_max"`
		WindSpeed10mMax         []float64 `json:"wind_speed_10m_max"`
		WeatherCode             []int     `json:"weather_code"`
	} `json:"daily"`
}

// GetForecast fetches daily weather from Open-Meteo (free, no API key).
func GetForecast(ctx context.Context, lat, lng float64, startDate time.Time, days int) ([]DailyForecast, error) {
	if days < 1 {
		days = 1
	}
	if days > 16 {
		days = 16
	}

	q := url.Values{}
	q.Set("latitude", fmt.Sprintf("%.4f", lat))
	q.Set("longitude", fmt.Sprintf("%.4f", lng))
	q.Set("daily", "temperature_2m_max,temperature_2m_min,precipitation_probability_max,wind_speed_10m_max,weather_code")
	q.Set("timezone", "auto")
	q.Set("forecast_days", fmt.Sprintf("%d", days))
	if !startDate.IsZero() {
		q.Set("start_date", startDate.Format("2006-01-02"))
		end := startDate.AddDate(0, 0, days-1)
		q.Set("end_date", end.Format("2006-01-02"))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, openMeteoURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("open-meteo status %d", resp.StatusCode)
	}

	var raw openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode open-meteo response: %w", err)
	}

	n := len(raw.Daily.Time)
	if n == 0 {
		return nil, fmt.Errorf("no forecast data returned")
	}
	if n > days {
		n = days
	}

	out := make([]DailyForecast, n)
	for i := 0; i < n; i++ {
		out[i] = DailyForecast{
			Date:              raw.Daily.Time[i],
			TempHighC:         at(raw.Daily.Temperature2mMax, i),
			TempLowC:          at(raw.Daily.Temperature2mMin, i),
			PrecipProbability: at(raw.Daily.PrecipitationProbabilityMax, i),
			WindSpeedKmh:      at(raw.Daily.WindSpeed10mMax, i),
			WeatherCode:       atInt(raw.Daily.WeatherCode, i),
		}
	}
	return out, nil
}

func at(vals []float64, i int) float64 {
	if i < len(vals) {
		return vals[i]
	}
	return 0
}

func atInt(vals []int, i int) int {
	if i < len(vals) {
		return vals[i]
	}
	return 0
}

// IsRainy returns true when precipitation probability suggests rain gear.
func (d DailyForecast) IsRainy() bool {
	return d.PrecipProbability >= 40
}

// IsCold returns true when low temps suggest layering.
func (d DailyForecast) IsCold() bool {
	return d.TempLowC < 12
}

// IsHot returns true when highs suggest lighter clothing.
func (d DailyForecast) IsHot() bool {
	return d.TempHighC >= 28
}
