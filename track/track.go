package track

import (
	"github.com/patdeg/go-appengine/common"
	"errors"
	"fmt"
	"github.com/mssola/user_agent"
	"golang.org/x/net/context"
	bigquery "google.golang.org/api/bigquery/v2"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/user"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const onePixelPNG = "\x89\x50\x4e\x47\x0d\x0a\x1a\x0a\x00\x00\x00\x0d\x49\x48" +
	"\x44\x52\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90\x77\x53" +
	"\xde\x00\x00\x00\x01\x73\x52\x47\x42\x00\xae\xce\x1c\xe9\x00\x00\x00\x04" +
	"\x67\x41\x4d\x41\x00\x00\xb1\x8f\x0b\xfc\x61\x05\x00\x00\x00\x09\x70\x48" +
	"\x59\x73\x00\x00\x0e\xc3\x00\x00\x0e\xc3\x01\xc7\x6f\xa8\x64\x00\x00\x00" +
	"\x0c\x49\x44\x41\x54\x18\x57\x63\xf8\xff\xff\x3f\x00\x05\xfe\x02\xfe\xa7" +
	"\x35\x81\x84\x00\x00\x00\x00\x49\x45\x4e\x44\xae\x42\x60\x82"

type Visit struct {
	DatastoreKey   *datastore.Key `json:"datastoreKey" datastore:"-"`
	Cookie         string         `json:"cookie,omitempty"`
	Session        string         `json:"session,omitempty"`
	URI            string         `json:"uri,omitempty"`
	Referer        string         `json:"referer,omitempty"`
	Time           time.Time      `json:"time,omitempty"`
	Host           string         `json:"host,omitempty"`
	RemoteAddr     string         `json:"remoteAddr,omitempty"`
	InstanceId     string         `json:"instanceId,omitempty"`
	VersionId      string         `json:"versionId,omitempty"`
	Scheme         string         `json:"scheme,omitempty"`
	Country        string         `json:"country,omitempty"`
	Region         string         `json:"region,omitempty"`
	City           string         `json:"city,omitempty"`
	Lat            float64        `json:"lat,omitempty"`
	Lon            float64        `json:"lon,omitempty"`
	AcceptLanguage string         `json:"acceptLanguage,omitempty"`
	UserAgent      string         `json:"userAgent,omitempty"`
	IsMobile       bool           `json:"isMobile,omitempty"`
	IsBot          bool           `json:"isBot,omitempty"`
	MozillaVersion string         `json:"mozillaVersion,omitempty"`
	Platform       string         `json:"platform,omitempty"`
	OS             string         `json:"os,omitempty"`
	EngineName     string         `json:"engineName,omitempty"`
	EngineVersion  string         `json:"engineVersion,omitempty"`
	BrowserName    string         `json:"browserName,omitempty"`
	BrowserVersion string         `json:"browserVersion,omitempty"`
	Category       string         `json:"category,omitempty"`
	Action         string         `json:"action,omitempty"`
	Label          string         `json:"label,omitempty"`
	Value          float64        `json:"value,omitempty"`
}

type RobotPage struct {
	Time       time.Time `json:"time,omitempty"`
	Name       string    `json:"name,omitempty"`
	URL        string    `json:"url,omitempty"`
	URI        string    `json:"uri,omitempty"`
	Host       string    `json:"host,omitempty"`
	RemoteAddr string    `json:"remoteAddr,omitempty"`
	UserAgent  string    `json:"userAgent,omitempty"`
	Country    string    `json:"country,omitempty"`
	Region     string    `json:"region,omitempty"`
	City       string    `json:"city,omitempty"`
	BotName    string    `json:"botName,omitempty"`
	BotVersion string    `json:"botVersion,omitempty"`
}

func createVisitsTableInBigQuery(c context.Context, d string) error {

	log.Infof(c, ">>>> createVisitsTableInBigQuery")

	log.Infof(c, "Create a new daily visits table in BigQuery")

	if len(d) != 8 {
		return errors.New("table name is badly formated - expected 8 characters")
	}
	newTable := &bigquery.Table{
		TableReference: &bigquery.TableReference{
			ProjectId: "myproject",
			DatasetId: "visits",
			TableId:   d,
		},
		FriendlyName: "Daily Visits table",
		Description:  "This table is created automatically to store daily visits to Deglon Consulting properties ",
		//ExpirationTime: expirationTime.Unix() * 1000,
		Schema: &bigquery.TableSchema{
			Fields: []*bigquery.TableFieldSchema{
				{Name: "Cookie", Type: "STRING", Description: "Cookie"},
				{Name: "Session", Type: "STRING", Description: "Session"},
				{Name: "URI", Type: "STRING", Description: "URI"},
				{Name: "Referer", Type: "STRING", Description: "Referer"},
				{Name: "Time", Type: "TIMESTAMP", Description: "Time"},
				{Name: "Host", Type: "STRING", Description: "Host"},
				{Name: "RemoteAddr", Type: "STRING", Description: "RemoteAddr"},
				{Name: "InstanceId", Type: "STRING", Description: "InstanceId"},
				{Name: "VersionId", Type: "STRING", Description: "VersionId"},
				{Name: "Scheme", Type: "STRING", Description: "Scheme"},
				{Name: "Country", Type: "STRING", Description: "Country"},
				{Name: "Region", Type: "STRING", Description: "Region"},
				{Name: "City", Type: "STRING", Description: "City"},
				{Name: "Lat", Type: "FLOAT", Description: "City latitude"},
				{Name: "Lon", Type: "FLOAT", Description: "City longitude"},
				{Name: "AcceptLanguage", Type: "STRING", Description: "AcceptLanguage"},
				{Name: "UserAgent", Type: "STRING", Description: "UserAgent"},
				{Name: "IsMobile", Type: "BOOLEAN", Description: "IsMobile"},
				{Name: "IsBot", Type: "BOOLEAN", Description: "IsBot"},
				{Name: "MozillaVersion", Type: "STRING", Description: "MozillaVersion"},
				{Name: "Platform", Type: "STRING", Description: "Platform"},
				{Name: "OS", Type: "STRING", Description: "OS"},
				{Name: "EngineName", Type: "STRING", Description: "EngineName"},
				{Name: "EngineVersion", Type: "STRING", Description: "EngineVersion"},
				{Name: "BrowserName", Type: "STRING", Description: "BrowserName"},
				{Name: "BrowserVersion", Type: "STRING", Description: "BrowserVersion"},
			},
		},
	}

	return common.CreateTableInBigQuery(c, newTable)
}

func createEventsTableInBigQuery(c context.Context, d string) error {

	log.Infof(c, ">>>> createEventsTableInBigQuery")

	log.Infof(c, "Create a new daily visits table in BigQuery")

	if len(d) != 8 {
		return errors.New("table name is badly formated - expected 8 characters")
	}
	newTable := &bigquery.Table{
		TableReference: &bigquery.TableReference{
			ProjectId: "myproject",
			DatasetId: "events",
			TableId:   d,
		},
		FriendlyName: "Daily Visits table",
		Description:  "This table is created automatically to store daily visits to Deglon Consulting properties ",
		//ExpirationTime: expirationTime.Unix() * 1000,
		Schema: &bigquery.TableSchema{
			Fields: []*bigquery.TableFieldSchema{
				{Name: "Cookie", Type: "STRING", Description: "Cookie"},
				{Name: "Session", Type: "STRING", Description: "Session"},
				{Name: "Category", Type: "STRING", Description: "Session"},
				{Name: "Action", Type: "STRING", Description: "Action"},
				{Name: "Label", Type: "STRING", Description: "Label"},
				{Name: "Value", Type: "FLOAT", Description: "Value"},
				{Name: "URI", Type: "STRING", Description: "URI"},
				{Name: "Referer", Type: "STRING", Description: "Referer"},
				{Name: "Time", Type: "TIMESTAMP", Description: "Time"},
				{Name: "Host", Type: "STRING", Description: "Host"},
				{Name: "RemoteAddr", Type: "STRING", Description: "RemoteAddr"},
				{Name: "InstanceId", Type: "STRING", Description: "InstanceId"},
				{Name: "VersionId", Type: "STRING", Description: "VersionId"},
				{Name: "Scheme", Type: "STRING", Description: "Scheme"},
				{Name: "Country", Type: "STRING", Description: "Country"},
				{Name: "Region", Type: "STRING", Description: "Region"},
				{Name: "City", Type: "STRING", Description: "City"},
				{Name: "Lat", Type: "FLOAT", Description: "City latitude"},
				{Name: "Lon", Type: "FLOAT", Description: "City longitude"},
				{Name: "AcceptLanguage", Type: "STRING", Description: "AcceptLanguage"},
				{Name: "UserAgent", Type: "STRING", Description: "UserAgent"},
				{Name: "IsMobile", Type: "BOOLEAN", Description: "IsMobile"},
				{Name: "IsBot", Type: "BOOLEAN", Description: "IsBot"},
				{Name: "MozillaVersion", Type: "STRING", Description: "MozillaVersion"},
				{Name: "Platform", Type: "STRING", Description: "Platform"},
				{Name: "OS", Type: "STRING", Description: "OS"},
				{Name: "EngineName", Type: "STRING", Description: "EngineName"},
				{Name: "EngineVersion", Type: "STRING", Description: "EngineVersion"},
				{Name: "BrowserName", Type: "STRING", Description: "BrowserName"},
				{Name: "BrowserVersion", Type: "STRING", Description: "BrowserVersion"},
			},
		},
	}

	return common.CreateTableInBigQuery(c, newTable)
}

func CreateTodayVisitsTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> CreateTodayVisitsTableInBigQueryHandler")

	isAdmin := false
	if user.Current(c) != nil {
		isAdmin = user.Current(c).Admin
	}

	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		log.Errorf(c, "Handler called without admin/cron priviledge")
		http.Error(w, "Handler called without admin/cron priviledge", http.StatusBadRequest)
		return
	}

	today := time.Now().Format("20060102")
	err := createVisitsTableInBigQuery(c, today)
	if err != nil {
		log.Errorf(c, "Error while creating table %v: %v", today, err)
		http.Error(w, "Error while creating today table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", today)

}

func CreateTomorrowVisitsTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> CreateTomorrowVisitsTableInBigQueryHandler")

	isAdmin := false
	if user.Current(c) != nil {
		isAdmin = user.Current(c).Admin
	}

	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		log.Errorf(c, "Handler called without admin/cron priviledge")
		http.Error(w, "Handler called without admin/cron priviledge", http.StatusBadRequest)
		return
	}

	tomorrow := time.Now().Add(time.Hour*23 + time.Minute*59).Format("20060102")
	err := createVisitsTableInBigQuery(c, tomorrow)
	if err != nil {
		log.Errorf(c, "Error while creating table %v: %v", tomorrow, err)
		http.Error(w, "Error while creating tomorrow table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", tomorrow)

}

func CreateTodayEventsTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> CreateTomorrowEventsTableInBigQueryHandler")

	isAdmin := false
	if user.Current(c) != nil {
		isAdmin = user.Current(c).Admin
	}

	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		log.Errorf(c, "Handler called without admin/cron priviledge")
		http.Error(w, "Handler called without admin/cron priviledge", http.StatusBadRequest)
		return
	}

	today := time.Now().Format("20060102")
	err := createEventsTableInBigQuery(c, today)
	if err != nil {
		log.Errorf(c, "Error while creating table %v: %v", today, err)
		http.Error(w, "Error while creating today table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", today)

}

func CreateTomorrowEventsTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> CreateTomorrowEventsTableInBigQueryHandler")

	isAdmin := false
	if user.Current(c) != nil {
		isAdmin = user.Current(c).Admin
	}

	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		log.Errorf(c, "Handler called without admin/cron priviledge")
		http.Error(w, "Handler called without admin/cron priviledge", http.StatusBadRequest)
		return
	}

	tomorrow := time.Now().Add(time.Hour*23 + time.Minute*59).Format("20060102")
	err := createEventsTableInBigQuery(c, tomorrow)
	if err != nil {
		log.Errorf(c, "Error while creating table %v: %v", tomorrow, err)
		http.Error(w, "Error while creating tomorrow table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", tomorrow)

}

func StoreVisitInBigQuery(c context.Context, v *Visit) error {
	log.Infof(c, ">>>> StoreVisitInBigQuery")

	insertId := strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + v.Cookie

	req := &bigquery.TableDataInsertAllRequest{
		Kind: "bigquery#tableDataInsertAllRequest",
		Rows: []*bigquery.TableDataInsertAllRequestRows{
			{
				InsertId: insertId,
				Json: map[string]bigquery.JsonValue{
					"Cookie":         v.Cookie,
					"Session":        v.Session,
					"URI":            v.URI,
					"Referer":        v.Referer,
					"Time":           v.Time,
					"Host":           v.Host,
					"RemoteAddr":     v.RemoteAddr,
					"InstanceId":     v.InstanceId,
					"VersionId":      v.VersionId,
					"Scheme":         v.Scheme,
					"Country":        v.Country,
					"Region":         v.Region,
					"City":           v.City,
					"Lat":            v.Lat,
					"Lon":            v.Lon,
					"AcceptLanguage": v.AcceptLanguage,
					"UserAgent":      v.UserAgent,
					"IsMobile":       v.IsMobile,
					"IsBot":          v.IsBot,
					"MozillaVersion": v.MozillaVersion,
					"Platform":       v.Platform,
					"OS":             v.OS,
					"EngineName":     v.EngineName,
					"EngineVersion":  v.EngineVersion,
					"BrowserName":    v.BrowserName,
					"BrowserVersion": v.BrowserVersion,
				},
			},
		},
	}

	tableName := time.Now().Format("20060102")

	err := common.StreamDataInBigquery(c, "myproject", "visits", tableName, req)
	if err != nil {
		log.Errorf(c, "Error while streaming visit to BigQuery: %v", err)
		return err
	}
	return nil
}

func StoreEventInBigQuery(c context.Context, v *Visit) error {

	log.Infof(c, ">>>> StoreVisitInBigQuery")

	insertId := strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + v.Cookie

	req := &bigquery.TableDataInsertAllRequest{
		Kind: "bigquery#tableDataInsertAllRequest",
		Rows: []*bigquery.TableDataInsertAllRequestRows{
			{
				InsertId: insertId,
				Json: map[string]bigquery.JsonValue{
					"Cookie":         v.Cookie,
					"Session":        v.Session,
					"URI":            v.URI,
					"Referer":        v.Referer,
					"Time":           v.Time,
					"Host":           v.Host,
					"RemoteAddr":     v.RemoteAddr,
					"InstanceId":     v.InstanceId,
					"VersionId":      v.VersionId,
					"Scheme":         v.Scheme,
					"Country":        v.Country,
					"Region":         v.Region,
					"City":           v.City,
					"Lat":            v.Lat,
					"Lon":            v.Lon,
					"AcceptLanguage": v.AcceptLanguage,
					"UserAgent":      v.UserAgent,
					"IsMobile":       v.IsMobile,
					"IsBot":          v.IsBot,
					"MozillaVersion": v.MozillaVersion,
					"Platform":       v.Platform,
					"OS":             v.OS,
					"EngineName":     v.EngineName,
					"EngineVersion":  v.EngineVersion,
					"BrowserName":    v.BrowserName,
					"BrowserVersion": v.BrowserVersion,
					"Category":       v.Category,
					"Action":         v.Action,
					"Label":          v.Label,
					"Value":          v.Value,
				},
			},
		},
	}

	tableName := time.Now().Format("20060102")

	err := common.StreamDataInBigquery(c, "myproject", "events", tableName, req)
	if err != nil {
		log.Errorf(c, "Error while streaming visit to BigQuery: %v", err)
		return err
	}
	return nil
}

func TrackVisit(w http.ResponseWriter, r *http.Request, cookie string) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>> TrackVisit")

	if _, err := memcache.Get(c, "visit-"+cookie); err == memcache.ErrCacheMiss {
		log.Infof(c, "Cookie not in memcache")
	} else if err != nil {
		log.Errorf(c, "Error getting item: %v", err)
	} else {
		log.Infof(c, "Cookie in memcache, do not track visit again")
		return
	}

	ua := user_agent.New(r.Header.Get("User-Agent"))
	engineName, engineversion := ua.Engine()
	browserName, browserVersion := ua.Browser()

	if common.IsBot(r.Header.Get("User-Agent")) {
		log.Infof(c, "TrackVisit: Events from Bots, ignoring")
		return
	}

	if r.Header.Get("X-AppEngine-Country") == "ZZ" {
		log.Infof(c, "TrackVisit: Country is ZZ - most likely a bot, ignoring")
		return
	}

	lat := float64(0)
	lon := float64(0)
	latlon := strings.Split(r.Header.Get("X-AppEngine-CityLatLong"), ",")
	if len(latlon) == 2 {
		lat = common.S2F(latlon[0])
		lon = common.S2F(latlon[1])
	}

	session := ""
	item, err := memcache.Get(c, "session-"+cookie)
	if err != nil {
		// Errror retrieving cookie, store new cookie
		session = strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + cookie
		item = &memcache.Item{
			Key:        "session-" + cookie,
			Value:      []byte(session),
			Expiration: time.Minute * 30,
		}
		if err := memcache.Add(c, item); err == memcache.ErrNotStored {
			log.Infof(c, "TrackEventDetails: item with key %q already exists", item.Key)
		} else if err != nil {
			log.Errorf(c, "TrackEventDetails: Error adding item: %v", err)
		}
	} else {
		// Cookie in memcache
		session = common.B2S(item.Value)
		log.Infof(c, "TrackEventDetails: cookie in memcache: %v", session)
	}
	log.Infof(c, "TrackEventDetails: Session = %v", session)

	visit := &Visit{
		Cookie:         cookie,
		Session:        session,
		URI:            r.RequestURI,
		Referer:        r.Header.Get("Referer"),
		Time:           time.Now(),
		Host:           r.Host,
		RemoteAddr:     r.RemoteAddr,
		InstanceId:     appengine.InstanceID(),
		VersionId:      appengine.VersionID(c),
		Scheme:         r.URL.Scheme,
		Country:        r.Header.Get("X-AppEngine-Country"),
		Region:         r.Header.Get("X-AppEngine-Region"),
		City:           r.Header.Get("X-AppEngine-City"),
		Lat:            lat,
		Lon:            lon,
		AcceptLanguage: r.Header.Get("Accept-Language"),
		UserAgent:      r.Header.Get("User-Agent"),
		IsMobile:       ua.Mobile(),
		IsBot:          ua.Bot(),
		MozillaVersion: ua.Mozilla(),
		Platform:       ua.Platform(),
		OS:             ua.OS(),
		EngineName:     engineName,
		EngineVersion:  engineversion,
		BrowserName:    browserName,
		BrowserVersion: browserVersion,
	}

	err = StoreVisitInBigQuery(c, visit)
	if err != nil {
		log.Errorf(c, "Error while storing visit in datastore: %v", err)
	} else {
		log.Infof(c, "Visit stored in datastore")
	}

}

func TrackEventDetails(w http.ResponseWriter, r *http.Request, cookie, category, action, label string, value float64) {

	c := appengine.NewContext(r)
	log.Infof(c, ">>>> TrackEventDetails")

	ua := user_agent.New(r.Header.Get("User-Agent"))
	engineName, engineversion := ua.Engine()
	browserName, browserVersion := ua.Browser()

	if common.IsBot(r.Header.Get("User-Agent")) {
		log.Infof(c, "TrackEventDetails: Events from Bots, ignoring")
		return
	}

	lat := float64(0)
	lon := float64(0)
	latlon := strings.Split(r.Header.Get("X-AppEngine-CityLatLong"), ",")
	if len(latlon) == 2 {
		lat = common.S2F(latlon[0])
		lon = common.S2F(latlon[1])
	}

	//uniqueId := cookie
	uniqueId := common.MD5(r.RemoteAddr + r.Header.Get("User-Agent"))
	session := ""
	item, err := memcache.Get(c, "s-"+uniqueId)
	if err != nil {
		// Errror retrieving cookie, store new cookie
		session = strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + uniqueId
		item = &memcache.Item{
			Key:        "s-" + uniqueId,
			Value:      []byte(session),
			Expiration: time.Minute * 30,
		}
		if err := memcache.Add(c, item); err == memcache.ErrNotStored {
			log.Infof(c, "TrackEventDetails: item with key %q already exists", item.Key)
		} else if err != nil {
			log.Errorf(c, "TrackEventDetails: Error adding item: %v", err)
		}
	} else {
		// Cookie in memcache
		session = common.B2S(item.Value)
		log.Infof(c, "TrackEventDetails: uniqueid in memcache: %v", session)
	}
	log.Infof(c, "TrackEventDetails: Unique Id = %v Session = %v", uniqueId, session)

	event := &Visit{
		Cookie:         cookie,
		Session:        session,
		URI:            r.RequestURI,
		Referer:        r.Header.Get("Referer"),
		Time:           time.Now(),
		Host:           r.Host,
		RemoteAddr:     r.RemoteAddr,
		InstanceId:     appengine.InstanceID(),
		VersionId:      appengine.VersionID(c),
		Scheme:         r.URL.Scheme,
		Country:        r.Header.Get("X-AppEngine-Country"),
		Region:         r.Header.Get("X-AppEngine-Region"),
		City:           r.Header.Get("X-AppEngine-City"),
		Lat:            lat,
		Lon:            lon,
		AcceptLanguage: r.Header.Get("Accept-Language"),
		UserAgent:      r.Header.Get("User-Agent"),
		IsMobile:       ua.Mobile(),
		IsBot:          ua.Bot(),
		MozillaVersion: ua.Mozilla(),
		Platform:       ua.Platform(),
		OS:             ua.OS(),
		EngineName:     engineName,
		EngineVersion:  engineversion,
		BrowserName:    browserName,
		BrowserVersion: browserVersion,
		Category:       common.Trunc500(category),
		Action:         common.Trunc500(action),
		Label:          common.Trunc500(label),
		Value:          value,
	}

	err = StoreEventInBigQuery(c, event)
	if err != nil {
		log.Errorf(c, "Error while storing event in BigQuery: %v", err)
	} else {
		log.Infof(c, "Event stored in BigQuery")
	}

	/*
		TrackEvent(c, cookie, Trunc500(r.FormValue("C")), Trunc500(r.FormValue("A")),
			Trunc500(r.FormValue("L")), S2F(r.FormValue("V")), r.RemoteAddr, r.Header.Get("User-Agent"))
	*/

}

func TrackEvent(w http.ResponseWriter, r *http.Request, cookie string) {
	log.Infof(appengine.NewContext(r), ">>>> TrackEvent")
	TrackEventDetails(w, r, cookie, r.FormValue("c"), r.FormValue("a"), r.FormValue("l"), common.S2F(r.FormValue("v")))
}

func TrackRobots(r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>> TrackRobots")

	userAgent := r.Header.Get("User-Agent")
	ua := user_agent.New(r.Header.Get("User-Agent"))
	botName, botVersion := ua.Browser()
	robotPage := RobotPage{
		Time:       time.Now(),
		URL:        r.URL.String(),
		URI:        r.RequestURI,
		Host:       r.Host,
		RemoteAddr: r.RemoteAddr,
		UserAgent:  userAgent,
		Country:    r.Header.Get("X-AppEngine-Country"),
		Region:     r.Header.Get("X-AppEngine-Region"),
		City:       r.Header.Get("X-AppEngine-City"),
		BotName:    botName,
		BotVersion: botVersion,
	}
	if strings.Contains(r.RequestURI, "_escaped_fragment_") {
		robotPage.Name = "escaped_fragment"
	}
	if strings.Contains(userAgent, "facebookexternalhit") {
		robotPage.Name = "Facebook"
	}
	if strings.Contains(userAgent, "LinkedInBot") {
		robotPage.Name = "Linkedin"
	}
	if strings.Contains(userAgent, "Googlebot") {
		robotPage.Name = "Google"
	}
	if strings.Contains(userAgent, "OrangeBot") {
		robotPage.Name = "Orange"
	}

	_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "RobotPages", nil), &robotPage)
	if err != nil {
		log.Errorf(c, "Error while storing robot page in datastore: %v", err)
	} else {
		log.Infof(c, "Robot page stored in datastore")
	}
}

func TrackHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> TrackHandler")

	log.Infof(c, "c=%v a=%v l=%v v=%v", r.FormValue("c"), r.FormValue("a"), r.FormValue("l"), r.FormValue("v"))
	TrackEvent(w, r, common.GetCookieID(w, r))
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Write([]byte(onePixelPNG))
}

func ClickHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> ClickHandler")

	log.Infof(c, "c=%v a=%v l=%v v=%v", r.FormValue("c"), r.FormValue("a"), r.FormValue("l"), r.FormValue("v"))
	TrackEvent(w, r, common.GetCookieID(w, r))
	url := r.FormValue("url")
	if url == "" {
		url = "http://myapp.appspot.com"
	}
	log.Infof(c, "Redirect to %v", url)
	http.Redirect(w, r, url, http.StatusFound)
}
