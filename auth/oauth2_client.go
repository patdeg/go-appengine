package auth

import (
	"github.com/patdeg/go-appengine/common"
	"golang.org/x/oauth2"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"net/http"
)

var DeglonConfig = &oauth2.Config{
	ClientID:     "my_client",
	ClientSecret: "my_secret",
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	},
	RedirectURL: "https://myapp.appspot.com/oauth2/callback",
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://myapp.appspot.com/oauth2/auth",
		TokenURL: "https://myapp.appspot.com/oauth2/token",
	},
}

func OAuth2LoginTestHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>> OAuth2LoginTestHandler")

	state := r.FormValue("redirect")
	if state == "" {
		state = "/"
	}

	url := DeglonConfig.AuthCodeURL(state)
	log.Infof(c, "Redirect to %v", url)

	http.Redirect(w, r, url, http.StatusFound)
}

func OAuth2CallbackTestHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>> OAuth2CallbackTestHandler")

	code := r.FormValue("code")
	state := r.FormValue("state")
	errorMessage := r.FormValue("error")

	log.Infof(c, "code: %v", code)
	log.Infof(c, "state: %v", state)
	if errorMessage != "" {
		log.Errorf(c, "Error while authentification: %v", errorMessage)
		url := "http://" + r.Host
		log.Infof(c, "Redirect to %v", url)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	tok, err := DeglonConfig.Exchange(c, code)
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

	SetCookieToken(c, w, "Deglon", r.Host, cookieValue, 1, false)

	err = common.SetObjMemCache(c, "d-token-"+tok.AccessToken, &tok, 24)
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

			u.LoginProvider = "Deglon"

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
