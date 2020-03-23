package nest

import (
	"bytes"
	"github.com/patdeg/go-appengine/common"
	"encoding/json"
	"errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

/*
	Info:
		- http://stackoverflow.com/questions/24601798/acquiring-and-changing-basic-data-on-the-nest-thermostat/24616406#24616406

	Weather Info: http://home.openweathermap.org/
	http://api.openweathermap.org/data/2.5/weather?q=34233,US

*/

type Token struct {
	// AccessToken is the token that authorizes and authenticates
	// the requests.
	AccessToken string `json:"access_token"`

	// TokenType is the type of token.
	// The Type method returns either this or "Bearer", the default.
	TokenType string `json:"token_type,omitempty"`

	// RefreshToken is a token that's used by the application
	// (as opposed to the user) to refresh the access token
	// if it expires.
	RefreshToken string `json:"refresh_token,omitempty"`

	// Expiry is the optional expiration time of the access token.
	//
	// If zero, TokenSource implementations will reuse the same
	// token forever and RefreshToken or equivalent
	// mechanisms for that TokenSource will not be used.
	Expiry time.Time `json:"expiry,omitempty"`
	// contains filtered or unexported fields

	ExpiresIn int32 `json:"expires_in,omitempty"`

	JSON string `json:"-"`
}

func NestAuthCodeURL(c context.Context, config *oauth2.Config, state string) string {
	authURL := config.Endpoint.AuthURL
	v := url.Values{}
	v.Set("client_id", config.ClientID)
	v.Set("state", state)
	authURL += "?" + v.Encode()

	return authURL
}

func NestExchange(c context.Context, config *oauth2.Config, code string) (*Token, error) {

	exchangeURL := config.Endpoint.TokenURL

	values := url.Values{}
	values.Set("client_id", config.ClientID)
	values.Set("code", code)
	values.Set("client_secret", config.ClientSecret)
	values.Set("grant_type", "authorization_code")

	log.Debugf(c, "Calling %v with %v", exchangeURL, values.Encode())

	client := urlfetch.Client(c)

	req, err := http.NewRequest("POST", exchangeURL, bytes.NewBufferString(values.Encode()))
	if err != nil {
		log.Errorf(c, "Error while creating request: %v", err)
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "Error while exchanging token: %v", err)
		return nil, err
	}

	common.DumpResponse(c, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf(c, "Error while reading body: %v", err)
		return nil, err
	}

	token, err := TokenDecode(c, body)
	if err != nil {
		return nil, err
	}

	token.Expiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Millisecond)
	DumpToken(c, token)

	return token, nil
}

func TokenDecode(c context.Context, value []byte) (*Token, error) {
	var token Token
	err := json.Unmarshal(value, &token)
	if err != nil {
		log.Errorf(c, "Error while decoing JSON: %v", err)
		log.Infof(c, "JSON: %v", common.B2S(value))
		return nil, err
	}
	token.JSON = common.B2S(value)
	return &token, nil
}

func TokenEncode(c context.Context, token *Token) ([]byte, error) {
	value, err := json.Marshal(token)
	if err != nil {
		log.Errorf(c, "Error while decoing JSON: %v", err)
		log.Infof(c, "JSON: %v", common.B2S(value))
		return []byte{}, err
	}
	return value, nil
}

func GetToken(c context.Context, cookieToken string) *Token {
	if cookieToken == "" {
		log.Errorf(c, "[GetToken] Warning, cookie token empty")
		return nil
	}

	token, err := TokenDecode(c, []byte(cookieToken))
	if err != nil {
		log.Errorf(c, "[GetToken] Error while decoing JSON: %v", err)
		return nil
	}
	return token
}

func DumpToken(c context.Context, token *Token) {
	if token != nil {
		log.Debugf(c, "Token:")
		log.Debugf(c, "  - AccessToken: %v", token.AccessToken)
		log.Debugf(c, "  - TokenType: %v", token.TokenType)
		log.Debugf(c, "  - RefreshToken: %v", token.RefreshToken)
		log.Debugf(c, "  - ExpiresIn: %v", token.ExpiresIn)
		log.Debugf(c, "  - Expiration: %v", token.Expiry)
		log.Debugf(c, "  - JSON: %v", token.JSON)
	} else {
		log.Debugf(c, "Token is null")
	}
}

func GetDevices(c context.Context, accessToken string) {
	client := urlfetch.Client(c)

	resp, err := client.Get("https://developer-api.nest.com/devices/thermostats?auth=" + accessToken)
	if err != nil {
		log.Errorf(c, "Error getting Devices Info: %v", err)
		return
	}
	common.DumpResponse(c, resp)
}

type Thermostat struct {
	DeviceId               string    `json:"device_id, omitempty"`
	Locale                 string    `json:"locale, omitempty"`
	SoftwareVersion        string    `json:"software_version, omitempty"`
	StructureId            string    `json:"structure_id, omitempty"`
	Name                   string    `json:"name, omitempty"`
	NameLong               string    `json:"name_long, omitempty"`
	LastConnection         time.Time `json:"last_connection, omitempty"`
	IsOnline               bool      `json:"is_online, omitempty"`
	CanCool                bool      `json:"can_cool, omitempty"`
	CanHeat                bool      `json:"can_heat, omitempty"`
	IsUsingEmergencyHeat   bool      `json:"is_using_emergency_heat, omitempty"`
	HasFan                 bool      `json:"has_fan, omitempty"`
	FanTimerActive         bool      `json:"fan_timer_active, omitempty"`
	FanTimerTimeout        time.Time `json:"fan_timer_timeout, omitempty"`
	HasLeaf                bool      `json:"has_leaf, omitempty"`
	TemperatureScale       string    `json:"temperature_scale, omitempty"`
	TargetTemperatureF     float32   `json:"target_temperature_f, omitempty"`
	TargetTemperatureC     float32   `json:"target_temperature_c, omitempty"`
	TargetTemperatureHighF float32   `json:"target_temperature_high_f, omitempty"`
	TargetTemperatureHighC float32   `json:"target_temperature_high_c, omitempty"`
	TargetTemperatureLowF  float32   `json:"target_temperature_low_f, omitempty"`
	TargetTemperatureLowC  float32   `json:"target_temperature_low_c, omitempty"`
	AwayTemperatureHighF   float32   `json:"away_temperature_high_f, omitempty"`
	AwayTemperatureHighC   float32   `json:"away_temperature_high_c, omitempty"`
	AwayTemperatureLowF    float32   `json:"away_temperature_low_f, omitempty"`
	AwayTemperatureLowC    float32   `json:"away_temperature_low_c, omitempty"`
	HvacMode               string    `json:"hvac_mode, omitempty"`
	AmbientTemperatureF    float32   `json:"ambient_temperature_f, omitempty"`
	AmbientTemperatureC    float32   `json:"ambient_temperature_c, omitempty"`
	Humidity               float32   `json:"humidity, omitempty"`
	HvacState              string    `json:"hvac_state, omitempty"`
	WhereId                string    `json:"where_id, omitempty"`
}

type Device struct {
	Thermostats map[string]Thermostat `json:"thermostats, omitempty"`
}

type StructureDevicesCompany struct {
	ProductType []string `json:"$product_type, omitempty"`
}

type StructureDevices struct {
	Company StructureDevicesCompany `json:"$company, omitempty"`
}

type ETA struct {
	TripId                      string    `json:"trip_id, omitempty"`
	EstimatedArrivalWindowBegin time.Time `json:"estimated_arrival_window_begin, omitempty"`
	EstimatedArrivalWindowEnd   time.Time `json:"estimated_arrival_window_end, omitempty"`
}

type Where struct {
	WhereId string `json:"where_id, omitempty"`
	Name    string `json:"name, omitempty"`
}

type Structure struct {
	StructureId         string           `json:"structure_id, omitempty"`
	Thermostats         []string         `json:"thermostats, omitempty"`
	SmokeCOAlarms       []string         `json:"smoke_co_alarms, omitempty"`
	Devices             StructureDevices `json:"devices, omitempty"`
	Away                string           `json:"away, omitempty"`
	Name                string           `json:"name, omitempty"`
	CountryCode         string           `json:"country_code, omitempty"`
	PostalCode          string           `json:"postal_code, omitempty"`
	PeakPeriodStartTime time.Time        `json:"peak_period_start_time, omitempty"`
	PeakPeriodEndTime   time.Time        `json:"peak_period_end_time, omitempty"`
	TimeZone            string           `json:"time_zone, omitempty"`
	ETA                 ETA              `json:"eta, omitempty"`
	Wheres              map[string]Where `json:"wheres, omitempty"`
}

type MetaData struct {
	AccessToken   string  `json:"access_token, omitempty"`
	ClientVersion float32 `json:"client_version, omitempty"`
}

type GetDataCall struct {
	MetaData   MetaData             `json:"metadata, omitempty"`
	Devices    Device               `json:"devices, omitempty"`
	Structures map[string]Structure `json:"structures, omitempty"`
}

func GetData(c context.Context, token *Token) (*GetDataCall, error) {

	if token == nil {
		log.Errorf(c, "Error no token")
		return nil, errors.New("Error no token")
	}

	if token.AccessToken == "" {
		log.Errorf(c, "Error no access token")
		return nil, errors.New("Error no access token")
	}

	client := urlfetch.Client(c)

	resp, err := client.Get("https://developer-api.nest.com/?auth=" + token.AccessToken)
	if err != nil {
		log.Errorf(c, "Error getting Devices Info: %v", err)
		return nil, err
	}

	var getDataCall GetDataCall
	common.UnmarshalResponse(c, resp, &getDataCall)

	log.Infof(c, "GetData: %v", getDataCall)
	return &getDataCall, nil
}
