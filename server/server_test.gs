/**
 * @fileoverview Test suite for verifying the functionality of the refactored server.gs library.
 * These test cases cover basic connectivity (Beacon), core API execution logic (ExecutionApi),
 * and Web App execution logic (WebApps) under different security configurations.
 */

/**
 * Runs all unit tests to validate the ggsrun server implementation.
 * Outputs status logs to the console and throws an error if any check fails.
 */
function runTests() {
  console.log("Starting test suite...");

  // Mock global ggsrunif context so that evaluated scripts calling ggsrunif.Log() function correctly in local environment
  globalThis.ggsrunif = {
    Log: Log,
    Beacon: Beacon
  };

  testBeacon();
  testExecutionApi();
  testExecutionApiWithGgsrunif();
  testWebAppsWithPassword();
  testWebAppsWithoutPassword();
  testWebAppsIncorrectPassword();
  testFolderTreeMock();

  console.log("All tests completed successfully!");
}

/**
 * Validates that the Beacon function returns a valid status string containing version metadata.
 */
function testBeacon() {
  console.log("Running testBeacon...");
  const result = Beacon();
  console.log("Beacon Result:", result);
  if (result.indexOf("Version") === -1) {
    throw new Error("testBeacon failed: version information missing");
  }
}

/**
 * Tests script evaluation through the Execution API pathway.
 * Validates that code execution outputs match, and global logging captures outputs correctly.
 */
function testExecutionApi() {
  console.log("Running testExecutionApi...");
  const payload = {
    com: JSON.stringify("(function() { ggsrunif.Log('API test log'); return 42; })()"),
    exefunc: "main",
    log: false
  };

  const resultObj = ExecutionApi(JSON.stringify(payload));
  console.log("ExecutionApi Result:", resultObj);
  if (resultObj.result !== 42) {
    throw new Error("testExecutionApi failed: unexpected return value");
  }
  if (!resultObj.logger || resultObj.logger[0] !== "API test log") {
    throw new Error("testExecutionApi failed: logging not captured");
  }
}

/**
 * Tests script evaluation using the global namespace alias "ggsrunif".
 * Validates that evaluated scripts calling ggsrunif.Beacon() resolve successfully.
 */
function testExecutionApiWithGgsrunif() {
  console.log("Running testExecutionApiWithGgsrunif...");
  const payload = {
    com: JSON.stringify("(function() { return ggsrunif.Beacon(); })()"),
    exefunc: "main",
    log: false
  };

  const resultObj = ExecutionApi(JSON.stringify(payload));
  console.log("ExecutionApi (ggsrunif.Beacon) Result:", resultObj);
  if (resultObj.result.indexOf("Version") === -1) {
    throw new Error("testExecutionApiWithGgsrunif failed: version information missing or call failed");
  }
}

/**
 * Tests Web App script evaluation when a password is required and the client supplies the correct one.
 */
function testWebAppsWithPassword() {
  console.log("Running testWebAppsWithPassword...");
  const mockEvent = {
    contentLength: 100,
    parameters: {
      com: [JSON.stringify("(function() { ggsrunif.Log('WebApps with password log'); return 'hello'; })()")],
      pass: ["my_secret_pass"],
      log: ["true"] // do not record spreadsheet log
    }
  };

  const response = WebApps(mockEvent, "my_secret_pass");
  const resultObj = JSON.parse(response.getContent());
  console.log("WebApps (with pass) Result:", resultObj);
  if (resultObj.result !== "hello") {
    throw new Error("testWebAppsWithPassword failed: unexpected result");
  }
  if (!resultObj.logger || resultObj.logger[0] !== "WebApps with password log") {
    throw new Error("testWebAppsWithPassword failed: logging not captured");
  }
}

/**
 * Tests Web App script evaluation when the server has no password configured.
 * Validates that execution is allowed without requiring a password parameter.
 */
function testWebAppsWithoutPassword() {
  console.log("Running testWebAppsWithoutPassword...");
  const mockEvent = {
    contentLength: 100,
    parameters: {
      com: [JSON.stringify("(function() { return 'no pass'; })()")],
      log: ["true"] // do not record spreadsheet log
    }
  };

  const response = WebApps(mockEvent); // No password passed to server library
  const resultObj = JSON.parse(response.getContent());
  console.log("WebApps (without pass) Result:", resultObj);
  if (resultObj.result !== "no pass") {
    throw new Error("testWebAppsWithoutPassword failed: unexpected result");
  }
}

/**
 * Tests Web App script evaluation when a password is required but the client supplies an incorrect one.
 * Validates that the server returns a "Bad password" error message.
 */
function testWebAppsIncorrectPassword() {
  console.log("Running testWebAppsIncorrectPassword...");
  const mockEvent = {
    contentLength: 100,
    parameters: {
      com: [JSON.stringify("(function() { return 'secret'; })()")],
      pass: ["wrong_pass"],
      log: ["true"]
    }
  };

  const response = WebApps(mockEvent, "correct_pass");
  const resultObj = JSON.parse(response.getContent());
  console.log("WebApps (wrong pass) Result:", resultObj);
  if (resultObj.result !== "Error on GAS side: Bad password.") {
    throw new Error("testWebAppsIncorrectPassword failed: did not catch incorrect password");
  }
}

/**
 * Verifies that the ggsrun class constructor and methods (e.g. foldertree) are correctly defined.
 */
function testFolderTreeMock() {
  console.log("Running testFolderTreeMock...");
  const instance = new ggsrun(null, null, []);
  if (typeof instance.foldertree !== 'function') {
    throw new Error("testFolderTreeMock failed: foldertree method is not defined");
  }
}
