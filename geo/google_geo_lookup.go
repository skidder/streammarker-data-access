package geo

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/net/context"
	"googlemaps.github.io/maps"
)

// GoogleGeoLookup to lookup timezone details
type GoogleGeoLookup struct {
	googleAPIKey  string
	mapsClient    *maps.Client
	locationCache *cache.Cache
}

// TimezoneInfo has timezone details
type TimezoneInfo struct {
	TimeZoneID   string
	TimeZoneName string
}

// NewGoogleGeoLookup constructs a new GoogleGeoLookup instance
func NewGoogleGeoLookup(googleAPIKey string) *GoogleGeoLookup {
	return &GoogleGeoLookup{
		googleAPIKey:  googleAPIKey,
		locationCache: cache.New(1*time.Hour, 1*time.Minute),
	}
}

// Initialize a GoogleGeoLookup instance
func (g *GoogleGeoLookup) Initialize() error {
	var err error
	g.mapsClient, err = maps.NewClient(maps.WithAPIKey(g.googleAPIKey))
	return err
}

// FindTimezoneForLocation get timezone for a given location
func (g *GoogleGeoLookup) FindTimezoneForLocation(latitude, longitude float64) (*TimezoneInfo, error) {
	// check local cache for timezone info for this location
	locationHash := fmt.Sprintf("%f,%f", latitude, longitude)
	if x, found := g.locationCache.Get(locationHash); found {
		return x.(*TimezoneInfo), nil
	}

	// cache-miss, request the timezone info from Google Maps API
	r := &maps.TimezoneRequest{
		Location: &maps.LatLng{
			Lat: latitude,
			Lng: longitude,
		},
		Timestamp: time.Now(),
	}

	resp, e := g.mapsClient.Timezone(context.Background(), r)
	if e == nil {
		// store results in local cache
		locationInfo := &TimezoneInfo{TimeZoneID: resp.TimeZoneID, TimeZoneName: resp.TimeZoneName}
		g.locationCache.Set(locationHash, locationInfo, cache.DefaultExpiration)
		return locationInfo, nil
	}
	return nil, e
}
