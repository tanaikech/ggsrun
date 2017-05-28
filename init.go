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

// GgsrunIni :
func (a *AuthContainer) ggsrunIni(c *cli.Context) *AuthContainer {
	if cfgdata, err := ioutil.ReadFile(filepath.Join(a.InitVal.cfgdir, cfgFile)); err == nil {
		err = json.Unmarshal(cfgdata, &a.GgsrunCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Format error of '%s'. ", cfgFile)
			os.Exit(1)
		}
		if c.Command.Names()[0] == "exe1" ||
			c.Command.Names()[0] == "exe2" {
			if len(c.String("scriptid")) == 0 && len(a.GgsrunCfg.Scriptid) == 0 {
				fmt.Fprintf(os.Stderr, "Error: No script id. Please use option '-i [Script ID]'. ")
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

func (a *AuthContainer) readClientSecret() *AuthContainer {
	if csecret, err := ioutil.ReadFile(filepath.Join(a.InitVal.workdir, clientsecretFile)); err == nil {
		err := json.Unmarshal(csecret, &a.Cs)
		if err != nil || (len(a.Cs.Cid.ClientID) == 0 && len(a.Cs.Ciw.ClientID) == 0) {
			fmt.Fprintf(os.Stderr, "Error: Please confirm '%s'. Error is %s.", clientsecretFile, err)
			os.Exit(1)
		}
		if len(a.Cs.Cid.ClientID) == 0 && len(a.Cs.Ciw.ClientID) > 0 {
			a.Cs.Cid = a.Cs.Ciw
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: No materials for retrieving accesstoken. Please download '%s'", clientsecretFile)
		os.Exit(1)
	}
	return a
}
