package strava

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	strava "github.com/strava/go.strava"
)

const basePath = "https://www.strava.com/api/v3"

var done = make(chan bool)
var token string
var errr error

// AuthToken represents an authorization token
type AuthToken struct {
	Token string
}

// GetAccessToken returns Strava access token to query APIs
func GetAccessToken(httpPort int) (*AuthToken, error) {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err = getTokenFromWeb(httpPort)
		if err != nil {
			return nil, err
		}

		saveToken(tokFile, tok)
	}

	return &AuthToken{Token: tok}, err
}

// Retrieves a token online.
func getTokenFromWeb(httpPort int) (string, error) {
	authenticator := &strava.OAuthAuthenticator{
		CallbackURL:            fmt.Sprintf("http://localhost:%d/exchange_token", httpPort),
		RequestClientGenerator: nil,
	}
	openbrowser(authorizationURL(authenticator, "activity:read"))

	path, err := authenticator.CallbackPath()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	http.HandleFunc(path, authenticator.HandlerFunc(oAuthSuccess, oAuthFailure))

	// start the server
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil)
		if err != nil {
			panic("ListenAndServe: " + err.Error())
		}
	}()

	<-done

	return token, errr
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (string, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Saves a token to a file path.
func saveToken(path string, token string) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	f.WriteString(token)
}

func oAuthSuccess(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	content, _ := json.MarshalIndent(auth.Athlete, "", " ")
	fmt.Fprint(w, string(content))

	token = auth.AccessToken
	errr = nil
	done <- true
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

	token = ""
	errr = err
	done <- true
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

// AuthorizationURL constructs the url a user should use to authorize this specific application.
func authorizationURL(auth *strava.OAuthAuthenticator, scope string) string {
	path := fmt.Sprintf("%s/oauth/authorize?client_id=%d&response_type=code&redirect_uri=%s&scope=%s&approval_prompt=force", basePath, strava.ClientId, auth.CallbackURL, scope)
	return path
}
