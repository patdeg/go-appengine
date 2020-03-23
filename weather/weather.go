package weather

import (
	"github.com/patdeg/go-appengine/common"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type Coord struct {
	Lon float32 `json:"lon,omitempty"`
	Lat float32 `json:"lat,omitempty"`
}

type Weather struct {
	Id          int64  `json:"id,omitempty"`
	Main        string `json:"id,omitempty"`
	description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

type Main struct {
	// Temperature. Unit Default: Kelvin, Metric: Celsius, Imperial: Fahrenheit.
	Temp float32 `json:"temp,omitempty"`

	// Atmospheric pressure (on the sea level, if there is no sea_level or grnd_level data), hPa
	Pressure float32 `json:"pressure,omitempty"`

	// Humidity, %
	Humidity float32 `json:"humidity,omitempty"`

	// Minimum temperature at the moment. This is deviation from current temp that is possible for large cities and megalopolises geographically expanded (use these parameter optionally). Unit Default: Kelvin, Metric: Celsius, Imperial: Fahrenheit.
	TempMin float32 `json:"temp_min,omitempty"`

	// Maximum temperature at the moment. This is deviation from current temp that is possible for large cities and megalopolises geographically expanded (use these parameter optionally). Unit Default: Kelvin, Metric: Celsius, Imperial: Fahrenheit.
	TempMax float32 `json:"temp_max,omitempty"`

	// Atmospheric pressure on the sea level, hPa
	SeaLevel float32 `json:"sea_level,omitempty"`

	// Atmospheric pressure on the ground level, hPa
	GrndLevel float32 `json:"grnd_level,omitempty"`
}

type Wind struct {
	// Wind speed. Unit Default: meter/sec, Metric: meter/sec, Imperial: miles/hour.
	speed float32 `json:"speed,omitempty"`

	// Wind direction, degrees (meteorological)
	deg float32 `json:"deg,omitempty"`
}

type Clouds struct {
	// Cloudiness, %
	All float32 `json:"all,omitempty"`
}

type Historic struct {
	// Rain or snow volume for the last 1 hour
	Last1Hour float32 `json:"1h,omitempty"`

	// Rain or snow volume for the last 3 hours
	Last3Hours float32 `json:"3h,omitempty"`
}

type Sys struct {
	// Internal parameter
	Type int64 `json:"type,omitempty"`

	// Internal parameter
	Id int64 `json:"id,omitempty"`

	// Internal parameter
	Message float32 `json:"message,omitempty"`

	// Country code (GB, JP etc.)
	Country string `json:"country,omitempty"`

	// Sunrise time, unix, UTC
	Sunrise int64 `json:"sunrise,omitempty"`

	// Sunset time, unix, UTC
	Sunset int64 `json:"sunset,omitempty"`
}

type CurrentCondition struct {
	//
	Coord Coord `json:"coord,omitempty"`

	//
	Weather []Weather `json:"weather,omitempty"`

	// Internal parameter
	base string `json:"base,omitempty"`

	//
	Main Main `json:"main,omitempty"`

	//
	Wind Wind `json:"wind,omitempty"`

	//
	Clouds Clouds `json:"clouds,omitempty"`

	//
	Rain Historic `json:"rain,omitempty"`

	//
	Snow Historic `json:"snow,omitempty"`

	// Time of data calculation, unix, UTC
	Time int64 `json:"dt,omitempty"`

	//
	Sys Sys `json:"sys,omitempty"`

	// City ID
	Id int64 `json:"id,omitempty"`

	// City name
	Name string `json:"name,omitempty"`

	// Internal parameter
	Cod int64 `json:"cod,omitempty"`
}

func GetWeatherByZipCode(c context.Context, zipcode string, country string) (*CurrentCondition, error) {

	client := urlfetch.Client(c)

	calURL := "http://api.openweathermap.org/data/2.5/weather?q=34233,US"

	resp, err := client.Get(calURL)
	if err != nil {
		log.Errorf(c, "Error getting weather Info: %v", err)
		return nil, err
	}

	var currentCondition CurrentCondition
	common.UnmarshalResponse(c, resp, &currentCondition)

	log.Infof(c, "currentCondition: %v", currentCondition)

	return &currentCondition, nil
}
