package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const nominatimURL = "https://nominatim.openstreetmap.org/search"

// GeoResult is a resolved city/location with coordinates.
type GeoResult struct {
	DisplayName string  `json:"display_name"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
}

type nominatimResponse []struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
}

// Geocode resolves a city or place name to coordinates using OpenStreetMap Nominatim.
func Geocode(ctx context.Context, city string) (*GeoResult, error) {
	city = strings.TrimSpace(city)
	if city == "" {
		return nil, fmt.Errorf("city name is required")
	}

	q := url.Values{}
	q.Set("q", city)
	q.Set("format", "json")
	q.Set("limit", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nominatimURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "FitCheck/1.0 (personal outfit app)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("nominatim status %d", resp.StatusCode)
	}

	var results nominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("decode nominatim response: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("location not found: %s", city)
	}

	var lat, lng float64
	if _, err := fmt.Sscanf(results[0].Lat, "%f", &lat); err != nil {
		return nil, fmt.Errorf("parse lat: %w", err)
	}
	if _, err := fmt.Sscanf(results[0].Lon, "%f", &lng); err != nil {
		return nil, fmt.Errorf("parse lng: %w", err)
	}

	return &GeoResult{
		DisplayName: results[0].DisplayName,
		Lat:         lat,
		Lng:         lng,
	}, nil
}
