# ggsrun Server Library

This directory contains the Google Apps Script (GAS) server library and helper files used by the `ggsrun` client tool to execute local GAS scripts on Google Drive / Web Apps.

## Files

- [server.gs](file:///home/adsam/GitHub/ggsrun/server/server.gs): The core server library implementation.
- [server_test.gs](file:///home/adsam/GitHub/ggsrun/server/server_test.gs): Local GAS test suite to verify server functionality.

## Deployment Instructions

### 1. Deploying as a Web App

To use the server library with `ggsrun w` (Web Apps execution mode), deploy the server script as a Web App:

1. Create a new Google Apps Script project.
2. Add the `ggsrunif` library using its Library ID (or copy the contents of `server.gs` directly into your project).
3. Create a server script file (e.g., `webapps_server.js`) with the following entry points:

   ```javascript
   // Secure execution using Google OAuth2 Access Token (Recommended)
   const doPost = e => ggsrunif.WebApps(e);

   // Alternative: Execution with password verification
   // const doPost = e => ggsrunif.WebApps(e, "your_password");

   const ExecutionApi = e => ggsrunif.ExecutionApi(e);
   ```

4. Configure the manifest file `appsscript.json`:
   ```json
   {
     "timeZone": "Asia/Tokyo",
     "exceptionLogging": "STACKDRIVER",
     "runtimeVersion": "V8",
     "dependencies": {
       "libraries": [
         {
           "userSymbol": "ggsrunif",
           "libraryId": "115-19njNHlbT-NI0hMPDnVO1sdrw2tJKCAJgOTIAPbi_jq3tOo4lVRov",
           "version": "0",
           "developmentMode": true
         }
       ]
     },
     "webapp": {
       "executeAs": "USER_DEPLOYING",
       "access": "MYSELF"
     },
     "executionApi": {
       "access": "MYSELF"
     }
   }
   ```
5. Deploy the script as a Web App:
   - **Execute as**: `User deploying the web app` (yourself)
   - **Who has access**: `Only myself` (highly recommended for access-token authenticated execution) or `Anyone` (if using password-only verification).

### 2. Testing the Server

You can test the server functions locally within the Apps Script editor:

1. Add the contents of [server_test.gs](file:///home/adsam/GitHub/ggsrun/server/server_test.gs) to your project.
2. Select the `runTests` function from the dropdown menu in the editor.
3. Click **Run**.
4. Check the execution logs to verify that all tests pass.

### 3. Testing client-side ggsrun commands

To verify the correct execution of each execution mode from your local CLI client, you can run the following test commands:

```bash
ggsrun e1 -ss "const main = (_) => ggsrunif.Beacon();" -j
ggsrun e2 -ss "const main = (_) => ggsrunif.Beacon();" -j
ggsrun w -ss "const main = (_) => ggsrunif.Beacon();" -j
```

