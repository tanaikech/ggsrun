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

## Code Review & Security Checklist

When reviewing or before executing generated GAS code, you must strictly perform a security code review.

### Security Checklist

1. **Google Drive & Document Access**: Does the script read, write, or delete files/folders (`DriveApp`, `SpreadsheetApp`, `DocumentApp`, `SlidesApp`)? Verify that only authorized files/folders are modified or deleted.
2. **Gmail & Mailing**: Does the script read emails, drafts, or send messages (`GmailApp`, `MailApp`)? Check if recipient addresses and message contents are safe and authorized.
3. **Calendar Events**: Does the script read, write, or delete calendar events (`CalendarApp`)? Confirm that only designated Calendars and Events are manipulated.
4. **Outbound Network Connections**: Does the script fetch external resources (`UrlFetchApp`)? Validate the target URL to ensure no credentials or sensitive data are being exfiltrated to untrusted endpoints.
5. **Hardcoded Secrets**: Are there any hardcoded API keys, OAuth tokens, or passwords? Ensure no credentials are exposed in the source code.
6. **Destructive Actions**: Does the code perform any bulk or irreversible deletion/overwriting of data?

### How to Proceed with Execution

Before executing the script using `ggsrun`'s `exe1`:
- **Summarize Accessed Services**: Clearly show the user which Google APIs and external URLs the script will access.
- **Request Confirmation**: Prompt the user to confirm whether they want to proceed with execution (e.g., "Would you like to execute this script?").
