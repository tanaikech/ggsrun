/**
 * ggsrun Manual Sandbox Test Script
 */

function testUrl(url, expectedSuccess) {
  try {
    const res = UrlFetchApp.fetch(url);
    if (expectedSuccess) {
      return "[PASS] URL connection succeeded (Expected): " + url;
    } else {
      return "[FAIL] URL connection succeeded (Should be blocked): " + url;
    }
  } catch (e) {
    if (!expectedSuccess && e.message.includes("Sandbox Runtime Blocked")) {
      return "[PASS] URL connection blocked (Expected): " + url + " -> " + e.message;
    } else {
      return "[FAIL] URL connection failed (Unexpected): " + url + " -> " + e.message;
    }
  }
}

function testDrive(fileId, expectedSuccess) {
  try {
    const file = DriveApp.getFileById(fileId);
    if (expectedSuccess) {
      return "[PASS] Drive file retrieval succeeded (Expected): " + fileId;
    } else {
      return "[FAIL] Drive file retrieval succeeded (Should be blocked): " + fileId;
    }
  } catch (e) {
    if (!expectedSuccess && e.message.includes("Sandbox Runtime Blocked")) {
      return "[PASS] Drive file retrieval blocked (Expected): " + fileId + " -> " + e.message;
    } else if (expectedSuccess && (e.message.includes("found") || e.message.includes("permission") || e.message.includes("Could not find"))) {
      return "[PASS] Drive file retrieval bypassed wrapper successfully (Expected API error since dummy ID is used): " + fileId + " -> " + e.message;
    } else {
      return "[FAIL] Drive file retrieval failed (Unexpected): " + fileId + " -> " + e.message + "\nStack: " + e.stack;
    }
  }
}

function testMail(email, expectedSuccess) {
  try {
    // Attempt to invoke sendEmail. We catch both expected validation block and scope errors.
    MailApp.sendEmail(email, "Test Subject", "Test Body");
    if (expectedSuccess) {
      return "[PASS] Mail send API call succeeded (Expected): " + email;
    } else {
      return "[FAIL] Mail send API call succeeded (Should be blocked): " + email;
    }
  } catch (e) {
    if (!expectedSuccess && e.message.includes("Sandbox Runtime Blocked")) {
      return "[PASS] Mail send blocked (Expected): " + email + " -> " + e.message;
    } else if (expectedSuccess && e.message.includes("sendEmail") && (e.message.includes("permission") || e.message.includes("scope") || e.message.includes("authorization"))) {
      return "[PASS] Mail send bypassed wrapper successfully (Expected API scope authorization error): " + email + " -> " + e.message;
    } else {
      return "[FAIL] Mail send failed (Unexpected): " + email + " -> " + e.message;
    }
  }
}

function main() {
  const results = [];

  results.push("--- UrlFetchApp Tests ---");
  // 1. Whitelisted URL pattern (wildcard matches, should bypass wrapper)
  results.push(testUrl("https://httpbin.org/anything/allowed", true));
  // 2. Explicitly blacklisted URL (should be blocked)
  results.push(testUrl("https://httpbin.org/anything/blocked", false));
  // 3. Unregistered URL (default block all policy, should be blocked)
  results.push(testUrl("https://example.com/unregistered", false));

  results.push("--- DriveApp Tests ---");
  // 1. Whitelisted File ID (should bypass wrapper)
  results.push(testDrive("1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P", true));
  // 2. Unregistered File ID (should be blocked)
  results.push(testDrive("9X8Y7Z6W5V4U3T2S1R0Q9P8O7N6M5L4K", false));

  results.push("--- MailApp Tests ---");
  // 1. Whitelisted recipient email (should bypass wrapper)
  results.push(testMail("allowed-tester@example.com", true));
  // 2. Unregistered recipient email (should be blocked)
  results.push(testMail("blocked-tester@example.com", false));

  return results;
}
