package auth

import (
	"errors"
	"github.com/patdeg/go-appengine/common"
	"github.com/patdeg/go-appengine/track"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"strings"
	"time"
)

var GoogleConfig = &oauth2.Config{
	ClientID:     "myclientid",
	ClientSecret: "myclientsecret",
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	},
	Endpoint: google.Endpoint,
}

var FacebookConfig = &oauth2.Config{
	ClientID:     "myclientid",
	ClientSecret: "myclientsecret",
	Scopes: []string{
		"email",
		"user_about_me",
		"user_photos",
	},
	Endpoint: facebook.Endpoint,
}

var (
	ISDEBUG = true
	VERSION = ""
)

var (
	ERROR_NO_EMAIL = errors.New("No Email")
)

var AdminUsers = []string{
	"myemail@gmail.com",
}

var WhiteListUsers = []string{
	"myemail@gmail.com",
}

func IsUserAdmin(c context.Context, email string) bool {
	return common.StringInSlice(strings.ToLower(email), AdminUsers)
}

func IsUserWhiteListed(c context.Context, email string) bool {
	return common.StringInSlice(strings.ToLower(email), WhiteListUsers)
}

func GetServiceAccountClient(c context.Context) *http.Client {
	return &http.Client{
		Transport: &oauth2.Transport{
			Source: google.AppEngineTokenSource(c,
				"https://www.googleapis.com/auth/userinfo.email"),
			Base: &urlfetch.Transport{
				Context: c,
			},
		},
	}
}

func RedirectIfNotLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>> RedirectIfNotLoggedIn")
	cookie, provider := GetCookieToken(r)
	if cookie == "" {
		if provider == "Google" {
			state := r.URL.Path
			GoogleConfig.RedirectURL = "http://" + r.Host + "/goog_callback"
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Methods", "PUT")
			w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
			url := GoogleConfig.AuthCodeURL(state)
			log.Infof(c, "Redirect to %v", url)
			http.Redirect(w, r, url, http.StatusFound)
			return true
		} else if provider == "Facebook" {
			state := r.URL.Path
			FacebookConfig.RedirectURL = "http://" + r.Host + "/fb_callback"
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Methods", "PUT")
			w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
			url := FacebookConfig.AuthCodeURL(state)
			log.Infof(c, "Redirect to %v", url)
			http.Redirect(w, r, url, http.StatusFound)
			return true
		} else {
			log.Errorf(c, "Error, unkown provider '%v'", provider)
			http.Error(w, "Internal Server Error: Wrong Provider", http.StatusInternalServerError)
			return true
		}
	}
	return false
}

func RedirectIfNotLoggedInAPI(w http.ResponseWriter, r *http.Request) bool {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>> RedirectIfNotLoggedInAPI")
	cookie, _ := GetCookieToken(r)
	if cookie == "" {
		log.Infof(c, "Cookie is null, requesting new login")
		http.Error(w, "Login required", http.StatusUnauthorized)
		return true
	}
	return false
}

func IsLoggedIn(r *http.Request) bool {
	cookie, _ := GetCookieToken(r)
	if cookie == "" {
		return false
	}
	return true
}

func ClearHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> Clear Handler")

	log.Infof(c, "New request Path:%v RemoteAddr:%v Method:%v Host:%v", r.URL.Path, r.RemoteAddr, r.Method, r.Host)
	log.Infof(c, "RequstURI:%v", r.RequestURI)

	http.SetCookie(w,
		&http.Cookie{
			Name:    "token",
			Value:   "",
			Secure:  true,
			Path:    "/",
			Expires: time.Now(),
			//MaxAge: 3600,
		})
	time.Sleep(time.Second * 5)

	url := "http://" + r.Host
	log.Infof(c, "Redirect to %v", url)
	http.Redirect(w, r, url, http.StatusFound)
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> GoogLoginHandler")

	state := r.FormValue("redirect")
	if state == "" {
		state = "/"
	}
	GoogleConfig.RedirectURL = "http://" + r.Host + "/goog_callback"
	url := GoogleConfig.AuthCodeURL(state)

	track.TrackEventDetails(w, r, common.GetCookieID(w, r), "Google Login", state, "", 0.0)

	log.Infof(c, "Redirect to %v", url)
	http.Redirect(w, r, url, http.StatusFound)
}

func GoogleLoginOfflineAccessHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> GoogLoginHandler")

	state := r.FormValue("redirect")
	if state == "" {
		state = "/"
	}
	GoogleConfig.RedirectURL = "http://" + r.Host + "/goog_callback"
	url := GoogleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	log.Infof(c, "Redirect to %v", url)
	http.Redirect(w, r, url, http.StatusFound)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> LogoutHandler")

	http.SetCookie(w, &http.Cookie{
		Name:  "fb-token",
		Value: "",
		//Secure:  true,
		Path:    "/",
		Domain:  r.Host,
		Expires: time.Now(),
	})
	time.Sleep(time.Millisecond * 100)

	http.SetCookie(w, &http.Cookie{
		Name:  "g-token",
		Value: "",
		//Secure:  true,
		Path:    "/",
		Domain:  r.Host,
		Expires: time.Now(),
	})
	time.Sleep(time.Millisecond * 100)

	cookieID := common.GetCookieID(w, r)
	if cookieID != "" {
		common.DeleteMemCache(c, "user-"+cookieID)
	}

	url := "http://" + r.Host
	log.Infof(c, "Redirect to %v", url)
	http.Redirect(w, r, url, http.StatusFound)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> Google Callback Handler")

	code := r.FormValue("code")
	state := r.FormValue("state")
	errorMessage := r.FormValue("error")

	track.TrackEventDetails(w, r, common.GetCookieID(w, r), "Google Callback", state, errorMessage, 0.0)

	log.Infof(c, "code: %v", code)
	log.Infof(c, "state: %v", state)
	if errorMessage != "" {
		log.Errorf(c, "Error while authentification: %v", errorMessage)
		url := "http://" + r.Host
		log.Infof(c, "Redirect to %v", url)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	GoogleConfig.RedirectURL = "http://" + r.Host + "/goog_callback"

	tok, err := GoogleConfig.Exchange(c, code)
	if err != nil {
		log.Errorf(c, "Error exchanging code for token: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := tok.AccessToken
	log.Infof(c, "Token Exchanged:")
	log.Infof(c, "  - AccessToken: %v", tok.AccessToken)
	log.Infof(c, "  - RefreshToken: %v", tok.RefreshToken)
	log.Infof(c, "  - Expiry: %v", tok.Expiry)

	cookieValue := common.Encrypt(c, r.RemoteAddr, token)
	log.Debugf(c, "cookie value = %v", cookieValue)

	SetCookieToken(c, w, "Google", r.Host, cookieValue, 1, false)

	err = common.SetObjMemCache(c, "g-token-"+tok.AccessToken, &tok, 24)
	if err != nil {
		log.Errorf(c, "Error setting token in memcache: %v", err)
	}

	// Switch LoginProvider to Facebook in user memcache
	cookieID := common.GetCookieID(w, r)
	if cookieID != "" {
		var u User
		err = common.GetObjMemCache(c, "user-"+cookieID, &u)
		if (err == nil) && (u.UserEmail != "") {
			// Found User in Memcache
			log.Debugf(c, "Found User in Memcache: %v", u)

			u.LoginProvider = "Google"

			if u.UserEmail != "" {
				log.Debugf(c, "Setting User in Memcache with key %v: %v", "user-"+cookieID, u)
				err = common.SetObjMemCache(c, "user-"+cookieID, &u, 24)
				if err != nil {
					log.Errorf(c, "Error setting user in memcache: %v", err)
				}
			}

		}
	}

	url := "http://" + r.Host + state

	log.Infof(c, "Redirect to %v", url)
	http.Redirect(w, r, url, http.StatusFound)

}

func GoogleCallbackDatastoreHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> Google Callback Datastore Handler")

	code := r.FormValue("code")
	state := r.FormValue("state")
	errorMessage := r.FormValue("error")

	track.TrackEventDetails(w, r, common.GetCookieID(w, r), "Google Callback", state, errorMessage, 0.0)

	log.Infof(c, "code: %v", code)
	log.Infof(c, "state: %v", state)
	if errorMessage != "" {
		log.Errorf(c, "Error while authentification: %v", errorMessage)
		url := "http://" + r.Host
		log.Infof(c, "Redirect to %v", url)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	GoogleConfig.RedirectURL = "http://" + r.Host + "/goog_callback"

	tok, err := GoogleConfig.Exchange(c, code)
	if err != nil {
		log.Errorf(c, "Error exchanging code for token: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := tok.AccessToken
	log.Infof(c, "Token Exchanged:")
	log.Infof(c, "  - AccessToken: %v", tok.AccessToken)
	log.Infof(c, "  - RefreshToken: %v", tok.RefreshToken)
	log.Infof(c, "  - Expiry: %v", tok.Expiry)

	cookieValue := common.Encrypt(c, r.RemoteAddr, token)
	log.Debugf(c, "cookie value = %v", cookieValue)

	SetCookieToken(c, w, "Google", r.Host, cookieValue, 1, false)

	err = common.SetObjMemCache(c, "g-token-"+tok.AccessToken, &tok, 24)
	if err != nil {
		log.Errorf(c, "Error setting token in memcache: %v", err)
	}

	// Switch LoginProvider to Facebook in user memcache
	cookieID := common.GetCookieID(w, r)
	if cookieID != "" {
		var u User
		err = common.GetObjMemCache(c, "user-"+cookieID, &u)
		if (err == nil) && (u.UserEmail != "") {
			// Found User in Memcache
			log.Debugf(c, "Found User in Memcache: %v", u)

			u.LoginProvider = "Google"

			if u.UserEmail != "" {
				log.Debugf(c, "Setting User in Memcache with key %v: %v", "user-"+cookieID, u)
				err = common.SetObjMemCache(c, "user-"+cookieID, &u, 24)
				if err != nil {
					log.Errorf(c, "Error setting user in memcache: %v", err)
				}
				log.Debugf(c, "GoogleCallbackDatastoreHandler: User Email: %v", u.UserEmail)
				log.Debugf(c, "GoogleCallbackDatastoreHandler: Create Date: %v", u.CreatedTime)
				err = StoreUsers(c, u, cookieID)
				if err != nil {
					log.Errorf(c, "Error storing user: %v", err)
				}
			}

		}
	}

	url := "http://" + r.Host + state

	log.Infof(c, "Redirect to %v", url)
	http.Redirect(w, r, url, http.StatusFound)

}

func FacebookLoginHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> FacebookLoginHandler")

	state := r.FormValue("redirect")
	if state == "" {
		state = "/"
	}
	FacebookConfig.RedirectURL = "http://" + r.Host + "/fb_callback"
	url := FacebookConfig.AuthCodeURL(state)

	track.TrackEventDetails(w, r, common.GetCookieID(w, r), "Facebook Login", state, "", 0.0)

	log.Infof(c, "Redirect to %v", url)
	http.Redirect(w, r, url, http.StatusFound)
}

func FacebookCallbackHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> FacebookCallbackHandler")

	code := r.FormValue("code")
	state := r.FormValue("state")
	errorMessage := r.FormValue("error")

	track.TrackEventDetails(w, r, common.GetCookieID(w, r), "Facebook Callback", state, errorMessage, 0.0)

	log.Infof(c, "code: %v", code)
	log.Infof(c, "state: %v", state)
	if errorMessage != "" {
		log.Errorf(c, "Error while authentification: %v", errorMessage)
		url := "http://" + r.Host
		log.Infof(c, "Redirect to %v", url)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	FacebookConfig.RedirectURL = "http://" + r.Host + "/fb_callback"

	tok, err := FacebookConfig.Exchange(c, code)
	if err != nil {
		log.Errorf(c, "Error exchanging code for token: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := tok.AccessToken
	log.Infof(c, "Token Exchanged:")
	log.Infof(c, "  - AccessToken: %v", tok.AccessToken)
	log.Infof(c, "  - RefreshToken: %v", tok.RefreshToken)
	log.Infof(c, "  - Expiry: %v", tok.Expiry)

	cookieValue := common.Encrypt(c, r.RemoteAddr, token)
	log.Debugf(c, "cookie value = %v", cookieValue)

	SetCookieToken(c, w, "Facebook", r.Host, cookieValue, 1, false)

	err = common.SetObjMemCache(c, "fb-token-"+tok.AccessToken, &tok, 24)
	if err != nil {
		log.Errorf(c, "Error setting token in memcache: %v", err)
	}

	// Switch LoginProvider to Facebook in user memcache
	cookieID := common.GetCookieID(w, r)
	if cookieID != "" {
		var u User
		err = common.GetObjMemCache(c, "user-"+cookieID, &u)
		if (err == nil) && (u.UserEmail != "") {
			// Found User in Memcache

			u.LoginProvider = "Facebook"

			if u.UserEmail != "" {

				log.Debugf(c, "Setting User in Memcache with key %v: %v", "user-"+cookieID, u)
				err = common.SetObjMemCache(c, "user-"+cookieID, &u, 24)
				if err != nil {
					log.Errorf(c, "Error setting user in memcache: %v", err)
				}
			}

		}
	}

	url := "http://" + r.Host + state

	log.Infof(c, "Redirect to %v", url)
	http.Redirect(w, r, url, http.StatusFound)

}

func FacebookCallbackDatastoreHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>>>>>> FacebookCallbackDatastoreHandler")

	code := r.FormValue("code")
	state := r.FormValue("state")
	errorMessage := r.FormValue("error")

	track.TrackEventDetails(w, r, common.GetCookieID(w, r), "Facebook Callback", state, errorMessage, 0.0)

	log.Infof(c, "code: %v", code)
	log.Infof(c, "state: %v", state)
	if errorMessage != "" {
		log.Errorf(c, "Error while authentification: %v", errorMessage)
		url := "http://" + r.Host
		log.Infof(c, "Redirect to %v", url)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	FacebookConfig.RedirectURL = "http://" + r.Host + "/fb_callback"

	tok, err := FacebookConfig.Exchange(c, code)
	if err != nil {
		log.Errorf(c, "Error exchanging code for token: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := tok.AccessToken
	log.Infof(c, "Token Exchanged:")
	log.Infof(c, "  - AccessToken: %v", tok.AccessToken)
	log.Infof(c, "  - RefreshToken: %v", tok.RefreshToken)
	log.Infof(c, "  - Expiry: %v", tok.Expiry)

	cookieValue := common.Encrypt(c, r.RemoteAddr, token)
	log.Debugf(c, "cookie value = %v", cookieValue)

	SetCookieToken(c, w, "Facebook", r.Host, cookieValue, 1, false)

	err = common.SetObjMemCache(c, "fb-token-"+tok.AccessToken, &tok, 24)
	if err != nil {
		log.Errorf(c, "Error setting token in memcache: %v", err)
	}

	// Switch LoginProvider to Facebook in user memcache
	cookieID := common.GetCookieID(w, r)
	if cookieID != "" {
		var u User
		err = common.GetObjMemCache(c, "user-"+cookieID, &u)
		if (err == nil) && (u.UserEmail != "") {
			// Found User in Memcache

			u.LoginProvider = "Facebook"

			if u.UserEmail != "" {
				log.Debugf(c, "Setting User in Memcache with key %v: %v", "user-"+cookieID, u)
				err = common.SetObjMemCache(c, "user-"+cookieID, &u, 24)
				if err != nil {
					log.Errorf(c, "Error setting user in memcache: %v", err)
				}
				log.Debugf(c, "FacebookCallbackDatastoreHandler: User Email: %v", u.UserEmail)
				log.Debugf(c, "FacebookCallbackDatastoreHandler: Create Date: %v", u.CreatedTime)
				err = StoreUsers(c, u, cookieID)
				if err != nil {
					log.Errorf(c, "Error storing user: %v", err)
				}
			}

		}
	}

	url := "http://" + r.Host + state

	log.Infof(c, "Redirect to %v", url)
	http.Redirect(w, r, url, http.StatusFound)

}

func StoreUsers(c context.Context, u User, cookieID string) error {

	log.Infof(c, ">>>> StoreUsers")

	var existingUser User

	// Check if email is valid
	log.Debugf(c, "StoreUsers: User Email: %v ", u.UserEmail)
	if u.UserEmail == "" {
		log.Errorf(c, "StoreUsers: Error, email is empty")
		return ERROR_NO_EMAIL
	}

	// Try to retrieve user from memcache
	err := common.GetObjMemCache(c, "user-"+cookieID, &existingUser)
	if err == nil {
		log.Debugf(c, "StoreUsers: Found user in memcache with key %v: %v", "user-"+cookieID, u)
		if (u.GlobalUserId == existingUser.GlobalUserId) &&
			(u.AccessToken == existingUser.AccessToken) &&
			(u.LoginProvider == existingUser.LoginProvider) &&
			(u.GoogleLoginURL == existingUser.GoogleLoginURL) &&
			(u.FacebookLoginURL == existingUser.FacebookLoginURL) &&
			(u.LogoutURL == existingUser.LogoutURL) &&
			(u.UserImage == existingUser.UserImage) &&
			(u.UserName == existingUser.UserName) &&
			(u.UserEmail == existingUser.UserEmail) &&
			(u.UserId == existingUser.UserId) &&
			(u.IsGoogle == existingUser.IsGoogle) &&
			(u.IsFacebook == existingUser.IsFacebook) &&
			(u.Custom1 == existingUser.Custom1) &&
			(u.Custom2 == existingUser.Custom2) &&
			(u.Custom3 == existingUser.Custom3) &&
			(u.CustomTime1 == existingUser.CustomTime1) &&
			(u.CustomTime2 == existingUser.CustomTime2) &&
			(u.CustomTime3 == existingUser.CustomTime3) {
			log.Debugf(c, "StoreUsers: Same user, ignore")
			return nil
		}
	} else if err != memcache.ErrCacheMiss {
		log.Errorf(c, "StoreUsers: Error reading memcache: %v", err)
	}

	// Try to retrieve user from datastore
	key := datastore.NewKey(c, "Users", u.UserEmail, 0, nil)
	err = datastore.Get(c, key, &existingUser)
	if err == datastore.ErrNoSuchEntity {
		log.Debugf(c, "StoreUsers: New User %v", u.UserEmail)
		u.CreatedTime = time.Now()
		u.UpdatedTime = time.Now()
		u.MailingOptIn = true
		// Create in datastore and memcache next...
	} else if err != nil {
		log.Errorf(c, "StoreUsers: Error getting user with key %v", key)
		return err
	} else {
		log.Debugf(c, "StoreUsers: Existing User %v", u.UserEmail)
		if (u.GlobalUserId == existingUser.GlobalUserId) &&
			(u.AccessToken == existingUser.AccessToken) &&
			(u.LoginProvider == existingUser.LoginProvider) &&
			(u.GoogleLoginURL == existingUser.GoogleLoginURL) &&
			(u.FacebookLoginURL == existingUser.FacebookLoginURL) &&
			(u.LogoutURL == existingUser.LogoutURL) &&
			(u.UserImage == existingUser.UserImage) &&
			(u.UserName == existingUser.UserName) &&
			(u.UserEmail == existingUser.UserEmail) &&
			(u.UserId == existingUser.UserId) &&
			(u.IsGoogle == existingUser.IsGoogle) &&
			(u.IsFacebook == existingUser.IsFacebook) &&
			(u.Custom1 == existingUser.Custom1) &&
			(u.Custom2 == existingUser.Custom2) &&
			(u.Custom3 == existingUser.Custom3) &&
			(u.CustomTime1 == existingUser.CustomTime1) &&
			(u.CustomTime2 == existingUser.CustomTime2) &&
			(u.CustomTime3 == existingUser.CustomTime3) {
			log.Debugf(c, "StoreUsers: Same user, ignore")
			return nil
		}
		log.Debugf(c, "StoreUsers: Something changed, updating datastore")
		u.CreatedTime = existingUser.CreatedTime
		u.MailingOptIn = existingUser.MailingOptIn
		u.UpdatedTime = time.Now()
		// Update datastore and memcache next...
	}

	log.Debugf(c, "Set User in Datastore with key %v: %v", key, u)
	_, err = datastore.Put(c, key, &u)
	if err != nil {
		log.Errorf(c, "StoreUsers: Error storing with key %v", key)
		return err
	}

	log.Debugf(c, "StoreUsers: User Email: %v", u.UserEmail)
	log.Debugf(c, "StoreUsers: Create Date: %v", u.CreatedTime)

	// Set user in memcache
	log.Debugf(c, "Set User in Memcache with key %v: %v", "user-"+cookieID, u)
	err = common.SetObjMemCache(c, "user-"+cookieID, &u, 24)
	if err != nil {
		log.Errorf(c, "Error setting user in memcache: %v", err)
		return err
	}

	return nil

}
