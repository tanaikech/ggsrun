# ggsrun

<a name="top"></a>

# Update History

## ggsrun

- v1.0.0 (April 24, 2017)

  Initial release.

- v1.1.0 (April 28, 2017)

  1. Added a command for updating existing project on Google Drive. The detail information is [here](help/README.md#updateproject).
  2. Added "TotalElapsedTime" for Show File List and Search Files.
  3. Some modifications.

- v1.2.0 (May 19, 2017)

  1. Added a command for retrieving revision files on Google Drive. The detail information is [here](help/README.md#revisionfile).
  2. Some modifications.

- v1.2.1 (May 28, 2017)

  1. ggsrun.cfg got be able to be read using the environment variable.
     - If the environment variable (**`GGSRUN_CFG_PATH`**) is set, ggsrun.cfg is read using it.
     - If it is not set, ggsrun.cfg is read from the current working directory. This is as has been the way until now.
     - This is the response for some requests.
     - This information was added to [here](help/README.md#environmentvariable).

- v1.2.2 (July 12, 2017)

  1. For Google Docs (spreadsheet, document, slide and drawing), since I noticed that the revision files would not be able to be retrieved using Drive API v3, I modified this using new workaround.
     - The new workaround is to use Drive API v2. `drive.revisions.get` of Drive API v2 can retrieve not only the revision list, but also the export links. I thought of the use of the export links. This became the new workaround.
     - For the files except for Google Docs, the revision files can be retrieved using Drive API v3.
     - The usage is [here](help/README.md#revisionfile).

  I don't know when this workaround will not be able to be used. But if this could not be used, I would like to investigate of other method.

- v1.3.0 (August 30, 2017)

  1. From this version, [container-bound scripts](https://developers.google.com/apps-script/guides/bound) can be downloaded. The container-bound script is the script created at the script editor on Google Sheets, Docs, or Forms file. The usage is [here](help/README.md#DownloadBoundScript).
     - In order to download container-bound scripts, the project ID of container-bound scripts is required. The project ID can be retrieved as follows.
       - Open the project. And please operate follows using click.
         - -> File
         - -> Project properties
         - -> Get Script ID (**This is the project ID.**)
  1. When a project is downloaded, the filename of HTML file had become `.gs`. This bug was modified.

- v1.3.1 (September 15, 2017)

  1. Recently, when scripts on local PC is uploaded to Google Drive as a new project, the time to create on Google became a bit long. (I think that this is due to Google Update.) Under this situation, when the script is uploaded, the timeout error occurs while the new project is created using the script. So the time until timeout of fetch was modified from 10 seconds to 30 seconds. By this, when the script is uploaded, no error occurs and the information of the created project is shown.
     - You can create a new project on Google Drive using scripts on local PC. The sample command is `ggsrun u -f sample.gs1,sample2.gs,sample3.html -pn newprojectname`

- v1.3.2 (October 20, 2017)

  1. Updated ggsrun's Install manual (README.md). Since I thought that the manual became too complicated, I separated it to [the simple version](https://github.com/tanaikech/ggsrun/) and [the detail version](README.md). And also , recently, since Google's specification was updated, about how to deploy API executable and enable APIs for ggsrun's Install manual were updated.
  1. From this version, scripts in a project can be rearranged. The rearrangement can be done by interactively on your terminal and/or a configuration file. The usage is [here](README.md#rearrangescripts)
     - For rearranging scripts, there is one important point. **When scripts in a project is rearranged, version history of scripts is reset once. So if you don't want to reset the version history, before rearranging, please copy the project.** By copying project, the project before rearranging is saved.

- v1.3.3 (October 30, 2017)

  1. [At October 24, 2017, "Manifests" which is new function for controlling the properties of Google Apps Script was added (GAS).](https://developers.google.com/apps-script/) You can see the detail of "Manifests" [here](https://developers.google.com/apps-script/concepts/manifests). **In order to modify the manifests from local PC, I added this new function to ggsrun. By using this, you can edit the manifests and update it from your local PC.** The usage is [here](README.md#modifymanifests)
     - I think that modifying manifests will be able to apply to various applications.
  1. Some modifications.

* v1.3.4 (January 2, 2018)

  1. Added new option for downloading 'bound-scripts' of Google Sheets, Docs, or Forms file.
     - When the bound-scripts are downloaded, the project name cannot be retrieved because Drive API cannot be used for the bound-scripts. So when the bound-scripts are downloaded, the project ID had been used previously. Such filename is not easily to be seen. By this additional option, users can give the filename when it downloads the bound-scripts.
     - The usage is [here](README.md#DownloadBoundScript)
  1. Removed a bug.
     - When a project is downloaded, script ID in the project is added to the top of each downloaded script as a comment. There was a problem at the character using for the comment out. This was modified.

- v1.4.0 (January 25, 2018)

  [Google Apps Script API](https://developers.google.com/apps-script/api/reference/rest/) was finally released. From this version, ggsrun uses this API. So ggsrun got to be able to use not only projects of standalone script type, but also projects of container-bound script type. I hope this updated ggsrun will be useful for you.

  1. **[To users which are using ggsrun with v1.3.4 and/or less](https://github.com/tanaikech/ggsrun/blob/master/README.md#from134to140).**
  1. For retrieving, downloading, creating and updating projects, [Apps Script API](https://developers.google.com/apps-script/api/reference/rest/) is used.
     - About retrieving information of projects, the information from Drive API is more than that from Apps Script API. So I used Drive API in this situation.
     - **[Please read how to enable APIs.](https://github.com/tanaikech/ggsrun/blob/master/README.md#BasicSettingFlow)**
  1. ggsrun got to be able to use both standalone scripts and container-bound scripts by Apps Script API.
     - [Create projects](README.md#uploadfiles)
     - [Update projects](README.md#updateproject)
     - There are some issues for creating projects.
       1. After Manifests was added to GAS, the time zone can be set by it. But when a new project is created by API, I noticed that the time zone is different from own local time zone. When a new project is manually created by browser, the time zone is the same to own local time zone. I think that this may be a bug. So I added an option for setting time zone when a new project is created. And also I reported about this to [Google Issue Tracker](https://issuetracker.google.com/issues/72019223).
       1. If you want to create a bound script in Slide, an error occurs. When a bound script can be created to Spreadsheet, Document and Form using Apps Script API. Furthermore, when the bound script in Slide is updated, it works fine. So I think that this may be also a bug. I reported about this to [Google Issue Tracker](https://issuetracker.google.com/issues/72238499).
          - About this, when you create a bound script in Slides, if ggsrun returns no errors, it means that this issue was solved.
  1. [Both standalone scripts and container-bound scripts can be rearranged.](README.md#rearrangescripts)
     - The file of `appsscript` for Manifests is always displayed to the top of files on the script editor, while the array of files can be changed. I think that this is the specification.
  1. For the option `exe1` for executing GAS, it can use for both standalone scripts and container-bound scripts.
  1. [Delete files using file ID on Google Drive.](README.md#downloadfiles)
  1. [Delete files in the project.](README.md#updateproject)
  1. [ggsrun can create new container-bound script in the new Google Docs.](README.md#uploadfiles)
     - For example, ggsrun creates a new Spreadsheet and uploads the script files to the Spreadsheet as a container-bound script.
  1. [Retrieve and create versions of projects.](README.md#revisionfile)
  1. [Unified the order of directories for searching `client_secret.json` and `ggsrun.cfg`.](README.md#QA7)
  1. Some modifications.

- v1.4.1 (February 9, 2018)
  1. [For uploading, the resumable-upload method was added.](README.md#ResumableUpload)
     - The resumable-upload method is automatically used by the size of file.
       - "multipart/form-data" can upload files with the size less than 5 MB.
       - "resumable-upload" can upload files with the size more than 5 MB.
     - The chunk for resumable-upload is 100 MB as the default.
       - Users can also give this chunk size using an option.
     - `$ ggsrun u -f filename -chunk 10`
       - This means that a file with filename is uploaded by each chunk of 10 MB.

<a name="v150"></a>

- v1.5.0 (October 27, 2018)
  1. [From this version, ggsrun got to be able to download all files and folders in the specific folder in Google Drive.](README.md#downloadfilesfromfolder) When all files are downloaded from a folder, the same folder structure of Google Drive is created to the local PC.
     - `$ ggsrun d -f folderName or folderId`
       - When the project file is downloaded, it is downloaded as a zip file. All scripts in the project is put in the zip file.
       - Also when you download a single project, you can use an option `--zip` or `-z`. By this, the downloaded project is saved as a zip file.
       - This new function can be also used for the shared folders. When you want to download the files from the shared folder, please use the folder ID of the shared folder.
  1. The file list with the folder tree in the specific folder got to be able to be retrieved.
  1. When the files are downloaded, the progression got to be able to be seen. When you want to see the progression, please use `-j` when you download files and folders.
  1. Files with large size got to be able to be used. In order to download files with large size (several gigabytes), files are saved by chunks.
  1. Some modifications.

<a name="v151"></a>

- v1.5.1 (November 2, 2018)
  1. Removed a bug.
     - When a file information was retrieved, createdTime and modifiedTime couldn't be seen and the information was incomplete.

<a name="v152"></a>

- v1.5.2 (November 4, 2018)
  1. About [downloading folders](https://github.com/tanaikech/ggsrun/blob/master/help/README.md#downloadfilesfromfolder), when files are downloaded from a folder, you can download Google Docs files with the mimeType you want. For example, when you download files from the folder, if `-e txt` is used, Google Docs are downloaded as the text file. When `-e pdf` is used, they are downloaded as the PDF file. Of course, there are mimeType which cannot be converted.
     - `$ ggsrun d -f [folderName] -e txt -j`
  1. About [uploading files](https://github.com/tanaikech/ggsrun/blob/master/help/README.md#uploadfiles), when files are uploaded from your local PC, the files got to be able to be converted to Google Docs. For this, new option of `--convertto`, `-c` is added. For example, when a text file is uploaded, if you use `-c doc`, the text file is uploaded as Google Document.
     - `$ ggsrun u -f [fileName] -c doc -j`

<a name="v160"></a>

- v1.6.0 (November 30, 2018)
  1. Although at ggsrun, files can be searched by filename and file ID, searching files using search query and regex couldn't be done. From version 1.6.0, files got to be able to be searched using the search query and regex.
     - `$ ggsrun sf -q "### search query ###" -f "### fields ###" -r "### regex ###"`
  1. Some modifications.

<a name="v170"></a>

- v1.7.0 (December 27, 2018)
  1. [Manage permissions of files.](https://github.com/tanaikech/ggsrun/blob/master/help/README.md#managepermissions)
  1. [Get Drive Information.](https://github.com/tanaikech/ggsrun/blob/master/help/README.md#getdriveinformation) By this, you can know the storage quotas.
  1. [**ggsrun got to be able to be used by not only OAuth2, but also Service Account. By this, using ggsrun, Google Drive for Service Account got to be able to be managed.**](https://github.com/tanaikech/ggsrun/blob/master/help/README.md#useserviceaccount)
  1. Some modifications.

<a name="v171"></a>

- v1.7.1 (December 30, 2018)
  1. A bug was removed.
     - When a project is downloaded and zipped, there was a case that "createdTime" and "modifiedTime" of the project cannot be retrieved by Apps Script API. This was modified.

<a name="v173"></a>

- v1.7.3 (January 3, 2020)

  1. It seems that the specification of `github.com/urfave/cli` was changed by the update of [https://github.com/urfave/cli](https://github.com/urfave/cli). By this, when `go get -u github.com/tanaikech/ggsrun` is run, an error occurred. So this error was removed.

<a name="v174"></a>

- v1.7.4 (March 11, 2020)

  1. Recently, I noticed that new Google Apps Script project of the standalone script type cannot be created by the create method of Drive API. From now, in order to create the standalone Google Apps Script project, only Google Apps Script API is required to be used. [Ref](https://gist.github.com/tanaikech/0609f2cd989c28d6bd49d211b70b453d) By this, I updated ggsrun. So the command for creating new GAS project is not changed.

     - `$ ggsrun u -p ###folderId### -f sample.gs -pn sampleGASProjectName`

<a name="v200"></a>

- v2.0.0 (February 25, 2022)

  1. Modified using the latest libraries. The specification of ggsrun is not changed.

<a name="v201"></a>

- v2.0.1 (February 24, 2023)

  1. Modified go.mod, go.sum.

## Server

- v1.0.0 (April 24, 2017)

  Initial release.

**You can read "How to install" at [here](https://github.com/tanaikech/ggsrun/blob/master/README.md#howtoinstall).**

[TOP](#top)
