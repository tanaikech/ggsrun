// scripts/after_tool_cleanup.js
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

function readStdin() {
  return new Promise((resolve) => {
    let data = "";
    process.stdin.on("data", (chunk) => {
      data += chunk;
    });
    process.stdin.on("end", () => resolve(data));
  });
}

async function main() {
  const rawInput = await readStdin();
  if (!rawInput) {
    process.stdout.write("{}\n");
    process.exit(0);
  }

  let payload;
  try {
    payload = JSON.parse(rawInput);
  } catch (e) {
    process.stdout.write("{}\n");
    process.exit(0);
  }

  const toolCall = payload.toolCall;
  if (!toolCall) {
    process.stdout.write("{}\n");
    process.exit(0);
  }

  // 1. Determine if this is a direct exe1 call or call_mcp_tool with ToolName: exe1
  const isDirectExe1 = toolCall.name === "exe1";
  const isMcpExe1 =
    toolCall.name === "call_mcp_tool" &&
    toolCall.args &&
    toolCall.args.ToolName === "exe1";

  if (!isDirectExe1 && !isMcpExe1) {
    process.stdout.write("{}\n");
    process.exit(0);
  }

  // 2. Retrieve arguments using the same logic as before_tool_sandbox.js
  let args = {};
  if (isDirectExe1) {
    args = toolCall.args || {};
  } else if (isMcpExe1) {
    const rawArgs = toolCall.args.Arguments;
    if (typeof rawArgs === "string") {
      try {
        args = JSON.parse(rawArgs);
      } catch (e) {
        console.error(
          "[LOG] PostToolUse Cleanup skipped: Failed to parse call_mcp_tool Arguments: " +
            e.message,
        );
        process.stdout.write("{}\n");
        process.exit(0);
      }
    } else {
      args = rawArgs || {};
    }
  }

  const scriptfile = args.scriptfile;

  if (!scriptfile) {
    process.stdout.write("{}\n");
    process.exit(0);
  }

  const absoluteScriptPath = path.isAbsolute(scriptfile)
    ? scriptfile
    : path.resolve(process.cwd(), scriptfile);

  // 3. Remove the injected sandbox security guard code
  if (fs.existsSync(absoluteScriptPath)) {
    try {
      let code = fs.readFileSync(absoluteScriptPath, "utf8");
      const regex =
        /\/\/ === SANDBOX SECURITY GUARD INJECTED ===.*?\/\/ === END OF SANDBOX SECURITY GUARD ===\n?/s;

      if (regex.test(code)) {
        const cleanedCode = code.replace(regex, "").trimStart();
        fs.writeFileSync(absoluteScriptPath, cleanedCode, "utf8");
        console.error(
          "[LOG] PostToolUse: Sandbox Guard successfully removed from scriptfile.",
        );
      }
    } catch (err) {
      console.error("[LOG] Cleanup Error: " + err.message);
    }
  }

  process.stdout.write("{}\n");
  process.exit(0);
}

main();
