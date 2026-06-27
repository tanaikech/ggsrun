# ggsrun - Onboarding, Setup, and Authentication Guide

This guide provides a comprehensive, step-by-step walkthrough to configure your environment, authenticate your local machine with Google APIs, link Google Cloud with Google Apps Script (GAS), and set up the execution server.

---

## Table of Contents
1. [Overview](#1-overview)
2. [Setup Method A: Simplified Quick Onboarding (Recommended)](#2-setup-method-a-simplified-quick-onboarding-recommended)
3. [Setup Method B: Traditional Manual Setup (Fallback)](#3-setup-method-b-traditional-manual-setup-fallback)
4. [Step 3: Link Your Google Cloud Project to Google Apps Script](#4-step-3-link-your-google-cloud-project-to-google-apps-script)
5. [Step 4: Set Up the Execution Server on Google Apps Script](#5-step-4-set-up-the-execution-server-on-google-apps-script)
   - [4.1 Add the Shared Server Library](#41-add-the-shared-server-library)
   - [4.2 Add the Gateway Code](#42-add-the-gateway-code)
   - [4.3 Deploy as an API Executable](#43-deploy-as-an-api-executable)
   - [4.4 Deploy as a Web App](#44-deploy-as-a-web-app)
6. [Advanced Configurations (Scopes Modification)](#6-advanced-configurations-scopes-modification)

---

## 1. Overview

`ggsrun` interacts directly with Google Drive and Google Apps Script APIs. To do this, your local binary must be authorized via OAuth2. `ggsrun` supports two main configuration paths:
* **Method A (Quick Onboarding)**: Automates API enablement on Google Cloud and streamlines credential generation.
* **Method B (Traditional Manual)**: Involves manually enabling APIs and creating Desktop client credentials inside the Google Cloud Console.

---

## 2. Setup Method A: Simplified Quick Onboarding (Recommended)

This method automates Google Cloud API configuration using tailored **GCP Quick Flows**, saving you many manual configuration steps.

1. **Initiate Onboarding**:
   Open your terminal, navigate to your workspace folder, and run:
   ```bash
   $ ggsrun setup
   ```
2. **Automated API Enablement**:
   - `ggsrun` will prompt you to open your browser to a tailored GCP Quick Flow link. Press **Enter** or **Y** to proceed.
   - This link automatically configures your Google Cloud project and enables all six required APIs instantly: **Google Drive API**, **Google Apps Script API**, **Google Sheets API**, **Gmail API**, **Google Slides API**, and **Google Docs API**.
   - Once enablement completes, the GCP Console will redirect you directly to the **Credentials Creation Screen**.
3. **Generate OAuth Credentials**:
   - On the GCP Console credentials screen, click **+ CREATE CREDENTIALS** > **OAuth client ID**.
   - Under **Application type**, choose **Desktop app**.
   - Name it `ggsrun Client` and click **Create**.
   - Click the download icon (JSON format) to save your credentials file.
4. **Register Credentials in ggsrun**:
   - Return to your terminal. `ggsrun` will ask how you want to register the credentials.
   - You can choose to provide the absolute file path to the downloaded JSON file (e.g., `/path/to/credentials.json`) or copy-paste the Client ID and Client Secret manually.
   - *Note: Unlike legacy versions, you do NOT need to rename the file to `client_secret.json`.*
5. **Grant OAuth Authorization**:
   - `ggsrun` will spawn a local loopback server and open your default browser.
   - Log in with your Google account.
   - If a warning stating *"Google hasn't verified this app"* appears, click **Advanced** > **Go to ggsrun Client (unsafe)** to proceed.
   - Click **Allow** to grant permissions.
   - The terminal will confirm success and securely write the configuration to `ggsrun.cfg`.
6. **Pre-fill Defaults (Optional)**:
   - `ggsrun` will prompt you to enter your **GAS Project Script ID** and **Web Apps URL** (optional). Pre-filling these settings saves you from passing the `-i` (script ID) or `-u` (Web App URL) flags in future commands.

---

## 3. Setup Method B: Traditional Manual Setup (Fallback)

If you prefer to configure Google Cloud manually, or have an existing GCP project, follow these steps:

### B.1 Google Cloud Platform (GCP) Configuration
1. Open the [Google Cloud Console](https://console.cloud.google.com/).
2. Select your project or click the dropdown at the top left > **New Project** to create one.
3. Go to **APIs & Services > Library**. Enable both the **Google Drive API** and **Google Apps Script API**.
4. Configure the **OAuth Consent Screen**:
   - Choose **External** user type and click **Create**.
   - Enter your App name (e.g., `ggsrun CLI`) and user support email.
   - **CRITICAL**: Scroll down to **Test users**, click **+ ADD USERS**, and enter your Gmail address. Only registered test users can authenticate unverified OAuth apps.
5. Generate Credentials:
   - Navigate to **APIs & Services > Credentials**.
   - Click **+ CREATE CREDENTIALS** > **OAuth client ID**.
   - Set the Application type to **Desktop app** and click **Create**.
   - Download the generated JSON credentials file, place it in your working directory, and rename it to exactly `client_secret.json`.

### B.2 Run Authorization Command
Execute the legacy authentication loop:
```bash
$ ggsrun auth
```
A browser tab will open. Complete the OAuth consent flow to authorize the client and generate `ggsrun.cfg`.

---

## 4. Step 3: Link Your Google Cloud Project to Google Apps Script

To invoke your script files through the Google Apps Script API, you must link your specific Google Apps Script project with your Google Cloud Project.

### 3.1 Retrieve Your GCP Project Number
1. Go to your [Google Cloud Console Dashboard](https://console.cloud.google.com/).
2. In the **Project Info** card at the top left, copy the **Project number** (e.g., `123456789012`). Do not confuse this with the alphanumeric *Project ID*.

### 3.2 Link Project in Google Apps Script Settings
1. Open your Apps Script project in the [Google Apps Script Editor](https://script.google.com/).
2. On the left sidebar, click the gear icon (**Project Settings** ⚙️).
3. Scroll down to the **Google Cloud Platform (GCP) Project** section and click **Change project**.
4. Paste your **Project number** into the text field and click **Set project**.

---

## 5. Step 4: Set Up the Execution Server on Google Apps Script

To execute stateless scripts remotely (`exe2` and `webapps` modes), you must establish a gateway server-side script using the official `ggsrunif` library.

### 4.1 Add the Shared Server Library
1. In the Apps Script Editor left sidebar, click the `+` icon next to **Libraries**.
2. Paste the following official `ggsrunif` Library ID:
   ```text
   115-19njNHlbT-NI0hMPDnVO1sdrw2tJKCAJgOTIAPbi_jq3tOo4lVRov
   ```
3. Click **Look up**.
4. Keep the default Identifier named `ggsrunif`, select the **latest version** from the version dropdown, and click **Add**.

### 4.2 Add the Gateway Code
Open your `Code.gs` script file and replace its contents with the following gateway wrapper endpoints:
```javascript
const doPost = (e) => ggsrunif.WebApps(e, "pass1");
const ExecutionApi = (e) => ggsrunif.ExecutionApi(e);
```
> [!NOTE]
> Change `"pass1"` to a secure custom password if you plan to execute webapps anonymously.

### 4.3 Deploy as an API Executable (For `exe1` and `exe2`)
1. At the top-right corner of the editor, click **Deploy > New Deployment**.
2. Click the gear icon next to "Select type" and choose **API Executable**.
3. In the description, type `ggsrun API Gateway`.
4. Set **Who has access** strictly to **Only myself** to keep your execution endpoint private.
5. Click **Deploy**.
6. Copy the **Script ID** (displayed on the deployment success screen or in Project Settings) to use with the `-i` flag.

### 4.4 Deploy as a Web App (For `webapps`)
1. At the top-right corner of the editor, click **Deploy > New Deployment**.
2. Click the gear icon next to "Select type" and choose **Web app**.
3. In the description, type `ggsrun Web Gateway`.
4. Set **Execute as** to **Me** (your email address).
5. Set **Who has access** depending on your target pipeline:
   - **Only myself** (Recommended): Highly secure setting. Requires `ggsrun` to be authenticated via OAuth2 with standard Google Drive scopes.
   - **Anyone**: Allows triggering the Web App anonymously from a remote environment (e.g., CI/CD) without OAuth tokens. Access is secured by the password parameter specified in `doPost` (passed via `-p`).
6. Click **Deploy**.
7. Copy the generated **Web app URL** (e.g., `https://script.google.com/macros/s/[WEB_APP_ID]/exec`) for the `-u` flag.

---

## 6. Advanced Configurations (Scopes Modification)

By default, `ggsrun` requests extensive permissions to handle file transfers and execution scopes. If you want to restrict or customize these scopes:
1. Open the `ggsrun.cfg` file generated in your configuration folder (located in `~/.config/` or specified by `GGSRUN_CFG_PATH`).
2. Locate the `"scopes"` JSON array:
   ```json
   "scopes": [
     "https://www.googleapis.com/auth/drive",
     "https://www.googleapis.com/auth/drive.readonly",
     ...
   ]
   ```
3. Add or remove scopes as needed.
4. Save the file and re-run:
   ```bash
   $ ggsrun auth
   ```
   The browser will open to prompt a new consent screen matching your exact custom scope definitions.
