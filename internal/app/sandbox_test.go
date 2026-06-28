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
	res, err := InjectSandbox("Logger.log('hello');", configPath)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !strings.Contains(res, "var _wrapped") {
		t.Errorf("expected wrapped script content, got: %s", res)
	}

	// Test 2: Bypass mode
	resBypass, err := InjectSandbox("Logger.log('hello');", "bypass")
	if err != nil {
		t.Errorf("expected no error in bypass mode, got: %v", err)
	}
	if resBypass != "Logger.log('hello');" {
		t.Errorf("expected unmodified script in bypass mode, got: %s", resBypass)
	}
}
