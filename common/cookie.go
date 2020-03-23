package common


import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	//"google.golang.org/appengine/datastore"
	"net/http"
	"time"
	//"github.com/mssola/user_agent"
	"strconv"
)

type Visitor struct {
	Key                   string `json:"key,omitempty"`
	Cookie                string `json:"cookie,omitempty"`
	Host                  string `json:"host,omitempty"`
	CreatedTimestamp      string `json:"createdTimestamp,omitempty"`
	CreatedIP             string `json:"createdIPO,omitempty"`
	CreatedReferer        string `json:"createdReferer,omitempty"`
	CreatedCountry        string `json:"createdCountry,omitempty"`
	CreatedRegion         string `json:"createdRegion,omitempty"`
	CreatedCity           string `json:"createdCity,omitempty"`
	CreatedUserAgent      string `json:"createdUserAgent,omitempty"`
	CreatedIsMobile       bool   `json:"createdIsMobile,omitempty"`
	CreatedIsBot          bool   `json:"createdIsBot,omitempty"`
	CreatedMozillaVersion string `json:"createdMozillaVersion,omitempty"`
	CreatedPlatform       string `json:"createdPlatform,omitempty"`
	CreatedOS             string `json:"createdOS,omitempty"`
	CreatedEngineName     string `json:"createdEngineName,omitempty"`
	CreatedEngineVersion  string `json:"createdEngineVersion,omitempty"`
	CreatedBrowserName    string `json:"createdBrowserName,omitempty"`
	CreatedBrowserVersion string `json:"createdBrowserVersion,omitempty"`
}

func ClearCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "ID",
		Value:   "",
		Path:    "/",
		Expires: time.Now(),
	})
}

func DoesCookieExists(r *http.Request) bool {
	cookie, err := r.Cookie("ID")
	if err != nil || cookie == nil || cookie.Value == "" {
		return false
	}
	return true
}

func GetCookieID(w http.ResponseWriter, r *http.Request) string {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>> GetCookieID")

	var id string
	cookie, err := r.Cookie("ID")
	log.Infof(c, "ID cookie: %v", cookie)
	if err != nil || cookie == nil || cookie.Value == "" {
		log.Infof(c, "Error: %v", err)
		log.Infof(c, "New Cookie...")
		ts := strconv.FormatInt(time.Now().UnixNano(), 10)
		id = MD5(ts + r.RemoteAddr)
		http.SetCookie(w, &http.Cookie{
			Name:    "ID",
			Value:   id,
			Path:    "/",
			Domain:  r.Host,
			Expires: time.Now().Add(time.Hour * 24 * 30),
		})
		log.Infof(c, "New Cookie = %v", id)
		/*
		key := datastore.NewKey(c, "Visitors", id, 0, nil)
		ua := user_agent.New(r.Header.Get("User-Agent"))
		engineName, engineversion := ua.Engine()
		browserName, browserVersion := ua.Browser()
		newVisitor := Visitor{
			Key:                   id,
			Cookie:                id,
			CreatedTimestamp:      ts,
			Host:                  r.Host,
			CreatedIP:             r.RemoteAddr,
			CreatedReferer:        r.Header.Get("Referer"),
			CreatedCountry:        r.Header.Get("X-AppEngine-Country"),
			CreatedRegion:         r.Header.Get("X-AppEngine-Region"),
			CreatedCity:           r.Header.Get("X-AppEngine-City"),
			CreatedUserAgent:      r.Header.Get("User-Agent"),
			CreatedIsMobile:       ua.Mobile(),
			CreatedIsBot:          ua.Bot(),
			CreatedMozillaVersion: ua.Mozilla(),
			CreatedPlatform:       ua.Platform(),
			CreatedOS:             ua.OS(),
			CreatedEngineName:     engineName,
			CreatedEngineVersion:  engineversion,
			CreatedBrowserName:    browserName,
			CreatedBrowserVersion: browserVersion,
		}
		key, err := datastore.Put(c, key, &newVisitor)
		if err != nil {
			log.Errorf(c, "Error while storing cookie %v in datastore: %v", id, err)
		} else {
			log.Infof(c, "New visitor %v stored in datastore under key %v", id,
				key.IntID())
		}
		*/
	} else {
		id = cookie.Value
		log.Infof(c, "Existing ID Cookie = %v", id)
	}
	return id
}