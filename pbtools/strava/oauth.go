package strava

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"

	strava "github.com/strava/go.strava"
)

const basePath = "https://www.strava.com/api/v3"
const tokenFilePath = "/tmp/pb-tools-token.json"

// AuthToken represents an authorization token
type AuthToken struct {
	Token     string
	ExpiresAt time.Time
}

type state struct {
	done      chan bool
	tokenResp authorizationResponse
	err       error
}

var s = state{}

type authorizationResponse struct {
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

// GetAccessToken returns Strava access token to query APIs
func GetAccessToken(httpPort int) (*AuthToken, error) {
	tok, err := tokenFromFile()
	if err != nil {
		tok, err = getTokenFromWeb(httpPort)
		if err != nil {
			return nil, err
		}

		saveToken(tok)
	}

	return &AuthToken{
		Token:     tok.AccessToken,
		ExpiresAt: time.Unix(tok.ExpiresAt, 0),
	}, err
}

// RefreshToken refresh access token
func RefreshToken() (*AuthToken, error) {
	tok, err := tokenFromFile()
	if err != nil {
		return nil, err
	}

	resp, err := execOauthTokenRequest(url.Values{"client_id": {fmt.Sprintf("%d", strava.ClientId)}, "client_secret": {strava.ClientSecret}, "grant_type": {"refresh_token"}, "refresh_token": {tok.RefreshToken}})
	if err != nil {
		return nil, err
	}

	saveToken(*resp)

	return &AuthToken{
		Token:     resp.AccessToken,
		ExpiresAt: time.Unix(resp.ExpiresAt, 0),
	}, nil
}

// Retrieves a token online.
func getTokenFromWeb(httpPort int) (authorizationResponse, error) {
	s.done = make(chan bool)

	callbackURL, _ := url.Parse(fmt.Sprintf("http://localhost:%d/exchange_token", httpPort))

	url := fmt.Sprintf("%s/oauth/authorize?client_id=%d&response_type=code&redirect_uri=%s&scope=activity:read&approval_prompt=force", basePath, strava.ClientId, callbackURL.String())
	openbrowser(url)

	http.HandleFunc(callbackURL.Path, handlerFunc())

	// start the server
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil)
		if err != nil {
			panic("ListenAndServe: " + err.Error())
		}
	}()

	<-s.done

	return s.tokenResp, s.err
}

// Authorize performs the second part of the OAuth exchange. The client has already been redirected to the
// Strava authorization page, has granted authorization to the application and has been redirected back to the
// defined URL. The code param was returned as a query string param in to the redirect_url.
func authorize(code string, client *http.Client) (*authorizationResponse, error) {
	// make sure a code was passed
	if code == "" {
		return nil, strava.OAuthInvalidCodeErr
	}

	resp, err := execOauthTokenRequest(url.Values{"client_id": {fmt.Sprintf("%d", strava.ClientId)}, "client_secret": {strava.ClientSecret}, "code": {code}})
	return resp, err
}

func execOauthTokenRequest(data url.Values) (*authorizationResponse, error) {
	resp, err := http.DefaultClient.PostForm(basePath+"/oauth/token", data)

	// this was a poor request, maybe strava servers down?
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// check status code, could be 500, or most likely the client_secret is incorrect
	if resp.StatusCode/100 == 5 {
		return nil, strava.OAuthServerErr
	}

	if resp.StatusCode/100 != 2 {
		var response strava.Error
		contents, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(contents, &response)

		if len(response.Errors) == 0 {
			return nil, strava.OAuthServerErr
		}

		if response.Errors[0].Resource == "Application" {
			return nil, strava.OAuthInvalidCredentialsErr
		}

		if response.Errors[0].Resource == "RequestToken" {
			return nil, strava.OAuthInvalidCodeErr
		}

		return nil, &response
	}

	var response authorizationResponse
	contents, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(contents, &response)

	if err != nil {
		return nil, err
	}

	return &response, nil
}

// handlerFunc builds a http.HandlerFunc that will complete the token exchange
// after a user authorizes an application on strava.com.
// This method handles the exchange and calls success or failure after it completes.
func handlerFunc() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// user denied authorization
		if r.FormValue("error") == "access_denied" {
			oAuthFailure(strava.OAuthAuthorizationDeniedErr, w, r)
			return
		}

		code := r.FormValue("code")
		if code == "" {
			oAuthFailure(strava.OAuthInvalidCodeErr, w, r)
		}

		resp, err := execOauthTokenRequest(url.Values{"client_id": {fmt.Sprintf("%d", strava.ClientId)}, "client_secret": {strava.ClientSecret}, "code": {code}})

		if err != nil {
			oAuthFailure(err, w, r)
			return
		}

		oAuthSuccess(resp, w, r)
	}
}

// Retrieves a token from a local file.
func tokenFromFile() (authorizationResponse, error) {
	b, err := ioutil.ReadFile(tokenFilePath)
	var token authorizationResponse
	if err != nil {
		return token, err
	}
	err = json.Unmarshal(b, &token)
	return token, err
}

// Saves a token to a file path.
func saveToken(token authorizationResponse) {
	data, _ := json.MarshalIndent(token, "", " ")
	err := ioutil.WriteFile(tokenFilePath, data, 0644)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
}

func oAuthSuccess(resp *authorizationResponse, w http.ResponseWriter, r *http.Request) {
	s.tokenResp = *resp
	s.err = nil
	s.done <- true
}

func oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Authorization Failure:\n")

	// some standard error checking
	if err == strava.OAuthAuthorizationDeniedErr {
		fmt.Fprint(w, "The user clicked the 'Do not Authorize' button on the previous page.\n")
		fmt.Fprint(w, "This is the main error your application should handle.")
	} else if err == strava.OAuthInvalidCredentialsErr {
		fmt.Fprint(w, "You provided an incorrect client_id or client_secret.\nDid you remember to set them at the begininng of this file?")
	} else if err == strava.OAuthInvalidCodeErr {
		fmt.Fprint(w, "The temporary token was not recognized, this shouldn't happen normally")
	} else if err == strava.OAuthServerErr {
		fmt.Fprint(w, "There was some sort of server error, try again to see if the problem continues")
	} else {
		fmt.Fprint(w, err)
	}

	s.tokenResp = authorizationResponse{}
	s.err = err
	s.done <- true
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
