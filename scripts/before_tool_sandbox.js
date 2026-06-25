// scripts/before_tool_sandbox.js
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

// Automatically install dependencies if node_modules doesn't exist
const nodeModulesPath = path.resolve(__dirname, "../node_modules");
if (!fs.existsSync(nodeModulesPath)) {
  try {
    const pluginDir = path.resolve(__dirname, "..");
    console.error("[ggsrun-plugin] Installing dependencies (npm install)...");
    execSync("npm install --production", { cwd: pluginDir, stdio: "ignore" });
    console.error("[ggsrun-plugin] Dependencies installed successfully.");
  } catch (err) {
    console.error("[ggsrun-plugin] Failed to install dependencies: " + err.message);
    process.exit(1);
  }
}

const acorn = require("acorn");
const walk = require("acorn-walk");

function readStdin() {
  return new Promise((resolve) => {
    let data = "";
    process.stdin.on("data", (chunk) => {
      data += chunk;
    });
    process.stdin.on("end", () => resolve(data));
  });
}

function respondAndExit(decision, reason) {
  const response = { decision, reason };
  process.stdout.write(JSON.stringify(response) + "\n");
  process.exit(0);
}

async function main() {
  const rawInput = await readStdin();
  if (!rawInput) respondAndExit("allow", "No input from stdin.");

  let payload;
  try {
    payload = JSON.parse(rawInput);
  } catch (e) {
    respondAndExit("deny", "Failed to parse hook input JSON: " + e.message);
  }

  const toolCall = payload.toolCall;
  if (!toolCall) respondAndExit("allow", "No toolCall in payload.");

  const isDirectExe1 = toolCall.name === "exe1";
  const isMcpExe1 =
    toolCall.name === "call_mcp_tool" &&
    toolCall.args &&
    toolCall.args.ToolName === "exe1";

  if (!isDirectExe1 && !isMcpExe1) {
    respondAndExit("allow", "Not target tool.");
  }

  let args = {};
  if (isDirectExe1) {
    args = toolCall.args || {};
  } else if (isMcpExe1) {
    const rawArgs = toolCall.args.Arguments;
    if (typeof rawArgs === "string") {
      try {
        args = JSON.parse(rawArgs);
      } catch (e) {
        respondAndExit(
          "deny",
          "Failed to parse call_mcp_tool Arguments: " + e.message,
        );
      }
    } else {
      args = rawArgs || {};
    }
  }

  const scriptfile = args.scriptfile;

  if (!scriptfile) {
    respondAndExit(
      "deny",
      "Security Restriction: Direct inline execution via 'stringscript' is disabled. Please write code to a local file and execute via 'scriptfile'.",
    );
  }

  const absoluteScriptPath = path.isAbsolute(scriptfile)
    ? scriptfile
    : path.resolve(process.cwd(), scriptfile);

  if (!fs.existsSync(absoluteScriptPath)) {
    respondAndExit("deny", `Script file not found: ${scriptfile}`);
  }

  let code = fs.readFileSync(absoluteScriptPath, "utf8");
  let hasEval = false;
  let hasObfuscatedAccess = false;

  // 1. AST obfuscation & eval detection
  try {
    const ast = acorn.parse(code, { ecmaVersion: 2020, sourceType: "script" });
    walk.simple(ast, {
      CallExpression(node) {
        if (node.callee.type === "Identifier" && node.callee.name === "eval") {
          hasEval = true;
        }
      },
      MemberExpression(node) {
        if (
          node.computed &&
          node.property.type !== "Literal" &&
          node.property.type !== "Identifier"
        ) {
          hasObfuscatedAccess = true;
        }
      },
    });
  } catch (parseError) {
    respondAndExit(
      "deny",
      "AST Validation Failed: Syntax error. " + parseError.message,
    );
  }

  if (hasEval)
    respondAndExit("deny", "Security Restriction: 'eval' usage is prohibited.");
  if (hasObfuscatedAccess)
    respondAndExit(
      "deny",
      "Security Restriction: Obfuscated or dynamic API member access is prohibited.",
    );

  // 2. Load whitelist configuration
  let baseDir = process.cwd();
  if (payload && payload.workspacePaths && payload.workspacePaths.length > 0) {
    baseDir = payload.workspacePaths[0];
  } else if (scriptfile) {
    baseDir = path.dirname(absoluteScriptPath);
  }
  const configPath = path.resolve(baseDir, "sandbox_config.json");
  let config;
  if (!fs.existsSync(configPath)) {
    const defaultConfig = {
      allowedFileIds: ["DEFAULT_TEST_SPREADSHEET_ID"],
      allowedFolderIds: ["DEFAULT_TEST_FOLDER_ID"],
      allowedEmails: ["sandbox-tester@example.com"],
      allowedCalendarIds: ["primary"],
      allowedEventIds: [],
      allowedUrls: ["https://httpbin.org/anything"]
    };
    try {
      fs.writeFileSync(configPath, JSON.stringify(defaultConfig, null, 2), "utf8");
      respondAndExit(
        "deny",
        `Security Notice: 'sandbox_config.json' was not found in the current working directory.
A template file has been created at '${configPath}'.

=== sandbox_config.json Detailed Explanation ===
This configuration file acts as a security whitelist for Google Apps Script execution.
Please edit the file and add your authorized resource IDs, emails, and URLs:
1. 'allowedFileIds': Array of Google Drive File/Spreadsheet IDs that the script is permitted to open.
2. 'allowedFolderIds': Array of Google Drive Folder IDs that the script is permitted to access.
3. 'allowedEmails': Array of recipient email addresses that GmailApp/MailApp can send emails to.
4. 'allowedCalendarIds': Array of Calendar IDs (e.g., 'primary') that the script can access.
5. 'allowedEventIds': Array of Calendar Event IDs that the script can access.
6. 'allowedUrls': Array of URL prefixes (e.g., 'https://api.github.com/') that UrlFetchApp is allowed to request.

Please review and configure this whitelist before running the script again.`
      );
    } catch (createErr) {
      respondAndExit(
        "deny",
        `Security Notice: 'sandbox_config.json' was not found in the current working directory, and failed to create a default configuration: ` + createErr.message
      );
    }
  }

  try {
    config = JSON.parse(fs.readFileSync(configPath, "utf8"));
  } catch (e) {
    respondAndExit("deny", "Failed to read sandbox_config.json: " + e.message);
  }

  // 3. Runtime Proxy Guard Code
  const guardCode = `// === SANDBOX SECURITY GUARD INJECTED ===
const DriveApp = (function() {
  const original = this.DriveApp || DriveApp;
  if (!original) return null;
  const allowedFileIds = ${JSON.stringify(config.allowedFileIds)};
  const allowedFolderIds = ${JSON.stringify(config.allowedFolderIds)};

  function wrapIterator(iter) {
    return {
      hasNext: function() { return iter.hasNext(); },
      next: function() {
        const item = iter.next();
        const id = item.getId();
        if (!allowedFileIds.includes(id) && !allowedFolderIds.includes(id)) {
          throw new Error("Sandbox Runtime Blocked: Accessed resource ID '" + id + "' is not whitelisted.");
        }
        return item;
      }
    };
  }

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'getFileById') {
        return function(id) {
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: File ID '" + id + "' is not in the whitelist.");
          }
          return original.getFileById(id);
        };
      }
      if (prop === 'getFolderById') {
        return function(id) {
          if (!allowedFolderIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Folder ID '" + id + "' is not in the whitelist.");
          }
          return original.getFolderById(id);
        };
      }
      if (prop === 'getRootFolder') {
        return function() {
          throw new Error("Sandbox Runtime Blocked: Direct access to Root Folder is prohibited.");
        };
      }
      if (prop === 'getFiles' || prop === 'searchFiles' || prop === 'getFilesByName') {
        return function(...args) {
          const rawIter = original[prop].apply(original, args);
          return wrapIterator(rawIter);
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'DriveApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const GmailApp = (function() {
  const original = this.GmailApp || GmailApp;
  if (!original) return null;
  const allowedEmails = ${JSON.stringify(config.allowedEmails)};

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'sendEmail' || prop === 'createDraft') {
        return function(recipient, ...args) {
          if (!allowedEmails.includes(recipient)) {
            throw new Error("Sandbox Runtime Blocked: Recipient address '" + recipient + "' is not whitelisted.");
          }
          return original[prop].apply(original, [recipient, ...args]);
        };
      }
      if (prop === 'getInboxThreads' || prop === 'getSpamThreads' || prop === 'getTrashThreads' ||
          prop === 'search' || prop === 'getChatThreads' || prop === 'getStarredThreads') {
        return function() {
          throw new Error("Sandbox Runtime Blocked: Inbox scanning is prohibited.");
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'GmailApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const MailApp = (function() {
  const original = this.MailApp || MailApp;
  if (!original) return null;
  const allowedEmails = ${JSON.stringify(config.allowedEmails)};

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'sendEmail') {
        return function(recipient, ...args) {
          if (!allowedEmails.includes(recipient)) {
            throw new Error("Sandbox Runtime Blocked: Recipient address '" + recipient + "' is not whitelisted.");
          }
          return original.sendEmail(recipient, ...args);
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'MailApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const SpreadsheetApp = (function() {
  const original = this.SpreadsheetApp || SpreadsheetApp;
  if (!original) return null;
  const allowedFileIds = ${JSON.stringify(config.allowedFileIds)};

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'openById') {
        return function(id) {
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Spreadsheet ID '" + id + "' is not whitelisted.");
          }
          return original.openById(id);
        };
      }
      if (prop === 'openByUrl') {
        return function(url) {
          const match = url.match(/\\/d\\/([^\\/]+)/);
          const id = match ? match[1] : url;
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Spreadsheet URL '" + url + "' is not whitelisted.");
          }
          return original.openByUrl(url);
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'SpreadsheetApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const DocumentApp = (function() {
  const original = this.DocumentApp || DocumentApp;
  if (!original) return null;
  const allowedFileIds = ${JSON.stringify(config.allowedFileIds)};

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'openById') {
        return function(id) {
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Document ID '" + id + "' is not whitelisted.");
          }
          return original.openById(id);
        };
      }
      if (prop === 'openByUrl') {
        return function(url) {
          const match = url.match(/\\/d\\/([^\\/]+)/);
          const id = match ? match[1] : url;
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Document URL '" + url + "' is not whitelisted.");
          }
          return original.openByUrl(url);
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'DocumentApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const SlidesApp = (function() {
  const original = this.SlidesApp || SlidesApp;
  if (!original) return null;
  const allowedFileIds = ${JSON.stringify(config.allowedFileIds)};

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'openById') {
        return function(id) {
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Presentation ID '" + id + "' is not whitelisted.");
          }
          return original.openById(id);
        };
      }
      if (prop === 'openByUrl') {
        return function(url) {
          const match = url.match(/\\/d\\/([^\\/]+)/);
          const id = match ? match[1] : url;
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Presentation URL '" + url + "' is not whitelisted.");
          }
          return original.openByUrl(url);
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'SlidesApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const CalendarApp = (function() {
  const original = this.CalendarApp || CalendarApp;
  if (!original) return null;
  const allowedCalendarIds = ${JSON.stringify(config.allowedCalendarIds || [])};
  const allowedEventIds = ${JSON.stringify(config.allowedEventIds || [])};

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'getCalendarById') {
        return function(id) {
          if (!allowedCalendarIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Calendar ID '" + id + "' is not whitelisted.");
          }
          return original.getCalendarById(id);
        };
      }
      if (prop === 'getDefaultCalendar') {
        return function() {
          if (!allowedCalendarIds.includes('primary')) {
            throw new Error("Sandbox Runtime Blocked: Default Calendar is not whitelisted.");
          }
          return original.getDefaultCalendar();
        };
      }
      if (prop === 'getEventById') {
        return function(id) {
          if (!allowedEventIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Event ID '" + id + "' is not whitelisted.");
          }
          return original.getEventById(id);
        };
      }
      if (prop === 'getEventSeriesById') {
        return function(id) {
          if (!allowedEventIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Event Series ID '" + id + "' is not whitelisted.");
          }
          return original.getEventSeriesById(id);
        };
      }
      if (prop === 'getOwnedCalendars' || prop === 'getSharedCalendars' || prop === 'getCalendarsByName') {
        return function(...args) {
          const rawArr = original[prop].apply(original, args);
          return rawArr.filter(cal => allowedCalendarIds.includes(cal.getId()));
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'CalendarApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const UrlFetchApp = (function() {
  const original = this.UrlFetchApp || UrlFetchApp;
  if (!original) return null;
  const allowedUrls = ${JSON.stringify(config.allowedUrls || [])};

  function checkUrl(url) {
    const isAllowed = allowedUrls.some(allowed => url.startsWith(allowed));
    if (!isAllowed) {
      throw new Error("Sandbox Runtime Blocked: URL '" + url + "' is not whitelisted.");
    }
  }

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'fetch') {
        return function(url, ...args) {
          checkUrl(url);
          return original.fetch(url, ...args);
        };
      }
      if (prop === 'fetchAll') {
        return function(requests, ...args) {
          if (Array.isArray(requests)) {
            requests.forEach(req => {
              if (typeof req === 'string') {
                checkUrl(req);
              } else if (req && typeof req.url === 'string') {
                checkUrl(req.url);
              }
            });
          }
          return original.fetchAll(requests, ...args);
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'UrlFetchApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();
// === END OF SANDBOX SECURITY GUARD ===
`;

  // 4. Automatic file injection
  try {
    fs.writeFileSync(absoluteScriptPath, guardCode + "\n" + code, "utf8");
  } catch (writeErr) {
    respondAndExit(
      "deny",
      "Failed to inject sandbox guard: " + writeErr.message,
    );
  }

  respondAndExit(
    "allow",
    "Sandbox AST verification cleared. Security Proxy successfully injected.",
  );
}

main();
