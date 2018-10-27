// Package main (init.go) :
// These methods are for reading and writing configuration file (ggsrun.cfg).
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

// GgsrunIni : Initialize ggsrun
func (a *AuthContainer) ggsrunIni(c *cli.Context) *AuthContainer {
	if cfgdata, err := a.chkInitFile(cfgFile); err == nil {
		err = json.Unmarshal(cfgdata, &a.GgsrunCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Format error of '%s'.\n", cfgFile)
			os.Exit(1)
		}
		if c.Command.Names()[0] == "exe1" ||
			c.Command.Names()[0] == "exe2" {
			if len(c.String("scriptid")) == 0 && len(a.GgsrunCfg.Scriptid) == 0 {
				fmt.Fprintf(os.Stderr, "Error: No script id. Please use option '-i [Script ID]'.\n")
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

// readClientSecret : Read client secret file
func (a *AuthContainer) readClientSecret() *AuthContainer {
	if csecret, err := a.chkInitFile(clientsecretFile); err == nil {
		err := json.Unmarshal(csecret, &a.Cs)
		if err != nil || (len(a.Cs.Cid.ClientID) == 0 && len(a.Cs.Ciw.ClientID) == 0) {
			fmt.Fprintf(os.Stderr, "Error: Please confirm '%s'.\nError is %s.\n", clientsecretFile, err)
			os.Exit(1)
		}
		if len(a.Cs.Cid.ClientID) == 0 && len(a.Cs.Ciw.ClientID) > 0 {
			a.Cs.Cid = a.Cs.Ciw
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: No materials for retrieving accesstoken. Please download '%s'.\n", clientsecretFile)
		os.Exit(1)
	}
	return a
}

// chkInitFile : Check initial files.
// By this method, at first, files are searched in working directory, and next, they are searched in the directory declared by the environment variable.
func (a *AuthContainer) chkInitFile(file string) ([]byte, error) {
	var err error
	var body []byte
	if body, err = ioutil.ReadFile(filepath.Join(a.InitVal.workdir, file)); err == nil {
		a.InitVal.usedDir = "work"
		return body, err
	}
	if a.InitVal.workdir != a.InitVal.cfgdir {
		if body, err = ioutil.ReadFile(filepath.Join(a.InitVal.cfgdir, file)); err == nil {
			a.InitVal.usedDir = "env"
			return body, err
		}
	}
	return nil, fmt.Errorf("error: %s was not found", file)
}
