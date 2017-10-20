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

    1. Updated ggsrun's Install manual (README.md). Since I thought that the manual became too complicated, I separated it to [the simple version](https://github.com/tanaikech/ggsrun/) and [the detail version](https://github.com/tanaikech/ggsrun/help). And also , recently, since Google's specification was updated, about how to deploy API executable and enable APIs for ggsrun's Install manual were updated.
    1. From this version, scripts in a project can be rearranged. The rearrangement can be done by interactively on your terminal and/or a configuration file. The usage is [here](help/README.md#rearrangescripts)
        - For rearranging scripts, there is one important point. **When scripts in a project is rearranged, version history of scripts is reset once. So if you don't want to reset the version history, before rearranging, please copy the project.** By copying project, the project before rearranging is saved.

## Server
* v1.0.0 (April 24, 2017)

    Initial release.

[TOP](#TOP)
