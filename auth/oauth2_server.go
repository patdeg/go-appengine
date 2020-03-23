package auth

// https://github.com/RangelReale/osin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/RangelReale/osin"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/mail"
	"html/template"
	"net/http"
	"net/url"
)

type Oauth2Server struct {
	server *osin.Server
}

var (
	oauth2Server *Oauth2Server
)

func start(c context.Context) {

	config := &osin.ServerConfig{
		AuthorizationExpiration:   250,
		AccessExpiration:          3600,
		TokenType:                 "Bearer",
		AllowedAuthorizeTypes:     osin.AllowedAuthorizeType{osin.CODE},
		AllowedAccessTypes:        osin.AllowedAccessType{osin.AUTHORIZATION_CODE},
		ErrorStatusCode:           200,
		AllowClientSecretInParams: false,
		AllowGetAccessRequest:     false,
	}

	storage := NewMyStorage()

	oauth2Server = &Oauth2Server{
		server: osin.NewServer(config, storage),
	}

}

func isPasswordGood(c context.Context, username string, password string) bool {

	if (username == "test") && (password == "test") {
		return true
	}
	return false
}

var loginHTML = `<html>
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<link href="/lib/bootstrap-3.3.4/css/bootstrap.min.css" rel="stylesheet">		
		<link href="/lib/bootstrap-3.3.4/css/bootstrap-theme.min.css" rel="stylesheet">
		<style>
			html {   
		   		height:100%;
			}

			body {
				min-height:100%;
				min-height:100vh;
				display:flex;
				align-items:center;
			}
		</style>
	</head>
	<body ng-app="">
		<div class="container">
			<div class="row">
				<div class="col-xs-12 col-sm-offset-1 col-sm-10 col-md-offset-2 col-lg-offset-3 col-md-8 col-lg-6">
					<p style="text-align:center">
						<img src="/img/logo.png" height="40" alt="Deglon Consulting" />						
					</p>
					<h2>Sign in to continue to [[.ClientId]]</h2>		
					<form method="POST" action="https://[[.Server]]/oauth2/auth?response_type=[[.Type]]&client_id=[[.ClientId]]&state=[[.State]]&redirect_uri=[[.Redirect]]">
						<div class="form-group" ng-show="create_account_mode">
							<label for="name">Your Name</label>							
	    					<input type="text" class="form-control" id="name" name="name" placeholder="Enter your name" ng-model="user.name">
	  					</div>
						<div class="form-group">
							<label for="email">Your Username (Email)</label>							
	    					<input type="email" class="form-control" id="email" name="email" autofocus placeholder="Enter your email" ng-model="user.email">
	  					</div>
	  					<div class="form-group">
	  						<label for="password">Your Password with Deglon Consulting</label>							
	    					<input type="password" class="form-control" id="password" name="password" placeholder="Password" ng-model="user.password">
	  					</div>
	  					<div class="form-group" ng-show="create_account_mode">
	  						<label for="password2">Enter your Password again</label>							
	    					<input type="password" class="form-control" id="password2" name="password2" placeholder="Password" ng-model="user.password2">
	  					</div>
	  					<input type="hidden" id="new_account" name="new_account" value="{{create_account_mode}}">

	  					<p style="text-align:center">
	  						<button type="submit" class="btn btn-primary" style="width:200px" ng-disabled="create_account_mode&&((user.password.length<5)||(user.password!=user.password2)||(user.email.length==0)||(user.name.length==0))">
	  							<span ng-hide="create_account_mode">Sign in</span>
	  							<span ng-show="create_account_mode">Create Account</span>
	  						</button>
	  					</p>	 
	  					<p style="text-align:center" ng-show="create_account_mode">
	  						<span ng-hide="user.name.length>0" style="color:red">
		  						Enter a valid name.
		  					</span>
		  					<span ng-hide="user.email.length>0" style="color:red">
		  						Enter a valid email.
		  					</span>	  					
		  					<span ng-hide="user.password.length>=5" style="color:red">
		  						Use at least 5 characters in your password.
		  					</span>
		  					<span ng-hide="user.password==user.password2" style="color:red">
		  						The second password doesn't match the first.
		  					</span> 
		  				</p>	
					</form>
					<p ng-hide="create_account_mode" style="text-align:center">
  						<button class="btn btn-info" style="width:200px" ng-click="create_account_mode=true">Create Account</button>
  					</p>
				</div>
			</div>
		</div>
	</body>
	<script type="text/javascript" src="/lib/angular-1.3.1/angular.min.js"></script>
</html>`

var validateHTML = `<html>
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<link href="/lib/bootstrap-3.3.4/css/bootstrap.min.css" rel="stylesheet">		
		<link href="/lib/bootstrap-3.3.4/css/bootstrap-theme.min.css" rel="stylesheet">
		<style>
			html {   
		   		height:100%;
			}

			body {
				min-height:100%;
				min-height:100vh;
				display:flex;
				align-items:center;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<div class="row">
				<div class="col-xs-12 col-sm-offset-1 col-sm-10 col-md-offset-2 col-lg-offset-3 col-md-8 col-lg-6">
					<p style="text-align:center">
						<img src="/img/logo.png" height="40" alt="Deglon Consulting" />
					</p>
					<h2>Thanks [[.Name]]. We have sent you an email at [[.Email]] to verify your account. Please click on the link in that email to continue.</h2>
					<h4>You can close this window.</h4>					
				</div>
			</div>
		</div>
	</body>	
</html>`

var emailHTML = `<html>
<body>
	<table style="width:440px;border:0px;border-collapse:collapse;">
		<colgroup>
			<col style="width:80px">
			<col style="width:360px">
		</colgroup>
		<tr>
			<td></td>
			<td>
				<p style="
					line-height:18px;
					margin-top:12px;
					margin-bottom:18px;
					text-align:right;
				">
					<img src="http://myapp.appspot.com/img/logo.png" width="100px">		
				</p>

				<p style="color:rgb(20, 0, 100);
					background-color:white;
					font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;
					font-size:25px;
					font-weight:700;
					line-height:24px;
					text-height:24px;
					margin-top:10px;
					margin-bottom:20px;
				">
					[[.Service]]
				</p>

				<p style="color:black;
					font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;
					font-size:12px;
					font-weight:400;
					line-height:18px;
				">
					Hi [[.Name]]<BR><BR>
					You just tried to login to [[.Service]] and requested to create an account. To finalize this process, we need to verify your email [[.Email]]. Please click on the link bellow to finish the process:.<BR>
				</p>

				<p style="color:black;
					font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;
					font-size:12px;
					font-weight:400;
					line-height:18px;
				">
					<a href="[[.URL]]">[[.URL]]</a>
				</p>
				
				<p style="color:black;
					font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;
					font-size:12px;
					font-weight:400;
					line-height:18px;
					margin-top:12px;
					margin-bottom:18px;
				">
					Patrick Deglon
				</p>
				
			</td> 
		</tr>
	</table>
	<P>
</body>
</html>`
var loginTemplate = template.Must(template.New("login.html").Delims("[[", "]]").Parse(loginHTML))
var validateTemplate = template.Must(template.New("validate.html").Delims("[[", "]]").Parse(validateHTML))
var emailTemplate = template.Must(template.New("email.html").Delims("[[", "]]").Parse(emailHTML))

func HandleLoginPage(ar *osin.AuthorizeRequest, w http.ResponseWriter, r *http.Request) bool {

	c := appengine.NewContext(r)

	r.ParseForm()

	for k, v := range r.Form {
		log.Debugf(c, "%v: %v", k, v)
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	name := r.FormValue("name")
	newAccount := r.FormValue("new_account")
	clientId := ar.Client.GetId()

	log.Infof(c, "Email: %v", email)
	log.Infof(c, "Password: %v", password)
	log.Infof(c, "Name: %v", name)
	log.Infof(c, "NewAccount: %v", newAccount)

	if (r.Method == "POST") && (email != "") {

		if newAccount != "" {
			log.Infof(c, "New Account Mode")

			buf := new(bytes.Buffer)
			if err := emailTemplate.Execute(buf, template.FuncMap{
				"Email":   email,
				"Name":    name,
				"Service": clientId,
				"URL":     fmt.Sprintf("https://myapp.appspot.com/oauth2/validate?email=%v&code=%v", email, "12345"),
			}); err != nil {
				log.Errorf(c, "emailTemplate: %v", err)
				http.Error(w, "emailTemplate error:"+err.Error(), http.StatusInternalServerError)
				return true
			}

			msg := &mail.Message{
				Sender: "myemail@gmail.com",
				To:     []string{email},
				Subject:  "Please verify your email for " + clientId,
				HTMLBody: buf.String(),
			}
			if err := mail.Send(c, msg); err != nil {
				log.Errorf(c, "Couldn't send email: %v", err)
				http.Error(w, "Couldn't send email:"+err.Error(), http.StatusInternalServerError)
				return true
			}
			log.Infof(c, "Email sent to %v", email)

			if err := validateTemplate.Execute(w, template.FuncMap{
				"Name":  name,
				"Email": email,
			}); err != nil {
				log.Infof(c, "Error with validateTemplate: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return true
			}
			return false
		}

		if isPasswordGood(c, email, password) {
			log.Infof(c, "Login Correct")
			return true
		}

	}

	if err := loginTemplate.Execute(w, template.FuncMap{
		"ClientId": ar.Client.GetId(),
		"Type":     ar.Type,
		"State":    ar.State,
		"Redirect": url.QueryEscape(ar.RedirectUri),
		"Server":   "myapp.appspot.com",
	}); err != nil {
		log.Infof(c, "Error with loginTemplate: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return true
	}

	return false
}

func DownloadAccessToken(url string, auth *osin.BasicAuth, output map[string]interface{}) error {
	// download access token
	preq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if auth != nil {
		preq.SetBasicAuth(auth.Username, auth.Password)
	}

	pclient := &http.Client{}
	presp, err := pclient.Do(preq)
	if err != nil {
		return err
	}

	if presp.StatusCode != 200 {
		return errors.New("Invalid status code")
	}

	jdec := json.NewDecoder(presp.Body)
	err = jdec.Decode(&output)
	return err
}

// Authorization code endpoint, for example /authorize
func OAuth2AuthorizeHandler(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	log.Debugf(c, ">>> OAuth2AuthorizeHandler")

	if oauth2Server == nil {
		log.Debugf(c, "Starting server...")
		start(c)
	}

	resp := oauth2Server.server.NewResponse()
	defer resp.Close()

	if ar := oauth2Server.server.HandleAuthorizeRequest(resp, r); ar != nil {
		log.Debugf(c, "Finished HandleAuthorizeRequest...")

		if !HandleLoginPage(ar, w, r) {
			log.Debugf(c, "HandleLoginPage return false, exiting")
			return
		}
		log.Debugf(c, "HandleLoginPage return true")

		ar.Authorized = true
		oauth2Server.server.FinishAuthorizeRequest(resp, r, ar)
		log.Debugf(c, "Finished FinishAuthorizeRequest")
	}
	log.Debugf(c, "Trying OutputJSON")
	osin.OutputJSON(resp, w, r)
}

// Access token endpoint, for example /token
func OAuth2TokenHandler(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	if oauth2Server == nil {
		start(c)
	}

	resp := oauth2Server.server.NewResponse()
	defer resp.Close()

	if ar := oauth2Server.server.HandleAccessRequest(resp, r); ar != nil {
		ar.Authorized = true
		oauth2Server.server.FinishAccessRequest(resp, r, ar)
	}
	osin.OutputJSON(resp, w, r)
}
