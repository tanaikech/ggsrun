package app

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjectSandbox(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "sandbox_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "sandbox_config.json")
	configContent := `{"allowedFileIds": ["test-id"]}`

	err = ioutil.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Test 1: Successful injection with valid config
	_, guard, err := InjectSandbox("Logger.log('hello');", configPath)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !strings.Contains(guard, "var _wrapped") {
		t.Errorf("expected wrapped guard content, got: %s", guard)
	}

	// Test 2: Bypass mode
	resBypass, guardBypass, err := InjectSandbox("Logger.log('hello');", "bypass")
	if err != nil {
		t.Errorf("expected no error in bypass mode, got: %v", err)
	}
	if resBypass != "Logger.log('hello');" {
		t.Errorf("expected unmodified script in bypass mode, got: %s", resBypass)
	}
	if guardBypass != "" {
		t.Errorf("expected empty guard in bypass mode, got: %s", guardBypass)
	}
}
