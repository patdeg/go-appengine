package common

import (
	"bytes"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/mail"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	PropertyID string = "UA-63208527-1"
	// PropertyID string = "UA-68699208-1"
)

// https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters
type GAEvent struct {
	CampaignContent        string `json:"CampaignContent,omitempty"`
	CustomDimension1       string `json:"CustomDimension1,omitempty"`
	CustomDimension2       string `json:"CustomDimension2,omitempty"`
	CustomDimension3       string `json:"CustomDimension3,omitempty"`
	CustomDimension4       string `json:"CustomDimension4,omitempty"`
	CustomDimension5       string `json:"CustomDimension5,omitempty"`
	CustomDimension6       string `json:"CustomDimension6,omitempty"`
	CustomDimension7       string `json:"CustomDimension7,omitempty"`
	CustomDimension8       string `json:"CustomDimension8,omitempty"`
	CustomDimension9       string `json:"CustomDimension9,omitempty"`
	Guid                   string `json:"Guid,omitempty"`
	UserId                 string `json:"UserId,omitempty"`
	CampaignKeyword        string `json:"CampaignKeyword,omitempty"`
	CampaignMedium         string `json:"CampaignMedium,omitempty"`
	CustomMetric1          string `json:"CustomMetric1,omitempty"`
	CustomMetric2          string `json:"CustomMetric2,omitempty"`
	CustomMetric3          string `json:"CustomMetric3,omitempty"`
	CustomMetric4          string `json:"CustomMetric4,omitempty"`
	CustomMetric5          string `json:"CustomMetric5,omitempty"`
	CustomMetric6          string `json:"CustomMetric6,omitempty"`
	CustomMetric7          string `json:"CustomMetric7,omitempty"`
	CustomMetric8          string `json:"CustomMetric8,omitempty"`
	CustomMetric9          string `json:"CustomMetric9,omitempty"`
	CampaignName           string `json:"CampaignName,omitempty"`
	CampaignSource         string `json:"CampaignSource,omitempty"`
	CurrencyCode           string `json:"CurrencyCode,omitempty"`
	DocumentHostName       string `json:"DocumentHostName,omitempty"`
	DocumentLocationURL    string `json:"DocumentLocationURL,omitempty"`
	DocumentPath           string `json:"DocumentPath,omitempty"`
	Referer                string `json:"Referer,omitempty"`
	DocumentTitle          string `json:"DocumentTitle,omitempty"`
	Action                 string `json:"Action,omitempty"`
	Category               string `json:"Category,omitempty"`
	Label                  string `json:"Label,omitempty"`
	Value                  string `json:"Value,omitempty"`
	ExceptionDescription   string `json:"ExceptionDescription,omitempty"`
	IsExceptionFatal       string `json:"IsExceptionFatal,omitempty"`
	GoogleAdWordsID        string `json:"GoogleAdWordsID,omitempty"`
	ItemCode               string `json:"ItemCode,omitempty"`
	ItemName               string `json:"ItemName,omitempty"`
	ItemPrice              string `json:"ItemPrice,omitempty"`
	ItemQuantity           string `json:"ItemQuantity,omitempty"`
	ItemCategory           string `json:"ItemCategory,omitempty"`
	SocialAction           string `json:"SocialAction,omitempty"`
	SocialNetwork          string `json:"SocialNetwork,omitempty"`
	SocialActionTarget     string `json:"SocialActionTarget,omitempty"`
	TransactionAffiliation string `json:"TransactionAffiliation,omitempty"`
	TransactionID          string `json:"TransactionID,omitempty"`
	TransactionShipping    string `json:"TransactionShipping,omitempty"`
	TransactionTax         string `json:"TransactionTax,omitempty"`
	Agent                  string `json:"Agent,omitempty"`
	IP                     string `json:"IP,omitempty"`
	UserLanguage           string `json:"UserLanguage,omitempty"`
	ExperimentID           string `json:"ExperimentID,omitempty"`
	ExperimentVariant      string `json:"ExperimentVariant,omitempty"`
}

func setIfNotEmpty(v *url.Values, key string, value string) {
	if value != "" {
		v.Set(key, value)
	}
}

func setEvent(etype string, event GAEvent) url.Values {
	v := url.Values{}
	v.Set("v", "1")
	v.Set("t", etype)

	setIfNotEmpty(&v, "cc", event.CampaignContent)
	setIfNotEmpty(&v, "cd1", event.CustomDimension1)
	setIfNotEmpty(&v, "cd2", event.CustomDimension2)
	setIfNotEmpty(&v, "cd3", event.CustomDimension3)
	setIfNotEmpty(&v, "cd4", event.CustomDimension4)
	setIfNotEmpty(&v, "cd5", event.CustomDimension5)
	setIfNotEmpty(&v, "cd6", event.CustomDimension6)
	setIfNotEmpty(&v, "cd7", event.CustomDimension7)
	setIfNotEmpty(&v, "cd8", event.CustomDimension8)
	setIfNotEmpty(&v, "cd9", event.CustomDimension9)
	setIfNotEmpty(&v, "cid", event.Guid)
	setIfNotEmpty(&v, "uid", event.UserId)
	setIfNotEmpty(&v, "ck", event.CampaignKeyword)
	setIfNotEmpty(&v, "cm", event.CampaignMedium)
	setIfNotEmpty(&v, "cm1", event.CustomMetric1)
	setIfNotEmpty(&v, "cm2", event.CustomMetric2)
	setIfNotEmpty(&v, "cm3", event.CustomMetric3)
	setIfNotEmpty(&v, "cm4", event.CustomMetric4)
	setIfNotEmpty(&v, "cm5", event.CustomMetric5)
	setIfNotEmpty(&v, "cm6", event.CustomMetric6)
	setIfNotEmpty(&v, "cm7", event.CustomMetric7)
	setIfNotEmpty(&v, "cm8", event.CustomMetric8)
	setIfNotEmpty(&v, "cm9", event.CustomMetric9)
	setIfNotEmpty(&v, "cn", event.CampaignName)
	setIfNotEmpty(&v, "cs", event.CampaignSource)
	setIfNotEmpty(&v, "cu", event.CurrencyCode)
	setIfNotEmpty(&v, "dh", event.DocumentHostName)
	setIfNotEmpty(&v, "dl", event.DocumentLocationURL)
	setIfNotEmpty(&v, "dp", event.DocumentPath)
	setIfNotEmpty(&v, "dr", event.Referer)
	setIfNotEmpty(&v, "dt", event.DocumentTitle)
	setIfNotEmpty(&v, "ea", event.Action)
	setIfNotEmpty(&v, "ec", event.Category)
	setIfNotEmpty(&v, "el", event.Label)
	setIfNotEmpty(&v, "ev", event.Value)
	setIfNotEmpty(&v, "exd", event.ExceptionDescription)
	setIfNotEmpty(&v, "exf", event.IsExceptionFatal)
	setIfNotEmpty(&v, "gclid", event.GoogleAdWordsID)
	setIfNotEmpty(&v, "ic", event.ItemCode)
	setIfNotEmpty(&v, "in", event.ItemName)
	setIfNotEmpty(&v, "ip", event.ItemPrice)
	setIfNotEmpty(&v, "iq", event.ItemQuantity)
	setIfNotEmpty(&v, "iv", event.ItemCategory)
	setIfNotEmpty(&v, "sa", event.SocialAction)
	setIfNotEmpty(&v, "sn", event.SocialNetwork)
	setIfNotEmpty(&v, "st", event.SocialActionTarget)
	setIfNotEmpty(&v, "ta", event.TransactionAffiliation)
	setIfNotEmpty(&v, "ti", event.TransactionID)
	setIfNotEmpty(&v, "ts", event.TransactionShipping)
	setIfNotEmpty(&v, "tt", event.TransactionTax)
	setIfNotEmpty(&v, "ua", event.Agent)
	setIfNotEmpty(&v, "uip", event.IP)
	setIfNotEmpty(&v, "ul", event.UserLanguage)
	setIfNotEmpty(&v, "xid", event.ExperimentID)
	setIfNotEmpty(&v, "xvar", event.ExperimentVariant)

	return v
}

func GetEvent(r *http.Request) GAEvent {
	guid := ""
	cookie, err := r.Cookie("ID")
	if err == nil {
		if cookie != nil {
			guid = cookie.Value
		}
	}
	if guid == "" {
		guid = Encrypt(appengine.NewContext(r), "", MD5(r.RemoteAddr+r.Header.Get("User-Agent")))
	}

	query := ""
	if r.Header.Get("Referer") != "" {
		if refererUrl, err := url.Parse(r.Header.Get("Referer")); err == nil {
			query = refererUrl.Query().Get("q")
		}
	}
	if query == "" {
		query = r.FormValue("k")
	}

	socialNetwork := ""
	referer := strings.ToLower(r.Referer())
	switch {
	case strings.Contains(referer, "facebook"):
		socialNetwork = "Facebook"
	case strings.Contains(referer, "twitter"):
		socialNetwork = "Twitter"
	case strings.Contains(referer, "linkedin"):
		socialNetwork = "LinkedIn"
	}

	event := GAEvent{
		Guid:                guid,
		IP:                  r.RemoteAddr,
		DocumentHostName:    r.Host,
		DocumentLocationURL: r.RequestURI,
		DocumentPath:        r.URL.Path,
		Referer:             r.Referer(),
		DocumentTitle:       r.URL.Path,
		Agent:               r.Header.Get("User-Agent"),
		UserLanguage:        r.Header.Get("Accept-Language"),
		CampaignKeyword:     query,
		CampaignName:        r.FormValue("cm"),
		SocialNetwork:       socialNetwork,
	}

	return event
}

func TrackGAPage(c context.Context, PropertyID string, event GAEvent) {
	endpointUrl := "https://www.google-analytics.com/collect?"
	v := setEvent("pageview", event)
	v.Set("tid", PropertyID)
	payload_data := v.Encode()
	log.Infof(c, "GA: Calling %v with %v", endpointUrl, payload_data)

	req, err := http.NewRequest("POST", endpointUrl, bytes.NewBufferString(payload_data))
	if err != nil {
		log.Errorf(c, "Error while tracking Google Analytics: %v", err)
		return
	}
	resp, err := urlfetch.Client(c).Do(req)
	if err != nil {
		log.Errorf(c, "Error while tracking Google Analytics: %v", err)
		return
	}

	log.Debugf(c, "GA status code %v", resp.StatusCode)
}

func TrackGAEvent(c context.Context, PropertyID string, event GAEvent) {
	endpointUrl := "https://www.google-analytics.com/collect?"
	v := setEvent("pageview", event)
	v.Set("tid", PropertyID)
	payload_data := v.Encode()
	log.Infof(c, "GA: Calling %v with %v", endpointUrl, payload_data)

	req, err := http.NewRequest("POST", endpointUrl, bytes.NewBufferString(payload_data))
	if err != nil {
		log.Errorf(c, "Error while tracking Google Analytics: %v", err)
		return
	}
	resp, err := urlfetch.Client(c).Do(req)
	if err != nil {
		log.Errorf(c, "Error while tracking Google Analytics: %v", err)
		return
	}

	log.Debugf(c, "GA status code %v", resp.StatusCode)
}

func GATrackServeError(w http.ResponseWriter, r *http.Request, PropertyID string,
	errorTitle, errorMessage string, err error, code int, isFatal bool, AppEngineEmail string) {
	c := appengine.NewContext(r)
	if err != nil {
		log.Errorf(c, "%v: %v", errorMessage, err.Error())
	}
	event := GetEvent(r)
	event.ExceptionDescription = errorMessage
	if err != nil {
		event.ExceptionDescription += ": " + err.Error()
	}
	if isFatal {
		event.IsExceptionFatal = "1"
		msg := &mail.Message{
			Sender:  AppEngineEmail,
			To:      []string{AppEngineEmail},
			Subject: "Fatal Error " + errorTitle,
			Body:    fmt.Sprintf(`There was a fatal error on %v at %v: %v`, r.Host, time.Now(), errorMessage),
		}
		if err := mail.Send(c, msg); err != nil {
			log.Errorf(c, "Couldn't send error email: %v", err)
		}
	} else {
		event.IsExceptionFatal = ""
	}
	TrackGAPage(c, PropertyID, event)
	http.Error(w, errorTitle, code)
}
