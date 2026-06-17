// Package main (handler_execute.go) :
// Handles Google Apps Script Execution API and Web Apps integrations.
package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ggsrun/internal/utl"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/urfave/cli"
)

// Helper: Check if standard input is redirected/piped
func isStdinPiped() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// Helper: Read all text from standard input
func readAllStdin() (string, error) {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Safeguard helpers
func showE1HelpAndExit(c *cli.Context) {
	cli.ShowCommandHelp(c, c.Command.Name)
	utl.Exit(1)
}

func showE2HelpAndExit(c *cli.Context) {
	cli.ShowCommandHelp(c, c.Command.Name)
	utl.Exit(1)
}

func showWHelpAndExit(c *cli.Context) {
	cli.ShowCommandHelp(c, c.Command.Name)
	utl.Exit(1)
}

// exeAPIWithout : exe1
func exeAPIWithout(c *cli.Context) error {
	// 4. Safeguard: check for required script source
	if c.String("scriptfile") == "" && c.String("stringscript") == "" && !isStdinPiped() {
		showE1HelpAndExit(c)
	}

	a := defAuthContainer(c)
	if !c.Bool("jsonparser") {
		a.Spinner, _ = pterm.DefaultSpinner.WithRemoveWhenDone(false).Start("Initializing ggsrun...")
	}
	e := a.ggsrunIni(c).
		goauth().
		defExecutionContainer()

	// Defer cleanup of temporary file if it was uploaded
	defer func() {
		if e.InitVal.tempFileNameToCleanup != "" {
			e.UpdateStatus("Cleaning up temporary script file from GAS project...")
			var newFiles []File
			for _, f := range e.Project.Files {
				if f.Name != e.InitVal.tempFileNameToCleanup {
					newFiles = append(newFiles, f)
				}
			}
			e.Project.Files = newFiles
			e.projectUpdate2()
			if !c.Bool("jsonparser") {
				pterm.Success.Printf("Cleaned up temporary file '%s' successfully.\n", e.InitVal.tempFileNameToCleanup)
			}
		}
	}()

	// 8. Manifest auto-validation
	if err := e.autoValidateAndDeployManifest(c, "e1"); err != nil {
		e.FailStatus("Manifest Validation Error")
		pterm.Error.Println(err)
		utl.Exit(1)
	}

	e.exe1Function(c).
		executionAPIwithoutServer(c).
		esenderForExe1(c).
		dispResult(c)

	// 3. Interactive script ID registration
	handleScriptIDRegistration(c, e)

	return nil
}

// exeAPIWith : exe2
func exeAPIWith(c *cli.Context) error {
	// 4. Safeguard: check for required script source or execution flag
	if c.String("scriptfile") == "" && c.String("stringscript") == "" && !isStdinPiped() && !c.Bool("foldertree") && !c.Bool("convert") {
		showE2HelpAndExit(c)
	}

	a := defAuthContainer(c)
	if !c.Bool("jsonparser") {
		a.Spinner, _ = pterm.DefaultSpinner.WithRemoveWhenDone(false).Start("Initializing ggsrun...")
	}
	e := a.ggsrunIni(c).
		goauth().
		defExecutionContainer()

	// 7. Verify library ggsrunif and function ExecutionApi presence in project
	if c.String("function") == "" {
		if !e.GgsrunCfg.ExecutionApiChecked {
			e.UpdateStatus("Verifying 'ExecutionApi' function and 'ggsrunif' library in GAS project...")
			e.projectBackup(c)
			libImported, funcExists := checkExecutionApiExists(e.Project)
			if libImported && funcExists {
				e.GgsrunCfg.ExecutionApiChecked = true
				cfgPath := e.resolveConfigFile()
				saveConfig(e, cfgPath)
			} else {
				if !libImported {
					pterm.Warning.Printf("Warning: The required GAS library 'ggsrunif' is not imported in your project (Script ID: %s).\n", e.GgsrunCfg.Scriptid)
				}
				if !funcExists {
					pterm.Warning.Printf("Warning: The target entry function 'ExecutionApi' is not found in your project.\n")
				}
				pterm.Info.Println("Please run 'ggsrun auth' or configure your project manifest and functions correctly.")
			}
		}
	}

	// 8. Manifest auto-validation
	if err := e.autoValidateAndDeployManifest(c, "e2"); err != nil {
		e.FailStatus("Manifest Validation Error")
		pterm.Error.Println(err)
		utl.Exit(1)
	}

	e.exe2Function(c).
		dispResult(c)

	// 3. Interactive script ID registration
	handleScriptIDRegistration(c, e)

	return nil
}

// Helper to check execution api exists in project
func checkExecutionApiExists(proj *Project) (bool, bool) {
	libImported := false
	funcExists := false
	for _, f := range proj.Files {
		if f.Name == "appsscript" && strings.ToUpper(f.Type) == "JSON" {
			var manifest map[string]interface{}
			if err := json.Unmarshal([]byte(f.Source), &manifest); err == nil {
				if deps, ok := manifest["dependencies"].(map[string]interface{}); ok {
					if libs, ok := deps["libraries"].([]interface{}); ok {
						for _, lib := range libs {
							if libMap, ok := lib.(map[string]interface{}); ok {
								if libMap["userSymbol"] == "ggsrunif" {
									libImported = true
									break
								}
							}
						}
					}
				}
			}
		}
		if strings.ToUpper(f.Type) == "SERVER_JS" {
			content := f.Source
			if strings.Contains(content, "function ExecutionApi") ||
				strings.Contains(content, "ExecutionApi =") {
				funcExists = true
			}
		}
	}
	return libImported, funcExists
}

// webAppsWith : exe3
func webAppsWith(c *cli.Context) error {
	// 4. Safeguard: check for required script source
	if c.String("scriptfile") == "" && c.String("stringscript") == "" && !isStdinPiped() {
		showWHelpAndExit(c)
	}

	a := defAuthContainer(c)
	if !c.Bool("jsonparser") {
		a.Spinner, _ = pterm.DefaultSpinner.WithRemoveWhenDone(false).Start("Initializing Web Apps execution...")
	}

	// 2. Pass cli.Context to tryLoadAuth
	a.tryLoadAuth(c)
	e := a.defExecutionContainer()

	// Handle script ID override if provided
	if c.String("scriptid") != "" {
		e.GgsrunCfg.Scriptid = c.String("scriptid")
	}

	e.UpdateStatus("Resolving script payload...")
	var rawScript string
	var isTemp bool
	var tempFileName string

	if c.String("stringscript") != "" {
		rawScript = c.String("stringscript")
		isTemp = true
	} else if isStdinPiped() {
		var err error
		rawScript, err = readAllStdin()
		if err != nil {
			e.FailStatus("Stdin Read Error")
			pterm.Error.Println(err)
			utl.Exit(1)
		}
		isTemp = true
	} else {
		scriptFile := c.String("scriptfile")
		if scriptFile == "" {
			e.FailStatus("Initialization failed")
			pterm.Error.Println("No script. Please set GAS script using '-s' or '--stringscript'.")
			utl.Exit(1)
		}
		rawScript = utl.ConvGasToPut(c)
		isTemp = true
	}

	// 5. If isTemp in webapps (w), upload and deploy
	if isTemp && e.GgsrunCfg.Scriptid != "" && e.GgsrunCfg.Accesstoken != "" && c.String("scriptfile") != "" {
		timestamp := time.Now().Format("20060102150405")
		tempFileName = "ggsrun_web_temp_" + timestamp

		e.UpdateStatus("Self-healing & deploying temporary script for Web Apps...")
		e.projectBackup(c)

		defer func() {
			e.UpdateStatus("Cleaning up temporary script file from Web Apps project...")
			var newFiles []File
			for _, f := range e.Project.Files {
				if f.Name != tempFileName {
					newFiles = append(newFiles, f)
				}
			}
			e.Project.Files = newFiles
			e.projectUpdate2()

			// Deploy the clean state
			e.autoValidateAndDeployManifest(c, "w")
			if !c.Bool("jsonparser") {
				pterm.Success.Printf("Cleaned up temporary file '%s' successfully.\n", tempFileName)
			}
		}()

		filedata := File{
			Name:   tempFileName,
			Type:   "SERVER_JS",
			Source: rawScript,
		}
		e.Project.Files = append(e.Project.Files, filedata)
		e.projectUpdate2()

		// Deploy web app version
		if err := e.autoValidateAndDeployManifest(c, "w"); err != nil {
			e.FailStatus("Deployment Error")
			pterm.Error.Println(err)
			utl.Exit(1)
		}
	} else {
		if err := e.autoValidateAndDeployManifest(c, "w"); err != nil {
			e.FailStatus("Manifest Validation Error")
			pterm.Error.Println(err)
			utl.Exit(1)
		}
	}

	e.UpdateStatus("Resolving script payload...")
	val := c.String("value")
	var argStr string
	if val != "" {
		if numRe.MatchString(val) || arrayRe.MatchString(val) || objRe.MatchString(val) || val == "true" || val == "false" || val == "null" {
			argStr = val
		} else {
			argStr = fmt.Sprintf("%q", val)
		}
	}

	wrappedScript := fmt.Sprintf(`(function() {
%s
if (typeof main !== 'undefined') {
	return main(%s);
}
return "Execution completed, but main() was not defined in the local script.";
})()`, rawScript, argStr)

	quotedBytes, _ := json.Marshal(wrappedScript)
	quotedScript := string(quotedBytes)

	e.webAppswithServerForExe3(quotedScript, c).
		dispResult(c)
	handleScriptIDRegistration(c, e)

	// Web Apps URL Registration/Overwrite Prompt
	if c.String("url") != "" && !c.Bool("jsonparser") {
		currentStoredURL := e.GgsrunCfg.WebappsUrl
		if c.String("url") != currentStoredURL {
			cfgPath := e.resolveConfigFile()
			absCfgPath, _ := filepath.Abs(cfgPath)

			var promptMsg string
			if currentStoredURL != "" {
				promptMsg = fmt.Sprintf("\nDo you want to overwrite the registered Web Apps URL '%s' with '%s' in the config file ('%s')? (y/n): ", currentStoredURL, c.String("url"), absCfgPath)
			} else {
				promptMsg = fmt.Sprintf("\nDo you want to register Web Apps URL '%s' to config file ('%s')? (y/n): ", c.String("url"), absCfgPath)
			}
			fmt.Print(promptMsg)
			var ans string
			fmt.Scanln(&ans)
			ans = strings.ToLower(strings.TrimSpace(ans))
			if ans == "y" || ans == "yes" {
				e.GgsrunCfg.WebappsUrl = c.String("url")
				saveConfig(e, cfgPath)
				pterm.Success.Printf("Successfully registered Web Apps URL to config file.\n")
			}
		}
	}
	return nil
}

// 3. Interactive script ID registration helper
func handleScriptIDRegistration(c *cli.Context, e *ExecutionContainer) {
	if e.InitVal.hasNewScriptID {
		if c.Bool("jsonparser") {
			return
		}
		cfgPath := e.resolveConfigFile()
		absCfgPath, _ := filepath.Abs(cfgPath)

		fmt.Printf("\nDo you want to register Script ID '%s' to config file ('%s')? (y/n): ", e.GgsrunCfg.Scriptid, absCfgPath)
		var ans string
		fmt.Scanln(&ans)
		ans = strings.ToLower(strings.TrimSpace(ans))
		if ans == "y" || ans == "yes" {
			e.InitVal.hasNewScriptID = false
			saveConfig(e, cfgPath)
			pterm.Success.Printf("Successfully registered Script ID '%s' to config file.\n", e.GgsrunCfg.Scriptid)
		}
	}
}

func saveConfig(e *ExecutionContainer, cfgPath string) {
	btok, _ := json.MarshalIndent(e.GgsrunCfg, "", "\t")
	err := os.WriteFile(cfgPath, btok, 0600)
	if err != nil {
		pterm.Error.Printf("Could not securely write configuration to '%s'. %v\n", cfgPath, err)
	}
}

// dispResult : Display result flexibly supporting pure JSON or Graphical TUI
func (e *ExecutionContainer) dispResult(c *cli.Context) {
	e.SuccessStatus("Execution completed successfully.")

	if len(e.Msg) > 0 {
		e.FeedBackData.Response.Result.Message = e.Msg
	}

	if c.Bool("jsonparser") {
		var target interface{}
		if c.Bool("onlyresult") {
			target = e.FeedBackData.Response.Result.Result
		} else {
			target = e.FeedBackData.Response.Result
		}

		// 2. Inject config path if target is a JSON object
		b, err := json.Marshal(target)
		if err == nil {
			absPath, _ := filepath.Abs(e.resolveConfigFile())
			var m map[string]interface{}
			if err2 := json.Unmarshal(b, &m); err2 == nil {
				m["config_path"] = absPath
				dispRes, _ := json.MarshalIndent(m, "", "  ")
				fmt.Printf("%s\n", string(dispRes))
				return
			}
			var arr []interface{}
			if err2 := json.Unmarshal(b, &arr); err2 == nil {
				wrapped := map[string]interface{}{
					"result":      arr,
					"config_path": absPath,
				}
				dispRes, _ := json.MarshalIndent(wrapped, "", "  ")
				fmt.Printf("%s\n", string(dispRes))
				return
			}
			var val interface{}
			if err2 := json.Unmarshal(b, &val); err2 == nil {
				wrapped := map[string]interface{}{
					"result":      val,
					"config_path": absPath,
				}
				dispRes, _ := json.MarshalIndent(wrapped, "", "  ")
				fmt.Printf("%s\n", string(dispRes))
				return
			}
		}

		dispRes, _ := json.MarshalIndent(target, "", "  ")
		fmt.Printf("%s\n", string(dispRes))
		return
	}

	// Dynamic Graphical TUI Output
	res := e.FeedBackData.Response.Result

	pterm.Println()
	pterm.DefaultHeader.
		WithFullWidth().
		WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).
		WithTextStyle(pterm.NewStyle(pterm.FgBlack)).
		Println("ggsrun Execution Report")

	tableData := pterm.TableData{
		{"API Endpoint", pterm.Green(res.Uapi)},
		{"Execution Time", pterm.Yellow(fmt.Sprintf("%.3f sec", res.TotalEt))},
	}
	if res.TokenAuthUsed {
		tableData = append(tableData, []string{"Security", pterm.LightGreen("Authenticated (" + res.TokenAuthMsg + ")")})
	}
	pterm.DefaultTable.WithData(tableData).Render()

	if len(res.Message) > 0 {
		pterm.Println()
		pterm.DefaultSection.Println("Execution Logs")
		list := pterm.BulletListPrinter{}
		for _, m := range res.Message {
			list.Items = append(list.Items, pterm.BulletListItem{Level: 0, Text: m})
		}
		list.Render()
	}

	pterm.Println()
	pterm.DefaultSection.Println("Result Payload")

	targetRes := res.Result
	if targetRes == nil {
		pterm.Info.Println("No output payload returned.")
	} else {
		b, err := json.MarshalIndent(targetRes, "", "  ")
		if err == nil {
			panel := pterm.DefaultBox.WithTitle("JSON Result").Sprint(pterm.LightCyan(string(b)))
			pterm.Println(panel)
		} else {
			pterm.Error.Println("Failed to render result payload.")
		}
	}
}
