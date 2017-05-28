// Package main (oauth.go) :
// Get accesstoken using refreshtoken, and confirm condition of accesstoken.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/tanaikech/ggsrun/utl"
)

// Goauth :
func (a *AuthContainer) goauth() *AuthContainer {
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
	return a
}

// ReAuth :
func (a *AuthContainer) reAuth() {
	a.readClientSecret().getNewAccesstoken().makecfgfile()
}

// makecfgfile :
func (a *AuthContainer) makecfgfile() {
	btok, _ := json.MarshalIndent(a.GgsrunCfg, "", "\t")
	ioutil.WriteFile(filepath.Join(a.InitVal.cfgdir, cfgFile), btok, 0777)
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
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
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
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
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

func (a *AuthContainer) chkRedirectURI() bool {
	for _, e := range a.Cs.Cid.Redirecturis {
		if strings.Contains(e, "localhost") {
			return true
		}
	}
	return false
}

func (a *AuthContainer) getCode() (string, error) {
	p := a.InitVal.Port
	if !a.chkRedirectURI() {
		return "", fmt.Errorf("Go manual mode.")
	}
	fmt.Printf("\n### This is a automatic input mode.\n### Please follow opened browser, login Google and click authentication.\n### Moves to a manual mode if you wait for 30 seconds under this situation.\n")
	a.Cs.Cid.Redirecturis = append(a.Cs.Cid.Redirecturis, "http://localhost:"+strconv.Itoa(p)+"/")
	codepara := url.Values{}
	codepara.Set("client_id", a.Cs.Cid.ClientID)
	codepara.Set("redirect_uri", a.Cs.Cid.Redirecturis[len(a.Cs.Cid.Redirecturis)-1])
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
				fmt.Fprintf(w, `<html><head><title>ggsrun status</title></head><body><p>Erorr.</p></body></html>`)
				s.Response <- authCode{Err: fmt.Errorf("Not found code.")}
				return
			}
			fmt.Fprintf(w, `<html><head><title>ggsrun status</title></head><body><p>The authentication was done. Please close this page.</p></body></html>`)
			s.Response <- authCode{Code: code}
		})
		var err error
		Listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			s.Response <- authCode{Err: err}
			return
		}
		server := http.Server{}
		server.Handler = mux
		go server.Serve(Listener)
		s.Start <- true
		<-s.End
		Listener.Close()
		s.Response <- authCode{Err: err}
		return
	}(p)
	<-s.Start
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", strings.Replace(codeurl, "&", `\&`, -1))
	case "linux":
		cmd = exec.Command("xdg-open", strings.Replace(codeurl, "&", `\&`, -1))
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", strings.Replace(codeurl, "&", `^&`, -1))
	default:
		return "", fmt.Errorf("Go manual mode.")
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("Go manual mode.")
	}
	var result authCode
	select {
	case result = <-s.Response:
	case <-time.After(time.Duration(30) * time.Second): // After 30 s, move to manual mode.
		return "", fmt.Errorf("Go manual mode.")
	}
	if result.Err != nil {
		return "", fmt.Errorf("Go manual mode.")
	}
	return result.Code, nil
}

// getNewAccesstoken : Retrieve accesstoken when there is no refreshtoken.
func (a *AuthContainer) getNewAccesstoken() *AuthContainer {
	var code string
	var err error
	code, err = a.getCode()
	if err != nil {
		codepara := url.Values{}
		codepara.Set("client_id", a.Cs.Cid.ClientID)
		codepara.Set("redirect_uri", a.Cs.Cid.Redirecturis[0])
		codepara.Set("scope", strings.Join(a.GgsrunCfg.Scopes, " "))
		codepara.Set("response_type", "code")
		codepara.Set("approval_prompt", "force")
		codepara.Set("access_type", "offline")
		codeurl := oauthurl + "auth?" + codepara.Encode()
		fmt.Printf("\n### This is a manual input mode.\n### Please input code retrieved by importing following URL to your browser.\n\n"+
			"[URL]==> %v\n"+
			"[CODE]==>", codeurl)
		if _, err := fmt.Scan(&code); err != nil {
			log.Fatalf("Error: %v.\n", err)
		}
		a.Cs.Cid.Redirecturis = append(a.Cs.Cid.Redirecturis, a.Cs.Cid.Redirecturis[0])
	}
	tokenparams := url.Values{}
	tokenparams.Set("client_id", a.Cs.Cid.ClientID)
	tokenparams.Set("client_secret", a.Cs.Cid.Clientsecret)
	tokenparams.Set("redirect_uri", a.Cs.Cid.Redirecturis[len(a.Cs.Cid.Redirecturis)-1])
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
		fmt.Fprintf(os.Stderr, "Error: [ %v ] - Code is wrong. ", err)
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
