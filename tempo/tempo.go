package tempo

import (
	"context"
	"encoding/json"
	"github.com/fatih/color"
	"github.com/skratchdot/open-golang/open"
	"net/http"
	"net/url"
	"time"
)

// TempoAPI defines parameters for connecting to the Tempo API
type TempoAPI struct {
	// Addr optionally specifies the TCP address for the server to listen on,
	// in the form "host:port". If empty "localhost:3000" is used.
	ServerAddr string

	// RedirectUri optionally specifies the redirect uri from the authorization
	// request. If empty "http://ServerAddr" is ued.
	RedirectUri string

	// ClientId is required and specifies the client id of the OAuth 2.0
	// application registered in Tempo
	ClientId string

	// ClientSecret is required and specifies the client secret of the OAuth 2.0
	// application registered in Tempo
	ClientSecret string

	// JiraUrl is required and specified the base url of the Jira instance in
	// the form "https://{jira-cloud-instance-name}.atlassian.net/"
	JiraUrl string
}

// Credentials for interacting with the Tempo API
type Credentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// Worklog obtained or sent to the Tempo API
type Worklog struct {
	IssueKey         string
	TimeSpentSeconds int
	StartDate        time.Time
	StartTime        time.Time
	AuthorAccountId  string

	// Optional
	Description string
}

// Login opens the authorization url in the broswer, starts a temporary web
// server which handles the callback and returns the authorization code
func (t TempoAPI) Login() (credentials Credentials, err error) {
	authorizationCode, err := t.getAuthorizationCode()
	if err != nil {
		return
	}

	response, err := http.PostForm("https://api.tempo.io/oauth/token/", urlValues(map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     t.ClientId,
		"client_secret": t.ClientSecret,
		"redirect_uri":  t.redirectUri(),
		"code":          authorizationCode,
	}))
	if err != nil {
		return
	}
	defer response.Body.Close()
	dec := json.NewDecoder(response.Body)
	err = dec.Decode(&credentials)
	return
}

// Refresh takes a refresh token from Credentials and retrieves a new
// access token for the Tempo API
func (t TempoAPI) Refresh(c Credentials) (credentials Credentials, err error) {
	response, err := http.PostForm("https://api.tempo.io/oauth/token/", urlValues(map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     t.ClientId,
		"client_secret": t.ClientSecret,
		"redirect_uri":  t.redirectUri(),
		"refresh_token": c.RefreshToken,
	}))
	if err != nil {
		return
	}
	defer response.Body.Close()
	dec := json.NewDecoder(response.Body)
	err = dec.Decode(&credentials)
	return
}

func urlValues(values map[string]string) url.Values {
	v := url.Values{}
	for key, value := range values {
		v.Set(key, value)
	}
	return v
}

func (t TempoAPI) getAuthorizationCode() (code string, err error) {
	color.Green("Opening login link in your browser...")
	err = open.Run(t.authUrl())
	if err != nil {
		color.Red("Couldn't open the link, error: %v", err.Error())
		return
	}
	server := http.Server{
		Addr: t.serverAddr(),
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if code = r.FormValue("code"); code != "" {
			server.Shutdown(context.Background())
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		color.Red("Server error: %v", err.Error())
	}
	color.Green("You may now close the broswer tab")
	return
}

func (t TempoAPI) serverAddr() string {
	if t.ServerAddr == "" {
		return "localhost:3000"
	}
	return t.ServerAddr
}

func (t TempoAPI) redirectUri() string {
	if t.RedirectUri == "" {
		return "http://" + t.serverAddr()
	}
	return t.RedirectUri
}

func (t TempoAPI) authUrl() string {
	return t.JiraUrl + "/plugins/servlet/ac/io.tempo.jira/oauth-authorize/?client_id=" + t.ClientId + "&redirect_uri=" + t.redirectUri() + "&access_type=tenant_user"
}
