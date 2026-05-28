// Package main (init.go) :
// These methods are for reading and writing configuration file (ggsrun.cfg).
package main

import (
	"os"
	"path/filepath"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/urfave/cli"
)

// resolveConfigFile determines the exact path to ggsrun.cfg based on strict priority.
func (a *AuthContainer) resolveConfigFile() string {
	// Priority 1: --config flag explicitly sets the path
	if a.InitVal.customConfig != "" {
		return filepath.Join(a.InitVal.customConfig, cfgFile)
	}
	// Priority 2: --credentials implicit path binding
	if a.InitVal.customCred != "" {
		return filepath.Join(filepath.Dir(a.InitVal.customCred), cfgFile)
	}
	// Priority 3: GGSRUN_CFG_PATH environment variable
	if a.InitVal.envConfig != "" {
		p := filepath.Join(a.InitVal.envConfig, cfgFile)
		// Return if exists, or if we are actively provisioning a new auth structure
		if _, err := os.Stat(p); err == nil || a.InitVal.isAuthCmd {
			return p
		}
	}
	// Priority 4: Final fallback to current working directory
	return filepath.Join(a.InitVal.workdir, cfgFile)
}

// resolveCredFile determines the exact path to the credentials file based on strict priority.
func (a *AuthContainer) resolveCredFile() string {
	// Priority 1: --credentials explicitly states exact file target
	if a.InitVal.customCred != "" {
		return a.InitVal.customCred
	}
	// Priority 2: GGSRUN_CFG_PATH directory search
	if a.InitVal.envConfig != "" {
		p := filepath.Join(a.InitVal.envConfig, clientsecretFile)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// Priority 3: Final fallback to current working directory
	return filepath.Join(a.InitVal.workdir, clientsecretFile)
}

// GgsrunIni : Initialize ggsrun
func (a *AuthContainer) ggsrunIni(c *cli.Context) *AuthContainer {
	cfgPath := a.resolveConfigFile()
	if cfgdata, err := os.ReadFile(cfgPath); err == nil {
		err = json.Unmarshal(cfgdata, &a.GgsrunCfg)
		if err != nil {
			pterm.Error.Printf("Format parsing failure for '%s'.\n", cfgPath)
			os.Exit(1)
		}
		if c.Command.Name == "exe1" || c.Command.Name == "exe2" {
			if len(c.String("scriptid")) == 0 && len(a.GgsrunCfg.Scriptid) == 0 {
				pterm.Error.Println("No script id. Please supply option '-i [Script ID]'.")
				os.Exit(1)
			}
			if len(c.String("scriptid")) > 0 {
				a.GgsrunCfg.Scriptid = c.String("scriptid")
				a.InitVal.update = true
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
			pterm.Error.Printf("Credentials schema mismatch in '%s'.\nError trace: %s.\n", credPath, err)
			os.Exit(1)
		}
		if len(a.Cs.Cid.ClientID) == 0 && len(a.Cs.Ciw.ClientID) > 0 {
			a.Cs.Cid = a.Cs.Ciw
		}
	} else {
		pterm.Error.Printf("No authentication materials located at '%s'.\n", credPath)
		os.Exit(1)
	}
	return a
}
