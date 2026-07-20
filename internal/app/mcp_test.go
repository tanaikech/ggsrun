package app

import (
	"bytes"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli"
)

// TestMCPServerToolsList verifies that the MCP Server's "tools/list" command
// successfully parses input and returns the correct list of tools, specifically
// checking for our new 'rawdata' option in the 'download' tool.
func TestMCPServerToolsList(t *testing.T) {
	// Create pipes to simulate stdin and stdout
	rIn, wIn, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	// Save original stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	// Redirect
	os.Stdin = rIn
	os.Stdout = wOut

	// Prepare json-rpc message for tools/list
	requestMsg := `{"jsonrpc": "2.0", "id": 42, "method": "tools/list"}` + "\n"
	_, err = wIn.Write([]byte(requestMsg))
	if err != nil {
		t.Fatalf("Failed to write to stdin: %v", err)
	}
	wIn.Close()

	// Use channel to read stdout asynchronously
	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		outChan <- buf.String()
	}()

	// Set up CLI context
	appObj := cli.NewApp()
	appObj.Version = "5.3.17"
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	cliCtx := cli.NewContext(appObj, set, nil)

	// Call the MCP server main loop. It will read the single line, respond, and exit when stdin closes.
	_ = runMCP(cliCtx)
	wOut.Close()

	// Wait for stdout content
	output := <-outChan

	// Verify JSON-RPC response
	if !strings.Contains(output, `"jsonrpc":"2.0"`) {
		t.Errorf("Expected JSON-RPC 2.0 response, but got: %s", output)
	}
	if !strings.Contains(output, `"id":42`) {
		t.Errorf("Expected request ID 42 to be returned, but got: %s", output)
	}

	// Verify download tool has 'rawdata' property
	if !strings.Contains(output, `"download"`) {
		t.Errorf("Expected 'download' tool to be defined")
	}
	if !strings.Contains(output, `"rawdata"`) {
		t.Errorf("Expected 'rawdata' parameter schema inside download tool")
	}

	// Verify upload tool has 'projectname' property
	if !strings.Contains(output, `"upload"`) {
		t.Errorf("Expected 'upload' tool to be defined")
	}
	if !strings.Contains(output, `"projectname"`) {
		t.Errorf("Expected 'projectname' parameter schema inside upload tool")
	}
}

// TestMCPServerToolsCall verifies that the MCP Server's "tools/call" command
// successfully executes tools via simulated JSON-RPC by compiling and calling a temp binary.
func TestMCPServerToolsCall(t *testing.T) {
	// Build the real ggsrun binary to a temporary path
	tempDir, err := os.MkdirTemp("", "ggsrun_mcp_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempBinary := filepath.Join(tempDir, "ggsrun")
	
	// Compile the real binary
	buildCmd := exec.Command("go", "build", "-o", tempBinary, "../..")
	var buildStderr bytes.Buffer
	buildCmd.Stderr = &buildStderr
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build ggsrun binary: %v\nStderr:\n%s", err, buildStderr.String())
	}

	// Create pipes to simulate stdin and stdout
	rIn, wIn, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	// Save original stdin/stdout and env
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldTestExe := os.Getenv("GGSRUN_TEST_EXE_PATH")
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		os.Setenv("GGSRUN_TEST_EXE_PATH", oldTestExe)
	}()

	// Redirect and configure
	os.Stdin = rIn
	os.Stdout = wOut
	os.Setenv("GGSRUN_TEST_EXE_PATH", tempBinary)

	// Prepare json-rpc message for tools/call (calling 'filelist' with an invalid ID parameter to simulate tool execution)
	requestMsg := `{"jsonrpc": "2.0", "id": 100, "method": "tools/call", "params": {"name": "filelist", "arguments": {"searchbyid": "invalid_id_test_string"}}}` + "\n"
	_, err = wIn.Write([]byte(requestMsg))
	if err != nil {
		t.Fatalf("Failed to write to stdin: %v", err)
	}
	wIn.Close()

	// Use channel to read stdout asynchronously
	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		outChan <- buf.String()
	}()

	// Set up CLI context
	appObj := cli.NewApp()
	appObj.Version = "5.3.17"
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	cliCtx := cli.NewContext(appObj, set, nil)

	// Call the MCP server main loop. It will execute, respond, and exit when stdin closes.
	_ = runMCP(cliCtx)
	wOut.Close()

	// Wait for stdout content
	output := <-outChan

	// Verify JSON-RPC response format and contents
	if !strings.Contains(output, `"jsonrpc":"2.0"`) {
		t.Errorf("Expected JSON-RPC 2.0 response, but got: %s", output)
	}
	if !strings.Contains(output, `"id":100`) {
		t.Errorf("Expected request ID 100 to be returned, but got: %s", output)
	}
	// The binary output under error/not found should still be returned in 'content' of the result
	if !strings.Contains(output, `"content"`) {
		t.Errorf("Expected 'content' field in tools/call response, but got: %s", output)
	}
}

// TestGASScriptStaticAnalysis verifies that the static safety analysis engine
// correctly identifies write operations vs read-only operations across Google APIs.
func TestGASScriptStaticAnalysis(t *testing.T) {
	writeScript := `
		function run() {
			var folder = DriveApp.createFolder("test");
			var file = DriveApp.createFile("data.txt", "content");
			MailApp.sendEmail("user@example.com", "Subject", "Body");
			SpreadsheetApp.getActiveSpreadsheet().appendRow([1, 2, 3]);
		}
	`
	readScript := `
		function run() {
			var file = DriveApp.getFileById("abc");
			var name = file.getName();
			var data = SpreadsheetApp.openById("xyz").getValues();
		}
	`

	reportWrite, hasWrite := analyzeGASScript(writeScript)
	if !hasWrite {
		t.Errorf("Expected writeScript to be flagged as containing write operations")
	}
	if !strings.Contains(reportWrite, ".createFolder") {
		t.Errorf("Expected report to mention .createFolder")
	}
	if !strings.Contains(reportWrite, ".sendEmail") {
		t.Errorf("Expected report to mention .sendEmail")
	}

	reportRead, hasReadWrite := analyzeGASScript(readScript)
	if hasReadWrite {
		t.Errorf("Expected readScript to be identified as read-only")
	}
	if !strings.Contains(reportRead, ".getFileById") {
		t.Errorf("Expected report to mention .getFileById read method")
	}
}

// TestMCPServerExe1SecurityGuardrail verifies that the MCP Server's "exe1" tool
// blocks execution and returns a safety report when 'confirm' is omitted or false.
func TestMCPServerExe1SecurityGuardrail(t *testing.T) {
	// Create pipes to simulate stdin/stdout
	rIn, wIn, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe error: %v", err)
	}
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe error: %v", err)
	}

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()
	os.Stdin = rIn
	os.Stdout = wOut

	// Inline string script with write operations - properly escaped for insertion into JSON string
	stringScript := `function run() { MailApp.sendEmail(\"a@b.com\", \"hello\", \"world\"); }`
	
	// Create arguments map with confirm omitted/false
	requestMsg := `{"jsonrpc": "2.0", "id": 101, "method": "tools/call", "params": {"name": "exe1", "arguments": {"function": "run", "stringscript": "` + stringScript + `", "confirm": false}}}` + "\n"
	_, _ = wIn.Write([]byte(requestMsg))
	wIn.Close()

	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		outChan <- buf.String()
	}()

	appObj := cli.NewApp()
	appObj.Version = "5.3.17"
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	cliCtx := cli.NewContext(appObj, set, nil)

	_ = runMCP(cliCtx)
	wOut.Close()

	output := <-outChan

	// Verify that safety warning and confirmation instruction were returned
	if !strings.Contains(output, "SECURITY WARNING") {
		t.Errorf("Expected security warning in response when confirm is false, but got: %s", output)
	}
	if !strings.Contains(output, `\"confirm\": true`) {
		t.Errorf("Expected prompt to run with confirm: true, but got: %s", output)
	}
}

