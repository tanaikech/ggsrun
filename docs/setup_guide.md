# ggsrun - Onboarding, Setup, and Authentication Guide

This guide provides a comprehensive, step-by-step walkthrough to configure your environment, install `ggsrun`, obtain Google Cloud Platform (GCP) credentials, authenticate your local machine, link GCP with Google Apps Script (GAS), deploy execution gateway endpoints, and verify end-to-end operation.

---

## Table of Contents
1. [Step 1: Install ggsrun on Your Computer](#step-1-install-ggsrun-on-your-computer)
   - [For Windows Users](#for-windows-users)
   - [For macOS & Linux Users](#for-macos--linux-users)
2. [Step 2: Choose Your Authorization Method](#step-2-choose-your-authorization-method)
   - [Option A: Simplified Quick Setup (Recommended)](#option-a-simplified-quick-setup-recommended)
   - [Option B: Traditional Manual Setup (Fallback)](#option-b-traditional-manual-setup-fallback)
3. [Step 3: Link Your Google Cloud Project to Google Apps Script](#step-3-link-your-google-cloud-project-to-google-apps-script)
   - [3.1 Get Your Project Number](#31-get-your-project-number)
   - [3.2 Link the Project in Your GAS Editor](#32-link-the-project-in-your-gas-editor)
4. [Step 4: Set Up the Execution Server on Google Apps Script](#step-4-set-up-the-execution-server-on-google-apps-script)
   - [4.1 Add the Shared Server Library](#41-add-the-shared-server-library)
   - [4.2 Add the Gateway Code](#42-add-the-gateway-code)
   - [4.3 Deploy the Script as an API Executable (For exe1 & exe2)](#43-deploy-the-script-as-an-api-executable-for-exe1--exe2)
   - [4.4 Deploy the Script as a Web App (For webapps)](#44-deploy-the-script-as-a-web-app-for-webapps)
5. [Step 5: Test Execution from Your Computer](#step-5-test-execution-from-your-computer)
   - [Test Option A: Execution via API Executable (exe2 mode)](#test-option-a-execution-via-api-executable-exe2-mode)
   - [Test Option B: Execution via Web App (webapps mode)](#test-option-b-execution-via-web-app-webapps-mode)
   - [Expected Output](#expected-output)
6. [Advanced Configurations (Modifying OAuth Scopes)](#6-advanced-configurations-modifying-oauth-scopes)
7. [Troubleshooting Diagnostics](#7-troubleshooting-diagnostics)
   - [1. Web Apps Returns Status Code 200, but output is HTML](#1-web-apps-returns-status-code-200-but-output-is-html)
   - [2. "Requested entity was not found" or 404 Errors](#2-requested-entity-was-not-found-or-404-errors)
   - [3. Headless Server Authentication](#3-headless-server-authentication)

---

## Step 1: Install ggsrun on Your Computer

`ggsrun` is a lightweight application that runs through your terminal (Command Prompt or PowerShell on Windows, or Terminal on macOS/Linux). You do not need to install complex development environments; you can simply download a ready-to-use version.

### For Windows Users

1. Navigate to the official [ggsrun Releases Page](https://github.com/tanaikech/ggsrun/releases).
2. Download the file named `ggsrun_windows_amd64.exe` (or `arm64` if using an ARM-based Windows device).
3. Rename the downloaded file to simply `ggsrun.exe`.
4. Move this file into a dedicated folder where you want to work (for example, create a new folder on your Desktop named `ggsrun-workspace`).

### For macOS & Linux Users

1. Navigate to the official [ggsrun Releases Page](https://github.com/tanaikech/ggsrun/releases).
2. Download the appropriate binary for your system:
   - **macOS (Intel)**: `ggsrun_darwin_amd64`
   - **macOS (Apple Silicon M1/M2/M3)**: `ggsrun_darwin_arm64`
   - **Linux**: `ggsrun_linux_amd64` (or matching CPU architecture, e.g., `linux_arm64`)
3. Rename the downloaded file to simply `ggsrun`.
4. Move the file to your working folder (e.g., a folder on your Desktop named `ggsrun-workspace`).
5. Open your Terminal, navigate to your working folder, and grant execution permissions to the binary:
   ```bash
   cd ~/Desktop/ggsrun-workspace
   chmod +x ggsrun
   ```

---

## Step 2: Choose Your Authorization Method

`ggsrun` provides two ways to authorize your local machine with Google Cloud. We strongly recommend **Option A: Simplified Quick Setup** as it saves you about 10 steps of manual configuration.

---

### Option A: Simplified Quick Setup (Recommended)
This method automates Google Cloud API configuration using tailored **GCP Quick Flows**, saving you many manual configuration steps. Alternatively, if you have the `gcloud` CLI installed and authenticated, `ggsrun` can automatically import your active credentials.

1. **Open your terminal** and navigate to your workspace folder where `ggsrun` is located:
   - **Windows Command Prompt**: `cd %USERPROFILE%\Desktop\ggsrun-workspace`
   - **macOS/Linux Terminal**: `cd ~/Desktop/ggsrun-workspace`
2. **Run the setup command**:
   - **Windows**: `ggsrun setup`
   - **macOS/Linux**: `./ggsrun setup`
3. **Follow Browser Instructions**:
   - `ggsrun` will ask to open your browser to a tailored GCP Quick Flow link. Choose **Y** (or press Enter) to proceed.
   - This link automatically configures your Google Cloud project and enables all seven required APIs instantly: **Google Drive API**, **Google Apps Script API**, **Google Sheets API**, **Gmail API**, **Google Slides API**, **Google Docs API**, and **Cloud Logging API**.
   - Once enablement completes, the GCP Console will redirect you straight to the **Create Credentials** page.
4. **Create Credentials**:
   - On the GCP Console, choose **+ CREATE CREDENTIALS** at the top > **OAuth client ID**.
   - Select **Desktop app** under "Application type". Name it `ggsrun Desktop Client` and click **Create**.
   - Download the JSON credential file to your computer.
     *(Note: It is NOT required to rename this file to "client_secret.json" for setup mode. You can leave it as its default downloaded name or save it to any path, such as `{your path}/{credential file name}.json`)*
5. **Register Credentials inside ggsrun**:
   - In your terminal, `ggsrun` will ask you how to load credentials.
   - If the `gcloud` CLI is detected on your system, you will see a third option:
     - **`[1] Provide path to downloaded client secret JSON file (Recommended)`**
     - **`[2] Enter Client ID and Client Secret manually`**
     - **`[3] Use active credentials from gcloud CLI (Auto-detected)`**
   - Choose **[3]** to automatically load your active GCP Project ID and access token. This option configures `ggsrun` to dynamically refresh tokens via the `gcloud` CLI, eliminating the need to download `client_secret.json` or manage `refresh_token`s.
   - If you choose **[1]**, paste the path to your downloaded JSON file.
   - If you choose **[2]**, paste the Client ID, Client Secret, and the GCP Project ID (optional, but required for log retrieval) manually.
6. **Launch Authorization**:
   - (If you chose Option 1 or 2) `ggsrun` will ask to launch the browser for authorization. Press **Y** to proceed.
   - Log in with your Google account. You may see a safety warning screen stating *"Google hasn't verified this app"*. Click **Advanced** and then click **Go to ggsrun Client (unsafe)** to proceed.
   - Click **Allow** to grant permission.
   - Your local environment is now fully authorized and securely saved to `ggsrun.cfg`!
7. **Configure Default Values (Optional)**:
   - Finally, `ggsrun` will prompt you to enter your **Google Apps Script Project Script ID** and your **Web Apps URL** (optional). Entering these now saves them in `ggsrun.cfg`, allowing you to run execution commands later without passing the `-i` or `-u` options every time!

---

### Option B: Traditional Manual Setup (Fallback)
If you already have a configured Google Cloud Project or prefer to manage everything manually:

#### B.1 Configure Your Google Cloud Project
1. Open your browser and navigate to the [Google Cloud Console](https://console.cloud.google.com/).
2. Log in, click the project dropdown menu at the top left, select **New Project**, and name it `My-ggsrun-Project`.
3. Enable **Google Drive API** and **Google Apps Script API** inside the API Library.
4. Set up the **OAuth consent screen** (choose External, fill in App info, and **CRITICAL**: add your own Gmail address as a **Test User**).
5. Navigate to **Credentials** > **+ CREATE CREDENTIALS** > **OAuth client ID**, choose **Desktop app**, click **Create**, download the JSON file, move it to your workspace, and rename it to exactly `client_secret.json`.

#### B.2 Perform the Automated Authorization
1. Execute in your workspace terminal:
   - **Windows**: `ggsrun auth`
   - **macOS/Linux**: `./ggsrun auth`
2. Google login will open automatically in your browser. Select your account, bypass the unverified warning (Click *Advanced* > *Go to ggsrun Client*), and click **Allow** to complete authorization and save `ggsrun.cfg`.

---

## Step 3: Link Your Google Cloud Project to Google Apps Script (Mandatory for both Methods)

No matter which setup option you used above, you must link your Apps Script project with your newly configured Google Cloud environment.

### 3.1 Get Your Project Number
1. Return to your [Google Cloud Console](https://console.cloud.google.com/).
2. Click on the **Dashboard** or **Welcome** page of your project.
3. Look for the **Project Info** card. Copy the **Project number** (this is a long string of numbers, e.g., `123456789012`). Do not confuse this with the alphanumeric *Project ID*.

### 3.2 Link the Project in Your GAS Editor
1. Go to the [Google Apps Script Dashboard](https://script.google.com/) and open the specific script project you wish to run, or create a **New Project**.
2. On the left sidebar of the modern GAS editor, click the gear icon (**Project Settings** ⚙️).
3. Scroll down to **Google Cloud Platform (GCP) Project** and click **Change project**.
4. Paste the **Project number** you copied in the previous step and click **Set project**. Your Apps Script project is now linked with your cloud infrastructure!

---

## Step 4: Set Up the Execution Server on Google Apps Script

To let `ggsrun` securely trigger your scripts remotely (necessary for stateless `exe2` and `webapps` modes), you must add a small gateway wrapper inside your Google Apps Script.

### 4.1 Add the Shared Server Library

1. Inside your Google Apps Script editor, look at the left sidebar and click the `+` icon next to **Libraries**.
2. Paste the following official `ggsrunif` Library ID into the box:
   ```text
   115-19njNHlbT-NI0hMPDnVO1sdrw2tJKCAJgOTIAPbi_jq3tOo4lVRov
   ```
3. Click **Look up**. Select the **latest version** from the dropdown menu, keep the Identifier named exactly as `ggsrunif`, and click **Add**.

### 4.2 Add the Gateway Code

Open your `Code.gs` file in the script editor and replace the default code with the following wrapper endpoints (supporting both `exe2` and `webapps` modes):

```javascript
const doPost = (e) => ggsrunif.WebApps(e, "pass1");
const ExecutionApi = (e) => ggsrunif.ExecutionApi(e);
```

> [!NOTE]
> Change `"pass1"` to a secure custom password if you plan to execute webapps anonymously.

### 4.3 Deploy the Script as an API Executable (For `exe1` & `exe2`)

1. At the top-right corner of the editor, click **Deploy > New Deployment**.
2. Click the gear icon next to "Select type" and choose **API Executable**.
3. In the description, type `ggsrun API Engine`.
4. In the **Who has access** dropdown, select **Only myself** (this keeps your automation completely private and secure).
5. Click **Deploy**.
6. Copy the **Script ID** (a long string of alphanumeric characters) displayed on this screen.

### 4.4 Deploy the Script as a Web App (For `webapps`)

1. At the top-right corner of the editor, click **Deploy > New Deployment**.
2. Click the gear icon next to "Select type" and choose **Web app**.
3. In the description, type `ggsrun Web Engine`.
4. Set **Execute as** to **Me** (your email address).
5. Set **Who has access** to:
   - **Only myself**: Highly recommended for secure execution. This requires the `ggsrun` CLI to be authenticated via `ggsrun auth` with Drive scopes enabled.
   - **Anyone**: Select this if you need to trigger the webapp anonymously from a CI/CD pipeline without a token. (Access is secured by the password flag `-p`).
6. Click **Deploy**.
7. Copy the generated **Web app URL** (e.g., `https://script.google.com/macros/s/[WEB_APP_ID]/exec`).

---

## Step 5: Test Execution from Your Computer

Let's test everything to ensure your computer can run scripts inside your Google Account successfully.

1. Create a simple text file on your local computer inside your working folder (`ggsrun-workspace`) and name it `test_script.js`.
2. Open it with any text editor (Notepad, VS Code, etc.) and paste the following simple test code inside:
   ```javascript
   function main(message) {
     return "Success! Received local message: " + message;
   }
   ```
3. Open your terminal window in your working folder (`cd ~/Desktop/ggsrun-workspace` or your Windows path).
4. Run the test commands below:

### Test Option A: Execution via API Executable (`exe2` mode)

Run the script using `ggsrun exe2`. Replace `[YOUR_SCRIPT_ID]` with the Script ID you copied during Step 4.3 (or omit the `-i` flag if you saved the Script ID in `ggsrun.cfg` during authentication):

* **Windows Command Prompt**:
  ```cmd
  ggsrun exe2 -i "[YOUR_SCRIPT_ID]" -f ExecutionApi -s test_script.js -v "Hello Google Apps Script!" -j
  ```
* **macOS/Linux Terminal**:
  ```bash
  ./ggsrun exe2 -i "[YOUR_SCRIPT_ID]" -f ExecutionApi -s test_script.js -v "Hello Google Apps Script!" -j
  ```

### Test Option B: Execution via Web App (`webapps` mode)

Run the script using `ggsrun webapps`. Replace `[YOUR_WEB_APP_URL]` with the Web App URL you copied during Step 4.4 (or omit the `-u` flag if you saved the Web App URL in `ggsrun.cfg` during authentication):

* **Windows Command Prompt**:
  ```cmd
  ggsrun webapps -u "[YOUR_WEB_APP_URL]" -p pass1 -s test_script.js -v "Hello Google Apps Script!" -j
  ```
* **macOS/Linux Terminal**:
  ```bash
  ./ggsrun webapps -u "[YOUR_WEB_APP_URL]" -p pass1 -s test_script.js -v "Hello Google Apps Script!" -j
  ```

### Expected Output

In both cases, `ggsrun` will securely upload the script payload, run it in the Google Cloud environment, and output a clean JSON result directly to your terminal showing the return string:

```json
"Success! Received local message: Hello Google Apps Script!"
```

Your local workspace automation is completely operational!

---

## 6. Advanced Configurations (Modifying OAuth Scopes)

By default, `ggsrun` requests extensive permissions to handle file transfers and execution scopes. If you want to restrict or customize these scopes:
1. Open the `ggsrun.cfg` file generated in your configuration folder (located in `~/.config/` or specified by `GGSRUN_CFG_PATH`).
2. Locate the `"scopes"` JSON array:
   ```json
   "scopes": [
     "https://www.googleapis.com/auth/drive",
     "https://www.googleapis.com/auth/drive.readonly",
     "https://www.googleapis.com/auth/script.projects",
     "https://www.googleapis.com/auth/script.external_request"
   ]
   ```
3. Add or remove scopes as needed.
4. Save the file and re-run:
   ```bash
   $ ggsrun auth
   ```
   The browser will open to prompt a new consent screen matching your exact custom scope definitions.

---

## 7. Troubleshooting Diagnostics

### 1. Web Apps Returns Status Code 200, but output is HTML
If you set your Web App to "Only myself" but the CLI returns a parsing error with HTML, it means your `ggsrun` lacks the proper OAuth token. Run `ggsrun auth` to generate a token with the `drive` scope, which the CLI will automatically use to authenticate the Web App request across the Google 302 Redirects.

### 2. "Requested entity was not found" or 404 Errors
If utilizing GAS execution (`exe1` / `exe2`), verify the target project is currently deployed as an **API Executable** on the latest version. Un-deployed or draft states cannot be invoked externally. Check that your GAS Project settings are linked to your exact GCP Project Number.

### 3. Headless Server Authentication
If `ggsrun auth` detects a headless Linux environment (where it cannot spawn a local browser loopback), it elegantly degrades into manual mode. It prints the URL; copy it into an external browser, authorize, and paste the code block back into standard input.

---

## 8. Advanced Configurations: Profiles and Unattended Auto-Setup (New in v5.3.11)

### 8.1 Profiles (`--profile`)
By default, `ggsrun` saves its settings to `ggsrun.cfg`. However, if you are managing multiple Google Apps Script projects, or working across different environments (such as `development` and `production`), you can use **Profiles** to keep your configurations separate.

To create or use a specific profile, simply append the `--profile {name}` flag to any command.

#### Example: Setting up a "dev" and a "prod" environment
1. **Set up the development environment**:
   ```bash
   $ ggsrun --profile dev setup
   ```
   This will guide you through the setup and save all tokens and parameters to a file named `ggsrun_dev.cfg` instead of `ggsrun.cfg`.

2. **Set up the production environment**:
   ```bash
   $ ggsrun --profile prod setup
   ```
   This saves everything to `ggsrun_prod.cfg`.

3. **Running scripts under a specific profile**:
   When you want to execute a script in development:
   ```bash
   $ ggsrun --profile dev exe1 -s "my_script.js" -f "myFunction"
   ```
   To execute in production:
   ```bash
   $ ggsrun --profile prod exe1 -s "my_script.js" -f "myFunction"
   ```

---

### 8.2 Unattended Auto-Setup (`--yes` / `-y`)
If you want to automate the authorization and setup process (for example, in a CI/CD pipeline, setup scripts, or if you simply do not want to press Enter multiple times), you can use the `--yes` (or `-y`) flag.

When running `ggsrun setup --yes` or `ggsrun auth --yes`, `ggsrun` will:
- **Auto-confirm directories**: Automatically save the config file to the default or resolved path.
- **Auto-detect credentials**: If multiple `client_secret.json` files are found, it will automatically select the most relevant one.
- **Auto-merge conflicts**: If there are mismatching Client IDs or Project IDs between your files, it will automatically resolve them using the newest values.
- **Auto-skip parameters**: If you already have a Script ID or Web Apps URL registered, it will automatically keep them without asking.

#### Example: One-click setup using an active gcloud CLI session
If you have `gcloud` authenticated on your machine, you can fully set up `ggsrun` in a single command with zero prompts:
```bash
$ ggsrun setup --yes
```

---

### Related Links:
- 📖 **[Command Reference Manual](commands_reference.md)** - Reference all CLI flags and transfer commands.
- 🛡️ **[Security Sandbox Guide](sandbox_guide.md)** - Set up execution sandboxing for `exe1`.
- 🧪 **[Local Development & Testing Guide](development_guide.md)** - Detailed guidance for testing `ggsrun` code.
- 🏡 **[Back to Home](../README.md)**
