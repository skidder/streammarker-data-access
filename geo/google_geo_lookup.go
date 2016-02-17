package geo

import (
	"time"

	"golang.org/x/net/context"
	"googlemaps.github.io/maps"
)

type GoogleGeoLookup struct {
	googleApiKey string
	mapsClient   *maps.Client
}

type TimezoneInfo struct {
	TimeZoneID   string
	TimeZoneName string
}

func NewGoogleGeoLookup(googleApiKey string) *GoogleGeoLookup {
	return &GoogleGeoLookup{googleApiKey: googleApiKey}
}

func (g *GoogleGeoLookup) Init() error {
	var err error
	g.mapsClient, err = maps.NewClient(maps.WithAPIKey(g.googleApiKey))
	return err
}

func (g *GoogleGeoLookup) FindTimezoneForLocation(latitude, longitude float64) (*TimezoneInfo, error) {
	r := &maps.TimezoneRequest{
		Location: &maps.LatLng{
			Lat: latitude,
			Lng: longitude,
		},
		Timestamp: time.Now(),
	}

	resp, e := g.mapsClient.Timezone(context.Background(), r)
	if e == nil {
		return &TimezoneInfo{TimeZoneID: resp.TimeZoneID, TimeZoneName: resp.TimeZoneName}, nil
	}
	return nil, e
}
