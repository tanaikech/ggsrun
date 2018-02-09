ggsrun
=====

<a name="TOP"></a>
# Update History

## ggsrun
* v1.0.0 (April 24, 2017)

    Initial release.

* v1.1.0 (April 28, 2017)

    1. Added a command for updating existing project on Google Drive. The detail information is [here](help/README.md#Update_Project).
    2. Added "TotalElapsedTime" for Show File List and Search Files.
    3. Some modifications.

* v1.2.0 (May 19, 2017)

    1. Added a command for retrieving revision files on Google Drive. The detail information is [here](help/README.md#RevisionFile).
    2. Some modifications.

* v1.2.1 (May 28, 2017)

    1. ggsrun.cfg got be able to be read using the environment variable.
        - If the environment variable (**``GGSRUN_CFG_PATH``**) is set, ggsrun.cfg is read using it.
        - If it is not set, ggsrun.cfg is read from the current working directory. This is as has been the way until now.
        - This is the response for some requests.
        - This incofmation was added to [here](help/README.md#environmentvariable).

* v1.2.2 (July 12, 2017)

    1. For Google Docs (spreadsheet, document, slide and drawing), since I noticed that the revision files would not be able to be retrieved using Drive API v3, I modified this using new workaround.
        - The new workaround is to use Drive API v2. ``drive.revisions.get`` of Drive API v2 can retrieve not only the revision list, but also the export links. I thought of the use of the export links. This became the new workaround.
        - For the files except for Google Docs, the revision files can be retrieved using Drive API v3.
        - The usage is [here](help/README.md#RevisionFile).

    I don't know when this workaround will not be able to be used. But if this could not be used, I would like to investigate of other method.

* v1.3.0 (August 30, 2017)

    1. From this version, [container-bound scripts](https://developers.google.com/apps-script/guides/bound) can be downloaded. The container-bound script is the script created at the script editor on Google Sheets, Docs, or Forms file. The usage is [here](help/README.md#DownloadBoundScript).
        - In order to download container-bound scripts, the project ID of container-bound scripts is required. The project ID can be retrieved as follows.
            - Open the project. And please operate follows using click.
                - -> File
                - -> Project properties
                - -> Get Script ID (**This is the project ID.**)
    1. When a project is downloaded, the filename of HTML file had become ``.gs``. This bug was modified.

* v1.3.1 (September 15, 2017)

    1. Recently, when scripts on local PC is uploaded to Google Drive as a new project, the time to create on Google became a bit long. (I think that this is due to Google Update.) Under this situation, when the script is uploaded, the timeout error occurs while the new project is created using the script. So the time until timeout of fetch was modified from 10 seconds to 30 seconds. By this, when the script is uploaded, no error occurs and the information of the created project is shown.
        - You can create a new project on Google Drive using scripts on local PC. The sample command is ``ggsrun u -f sample.gs1,sample2.gs,sample3.html -pn newprojectname``

* v1.3.2 (October 20, 2017)

    1. Updated ggsrun's Install manual (README.md). Since I thought that the manual became too complicated, I separated it to [the simple version](https://github.com/tanaikech/ggsrun/) and [the detail version](README.md). And also , recently, since Google's specification was updated, about how to deploy API executable and enable APIs for ggsrun's Install manual were updated.
    1. From this version, scripts in a project can be rearranged. The rearrangement can be done by interactively on your terminal and/or a configuration file. The usage is [here](README.md#rearrangescripts)
        - For rearranging scripts, there is one important point. **When scripts in a project is rearranged, version history of scripts is reset once. So if you don't want to reset the version history, before rearranging, please copy the project.** By copying project, the project before rearranging is saved.

* v1.3.3 (October 30, 2017)

    1. [At October 24, 2017, "Manifests" which is new function for controlling the properties of Google Apps Script was added (GAS).](https://developers.google.com/apps-script/) You can see the detail of "Manifests" [here](https://developers.google.com/apps-script/concepts/manifests). **In order to modify the manifests from local PC, I added this new function to ggsrun. By using this, you can edit the manifests and update it from your local PC.** The usage is [here](README.md#ModifyManifests)
        - I think that modifying manifests will be able to apply to various applications.
    1. Some modifications.


* v1.3.4 (January 2, 2018)

    1. Added new option for downloading 'bound-scripts' of Google Sheets, Docs, or Forms file.
        - When the bound-scripts are downloaded, the project name cannot be retrieved because Drive API cannot be used for the bound-scripts. So when the bound-scripts are downloaded, the project ID had been used previously. Such filename is not easily to be seen. By this additional option, users can give the filename when it downloads the bound-scripts.
        - The usage is [here](README.md#DownloadBoundScript)
    1. Removed a bug.
        - When a project is downloaded, script ID in the project is added to the top of each downloaded script as a comment. There was a problem at the character using for the comment out. This was modified.


* v1.4.0 (January 25, 2018)

    [Google Apps Script API](https://developers.google.com/apps-script/api/reference/rest/) was finally released. From this version, ggsrun uses this API. So ggsrun got to be able to use not only projects of standalone script type, but also projects of container-bound script type. I hope this updated ggsrun will be useful for you.

    1. **[To users which are using ggsrun with v1.3.4 and/or less](https://github.com/tanaikech/ggsrun/blob/master/README.md#from134to140).**
    1. For retrieving, downloading, creating and updating projects, [Apps Script API](https://developers.google.com/apps-script/api/reference/rest/) is used.
        - About retrieving information of projects, the information from Drive API is more than that from Apps Script API. So I used Drive API in this situation.
        - **[Please read how to enable APIs.](https://github.com/tanaikech/ggsrun/blob/master/README.md#BasicSettingFlow)**
    1. ggsrun got to be able to use both standalone scripts and container-bound scripts by Apps Script API.
        - [Create projects](README.md#UploadFiles)
        - [Update projects](README.md#Update_Project)
        - There are some issues for creating projects.
            1. After Manifests was added to GAS, the time zone can be set by it. But when a new project is created by API, I noticed that the time zone is different from own local time zone. When a new project is manually created by browser, the time zone is the same to own local time zone. I think that this may be a bug. So I added an option for setting time zone when a new project is created. And also I reported about this to [Google Issue Tracker](https://issuetracker.google.com/issues/72019223).
            1. If you want to create a bound script in Slide, an error occurs. When a bound script can be created to Spreadsheet, Document and Form using Apps Script API. Furthermore, when the bound script in Slide is updated, it works fine. So I think that this may be also a bug. I reported about this to [Google Issue Tracker](https://issuetracker.google.com/issues/72238499).
                - About this, when you create a bound script in Slides, if ggsrun returns no errors, it means that this issue was solved.
    1. [Both standalone scripts and container-bound scripts can be rearranged.](README.md#rearrangescripts)
        - The file of ``appsscript`` for Manifests is always displayed to the top of files on the script editor, while the array of files can be changed. I think that this is the specification.
    1. For the option ``exe1`` for executing GAS, it can use for both standalone scripts and container-bound scripts.
    1. [Delete files using file ID on Google Drive.](README.md#DownloadFiles)
    1. [Delete files in the project.](README.md#Update_Project)
    1. [ggsrun can create new container-bound script in the new Google Docs.](README.md#UploadFiles)
        - For example, ggsrun creates a new Spreadsheet and uploads the script files to the Spreadsheet as a container-bound script.
    1. [Retrieve and create versions of projects.](README.md#RevisionFile)
    1. [Unified the order of directories for searching ``client_secret.json`` and ``ggsrun.cfg``.](README.md#QA7)
    1. Some modifications.

* v1.4.1 (February 9, 2018)
    1. [For uploading, the resumable-upload method was added.](README.md#ResumableUpload)
        - The resumable-upload method is automatically used by the size of file.
            - "multipart/form-data" can upload files with the size less than 5 MB.
            - "resumable-upload" can upload files with the size more than 5 MB.
        - The chunk for resumable-upload is 100 MB as the default.
            - Users can also give this chunk size using an option.
        - ``$ ggsrun u -f filename -chunk 10``
            - This means that a file with filename is uploaded by each chunk of 10 MB.


**You can read "How to install" at [here](https://github.com/tanaikech/ggsrun/blob/master/README.md#How_to_Install).**

## Server
* v1.0.0 (April 24, 2017)

    Initial release.

[TOP](#TOP)
