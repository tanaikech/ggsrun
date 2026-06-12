// Package main (init.go) :
// These methods are for reading and writing configuration file (ggsrun.cfg).
package app

import (
	"fmt"
	"os"
	"path/filepath"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/urfave/cli"
)

// resolveConfigFile determines the exact path to ggsrun.cfg based on strict priority.
func (i *InitVal) resolveConfigFile() string {
	// Priority 1: --config flag explicitly sets the path
	if i.customConfig != "" {
		return filepath.Join(i.customConfig, cfgFile)
	}
	// Priority 2: --credentials implicit path binding
	if i.customCred != "" {
		return filepath.Join(filepath.Dir(i.customCred), cfgFile)
	}
	// Priority 3: GGSRUN_CFG_PATH environment variable
	if i.envConfig != "" {
		p := filepath.Join(i.envConfig, cfgFile)
		// Return if exists, or if we are actively provisioning a new auth structure
		if _, err := os.Stat(p); err == nil || i.isAuthCmd {
			return p
		}
	}
	// Priority 4: Final fallback to current working directory
	return filepath.Join(i.workdir, cfgFile)
}

// resolveCredFile determines the exact path to the credentials file based on strict priority.
func (i *InitVal) resolveCredFile() string {
	// Priority 1: --credentials explicitly states exact file target
	if i.customCred != "" {
		return i.customCred
	}
	// Priority 2: GGSRUN_CFG_PATH directory search
	if i.envConfig != "" {
		p := filepath.Join(i.envConfig, clientsecretFile)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// Priority 3: Final fallback to current working directory
	return filepath.Join(i.workdir, clientsecretFile)
}

// GgsrunIni : Initialize ggsrun
func (a *AuthContainer) ggsrunIni(c *cli.Context) *AuthContainer {
	a.UpdateStatus("Reading configuration...")
	cfgPath := a.resolveConfigFile()
	if !c.Bool("jsonparser") {
		absCfgPath, _ := filepath.Abs(cfgPath)
		fmt.Fprintf(os.Stdout, "[INFO] Using config file: %s\n", absCfgPath)
	}
	if cfgdata, err := os.ReadFile(cfgPath); err == nil {
		err = json.Unmarshal(cfgdata, &a.GgsrunCfg)
		if err != nil {
			a.FailStatus("Configuration Error")
			pterm.Error.Printf("Format parsing failure for '%s'.\n", cfgPath)
			os.Exit(1)
		}
		if c.Command.Name == "exe1" || c.Command.Name == "exe2" || c.Command.Name == "webapps" {
			if c.Command.Name != "webapps" {
				if len(c.String("scriptid")) == 0 && len(a.GgsrunCfg.Scriptid) == 0 {
					a.FailStatus("Validation Error")
					pterm.Error.Println("No script id. Please supply option '-i [Script ID]'.")
					os.Exit(1)
				}
			}
			originalScriptID := a.GgsrunCfg.Scriptid
			if len(c.String("scriptid")) > 0 {
				a.GgsrunCfg.Scriptid = c.String("scriptid")
				a.InitVal.hasNewScriptID = true
				a.InitVal.originalScriptID = originalScriptID
			}
			if len(c.String("function")) > 0 {
				a.Param.Function = c.String("function")
			}
		}
	} else {
		return a.readClientSecret()
	}
	return a
}

// readClientSecret : Read client secret file based on hierarchical priority
func (a *AuthContainer) readClientSecret() *AuthContainer {
	credPath := a.resolveCredFile()
	if csecret, err := os.ReadFile(credPath); err == nil {
		err := json.Unmarshal(csecret, &a.Cs)
		if err != nil || (len(a.Cs.Cid.ClientID) == 0 && len(a.Cs.Ciw.ClientID) == 0) {
			a.FailStatus("Configuration Error")
			pterm.Error.Printf("Credentials schema mismatch in '%s'.\nError trace: %s.\n", credPath, err)
			os.Exit(1)
		}
		if len(a.Cs.Cid.ClientID) == 0 && len(a.Cs.Ciw.ClientID) > 0 {
			a.Cs.Cid = a.Cs.Ciw
		}
	} else {
		a.FailStatus("Configuration Error")
		pterm.Error.Printf("No authentication materials located at '%s'.\n", credPath)
		os.Exit(1)
	}
	return a
}
