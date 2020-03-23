package auth

import (
	"bytes"
	"github.com/patdeg/go-appengine/common"
	"encoding/json"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	plus "google.golang.org/api/plus/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/appengine/user"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	//"errors"
)

type User struct {
	GlobalUserId     string    `json:"global_user_id,omitempty"`
	AccessToken      string    `json:"access_token,omitempty"`
	LoginProvider    string    `json:"login_provider,omitempty"`
	GoogleLoginURL   string    `json:"google_login_url,omitempty"`
	FacebookLoginURL string    `json:"facebook_login_url,omitempty"`
	LogoutURL        string    `json:"logout_url,omitempty"`
	UserImage        string    `json:"user_image,omitempty"`
	UserName         string    `json:"user_name,omitempty"`
	UserEmail        string    `json:"user_Email,omitempty"`
	UserId           string    `json:"user_id,omitempty"`
	IsGoogle         bool      `json:"is_google,omitempty"`
	IsFacebook       bool      `json:"is_facebook,omitempty"`
	CreatedTime      time.Time `json:"created_time,omitempty"`
	UpdatedTime      time.Time `json:"updated_time,omitempty"`
	Custom1          string    `json:"custom1,omitempty"`
	Custom2          string    `json:"custom2,omitempty"`
	Custom3          string    `json:"custom3,omitempty"`
	CustomTime1      time.Time `json:"custom_time1,omitempty"`
	CustomTime2      time.Time `json:"custom_time1,omitempty"`
	CustomTime3      time.Time `json:"custom_time1,omitempty"`
	MailingOptIn     bool      `json:"mailing_opt_in,omitempty"`
	CookieID         string    `json:"cookie_id,omitempty" datastore:"-"`
}

type FacebookUserPicture struct {
	URL          string `json:"url,omitempty"`
	IsSilhouette bool   `json:"is_silhouette,omitempty"`
	Height       int    `json:"height,omitempty"`
	Width        int    `json:"width,omitempty"`
}

type FacebookUserPictureData struct {
	Data FacebookUserPicture `json:"data,omitempty"`
}

type FacebookCover struct {
	ID      string `json:"id,omitempty"`
	OffsetX int    `json:"offset_x,omitempty"`
	OffsetY int    `json:"offset_y,omitempty"`
	Source  string `json:"source,omitempty"`
}

type FacebookUser struct {
	Id      string                  `json:"id,omitempty"`
	Name    string                  `json:"name,omitempty"`
	Email   string                  `json:"email,omitempty"`
	Image   string                  `json:"image,omitempty"`
	Picture FacebookUserPictureData `json:"picture,omitempty"`
	Cover   FacebookCover           `json:"cover,omitempty"`

	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Link        string `json:"link,omitempty"`
	Locale      string `json:"locale,omitempty"`
	Timezone    int    `json:"timezone,omitempty"`
	UpdatedTime string `json:"updated_time,omitempty"`
	Verified    bool   `json:"verified,omitempty"`
}

type TokenInfo struct {
	// AccessType: The access type granted with this token. It can be
	// offline or online.
	AccessType string `json:"access_type,omitempty"`
	// Audience: Who is the intended audience for this token. In general the
	// same as issued_to.
	Audience string `json:"audience,omitempty"`

	// Email: The email address of the user. Present only if the email scope
	// is present in the request.
	Email string `json:"email,omitempty"`

	// EmailVerified: Boolean flag which is true if the email address is
	// verified. Present only if the email scope is present in the request.
	EmailVerified bool `json:"email_verified,omitempty"`

	// ExpiresIn: The expiry time of the token, as number of seconds left
	// until expiry.
	ExpiresIn int64 `json:"expires_in,omitempty"`

	// IssuedAt: The issue time of the token, as number of seconds.
	IssuedAt int64 `json:"issued_at,omitempty"`

	// IssuedTo: To whom was the token issued to. In general the same as
	// audience.
	IssuedTo string `json:"issued_to,omitempty"`

	// Issuer: Who issued the token.
	Issuer string `json:"issuer,omitempty"`

	// Nonce: Nonce of the id token.
	Nonce string `json:"nonce,omitempty"`

	// Scope: The space separated list of scopes granted to this token.
	Scope string `json:"scope,omitempty"`

	// UserId: The obfuscated user id.
	UserId string `json:"user_id,omitempty"`

	// VerifiedEmail: Boolean flag which is true if the email address is
	// verified. Present only if the email scope is present in the request.
	VerifiedEmail bool `json:"verified_email,omitempty"`
}

func CheckToken(c context.Context, token string) (*TokenInfo, error) {

	log.Infof(c, ">>>> CheckToken")

	client := urlfetch.Client(c)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v1/tokeninfo?access_token=" + token)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	var tokenInfo TokenInfo
	err = json.Unmarshal(buf.Bytes(), &tokenInfo)
	if err != nil {
		log.Errorf(c, "Error while decoing JSON: %v", err)
		log.Infof(c, "JSON: %v", buf.String())
		return nil, err
	}

	return &tokenInfo, nil
}

func SetCookieToken(c context.Context, w http.ResponseWriter, provider string, host string, cookieValue string, hours int32, isSecure bool) {
	log.Infof(c, ">>>> SetCookieToken")

	key := "token"
	if provider == "Google" {
		key = "g-token"
	} else if provider == "Facebook" {
		key = "fb-token"
	} else if provider == "Deglon" {
		key = "d-token"
	} else {
		log.Errorf(c, "Error, unkown provider '%v'", provider)
	}
	http.SetCookie(w, &http.Cookie{
		Name:    key,
		Value:   url.QueryEscape(cookieValue),
		Secure:  isSecure,
		Path:    "/",
		Domain:  host,
		Expires: time.Now().Add(time.Hour * time.Duration(hours)),
	})
}

func GetCookieToken(r *http.Request) (token string, provider string) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>> GetCookieToken")

	provider = "Google"
	cookie, err := r.Cookie("g-token")
	if err == http.ErrNoCookie {
		provider = "Facebook"
		cookie, err = r.Cookie("fb-token")
		if (err != nil) && (err != http.ErrNoCookie) {
			log.Errorf(c, "[GetCookieToken] Error reading Facebook cookie 'token': %v", err)
			return "", ""
		}
	} else if err != nil {
		log.Errorf(c, "[GetCookieToken] Error reading Google cookie 'token': %v", err)
		return "", ""
	}

	if cookie == nil {
		if ISDEBUG {
			log.Debugf(c, "[GetCookieToken] Cookie is null")
		}
		return "", ""
	}

	if ISDEBUG {
		common.DumpCookie(c, cookie)
	}

	if cookie.Value == "" {
		if ISDEBUG {
			log.Debugf(c, "[GetCookieToken] Cookie is empty")
		}
		return "", ""
	}

	value, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		log.Errorf(c, "[GetCookieToken] Error while unescaping: %v", err)
		return "", ""
	}

	token = common.Decrypt(c, r.RemoteAddr, value)
	if token == "" {
		log.Errorf(c, "getToken: Error decrypting cookie")
		return "", ""
	}

	if provider == "Google" {
		// Check if token is valid
		tokenInfo, err := CheckToken(c, token)
		if err != nil {
			log.Errorf(c, "getToken: Error checking token %v: %v", token, err)
			return "", ""
		}

		if tokenInfo.IssuedTo == "" {
			log.Errorf(c, "getToken: Error, IssueTo is empty")
			return "", ""
		}

	} else if provider == "Facebook" {
		//TODO: add Facebook token check
	}

	return token, provider
}

func GetUserAndCookieID(w http.ResponseWriter, r *http.Request) (*User, string) {
	c := appengine.NewContext(r)
	log.Infof(c, ">>>> GetUserAndCookieID")

	cookieID := common.GetCookieID(w, r)

	var u User

	u.CookieID = cookieID

	if common.DoesCookieExists(r) == false {
		log.Debugf(c, "GetUserAndCookieID: No Cookie ID, returning limited user")
		if user.Current(c) != nil {
			log.Debugf(c, "GetUserAndCookieID: AppEngine User is logged in: %v", user.Current(c).Email)
			u.UserEmail = user.Current(c).Email
		}
		u.GoogleLoginURL = "/goog_login"
		u.FacebookLoginURL = "/fb_login"
		log.Debugf(c, "GetUserAndCookieID: User Id: %v", u.GlobalUserId)
		return &u, cookieID
	}

	// Try to retrieve user from memcache
	err := common.GetObjMemCache(c, "user-"+cookieID, &u)
	if err == nil {
		//log.Errorf(c, "MEMCACHE DISABLED")
		log.Debugf(c, "GetUserAndCookieID: Found user in memcache with key %v: %v", "user-"+cookieID, u)
		if u.UserEmail == "" {
			log.Errorf(c, "GetUserAndCookieID: User found in memcache has no email, force refresh")
		} else {
			log.Debugf(c, "GetUserAndCookieID: User Id: %v", u.GlobalUserId)
			return &u, cookieID
		}
	}

	u.AccessToken, u.LoginProvider = GetCookieToken(r)
	u.GoogleLoginURL = ""
	u.FacebookLoginURL = ""
	u.LogoutURL = ""
	u.UserImage = ""
	u.UserName = ""
	u.UserEmail = ""
	u.UserId = ""
	u.IsGoogle = false
	u.IsFacebook = false
	if u.AccessToken == "" {
		log.Debugf(c, "AccessToken is empty")
		u.GoogleLoginURL = "/goog_login"
		u.FacebookLoginURL = "/fb_login"
	} else {
		u.LogoutURL = "/logout"
		log.Debugf(c, "GetUserAndCookieID: LoginProvider: %v", u.LoginProvider)
		if u.LoginProvider == "Google" {
			log.Debugf(c, "GetUserAndCookieID: Google User")
			me, err := GoogleUserInfo(c, u.AccessToken)
			if err == nil {
				log.Debugf(c, "GetUserAndCookieID: GoogleUserInfo successful")
				u.IsGoogle = true
				if me.Image != nil {
					u.UserImage = me.Image.Url
				}
				u.UserName = me.DisplayName
				// Loop through user's emails and find the first which type is "account" (Google account email address)
				// https://godoc.org/google.golang.org/api/plus/v1#PersonEmails
				for _, e := range me.Emails {
					if e.Type == "account" {
						u.UserEmail = e.Value
						break
					}
				}
				// If userEmail is empty (i.e. no email type "account"), take first email in list
				if u.UserEmail == "" {
					if me.Emails[0] != nil {
						u.UserEmail = me.Emails[0].Value
					}
				}
				u.UserId = common.Encrypt(c, "", me.Id)
				u.GlobalUserId = common.Encrypt(c, "", "G-"+me.Id)
			} else {
				log.Debugf(c, "GetUserAndCookieID: GoogleUserInfo error: %v", err)
			}
		} else if u.LoginProvider == "Facebook" {
			log.Debugf(c, "GetUserAndCookieID: Facebook User")
			me, err := FacebookUserInfo(c, u.AccessToken)
			if err == nil {
				log.Debugf(c, "GetUserAndCookieID: FacebookUserInfo successful")
				u.IsFacebook = true
				u.UserImage = me.Image
				u.UserName = me.Name
				u.UserEmail = me.Email
				u.UserId = common.Encrypt(c, "", me.Id)
				u.GlobalUserId = common.Encrypt(c, "", "FB-"+me.Id)
			} else {
				log.Debugf(c, "GetUserAndCookieID: FacebookUserInfo error: %v", err)
			}
		}
	}

	if user.Current(c) != nil {
		if u.UserEmail == "" {
			u.UserEmail = user.Current(c).Email
		}
	}

	if u.UserEmail != "" {
		// Try to retrieve user from datastore
		var dsUser User
		key := datastore.NewKey(c, "Users", u.UserEmail, 0, nil)
		err = datastore.Get(c, key, &dsUser)
		if err == nil {
			log.Debugf(c, "GetUserAndCookieID: Found user in Datastore")
			u.CreatedTime = dsUser.CreatedTime
			u.UpdatedTime = dsUser.UpdatedTime
			u.MailingOptIn = dsUser.MailingOptIn
			log.Debugf(c, "GetUserAndCookieID: User Id DS: %v", dsUser.GlobalUserId)
			log.Debugf(c, "GetUserAndCookieID: User Email DS: %v", dsUser.UserEmail)
			log.Debugf(c, "GetUserAndCookieID: Create Date DS: %v", dsUser.CreatedTime)
		} else if err == datastore.ErrNoSuchEntity {
			log.Debugf(c, "GetUserAndCookieID: User not found in Datastore")
			u.CreatedTime = time.Now()
			u.UpdatedTime = time.Now()
			u.MailingOptIn = true
		} else {
			log.Errorf(c, "GetUserAndCookieID: Error getting user in Datastore: %v", err)
		}
	} else {
		log.Debugf(c, "GetUserAndCookieID: Email is empty")
		u.CreatedTime = time.Now()
		u.UpdatedTime = time.Now()
		u.MailingOptIn = true
	}

	log.Debugf(c, "GetUserAndCookieID: User Id: %v", u.GlobalUserId)
	log.Debugf(c, "GetUserAndCookieID: User Email: %v", u.UserEmail)
	log.Debugf(c, "GetUserAndCookieID: Create Date: %v", u.CreatedTime)

	if u.UserEmail != "" {
		err = StoreUsers(c, u, cookieID)
		if err != nil {
			log.Errorf(c, "Error storing user: %v", err)
		}
		log.Debugf(c, "GetUserAndCookieID: Get User successful: %v", u)
	}

	return &u, cookieID
}

func GetUser(w http.ResponseWriter, r *http.Request) *User {
	log.Infof(appengine.NewContext(r), ">>>> GetUser")
	u, _ := GetUserAndCookieID(w, r)
	return u
}

func GoogleUserInfo(c context.Context, accessToken string) (*plus.Person, error) {

	log.Infof(c, ">>>> GoogleUserInfo")

	token := &oauth2.Token{
		AccessToken: accessToken,
	}

	client := GoogleConfig.Client(c, token)

	plusService, err := plus.New(client)
	if err != nil {
		log.Errorf(c, "Error getting Google Plus Service: %v", err)
		return nil, err
	}

	log.Infof(c, "Get User (me) info")
	googlePerson, err := plus.NewPeopleService(plusService).Get("me").Do()
	if err != nil {
		log.Errorf(c, "Error getting 'me' through People Service: %v", err)
		return nil, err
	}
	log.Infof(c, "User name: %v:", googlePerson.DisplayName)
	log.Debugf(c, "googlePerson: %v:", *googlePerson)

	return googlePerson, nil
}

func GoogleUserInfoEmail(c context.Context, accessToken string) (*plus.Person, string, error) {
	log.Infof(c, ">>>> GoogleUserInfoEmail")

	token := &oauth2.Token{
		AccessToken: accessToken,
	}

	client := GoogleConfig.Client(c, token)

	plusService, err := plus.New(client)
	if err != nil {
		log.Errorf(c, "Error getting Google Plus Service: %v", err)
		return nil, "", err
	}

	log.Infof(c, "Get User (me) info")
	googlePerson, err := plus.NewPeopleService(plusService).Get("me").Do()
	if err != nil {
		log.Errorf(c, "Error getting 'me' through People Service: %v", err)
		return nil, "", err
	}
	log.Infof(c, "User name: %v:", googlePerson.DisplayName)
	log.Debugf(c, "googlePerson: %v:", *googlePerson)

	email := ""
	// Loop through user's emails and find the first which type is "account" (Google account email address)
	// https://godoc.org/google.golang.org/api/plus/v1#PersonEmails
	for _, e := range googlePerson.Emails {
		if e.Type == "account" {
			email = e.Value
			break
		}
	}
	// If userEmail is empty (i.e. no email type "account"), take first email in list
	if email == "" {
		if googlePerson.Emails[0] != nil {
			email = googlePerson.Emails[0].Value
		}
	}

	return googlePerson, email, nil
}

func FacebookUserInfo(c context.Context, accessToken string) (*FacebookUser, error) {
	log.Infof(c, ">>>> FacebookUserInfo")

	/*
		var token oauth2.Token

		// Try to retrieve user from memcache
		err := common.GetObjMemCache(c, "fb-token-"+accessToken, &token)
		if err == nil {
			log.Errorf("Error recovering Facebook token for Facebook User Info: %v", err)
			return nil, err
		}
	*/

	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	client := FacebookConfig.Client(c, token)

	URL := "https://graph.facebook.com/v2.5/me?access_token=" + accessToken
	URL += "&fields=id,name,email,picture"
	log.Infof(c, "Calling %v", URL)
	resp, err := client.Get(URL)
	if err != nil {
		log.Errorf(c, "FacebookUserInfo: Error at PostForm: %v", err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf(c, "FacebookUserInfo: Error at ReadAll: %v", err)
		return nil, err
	}
	log.Debugf(c, "Facebook /me body: %v", string(body))

	var facebookUser FacebookUser
	err = json.Unmarshal(body, &facebookUser)
	if err != nil {
		log.Errorf(c, "FacebookUserInfo: Error reading Facebook user %v: %v", string(body), err)
		return nil, err
	}

	facebookUser.Image = facebookUser.Picture.Data.URL

	log.Debugf(c, "Facebook User Id: %v", facebookUser.Id)
	log.Debugf(c, "Facebook User Name: %v", facebookUser.Name)
	log.Debugf(c, "Facebook User Email: %v", facebookUser.Email)
	log.Debugf(c, "Facebook User Picture: %v", facebookUser.Image)

	return &facebookUser, nil

}

func GetUserByEmail(c context.Context, email string) (*User, error) {
	var user User
	key := datastore.NewKey(c, "Users", email, 0, nil)
	err := datastore.Get(c, key, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByGlobalId(c context.Context, globalId string) (*User, error) {
	log.Debugf(c, "GetUserByGlobalId for id %v", globalId)
	var user User
	q := datastore.NewQuery("Users").Filter("GlobalUserId =", globalId).Limit(1)
	_, err := common.GetFirst(c, q, &user)
	if err != nil {
		log.Errorf(c, "GetUserByGlobalId: Error %v", err)
		return nil, err
	}
	return &user, nil
}
