package common

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
	"net/http"
	"net/http/httputil"
)

func DumpRequest(r *http.Request, withBody bool) {
	c := appengine.NewContext(r)

	request, err := httputil.DumpRequestOut(r, withBody)
	if err != nil {
		log.Errorf(c, "Error while dumping request: %v", err)
		return
	}
	log.Debugf(c, "Request: %v", B2S(request))
}

func DumpResponse(c context.Context, r *http.Response) {
	resp, err := httputil.DumpResponse(r, true)
	if err != nil {
		log.Errorf(c, "Error while dumping response: %v", err)
		return
	}
	log.Debugf(c, "Response: %v", B2S(resp))
}

func DumpCookie(c context.Context, cookie *http.Cookie) {
	if cookie != nil {
		log.Infof(c, "Cookie:")
		log.Infof(c, "  - Name: %v", cookie.Name)
		log.Infof(c, "  - Value: %v", cookie.Value)
		log.Infof(c, "  - Path: %v", cookie.Path)
		log.Infof(c, "  - Domain: %v", cookie.Domain)
		log.Infof(c, "  - Expires: %v", cookie.Expires)
		log.Infof(c, "  - RawExpires: %v", cookie.RawExpires)
		log.Infof(c, "  - MaxAge: %v", cookie.MaxAge)
		log.Infof(c, "  - Secure:%v", cookie.Secure)
		log.Infof(c, "  - HttpOnly: %v", cookie.HttpOnly)
		log.Infof(c, "  - Raw: %v", cookie.Raw)
	} else {
		log.Debugf(c, "Cookie is null")
	}
}

func DumpCookies(r *http.Request) {
	c := appengine.NewContext(r)

	for _, v := range r.Cookies() {
		log.Debugf(c, "Cookie %v = %v", v.Name, v.Value)
	}

}

func DebugInfo(r *http.Request) {
	c := appengine.NewContext(r)

	log.Debugf(c, "Request %v ", r)
	log.Debugf(c, "URL:%v ", r.URL)
	log.Debugf(c, "Method:%v ", r.Method)
	log.Debugf(c, "Proto:%v ", r.Proto)
	log.Debugf(c, "Header:%v ", r.Header)
	log.Debugf(c, "ContentLength:%v ", r.ContentLength)
	log.Debugf(c, "Host:%v ", r.Host)
	log.Debugf(c, "Form:%v ", r.Form)
	log.Debugf(c, "PostForm:%v ", r.PostForm)
	log.Debugf(c, "MultipartForm:%v ", r.MultipartForm)
	log.Debugf(c, "RemoteAddr:%v ", r.RemoteAddr)
	log.Debugf(c, "RequestURI:%v ", r.RequestURI)

	log.Debugf(c, "AppID:%v ", appengine.AppID(c))
	log.Debugf(c, "Datacenter:%v ", appengine.Datacenter(c))
	log.Debugf(c, "DefaultVersionHostname:%v ", appengine.DefaultVersionHostname(c))
	log.Debugf(c, "InstanceID:%v ", appengine.InstanceID())
	log.Debugf(c, "IsDevAppServer:%v ", appengine.IsDevAppServer())
	log.Debugf(c, "ModuleName:%v ", appengine.ModuleName(c))
	log.Debugf(c, "RequestID:%v ", appengine.RequestID(c))
	log.Debugf(c, "ServerSoftware:%v ", appengine.ServerSoftware())
	log.Debugf(c, "VersionID:%v ", appengine.VersionID(c))
	account, err := appengine.ServiceAccount(c)
	if err != nil {
		log.Errorf(c, "Error getting ServiceAccount: %v", err)
	} else {
		log.Debugf(c, "Service account:%v ", account)
	}

	for k, v := range r.Header {
		log.Debugf(c, "Header(%v):%v ", k, v)
	}

	if user.Current(c) == nil {
		log.Debugf(c, "User:not logged in")

		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			log.Errorf(c, "Error getting login URL %v", err.Error())
		} else {
			log.Debugf(c, "Login URL:%v ", url)
		}
	} else {
		log.Debugf(c, "User:%v ", user.Current(c))
		log.Debugf(c, "Email:%v ", user.Current(c).Email)
		log.Debugf(c, "AuthDomain:%v ", user.Current(c).AuthDomain)
		log.Debugf(c, "Admin:%v ", user.Current(c).Admin)
		log.Debugf(c, "ID:%v ", user.Current(c).ID)
		log.Debugf(c, "FederatedIdentity:%v ", user.Current(c).FederatedIdentity)
		log.Debugf(c, "FederatedProvider:%v ", user.Current(c).FederatedProvider)
	}

}
