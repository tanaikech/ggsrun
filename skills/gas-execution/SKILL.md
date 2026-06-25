---
name: gas-execution
description: Guidelines for writing and executing Google Apps Script code using ggsrun and finding documentation via workspace-developer. Use when developing or running GAS applications.
---

# Google Apps Script Execution and Development Skill

Follow these guidelines when writing, reviewing, and executing Google Apps Script (GAS) applications:

## Script Execution via ggsrun
* **Execution Tool**: Always use the `exe1` tool of the `ggsrun` MCP server to synchronize local code and run the entry function in a single step.
* **Designing Return Values**: The execution response displays only the value returned by the `return` statement of the executed function. Design your script to return a meaningful value representing the execution result.
* **Detailed Logs**: You can return detailed logs or execution summaries as the final return string to provide context and results to the user.
* **JSON Serialization**: If returning structured data (such as objects, arrays, or status maps), use `JSON.stringify()` to serialize it before returning.

## Verifying GAS API Usage
* **Documentation Search**: When you are unsure about the methods, behaviors, or parameters of built-in GAS classes (e.g., `DriveApp`, `GmailApp`, `SpreadsheetApp`), query the `workspace-developer` MCP server to fetch detailed class references and usage examples.
