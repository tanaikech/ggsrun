# ggsrun

<a name="TOP"></a>
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENCE)

<a name="Overview"></a>

# Overview

This is a CLI tool to execute Google Apps Script (GAS) on a terminal. Also this CLI tool can be used for managing files in Google Drive for OAuth2 and Service Account.

<a name="Demo"></a>

# Demo

![](help/images/spreadsheetdemo.gif)

<a name="Description"></a>

# Description

Will you want to develop GAS on your local PC? Generally, when we develop GAS, we have to login to Google using own browser and develop it on the Script Editor. Recently, I have wanted to have more convenient local-environment for developing GAS. So I created this "ggsrun". The main work is to execute GAS on local terminal and retrieve the results from Google. Furthermore, this tool can be also used for managing files in own Google Drive (OAuth2) and Google Drive for Service Account.

Features of "ggsrun" are as follows.

1. **[Develops GAS using your local terminal and text editor which got accustomed to using.](help/README.md#demosublime)**<sup><font color="Red">Updated! (v1.4.0)</font></sup>
1. **[Executes GAS by giving values to your script.](help/README.md#executesgasandretrievesresultvalues)**
1. **[Executes GAS made of CoffeeScript.](help/README.md#coffeescript)**
1. **[Downloads spreadsheet, document and presentation, while executes GAS, simultaneously.](help/README.md#downloadfiles)**
1. **[Downloads files from Google Drive and Uploads files to Google Drive.](help/README.md#uploadfiles)** <sup><font color="Red">Updated! (v1.4.1)</font></sup>
1. **[Downloads standalone script and bound script.](help/README.md#downloadfiles)** <sup><font color="Red">Updated! (v1.4.0)</font></sup>
1. **[Downloads all files and folders in a specific folder.](help/README.md#downloadfilesfromfolder)** <sup><font color="Red">Updated! (v1.5.2)</font></sup>
1. **[Upload script files and create project as standalone script and container-bound script.](help/README.md#uploadfiles)** <sup><font color="Red">Updated! (v1.5.2)</font></sup>
1. **[Update project.](help/README.md#updateproject)** <sup><font color="Red">Updated! (v1.4.0)</font></sup>
1. **[Retrieve revision files of Google Docs and retrieve versions of projects.](help/README.md#revisionfile)** <sup><font color="Red">Updated! (v1.4.0)</font></sup>
1. **[Rearranges scripts in project.](help/README.md#rearrangescripts)** <sup><font color="Red">Updated! (v1.4.0)</font></sup>
1. **[Modifies Manifests in project.](help/README.md#modifymanifests)**
1. **[Seach files in Google Drive using search query and regex](help/README.md#searchfilesusingregex)** <sup><font color="Red">Updated! (v1.6.0)</font></sup>
1. **[Manage Permissions of files](help/README.md#managepermissions)** <sup><font color="Red">Updated! (v1.7.0)</font></sup>
1. **[Get Drive Information.](help/README.md#getdriveinformation)** <sup><font color="Red">Updated! (v1.7.0)</font></sup>
1. **[ggsrun got to be able to be used by not only OAuth2, but also Service Account.](help/README.md#useserviceaccount)** <sup><font color="Red">Updated! (v1.7.0)</font></sup>

<a name="howtoinstall"></a>

# How to Install

December 28, 2021: If you want to simply test ggsrun, you can also use [this method](https://gist.github.com/tanaikech/695f7016b04e4c4156ad928e9482ead9).

## 1. Get ggsrun

Download an executable file of ggsrun from [the release page](https://github.com/tanaikech/ggsrun/releases) and import to a directory with path.

or

Use go get.

```bash
$ go install github.com/tanaikech/ggsrun@latest
```

- `GO111MODULE=on`

<a name="BasicSettingFlow"></a>

## Before Basic setting flow

**IMPORTANT: Please check this.**

At April 8, 2019, the specification of Google Apps Script Project was changed. By this, when new GAS project is created after April 8, 2019, in order to use Goole API (Google Apps Script API and Drive API which can be used by ggsrun.), the GAS project is required to link to Cloud Platform Project. When you use the GAS project creted after April 8, 2019, at first, please do [this flow](https://gist.github.com/tanaikech/e945c10917fac34a9d5d58cad768832c).

After the GAS project was linked to Cloud Platform Project by above flow, please do "Basic setting flow" at the next section.

- [Ref1: Default Cloud Platform projects](https://developers.google.com/apps-script/guides/cloud-platform-projects#default_cloud_platform_projects)
- [Ref2: Linking Cloud Platform Project to Google Apps Script Project](https://gist.github.com/tanaikech/e945c10917fac34a9d5d58cad768832c)

## 2. Basic setting flow

When you click each link of title, you can see the detail information.

1. [Setup ggsrun Server (at Google side)](help/README.md#setupggsrunserver)
   - Create new project and install the server as a library.
   - [Deploy API executable](https://developers.google.com/apps-script/api/how-tos/execute#step_1_deploy_the_script_as_an_api_executable). Choose "Only myself" as "Who has access to the script"
   - [Install the server as a library.](https://developers.google.com/apps-script/guides/libraries#managing_libraries) Script ID of the library is
     - **`115-19njNHlbT-NI0hMPDnVO1sdrw2tJKCAJgOTIAPbi_jq3tOo4lVRov`**
   - **<u>After installed the library, please push the save button at the script editor.</u>** This is very important! By this, the library is completely reflected.
1. [Get Client ID, Client Secret](help/README.md#getclientid)
   - On the Script Editor
     - Resources -> Cloud Platform Project
     - Click the lower part of "This script is currently associated with project:"
     - In "Getting Started", Click "Enable APIs and get credentials like keys".
     - On "API APIs&services"
     - Click "Credentials" at left side.
     - At "Create Credentials", Click OAuth client ID.
     - Choose **Other**
     - Input Name (This is a name you want.)
     - done
     - Download a JSON file with Client ID and Client Secret as **`client_secret.json`** using download button.
1. [Enable APIs](help/README.md#onstallexecutionapi)
   - ggsrun uses Google Apps Script API and Drive API. Please enable them at API console. You can directly access them as follows. Project ID can be seen at downloaded `client_secret.json`.
     - `https://console.cloud.google.com/apis/library/script.googleapis.com/?project=### project ID ###`
       - **Also here [https://script.google.com/home/usersettings](https://script.google.com/home/usersettings) has to be enabled. Please turn ON.**
     - `https://console.cloud.google.com/apis/api/drive.googleapis.com/?project=### project ID ###`
1. [Create configure file for ggsrun](help/README.md#Createconfigurefile)
   - Run `$ ggsrun auth` at the directory with `client_secret.json`.
1. [Test Run](help/README.md#runggsrun)
   - Create a sample script `function main(){return Beacon()}` as `sample.gs`.
   - Run `$ ggsrun e2 -s sample.gs -i [Script ID] -j`. Script ID is ID of the project installed the server.

Congratulation! You got ggsrun!

<a name="from134to140"></a>

# To users which are using ggsrun with v1.3.4 and/or less <sup><font color="Red">Updated! (v1.4.0)</font></sup>

Please reauthorize to include a new scope to the access token as follows.

1. Confirm whether Google Apps Script API is enabled. You can directly access it as follows. Project ID can be seen at the downloaded `client_secret.json`.
   - `https://console.cloud.google.com/apis/library/script.googleapis.com/?project=### project ID ###`
   - Also here [https://script.google.com/home/usersettings](https://script.google.com/home/usersettings) has to be enabled. Please turn ON.
1. Add a scope of `https://www.googleapis.com/auth/script.projects` to `ggsrun.cfg`.
1. Run the following command under the directory with `client_secret.json` and `ggsrun.cfg`.
   - `$ ggsrun auth`

Completed!

<a name="from170"></a>

# From version 1.7.0, ggsrun can access to Google Drive using Service Account. <sup><font color="Red">Updated! (v1.7.0)</font></sup>

ggsrun can access to Google Drive using [Service Account](https://developers.google.com/identity/protocols/OAuth2ServiceAccount). When OAuth2 is used, you can see the files and folders in own Google Drive. When Service Account is used, you can see them in Google Drive for Service Account. Namely, the Drive for OAuth2 is different from that for Service Account. Please be careful this. And when Service Account is used, there are what it can do and what it can't do. About them, please read [here](help/README.md#useserviceaccount).

# How to use ggsrun

1. [Executes GAS and Retrieves Result Values](help/README.md#executesgasandretrievesresultvalues)
1. [Executes GAS with Values and Retrieves Feedbacked Values](help/README.md#executesgaswithvaluesandretrievesfeedbackedvalues)
1. [For Debug](help/README.md#fordebug)
1. [Executes GAS with Values and Downloads File](help/README.md#executesgaswithvaluesanddownloadsfile)
1. [Executes Existing Functions on Project](help/README.md#ExecutesExistingFunctionsonProject)
1. [Download Files](help/README.md#downloadfiles)
1. [Downloads all files and folders in a specific folder.](help/README.md#downloadfilesfromfolder)
1. [Upload Files](help/README.md#uploadfiles)
1. [Show File List](help/README.md#showfilelist)
1. [Search Files](help/README.md#searchfiles)
1. [Update Project](help/README.md#updateproject)
1. [Retrieve Revision Files and Versions of Projects](help/README.md#revisionfile)
1. [Rearrange Script in Project](help/README.md#rearrangescripts)
1. [Modify Manifests](help/README.md#modifymanifests)
1. [Seach Files using Query and Regex](help/README.md#searchfilesusingregex)
1. [Manage Permissions of files](#managepermissions)
1. [Get Drive Information.](#getdriveinformation)
1. [ggsrun got to be able to be used by not only OAuth2, but also Service Account.](#useserviceaccount)

# Applications

1. [For Sublime Text](help/README.md#demosublime)
1. [For CoffeeScript](help/README.md#coffeescript)
1. [Create Triggers](help/README.md#createtriggers)
1. [Link to Python script](help/README.md#linktovariousresources)

# [Q&A](help/README.md#qa)

1. [Authorization for Google Services for your script](help/README.md#qa1)
1. [In the case that result is "Script Error on GAS side: Insufficient Permission"](help/README.md#qa2)
1. [In the case that result is "message": "Requested entity was not found."](help/README.md#qa3)
1. [In the case that result is "Script Error on GAS side: Script has attempted to perform an action that is not allowed when invoked through the Google Apps Script Execution API."](help/README.md#qa4)
1. [In the case that result is "Missing ';' before statement."](help/README.md#qa5)
1. [About library](help/README.md#qa6)
1. [Order of directories for searching](help/README.md#qa7)

---

<a name="Licence"></a>

# Licence

[MIT](LICENCE)

<a name="Author"></a>

# Author

[Tanaike](https://tanaikech.github.io/about/)

If you have any questions and commissions for me, feel free to tell me using e-mail tanaike@hotmail.com

<a name="Update_History"></a>

# Update History

You can see the Update History at **[here](help/UpdateHistory.md)**.

<u>If you want to read the detail manual, please check [here](help/README.md).</u>

[TOP](#TOP)
