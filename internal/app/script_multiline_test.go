package app

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
)

// TestMultilineAndComplexScriptHandling tests that complex scripts with newlines,
// comments, regular expressions, escaped characters, and Unicode are preserved
// without corruption during payload packaging and JSON marshaling.
func TestMultilineAndComplexScriptHandling(t *testing.T) {
	complexScript := `function test(e) {
  // Single-line comment with 'quotes' and "double quotes" and https://example.com
  /* Multi-line comment
     with special characters: $ \ ' " ` + "`" + `
  */
  const urlRegex = /^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$/gi;
  const rawText = "Line 1\nLine 2\t\"Escaped Quotes\" and \\Backslashes\\";
  const formatted = ` + "`Processed: ${e.key}, Regex: ${urlRegex.test('https://google.com')}`" + `;

  return {
    message: rawText,
    formatted: formatted,
    input: e.key,
    unicode: "日本語テスト 🚀 特殊記号: ※§±"
  };
}`

	// 1. Verify JSON serialization round-trip in Project file payload
	origFile := File{
		Name:   "ggsrun_exe1_temp_20260815",
		Type:   "SERVER_JS",
		Source: complexScript,
	}
	proj := Project{
		Files: []File{origFile},
	}

	data, err := json.Marshal(proj)
	if err != nil {
		t.Fatalf("json.Marshal failed on complex script: %v", err)
	}

	var unmarshaledProj Project
	if err := json.Unmarshal(data, &unmarshaledProj); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(unmarshaledProj.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(unmarshaledProj.Files))
	}

	if unmarshaledProj.Files[0].Source != complexScript {
		t.Errorf("Script source was corrupted after JSON roundtrip.\nExpected:\n%s\nGot:\n%s", complexScript, unmarshaledProj.Files[0].Source)
	}

	// 2. Verify sandbox injection on complex multiline script
	sanitizedScript, guard, err := InjectSandbox(complexScript, "bypass")
	if err != nil {
		t.Fatalf("InjectSandbox bypass failed: %v", err)
	}
	if guard != "" {
		t.Errorf("Expected empty guard on bypass, got %s", guard)
	}
	if sanitizedScript != complexScript {
		t.Errorf("Script altered during bypass sandbox injection")
	}

	// 3. Verify JSON arguments decoding in Execution API parameter structure
	argJSON := `{"key":"value1","nested":{"arr":[1,2,3],"flag":true},"str":"Hello \"World\" \n \t \\"}`
	var parsedVal interface{}
	if err := json.Unmarshal([]byte(argJSON), &parsedVal); err != nil {
		t.Fatalf("json.Unmarshal failed for argument: %v", err)
	}

	epara := &e1para{
		Function:   "test",
		Parameters: []interface{}{parsedVal},
		DevMode:    true,
	}

	eparaBytes, err := json.Marshal(epara)
	if err != nil {
		t.Fatalf("json.Marshal for e1para failed: %v", err)
	}

	var parsedEpara e1para
	if err := json.Unmarshal(eparaBytes, &parsedEpara); err != nil {
		t.Fatalf("json.Unmarshal for e1para failed: %v", err)
	}

	if parsedEpara.Function != "test" {
		t.Errorf("Expected function 'test', got '%s'", parsedEpara.Function)
	}
	if len(parsedEpara.Parameters) != 1 {
		t.Fatalf("Expected 1 parameter, got %d", len(parsedEpara.Parameters))
	}
}

// TestPipedStdinSimulation verifies that io.ReadAll correctly captures
// multi-line text with indentations, newlines, and quotes intact.
func TestPipedStdinSimulation(t *testing.T) {
	multilineInput := `function test(e) {
  return 'ok ' + e.key;
}`
	reader := strings.NewReader(multilineInput)
	captured, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll failed: %v", err)
	}

	if string(captured) != multilineInput {
		t.Errorf("Expected stdin string %q, got %q", multilineInput, string(captured))
	}
}
