// Package main (handler_execute.go) :
// Handles Google Apps Script Execution API and Web Apps integrations.
package main

import (
	"fmt"
	"os"
	"regexp"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/urfave/cli"
)

// exeAPIWithout : exe1
// Update project and Execution API withour server script.
func exeAPIWithout(c *cli.Context) error {
	a := defAuthContainer(c)
	if !c.Bool("jsonparser") {
		a.Spinner, _ = pterm.DefaultSpinner.WithRemoveWhenDone(false).Start("Initializing ggsrun...")
	}
	e := a.ggsrunIni(c).
		goauth().
		defExecutionContainer()

	e.exe1Function(c).
		executionAPIwithoutServer(c).
		esenderForExe1(c).
		dispResult(c)
	return nil
}

// exeAPIWith : exe2
// No update project. Only execute GAS using Execution API with server script.
func exeAPIWith(c *cli.Context) error {
	a := defAuthContainer(c)
	if !c.Bool("jsonparser") {
		a.Spinner, _ = pterm.DefaultSpinner.WithRemoveWhenDone(false).Start("Initializing ggsrun...")
	}
	e := a.ggsrunIni(c).
		goauth().
		defExecutionContainer()

	e.exe2Function(c).
		dispResult(c)
	return nil
}

// webAppsWith : exe3
// Executes GAS script via Web Apps, with token authentication support for 'Only myself' security.
func webAppsWith(c *cli.Context) error {
	a := defAuthContainer(c)
	if !c.Bool("jsonparser") {
		a.Spinner, _ = pterm.DefaultSpinner.WithRemoveWhenDone(false).Start("Initializing Web Apps execution...")
	}
	// Try loading auth gracefully. If tokens/scopes exist, it secures the Web App call.
	a.tryLoadAuth()
	e := a.defExecutionContainer()

	e.UpdateStatus("Resolving script payload...")
	var rawScript string
	if c.String("stringscript") != "" {
		rawScript = c.String("stringscript")
	} else {
		scriptFile := c.String("scriptfile")
		if scriptFile == "" {
			e.FailStatus("Initialization failed")
			pterm.Error.Println("No script. Please set GAS script using '-s' or '--stringscript'.")
			os.Exit(1)
		}
		b, err := os.ReadFile(scriptFile)
		if err != nil {
			e.FailStatus("Initialization failed")
			pterm.Error.Printf("Failed to read script file: %v\n", err)
			os.Exit(1)
		}
		rawScript = string(b)
	}

	val := c.String("value")
	var argStr string
	if val != "" {
		// Detect numeric, boolean, null, array, or objects to prevent string wrapping
		if regexp.MustCompile(`^[+-]?[0-9]*[\.]?[0-9]+$`).MatchString(val) ||
			regexp.MustCompile(`^\[.*\]$`).MatchString(val) ||
			regexp.MustCompile(`^{.*}$`).MatchString(val) ||
			val == "true" || val == "false" || val == "null" {
			argStr = val
		} else {
			argStr = fmt.Sprintf("%q", val) // Wrap strings securely
		}
	}

	// Double-Eval prevention logic ported to Web Apps mode
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
	return nil
}

// dispResult : Display result flexibly supporting pure JSON or Graphical TUI
func (e *ExecutionContainer) dispResult(c *cli.Context) {
	e.SuccessStatus("Execution completed successfully.")

	if len(e.Msg) > 0 {
		e.FeedBackData.Response.Result.Message = e.Msg
	}

	if c.Bool("jsonparser") {
		var dispRes []byte
		if c.Bool("onlyresult") {
			dispRes, _ = json.MarshalIndent(e.FeedBackData.Response.Result.Result, "", "  ")
		} else {
			dispRes, _ = json.MarshalIndent(e.FeedBackData.Response.Result, "", "  ")
		}
		// Push pure JSON straight to standard output
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
