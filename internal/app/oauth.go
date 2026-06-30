// Package main (oauth.go) :
// Get accesstoken using refreshtoken, and confirm condition of accesstoken.
package app

import (
	"fmt"
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

	"ggsrun/internal/utl"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	gettokenbyserviceaccount "github.com/tanaikech/go-gettokenbyserviceaccount"
	"github.com/urfave/cli"
)

// Goauth :
func (a *AuthContainer) goauth() *AuthContainer {
	a.UpdateStatus("Authenticating...")
	if a.useServiceAccount != "" {
		if err := a.getAtFromSa(); err != nil {
			a.FailStatus("Authentication Failed")
			pterm.Error.Println(err)
			utl.Exit(1)
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

// tryLoadAuth : Attempt to load authentication configuration securely without crashing.
// Used primarily for webapps which can function both anonymously and securely authenticated.
func (a *AuthContainer) tryLoadAuth(c *cli.Context) {
	cfgPath := a.resolveConfigFile()
	if !c.Bool("jsonparser") {
		absCfgPath, _ := filepath.Abs(cfgPath)
		fmt.Fprintf(os.Stdout, "[INFO] Using config file: %s\n", absCfgPath)
	}
	if cfgdata, err := os.ReadFile(cfgPath); err == nil {
		json.Unmarshal(cfgdata, &a.GgsrunCfg)

		// Detect if Drive scope exists for secure execution
		hasScope := false
		for _, s := range a.GgsrunCfg.Scopes {
			if strings.Contains(s, "auth/drive") || strings.Contains(s, "auth/drive.readonly") {
				hasScope = true
				break
			}
		}

		if hasScope {
			if a.useServiceAccount != "" {
				a.getAtFromSa()
			} else if a.GgsrunCfg.Clientid != "" && a.GgsrunCfg.Refreshtoken != "" {
				// Refresh the token automatically if it has expired
				if (a.InitVal.pstart.Unix() - a.GgsrunCfg.Expiresin) > 0 {
					a.getAtoken()
					a.makecfgfile()
				}
			}
		}
	} else if a.useServiceAccount != "" {
		a.getAtFromSa()
	}
}

// ReAuth : Overhauled interactive auth configuration command for v5.2.0
func (a *AuthContainer) reAuth() {
	// Enforce normalized 14 scopes for new authorization including Cloud Logging
	a.GgsrunCfg.Scopes = []string{
		"https://www.googleapis.com/auth/drive",
		"https://www.googleapis.com/auth/drive.file",
		"https://www.googleapis.com/auth/drive.scripts",
		"https://www.googleapis.com/auth/script.external_request",
		"https://www.googleapis.com/auth/script.scriptapp",
		"https://www.googleapis.com/auth/spreadsheets",
		"https://www.googleapis.com/auth/documents",
		"https://www.googleapis.com/auth/script.projects",
		"https://www.googleapis.com/auth/script.deployments",
		"https://www.googleapis.com/auth/presentations",
		"https://www.googleapis.com/auth/forms",
		"https://mail.google.com/",
		"https://www.googleapis.com/auth/script.webapp.deploy",
		"https://www.googleapis.com/auth/logging.read",
	}

	cfgPath := a.resolveConfigFile()
	absCfgPath, _ := filepath.Abs(cfgPath)
	credPath := a.resolveCredFile()
	absCredPath, _ := filepath.Abs(credPath)
	currentEnvPath := os.Getenv("GGSRUN_CFG_PATH")

	fmt.Println("==================================================")
	fmt.Println("ggsrun OAuth Authorization Setup")
	fmt.Println("==================================================")
	fmt.Printf("Config File Path Resolved ('ggsrun.cfg'): %s\n", absCfgPath)
	fmt.Printf("Credentials File Path Resolved ('client_secret.json'): %s\n", absCredPath)
	if currentEnvPath != "" {
		fmt.Printf("Active GGSRUN_CFG_PATH: %s\n", currentEnvPath)
	}
	fmt.Println("==================================================")
	fmt.Printf("Client Secret Path ('client_secret.json'): %s\n", absCredPath)
	fmt.Printf("GGSRUN_CFG_PATH Environment Variable: '%s'\n", currentEnvPath)
	fmt.Println("==================================================")

	// Target Location Alteration
	defaultDir := a.InitVal.workdir
	label := "Default"
	if currentEnvPath != "" {
		defaultDir = currentEnvPath
		label = "Current"
	}

	var newDir string
	if a.InitVal.autoConfirm {
		newDir = defaultDir
		pterm.Info.Printf("Enter directory path to save 'ggsrun.cfg' (%s: %s) [Auto-confirmed]\n", label, defaultDir)
	} else {
		fmt.Printf("Enter directory path to save 'ggsrun.cfg' (%s: %s): ", label, defaultDir)
		fmt.Scanln(&newDir)
		newDir = strings.TrimSpace(newDir)
		if newDir == "" {
			newDir = defaultDir
		}
	}
	absNewDir, err := filepath.Abs(newDir)
	if err == nil {
		a.InitVal.customConfig = absNewDir
		cfgPath = filepath.Join(absNewDir, cfgFile)
		absCfgPath, _ = filepath.Abs(cfgPath)
		fmt.Printf("Config file will be saved to: %s\n", absCfgPath)

		// Reload configuration from the newly selected path if it exists
		if cfgdata, err := os.ReadFile(cfgPath); err == nil {
			var tempCfg GgsrunCfg
			if err2 := json.Unmarshal(cfgdata, &tempCfg); err2 == nil {
				if tempCfg.Scriptid != "" {
					a.GgsrunCfg.Scriptid = tempCfg.Scriptid
				}
				if tempCfg.WebappsUrl != "" {
					a.GgsrunCfg.WebappsUrl = tempCfg.WebappsUrl
				}
				if tempCfg.Projectid != "" {
					a.GgsrunCfg.Projectid = tempCfg.Projectid
				}
				if tempCfg.Clientid != "" && a.GgsrunCfg.Clientid == "" {
					a.GgsrunCfg.Clientid = tempCfg.Clientid
				}
				if tempCfg.Clientsecret != "" && a.GgsrunCfg.Clientsecret == "" {
					a.GgsrunCfg.Clientsecret = tempCfg.Clientsecret
				}
			}
		}

		// Check deviation from environmental GGSRUN_CFG_PATH (only if GGSRUN_CFG_PATH is set)
		if currentEnvPath != "" {
			absEnvDir, _ := filepath.Abs(currentEnvPath)
			if absNewDir != absEnvDir {
				pterm.Warning.Println("==================================================")
				pterm.Warning.Println("WARNING: CONFIGURATION PATH MISMATCH")
				pterm.Warning.Printf("The selected directory '%s' deviates from the environmental %s='%s'.\n", absNewDir, cfgpathenv, currentEnvPath)
				pterm.Warning.Printf("Please align the environment variable by running:\n  export %s=\"%s\"\n", cfgpathenv, absNewDir)
				pterm.Warning.Println("==================================================")
			}
		}
	}

	// Project ID Setup
	var promptMsg string
	if a.GgsrunCfg.Scriptid != "" {
		promptMsg = fmt.Sprintf("Please enter your Google Apps Script project Script ID (Press Enter to skip, you can register this later) [Current: '%s']: ", a.GgsrunCfg.Scriptid)
	} else {
		promptMsg = "Please enter your Google Apps Script project Script ID (Press Enter to skip, you can register this later): "
	}
	var scriptID string
	if a.InitVal.autoConfirm {
		scriptID = ""
		pterm.Info.Println("Script ID Setup [Auto-confirmed]")
	} else {
		fmt.Print(promptMsg)
		fmt.Scanln(&scriptID)
		scriptID = strings.TrimSpace(scriptID)
	}
	if scriptID == "" {
		if a.GgsrunCfg.Scriptid != "" {
			pterm.Info.Printf("Keeping current Script ID: %s\n", a.GgsrunCfg.Scriptid)
		} else {
			pterm.Warning.Println("Warning: Script ID registration skipped. Google Apps Script capabilities are blocked until configured or run with option '-i'.")
		}
	} else {
		a.GgsrunCfg.Scriptid = scriptID
		pterm.Success.Printf("Registered Script ID: %s\n", scriptID)
	}

	// Web Apps URL Setup
	var webAppsPrompt string
	if a.GgsrunCfg.WebappsUrl != "" {
		webAppsPrompt = fmt.Sprintf("Please enter your Google Apps Script Web Apps URL (Press Enter to skip, you can register this later) [Current: '%s']: ", a.GgsrunCfg.WebappsUrl)
	} else {
		webAppsPrompt = "Please enter your Google Apps Script Web Apps URL (Press Enter to skip, you can register this later): "
	}
	var webappsURL string
	if a.InitVal.autoConfirm {
		webappsURL = ""
		pterm.Info.Println("Web Apps URL Setup [Auto-confirmed]")
	} else {
		fmt.Print(webAppsPrompt)
		fmt.Scanln(&webappsURL)
		webappsURL = strings.TrimSpace(webappsURL)
	}
	if webappsURL == "" {
		if a.GgsrunCfg.WebappsUrl != "" {
			pterm.Info.Printf("Keeping current Web Apps URL: %s\n", a.GgsrunCfg.WebappsUrl)
		} else {
			pterm.Info.Println("Web Apps URL registration skipped.")
		}
	} else {
		a.GgsrunCfg.WebappsUrl = webappsURL
		pterm.Success.Printf("Registered Web Apps URL: %s\n", webappsURL)
	}

	// Consent Pre-flight Disclosure
	a.readClientSecret()
	if a.Cs.Cid.ClientID == "" && a.Cs.Ciw.ClientID != "" {
		a.Cs.Cid = a.Cs.Ciw
	}
	a.resolveCredentialsConflict()

	fmt.Println("\nOAuth 2.0 Credentials:")
	fmt.Printf("  Client ID: %s\n", a.Cs.Cid.ClientID)
	fmt.Printf("  Client Secret: %s\n", maskClientSecret(a.Cs.Cid.Clientsecret))
	if a.GgsrunCfg.Projectid != "" {
		fmt.Printf("  GCP Project ID: %s\n", a.GgsrunCfg.Projectid)
	}
	fmt.Println("Requested OAuth Scopes:")
	for _, scope := range a.GgsrunCfg.Scopes {
		fmt.Printf("  - %s\n", scope)
	}
	fmt.Println("==================================================")

	fmt.Print("Proceed to launch browser for authentication? [y/N]: ")
	var confirm string
	if a.InitVal.autoConfirm {
		confirm = "y"
		pterm.Info.Println("Proceed to launch browser for authentication? [y/N] [Auto-confirmed]")
	} else {
		fmt.Scanln(&confirm)
	}
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		pterm.Info.Println("Authentication setup aborted.")
		utl.Exit(0)
	}

	a.getNewAccesstoken().makecfgfile()
}

func maskClientSecret(secret string) string {
	if len(secret) <= 10 {
		return "********"
	}
	return secret[:6] + "********" + secret[len(secret)-4:]
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

	// Temporarily restore original script ID to prevent premature configuration pollution
	currentScriptID := a.GgsrunCfg.Scriptid
	if a.InitVal.hasNewScriptID {
		a.GgsrunCfg.Scriptid = a.InitVal.originalScriptID
	}

	btok, _ := json.MarshalIndent(a.GgsrunCfg, "", "\t")

	// Restore current script ID for execution runtime context
	if a.InitVal.hasNewScriptID {
		a.GgsrunCfg.Scriptid = currentScriptID
	}

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
	if a.GgsrunCfg.UseGcloudAuth {
		a.UpdateStatus("Refreshing Access Token via gcloud CLI...")
		cmd := exec.Command("gcloud", "auth", "print-access-token")
		out, err := cmd.Output()
		if err != nil {
			a.FailStatus("Token Refresh via gcloud Failed")
			pterm.Error.Printf("Failed to refresh token via gcloud: %v\n", err)
			utl.Exit(1)
		}
		a.GgsrunCfg.Accesstoken = strings.TrimSpace(string(out))
		a.GgsrunCfg.Expiresin = time.Now().Unix() + 3540 // 59 minutes
		return a
	}
	a.UpdateStatus("Refreshing Access Token...")
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
		a.FailStatus("Token Refresh Failed")
		pterm.Error.Printf("%v. %s\n", err, body)
		pterm.Info.Println("Hint: Try clearing your existing config manually or invoke 'ggsrun auth'.")
		utl.Exit(1)
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
		a.FailStatus("Token Check Failed")
		pterm.Error.Printf("%v. ", err)
		utl.Exit(1)
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
	pterm.Info.Printf("Please open this URL in your browser if it does not open automatically:\n%s\n", codeurl)
	var cmd *exec.Cmd
	if isWSL() {
		fmt.Println("\nWSL 2 environment detected.")
		fmt.Println("Please choose which browser to open the authentication page:")
		fmt.Println("  [1] Windows default browser (Recommended)")
		fmt.Println("  [2] WSL/Ubuntu native browser")
		fmt.Println("  [3] Do not open automatically (Manual copy-paste)")
		fmt.Print("Enter choice [1-3] (Default: 1): ")
		var choice string
		fmt.Scanln(&choice)
		choice = strings.TrimSpace(choice)
		if choice == "" {
			choice = "1"
		}

		switch choice {
		case "1":
			cmd = getWslBrowserCmd(codeurl)
		case "2":
			cmd = exec.Command("xdg-open", codeurl)
		case "3":
			pterm.Info.Println("Automatic browser launch skipped. Please use the URL printed above.")
		default:
			pterm.Warning.Println("Invalid choice. Skipping automatic launch.")
		}
	} else {
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
	}

	if cmd != nil {
		if err := cmd.Start(); err != nil {
			pterm.Warning.Println("Could not open browser automatically.")
		}
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
	a.UpdateStatus("Requesting new Access Token...")
	pterm.Info.Println("Authorization process initiated...")
	code, err := a.getCode()
	if err != nil {
		a.FailStatus("Authorization Flow Error")
		pterm.Error.Printf("Error during authorization flow: %v\n", err)
		utl.Exit(1)
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
		a.FailStatus("Token Issuance Failed")
		pterm.Error.Printf("[ %v ] - Authorization token issuance failed. ", err)
		utl.Exit(1)
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

// isWSL checks if the current environment is Windows Subsystem for Linux (WSL).
func isWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}
	content := strings.ToLower(string(data))
	return strings.Contains(content, "microsoft") || strings.Contains(content, "wsl")
}

// getWslBrowserCmd resolves the appropriate command to launch a browser in the Windows host from WSL.
func getWslBrowserCmd(codeurl string) *exec.Cmd {
	if _, err := exec.LookPath("wslview"); err == nil {
		return exec.Command("wslview", codeurl)
	}
	if _, err := exec.LookPath("cmd.exe"); err == nil {
		return exec.Command("cmd.exe", "/c", "start", "", codeurl)
	}
	if _, err := exec.LookPath("powershell.exe"); err == nil {
		escapedUrl := strings.ReplaceAll(codeurl, "'", "`'")
		return exec.Command("powershell.exe", "-NoProfile", "-Command", fmt.Sprintf("Start-Process '%s'", escapedUrl))
	}
	return exec.Command("xdg-open", codeurl)
}

// quickSetup : Interactive, user-friendly setup flow utilizing GCP Quick Flows
func (a *AuthContainer) quickSetup() {
	// Target Scopes Definition
	a.GgsrunCfg.Scopes = []string{
		"https://www.googleapis.com/auth/drive",
		"https://www.googleapis.com/auth/drive.file",
		"https://www.googleapis.com/auth/drive.scripts",
		"https://www.googleapis.com/auth/script.external_request",
		"https://www.googleapis.com/auth/script.scriptapp",
		"https://www.googleapis.com/auth/spreadsheets",
		"https://www.googleapis.com/auth/documents",
		"https://www.googleapis.com/auth/script.projects",
		"https://www.googleapis.com/auth/script.deployments",
		"https://www.googleapis.com/auth/presentations",
		"https://www.googleapis.com/auth/forms",
		"https://mail.google.com/",
		"https://www.googleapis.com/auth/script.webapp.deploy",
		"https://www.googleapis.com/auth/logging.read",
	}

	cfgPath := a.resolveConfigFile()
	absCfgPath, _ := filepath.Abs(cfgPath)
	credPath := a.resolveCredFile()
	absCredPath, _ := filepath.Abs(credPath)
	currentEnvPath := os.Getenv("GGSRUN_CFG_PATH")

	fmt.Println("==================================================")
	fmt.Println("ggsrun Simplified Onboarding Quick Setup")
	fmt.Println("==================================================")
	fmt.Printf("Config File Path Resolved ('ggsrun.cfg'): %s\n", absCfgPath)
	fmt.Printf("Credentials File Path Resolved ('client_secret.json'): %s\n", absCredPath)
	if currentEnvPath != "" {
		fmt.Printf("Active GGSRUN_CFG_PATH: %s\n", currentEnvPath)
	}
	fmt.Println("==================================================")
	pterm.Info.Println("Step 1: Enable Google APIs & Create Credentials")
	pterm.Info.Println("We will open a custom GCP setup link in your browser to:")
	pterm.Info.Println("  1. Automatically enable Drive, Google Apps Script and Cloud Logging APIs.")
	pterm.Info.Println("  2. Redirect you straight to the Credentials creation page.")
	fmt.Println()

	quickFlowURL := "https://console.cloud.google.com/flows/enableapi?apiid=drive.googleapis.com,script.googleapis.com,sheets.googleapis.com,gmail.googleapis.com,slides.googleapis.com,docs.googleapis.com,logging.googleapis.com&redirect_to=https://console.cloud.google.com/apis/credentials"

	var launchConfirm string
	if a.InitVal.autoConfirm {
		launchConfirm = "y"
		pterm.Info.Println("Proceed to launch browser to set up APIs? [Y/n] [Auto-confirmed]")
	} else {
		fmt.Print("Proceed to launch browser to set up APIs? [Y/n]: ")
		fmt.Scanln(&launchConfirm)
		launchConfirm = strings.ToLower(strings.TrimSpace(launchConfirm))
	}
	if launchConfirm == "" || launchConfirm == "y" {
		pterm.Info.Println("Launching browser...")
		var cmd *exec.Cmd
		if isWSL() {
			cmd = getWslBrowserCmd(quickFlowURL)
		} else {
			switch runtime.GOOS {
			case "darwin":
				cmd = exec.Command("open", quickFlowURL)
			case "linux":
				cmd = exec.Command("xdg-open", quickFlowURL)
			case "windows":
				cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", quickFlowURL)
			default:
				cmd = exec.Command("xdg-open", quickFlowURL)
			}
		}
		if cmd != nil {
			if err := cmd.Start(); err != nil {
				pterm.Warning.Println("Could not open browser automatically.")
			}
		}
	}

	pterm.Info.Println("If the browser did not open, please navigate to this URL manually:")
	pterm.Info.Println(quickFlowURL)
	fmt.Println("==================================================")

	pterm.Info.Println("Step 2: Load OAuth 2.0 Credentials")
	pterm.Info.Println("Once you have created the 'Desktop app' credentials, choose an option:")
	fmt.Println("  [1] Provide path to downloaded client secret JSON file (Recommended)")
	fmt.Println("  [2] Enter Client ID and Client Secret manually")
	hasGcloud := isGcloudAvailable()
	if hasGcloud {
		fmt.Println("  [3] Use active credentials from gcloud CLI (Auto-detected)")
		fmt.Print("Enter choice [1-3] (Default: 1): ")
	} else {
		fmt.Print("Enter choice [1-2] (Default: 1): ")
	}
	var choice string
	fmt.Scanln(&choice)
	choice = strings.TrimSpace(choice)
	if choice == "" {
		choice = "1"
	}

	if choice == "1" {
		var candidates []string
		if a.InitVal.customCred != "" {
			candidates = append(candidates, a.InitVal.customCred)
		} else {
			cwdPath := filepath.Join(a.InitVal.workdir, clientsecretFile)
			if _, err := os.Stat(cwdPath); err == nil {
				candidates = append(candidates, cwdPath)
			}
			if a.InitVal.envConfig != "" {
				envPath := filepath.Join(a.InitVal.envConfig, clientsecretFile)
				if _, err := os.Stat(envPath); err == nil {
					if absCwd, err1 := filepath.Abs(cwdPath); err1 == nil {
						if absEnv, err2 := filepath.Abs(envPath); err2 == nil && absCwd != absEnv {
							candidates = append(candidates, envPath)
						}
					} else {
						candidates = append(candidates, envPath)
					}
				}
			}
		}

		var jsonPath string
		if len(candidates) > 0 {
			pterm.Info.Println("Found potential credentials file(s):")
			for idx, cand := range candidates {
				absCand, _ := filepath.Abs(cand)
				fmt.Printf("  [%d] %s\n", idx+1, absCand)
			}
			customIdx := len(candidates) + 1
			var sel string
			if a.InitVal.autoConfirm {
				sel = "1"
				pterm.Info.Printf("Select an option [1-%d] (Default: 1) [Auto-confirmed]: 1\n", customIdx)
			} else {
				fmt.Printf("Select an option [1-%d] (Default: 1): ", customIdx)
				fmt.Scanln(&sel)
				sel = strings.TrimSpace(sel)
			}
			if sel == "" {
				sel = "1"
			}
			selNum, err := strconv.Atoi(sel)
			if err == nil && selNum >= 1 && selNum <= len(candidates) {
				jsonPath = candidates[selNum-1]
			} else if err == nil && selNum == customIdx {
				fmt.Print("Enter path to JSON file: ")
				fmt.Scanln(&jsonPath)
				jsonPath = strings.TrimSpace(jsonPath)
			} else {
				pterm.Error.Println("Invalid option selection.")
				utl.Exit(1)
			}
		} else {
			pterm.Warning.Println("No credentials file found in standard locations.")
			fmt.Print("Enter path to JSON file: ")
			fmt.Scanln(&jsonPath)
			jsonPath = strings.TrimSpace(jsonPath)
		}

		if jsonPath == "" {
			pterm.Error.Println("JSON file path cannot be empty.")
			utl.Exit(1)
		}
		if strings.HasPrefix(jsonPath, "~") {
			home, _ := os.UserHomeDir()
			jsonPath = filepath.Join(home, jsonPath[1:])
		}
		credentialsData, err := os.ReadFile(jsonPath)
		if err != nil {
			pterm.Error.Printf("Failed to read JSON file: %v\n", err)
			utl.Exit(1)
		}
		err = json.Unmarshal(credentialsData, &a.Cs)
		if err != nil || (len(a.Cs.Cid.ClientID) == 0 && len(a.Cs.Ciw.ClientID) == 0) {
			pterm.Error.Println("Invalid credentials JSON format.")
			utl.Exit(1)
		}
		if len(a.Cs.Cid.ClientID) == 0 && len(a.Cs.Ciw.ClientID) > 0 {
			a.Cs.Cid = a.Cs.Ciw
		}
		a.resolveCredentialsConflict()
	} else if choice == "2" {
		fmt.Print("Enter Client ID: ")
		var clientID string
		fmt.Scanln(&clientID)
		clientID = strings.TrimSpace(clientID)

		fmt.Print("Enter Client Secret: ")
		var clientSecret string
		fmt.Scanln(&clientSecret)
		clientSecret = strings.TrimSpace(clientSecret)

		if clientID == "" || clientSecret == "" {
			pterm.Error.Println("Client ID and Client Secret cannot be empty.")
			utl.Exit(1)
		}
		a.Cs.Cid.ClientID = clientID
		a.Cs.Cid.Clientsecret = clientSecret

		fmt.Print("Enter GCP Project ID (Optional, press Enter to skip. Required for log retrieval): ")
		var projectID string
		fmt.Scanln(&projectID)
		projectID = strings.TrimSpace(projectID)
		a.Cs.Cid.Projectid = projectID
		a.GgsrunCfg.Projectid = projectID
	} else if choice == "3" && hasGcloud {
		pterm.Info.Println("Detecting credentials from gcloud CLI...")
		
		// Get Project ID
		pCmd := exec.Command("gcloud", "config", "get-value", "project")
		pOut, err := pCmd.Output()
		if err != nil {
			pterm.Error.Printf("Failed to retrieve project ID from gcloud: %v\n", err)
			utl.Exit(1)
		}
		projectID := strings.TrimSpace(string(pOut))
		if projectID == "" {
			pterm.Error.Println("No active project configured in gcloud CLI.")
			utl.Exit(1)
		}

		// Get Access Token
		tCmd := exec.Command("gcloud", "auth", "print-access-token")
		tOut, err := tCmd.Output()
		if err != nil {
			pterm.Error.Printf("Failed to retrieve access token from gcloud: %v\n", err)
			utl.Exit(1)
		}
		accessToken := strings.TrimSpace(string(tOut))

		a.GgsrunCfg.Projectid = projectID
		a.GgsrunCfg.Accesstoken = accessToken
		a.GgsrunCfg.UseGcloudAuth = true
		a.GgsrunCfg.Expiresin = time.Now().Unix() + 3540 // 59 minutes

		// Create dummy client secret to bypass following checks
		a.Cs.Cid.ClientID = "gcloud-cli-client"
		a.Cs.Cid.Clientsecret = "gcloud-cli-secret"
	} else {
		pterm.Error.Println("Invalid choice.")
		utl.Exit(1)
	}

	fmt.Println("==================================================")
	pterm.Info.Println("Step 3: Setup Google Apps Script Parameters")
	targetCfgPath := a.resolveConfigFile()
	absTargetCfg, _ := filepath.Abs(targetCfgPath)
	pterm.Info.Printf("(Saving parameters directly to: %s)\n", absTargetCfg)
	fmt.Println("==================================================")
	
	// Script ID Setup
	var scriptIDPrompt string
	if a.GgsrunCfg.Scriptid != "" {
		scriptIDPrompt = fmt.Sprintf("Enter your target Google Apps Script Script ID\n(Press Enter to skip, you can register this later) [Current: '%s']: ", a.GgsrunCfg.Scriptid)
	} else {
		scriptIDPrompt = "Enter your target Google Apps Script Script ID\n(Press Enter to skip, you can register this later): "
	}
	var scriptID string
	if a.InitVal.autoConfirm {
		scriptID = ""
		pterm.Info.Println("Script ID Setup [Auto-confirmed]")
	} else {
		fmt.Print(scriptIDPrompt)
		fmt.Scanln(&scriptID)
		scriptID = strings.TrimSpace(scriptID)
	}
	if scriptID != "" {
		a.GgsrunCfg.Scriptid = scriptID
		pterm.Success.Printf("Registered Script ID: %s\n", scriptID)
	} else if a.GgsrunCfg.Scriptid != "" {
		pterm.Info.Printf("Keeping current Script ID: %s\n", a.GgsrunCfg.Scriptid)
	}

	// Web Apps URL Setup
	var webAppsPrompt string
	if a.GgsrunCfg.WebappsUrl != "" {
		webAppsPrompt = fmt.Sprintf("Enter your Google Apps Script Web Apps URL\n(Press Enter to skip, you can register this later) [Current: '%s']: ", a.GgsrunCfg.WebappsUrl)
	} else {
		webAppsPrompt = "Enter your Google Apps Script Web Apps URL\n(Press Enter to skip, you can register this later): "
	}
	var webappsURL string
	if a.InitVal.autoConfirm {
		webappsURL = ""
		pterm.Info.Println("Web Apps URL Setup [Auto-confirmed]")
	} else {
		fmt.Print(webAppsPrompt)
		fmt.Scanln(&webappsURL)
		webappsURL = strings.TrimSpace(webappsURL)
	}
	if webappsURL != "" {
		a.GgsrunCfg.WebappsUrl = webappsURL
		pterm.Success.Printf("Registered Web Apps URL: %s\n", webappsURL)
	} else if a.GgsrunCfg.WebappsUrl != "" {
		pterm.Info.Printf("Keeping current Web Apps URL: %s\n", a.GgsrunCfg.WebappsUrl)
	}

	fmt.Println("==================================================")
	if a.GgsrunCfg.UseGcloudAuth {
		pterm.Info.Println("Step 4: Save Configuration")
		pterm.Success.Printf("Auto-detected GCP Project ID: %s\n", a.GgsrunCfg.Projectid)
		a.makecfgfile()
	} else {
		pterm.Info.Println("Step 4: Launch Consent Authorization")
		fmt.Printf("  Client ID: %s\n", a.Cs.Cid.ClientID)
		fmt.Printf("  Client Secret: %s\n", maskClientSecret(a.Cs.Cid.Clientsecret))
		fmt.Print("Proceed to launch browser to authorize ggsrun? [Y/n]: ")
		var confirm string
		if a.InitVal.autoConfirm {
			confirm = "y"
			pterm.Info.Println("Proceed to launch browser to authorize ggsrun? [Y/n] [Auto-confirmed]")
		} else {
			fmt.Scanln(&confirm)
			confirm = strings.ToLower(strings.TrimSpace(confirm))
		}
		if confirm != "" && confirm != "y" {
			pterm.Info.Println("Setup aborted.")
			utl.Exit(0)
		}

		a.getNewAccesstoken().makecfgfile()
	}
	pterm.Success.Println("Simplified Quick Setup completed successfully!")
}

// isGcloudAvailable : Checks if gcloud CLI is installed and has an active account.
func isGcloudAvailable() bool {
	_, err := exec.LookPath("gcloud")
	if err != nil {
		return false
	}
	cmd := exec.Command("gcloud", "config", "get-value", "account")
	out, err := cmd.Output()
	return err == nil && len(strings.TrimSpace(string(out))) > 0
}

// resolveCredentialsConflict compares client_secret.json and existing ggsrun.cfg.
// It prompts the user if there are mismatches and updates a.GgsrunCfg.
func (a *AuthContainer) resolveCredentialsConflict() {
	// 1. Resolve Client ID
	newClientID := a.Cs.Cid.ClientID
	existingClientID := a.GgsrunCfg.Clientid
	if existingClientID != "" && newClientID != "" && existingClientID != newClientID {
		pterm.Warning.Println("Mismatch detected for OAuth Client ID:")
		pterm.Println(fmt.Sprintf("  [1] From client_secret.json: %s", newClientID))
		pterm.Println(fmt.Sprintf("  [2] From existing ggsrun.cfg:  %s", existingClientID))
		var choice string
		if a.InitVal.autoConfirm {
			choice = "1"
			pterm.Info.Println("Which Client ID would you like to use? [1-2] (Default: 1) [Auto-confirmed]: 1")
		} else {
			fmt.Print("Which Client ID would you like to use? [1-2] (Default: 1): ")
			fmt.Scanln(&choice)
		}
		if strings.TrimSpace(choice) == "2" {
			a.Cs.Cid.ClientID = existingClientID
			a.Cs.Cid.Clientsecret = a.GgsrunCfg.Clientsecret
			pterm.Info.Println("Using Client ID from existing ggsrun.cfg.")
		} else {
			a.GgsrunCfg.Clientid = newClientID
			a.GgsrunCfg.Clientsecret = a.Cs.Cid.Clientsecret
			pterm.Info.Println("Using Client ID from client_secret.json.")
		}
	} else if newClientID != "" {
		if existingClientID == newClientID {
			if a.InitVal.autoConfirm {
				pterm.Info.Printf("OAuth Client ID is identical ('%s'). Keep using it? [Y/n] [Auto-confirmed]\n", newClientID)
			} else {
				fmt.Printf("OAuth Client ID is identical ('%s'). Keep using it? [Y/n]: ", newClientID)
				var ans string
				fmt.Scanln(&ans)
				if strings.ToLower(strings.TrimSpace(ans)) == "n" {
					pterm.Warning.Println("Please update your credentials file or enter manually.")
				}
			}
		}
		a.GgsrunCfg.Clientid = a.Cs.Cid.ClientID
		a.GgsrunCfg.Clientsecret = a.Cs.Cid.Clientsecret
	}

	// 2. Resolve GCP Project ID
	newProjectID := a.Cs.Cid.Projectid
	existingProjectID := a.GgsrunCfg.Projectid
	if existingProjectID != "" && newProjectID != "" && existingProjectID != newProjectID {
		pterm.Warning.Println("Mismatch detected for GCP Project ID:")
		pterm.Println(fmt.Sprintf("  [1] From client_secret.json: %s", newProjectID))
		pterm.Println(fmt.Sprintf("  [2] From existing ggsrun.cfg:  %s", existingProjectID))
		var choice string
		if a.InitVal.autoConfirm {
			choice = "1"
			pterm.Info.Println("Which Project ID would you like to use? [1-2] (Default: 1) [Auto-confirmed]: 1")
		} else {
			fmt.Print("Which Project ID would you like to use? [1-2] (Default: 1): ")
			fmt.Scanln(&choice)
		}
		if strings.TrimSpace(choice) == "2" {
			a.GgsrunCfg.Projectid = existingProjectID
			pterm.Info.Println("Using GCP Project ID from existing ggsrun.cfg.")
		} else {
			a.GgsrunCfg.Projectid = newProjectID
			pterm.Info.Println("Using GCP Project ID from client_secret.json.")
		}
	} else if newProjectID != "" {
		if existingProjectID == newProjectID {
			if a.InitVal.autoConfirm {
				pterm.Info.Printf("GCP Project ID is identical ('%s'). Keep using it? [Y/n] [Auto-confirmed]\n", newProjectID)
			} else {
				fmt.Printf("GCP Project ID is identical ('%s'). Keep using it? [Y/n]: ", newProjectID)
				var ans string
				fmt.Scanln(&ans)
				if strings.ToLower(strings.TrimSpace(ans)) == "n" {
					pterm.Warning.Println("Proceeding with caution. Project ID can be modified in ggsrun.cfg.")
				}
			}
		}
		a.GgsrunCfg.Projectid = newProjectID
	}
}
