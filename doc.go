/*
Package main (doc.go) :
This is a CLI tool to execute Google Apps Script (GAS) on a terminal.

Will you want to develop GAS on your local PC? Generally, when we develop GAS, we have to login to Google using own browser and develop it on the Script Editor. Recently, I have wanted to have more convenient local-environment for developing GAS. So I created this "ggsrun". The main work is to execute GAS on local terminal and retrieve the results from Google.

# Features of "ggsrun" are as follows.

1. Develops GAS using your terminal and text editor which got accustomed to using.

2. Executes GAS by giving values to your script.

3. Executes GAS made of CoffeeScript.

4. Downloads spreadsheet, document and presentation, while executes GAS, simultaneously.

5. Upload files to Google Drive. When files are uploaded, also they can be converted by options.

6. Creates, updates and backs up project of both standalone type and bound script type.

7. Creates Google Docs (Spreadsheet, Document, Slide and Form) and create bound script in the created Google Docs.

8. Downloads files from Google Drive and Uploads files to Google Drive. Also container-bound scripts can be downloaded.

9. Downloads revision files from Google Drive.

10. Rearranges files in project of both standalone type and bound script type.

11. Modifies Manifests (appsscript.json) in project.

12. Remove files in the project of both standalone type and bound script type.

13. Retrieve revision file list and revision data.


You can see the release page https://github.com/tanaikech/ggsrun/releases

# Google API

ggsrun uses Execution API, Web Apps and Drive API on Google. About how to install ggsrun, please check my github repository.

https://github.com/tanaikech/ggsrun/

You can read the detail information there.


---------------------------------------------------------------

# How to Execute Google Apps Script Using ggsrun
When you have the configure file `ggsrun.cfg`, you can execute GAS. If you cannot find it, please download `client_secret.json` and run

$ ggsrun auth

In the case of using Execution API,

$ ggsrun e1 -s sample.gs

If you want to execute a function except for `main()` of default, you can use an option like `-f foo`. This command `exe1` can be used to execute a function on project.

$ ggsrun e1 -f foo

$ ggsrun e2 -s sample.gs

At `e2`, you cannot select the executing function except for `main()` of default.

`e1`, `e2` and `-s` mean using Execution API and GAS script file name, respectively. Sample codes which are shown here will be used Execution API. At this time, the executing function is `main()`,  which is a default, in the script.

In the case of using Web Apps,

$ ggsrun w -s sample.gs -p password -u [ WebApps URL ]

`w` and `-p` mean using Web Apps and password you set at the server side, respectively. Using `-u` it imports Web Apps URL like `-u https://script.google.com/macros/s/#####/exec`.


---------------------------------------------------------------
*/
package main
