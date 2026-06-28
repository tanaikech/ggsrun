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
  let guardCodeTemplate = "";
  try {
    const templatePath = path.resolve(__dirname, "for_sandbox_gas.js");
    guardCodeTemplate = fs.readFileSync(templatePath, "utf8");
  } catch (err) {
    respondAndExit(
      "deny",
      "Failed to read sandbox guard template 'for_sandbox_gas.js': " + err.message
    );
  }

  const guardCode = guardCodeTemplate
    .replace(/__ALLOWED_FILE_IDS__/g, JSON.stringify(config.allowedFileIds))
    .replace(/__ALLOWED_FOLDER_IDS__/g, JSON.stringify(config.allowedFolderIds))
    .replace(/__ALLOWED_EMAILS__/g, JSON.stringify(config.allowedEmails))
    .replace(/__ALLOWED_CALENDAR_IDS__/g, JSON.stringify(config.allowedCalendarIds || []))
    .replace(/__ALLOWED_EVENT_IDS__/g, JSON.stringify(config.allowedEventIds || []))
    .replace(/__ALLOWED_URLS__/g, JSON.stringify(config.allowedUrls || []))
    .replace(/__BLOCKED_URLS__/g, JSON.stringify(config.blockedUrls || []));

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
