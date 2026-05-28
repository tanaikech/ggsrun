// Package main (oauth.go) :
// Get accesstoken using refreshtoken, and confirm condition of accesstoken.
package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"ggsrun/utl"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	gettokenbyserviceaccount "github.com/tanaikech/go-gettokenbyserviceaccount"
)

// Goauth :
func (a *AuthContainer) goauth() *AuthContainer {
	if a.useServiceAccount != "" {
		if err := a.getAtFromSa(); err != nil {
			pterm.Error.Println(err)
			os.Exit(1)
		}
		a.Msg = append(a.Msg, "Service Account was used.")
		return a
	}
	if len(a.GgsrunCfg.Clientid) > 0 &&
		len(a.GgsrunCfg.Clientsecret) > 0 &&
		len(a.GgsrunCfg.Refreshtoken) > 0 {
		if (a.InitVal.pstart.Unix()-a.GgsrunCfg.Expiresin) > 0 ||
			len(a.GgsrunCfg.Accesstoken) == 0 {
			a.getAtoken().makecfgfile()
		} else {
			if a.InitVal.update {
				a.makecfgfile()
			}
		}
	} else {
		a.readClientSecret().getNewAccesstoken().makecfgfile()
	}
	a.Msg = append(a.Msg, "Access Token was was used.")
	return a
}

// ReAuth :
func (a *AuthContainer) reAuth() {
	a.readClientSecret().getNewAccesstoken().makecfgfile()
}

// makecfgfile : Generates and saves ggsrun.cfg to the strictly resolved path
func (a *AuthContainer) makecfgfile() {
	cfgPath := a.resolveConfigFile()

	if a.InitVal.isAuthCmd {
		if _, err := os.Stat(cfgPath); err == nil {
			pterm.Warning.Printf("Configuration file already exists at '%s'.\n", cfgPath)
			fmt.Print("### Do you want to overwrite it? [y/N]: ")
			var ans string
			fmt.Scanln(&ans)
			if strings.ToLower(strings.TrimSpace(ans)) != "y" {
				pterm.Info.Println("Aborted saving configuration.")
				return
			}
		}
	}

	btok, _ := json.MarshalIndent(a.GgsrunCfg, "", "\t")
	err := os.WriteFile(cfgPath, btok, 0600)
	if err != nil {
		pterm.Error.Printf("Could not securely write configuration to '%s'. %v\n", cfgPath, err)
	} else {
		if a.InitVal.isAuthCmd {
			pterm.Success.Printf("Successfully provisioned configuration file at: %s\n", cfgPath)
		}
	}
}

// getAtoken : Retrieves accesstoken from refreshtoken.
func (a *AuthContainer) getAtoken() *AuthContainer {
	a.Msg = append(a.Msg, "Got a new accesstoken.")
	values := url.Values{}
	values.Set("client_id", a.GgsrunCfg.Clientid)
	values.Set("client_secret", a.GgsrunCfg.Clientsecret)
	values.Set("refresh_token", a.GgsrunCfg.Refreshtoken)
	values.Set("grant_type", "refresh_token")
	r := &utl.RequestParams{
		Method:      "POST",
		APIURL:      oauthurl + "token",
		Data:        strings.NewReader(values.Encode()),
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: "",
		Dtime:       10,
	}
	body, err := r.FetchAPI()
	if err != nil {
		pterm.Error.Printf("%v. %s\n", err, body)
		pterm.Info.Println("Hint: Try clearing your existing config manually or invoke 'ggsrun auth'.")
		os.Exit(1)
	}
	json.Unmarshal(body, &a.Atoken)
	a.GgsrunCfg.Accesstoken = a.Atoken.Accesstoken
	a.GgsrunCfg.Expiresin = a.chkAtoken() - 360 // 6 minutes as adjustment time
	return a
}

// chkAtoken : For AuthContainer
func (a *AuthContainer) chkAtoken() int64 {
	r := &utl.RequestParams{
		Method:      "GET",
		APIURL:      chkatutl + "tokeninfo?access_token=" + a.GgsrunCfg.Accesstoken,
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: "",
		Dtime:       10,
	}
	body, err := r.FetchAPI()
	if err != nil {
		pterm.Error.Printf("%v. ", err)
		os.Exit(1)
	}
	json.Unmarshal(body, &a.ChkAt)
	if len(a.ChkAt.Error) > 0 {
		a.getAtoken()
	}
	exp, _ := strconv.ParseInt(a.ChkAt.Exp, 10, 64)
	return exp
}

// chkAtoken : For ExecutionContainer
func (e *ExecutionContainer) chkAtoken() *ChkAt {
	r := &utl.RequestParams{
		Method:      "GET",
		APIURL:      chkatutl + "tokeninfo?access_token=" + e.GgsrunCfg.Accesstoken,
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: "",
		Dtime:       10,
	}
	body, _ := r.FetchAPI()
	var c ChkAt
	json.Unmarshal(body, &c)
	return &c
}

func (a *AuthContainer) getCode() (string, error) {
	p := a.InitVal.Port
	redirectURI := "http://localhost:" + strconv.Itoa(p) + "/"
	a.Cs.Cid.Redirecturis = []string{redirectURI}

	codepara := url.Values{}
	codepara.Set("client_id", a.Cs.Cid.ClientID)
	codepara.Set("redirect_uri", redirectURI)
	codepara.Set("scope", strings.Join(a.GgsrunCfg.Scopes, " "))
	codepara.Set("response_type", "code")
	codepara.Set("approval_prompt", "force")
	codepara.Set("access_type", "offline")
	codeurl := oauthurl + "auth?" + codepara.Encode()

	s := &serverInfToGetCode{
		Response: make(chan authCode, 1),
		Start:    make(chan bool, 1),
		End:      make(chan bool, 1),
	}
	defer func() {
		s.End <- true
	}()
	go func(port int) {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			if len(code) == 0 {
				fmt.Fprintf(w, `<html><head><title>ggsrun Auth Error</title></head><body style="font-family: sans-serif; text-align: center; margin-top: 50px;"><h2>Authentication Error</h2><p>No code found in request.</p></body></html>`)
				s.Response <- authCode{Err: fmt.Errorf("not found code")}
				return
			}
			fmt.Fprintf(w, `<html><head><title>ggsrun Auth Success</title></head><body style="font-family: sans-serif; text-align: center; margin-top: 50px; background-color: #f0fdf4;"><h2>Authentication Successful!</h2><p>You can safely close this window and return to your terminal.</p></body></html>`)
			s.Response <- authCode{Code: code}
		})
		Listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			s.Response <- authCode{Err: err}
			return
		}
		server := http.Server{Handler: mux}
		go server.Serve(Listener)
		s.Start <- true
		<-s.End
		Listener.Close()
	}(p)

	<-s.Start

	pterm.Info.Println("Launching browser for automatic authentication...")
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", codeurl)
	case "linux":
		cmd = exec.Command("xdg-open", codeurl)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", codeurl)
	default:
		cmd = exec.Command("xdg-open", codeurl)
	}

	if err := cmd.Start(); err != nil {
		pterm.Warning.Printf("Could not open browser automatically. Please open this URL manually:\n%s\n", codeurl)
	}

	var result authCode
	select {
	case result = <-s.Response:
	case <-time.After(120 * time.Second):
		return "", fmt.Errorf("timeout waiting for authorization code")
	}

	if result.Err != nil {
		return "", result.Err
	}
	return result.Code, nil
}

// getNewAccesstoken : Retrieve accesstoken when there is no refreshtoken.
func (a *AuthContainer) getNewAccesstoken() *AuthContainer {
	pterm.Info.Println("Authorization process initiated...")
	code, err := a.getCode()
	if err != nil {
		pterm.Error.Printf("Error during authorization flow: %v\n", err)
		os.Exit(1)
	}

	tokenparams := url.Values{}
	tokenparams.Set("client_id", a.Cs.Cid.ClientID)
	tokenparams.Set("client_secret", a.Cs.Cid.Clientsecret)
	tokenparams.Set("redirect_uri", a.Cs.Cid.Redirecturis[0])
	tokenparams.Set("code", code)
	tokenparams.Set("grant_type", "authorization_code")
	r := &utl.RequestParams{
		Method:      "POST",
		APIURL:      oauthurl + "token",
		Data:        strings.NewReader(tokenparams.Encode()),
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: "",
		Dtime:       10,
	}
	body, err := r.FetchAPI()
	if err != nil {
		pterm.Error.Printf("[ %v ] - Authorization token issuance failed. ", err)
		os.Exit(1)
	}
	json.Unmarshal(body, &a.Atoken)
	a.GgsrunCfg.Clientid = a.Cs.Cid.ClientID
	a.GgsrunCfg.Clientsecret = a.Cs.Cid.Clientsecret
	a.GgsrunCfg.Refreshtoken = a.Atoken.Refreshtoken
	a.GgsrunCfg.Accesstoken = a.Atoken.Accesstoken
	a.GgsrunCfg.Expiresin = a.chkAtoken() - 360 // 6 minutes as adjustment time
	return a
}

// getAtFromSa : Retrieve access token from Service Account
func (a *AuthContainer) getAtFromSa() error {
	credentialsData, err := os.ReadFile(a.useServiceAccount)
	if err != nil {
		return err
	}
	para := struct {
		PrivateKey  string `json:"private_key"`
		ClientEmail string `json:"client_email"`
	}{}
	json.Unmarshal(credentialsData, &para)
	scopes := strings.Join(a.Scopes, " ")
	res, err := gettokenbyserviceaccount.Do(para.PrivateKey, para.ClientEmail, "", scopes)
	if err != nil {
		return err
	}
	a.GgsrunCfg.Accesstoken = res.AccessToken
	return nil
}
