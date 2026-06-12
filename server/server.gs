/**
 * @fileoverview Server library script for ggsrun.
 * This library functions as a gateway (interface) for the "ggsrun" CLI client,
 * enabling execution of local Google Apps Script files on Google servers via
 * the Google Apps Script Execution API (API Executable) or Web Apps deployment.
 *
 * Web: https://github.com/tanaikech/ggsrun
 */

/**
 * Global variable referencing the library's global scope, enabling evaluated
 * script payloads to call methods via the namespace alias (e.g. ggsrunif.Beacon()).
 * @type {Object}
 */
var ggsrunif = this;

/**
 * Global variable storing the active ggsrun execution instance.
 * Allows inner evaluated scripts to access the logging collector.
 * @type {ggsrun}
 */
var IGGSRUN;

/**
 * Current version of the ggsrun server library.
 * @type {string}
 */
var VERSION = "1.2.1";

/**
 * Entry point for executing scripts via the Google Apps Script API (formerly Execution API).
 * Parses the incoming JSON payload and evaluates the contained script securely.
 *
 * @param {string} e JSON string containing the script payload, execution functions, and log options.
 * @return {Object} An object containing the script execution result, log output, execution time, and date.
 */
function ExecutionApi(e) {
  IGGSRUN = new ggsrun(e, null, []);
  return IGGSRUN.executionapi();
}

/**
 * Entry point for executing scripts via Google Apps Script Web Apps.
 * Traverses request parameters, verifies security passwords (if configured),
 * and evaluates the script, returning the output as a JSON ContentService text output.
 *
 * @param {Object} e The event object containing request parameters from the POST request.
 * @param {Object} e.parameters Query parameters and post data arrays.
 * @param {string[]} e.parameters.com Script payload to evaluate.
 * @param {string[]} [e.parameters.pass] Security password sent by the client.
 * @param {string[]} [e.parameters.log] Flag indicating whether to record or skip execution logging.
 * @param {number} e.contentLength Length of the incoming payload.
 * @param {string} [password] The password required to run the Web App. Bypassed if null, undefined, or empty string.
 * @return {ContentService.TextOutput} JSON response content with mimeType set to JSON.
 */
function WebApps(e, password) {
  IGGSRUN = new ggsrun(e, password, []);
  return IGGSRUN.webapps();
}

/**
 * Helper logging method called within evaluated scripts.
 * Collects log entries during script runtime.
 *
 * @param {*} l The log message, object, or value to record.
 */
function Log(l) {
  if (IGGSRUN) {
    IGGSRUN.logg(l);
  }
}

/**
 * Verification beacon to confirm that the server script is reachable and functional.
 *
 * @return {string} A validation message containing the library version and metadata.
 */
function Beacon() {
  return new ggsrun(null, null, null).beacon();
}

// Private scoping for the ggsrun engine to keep helper functions encapsulated
(function(globalScope) {
  /**
   * The core execution and utility engine for ggsrun.
   */
  class ggsrun {
    /**
     * Initializes a new instance of the ggsrun engine.
     *
     * @param {Object|string|null} e Raw request event object (Web Apps) or JSON payload string (Execution API).
     * @param {string|null} pass Server-configured security password for Web Apps verification.
     * @param {Array<*>} [logar] Array to collect execution logs during run time.
     */
    constructor(e, pass, logar) {
      /** @type {Object|string|null} */
      this.e = e;
      /** @type {string|null} */
      this.pass = pass;
      /** @type {Array<*>} */
      this.logar = logar || [];
      /** @type {GoogleAppsScript.Spreadsheet.Sheet|null} */
      this.ss = null;
    }

    /**
     * Confirms the server exists and returns version details.
     *
     * @return {string} A human-readable verification string.
     */
    beacon() {
      return `This is a server for ggsrun. Version is ${VERSION}. Author is https://github.com/tanaikech .`;
    }

    /**
     * Collects log statements and appends them to the internal log array.
     *
     * @param {*} val The value or message to record.
     */
    logg(val) {
      this.logar.push(val);
    }

    /**
     * Downloads file content by ID (without converting to Google Workspace document formats).
     * Used primarily for fetching raw bytes of non-Google documents.
     *
     * @return {Object} Payload containing file raw bytes (as array of numbers), name, and mimeType.
     * @return {number[]} return.result The byte array of the downloaded file.
     * @return {string} return.name The file name on Google Drive.
     * @return {string} return.mimeType The file MIME type.
     */
    nodocsdownloader() {
      try {
        const file = DriveApp.getFileById(this.e);
        const blob = file.getBlob();
        return {
          result: blob.getBytes(),
          name: file.getName(),
          mimeType: blob.getContentType()
        };
      } catch (err) {
        return {
          result: "Error"
        };
      }
    }

    /**
     * Traverses the entire directory tree from Drive root recursively and returns paths.
     *
     * @return {string[][]} Array of split path arrays representing folder trees.
     */
    foldertree() {
      /**
       * Recursively traverses folder structures.
       *
       * @param {GoogleAppsScript.Drive.Folder} folder The current directory to scan.
       * @param {string} path Cumulative path string delimited by special token.
       * @param {string[][]} results Container for accumulated paths.
       * @return {string[][]} Accumulated paths.
       */
      const traverse = (folder, path, results) => {
        const subFolders = [];
        const folderIterator = folder.getFolders();
        while (folderIterator.hasNext()) {
          subFolders.push(folderIterator.next());
        }
        
        const nextPath = path + folder.getName() + "(" + folder.getId() + ")#_aabbccddee_#";
        const pathParts = nextPath.split("#_aabbccddee_#");
        pathParts.pop(); // Remove trailing empty split part
        results.push(pathParts);
        
        for (let i = 0; i < subFolders.length; i++) {
          traverse(subFolders[i], nextPath, results);
        }
        return results;
      };
      return traverse(DriveApp.getRootFolder(), "", []);
    }

    /**
     * Process and evaluate code sent via the Apps Script API (Execution API).
     *
     * @return {Object} Response object containing evaluation results and metadata.
     */
    executionapi() {
      const startTime = Date.now();
      const dateDat = Utilities.formatDate(new Date(), "GMT", "yyyy-MM-dd_HH:mm:ss'_GMT'");
      
      let rec;
      try {
        rec = JSON.parse(this.e);
      } catch (err) {
        return emessage.call(this, "Script Error on GAS side: Invalid JSON payload: " + err.message, startTime, dateDat);
      }

      if (rec && rec.log) {
        try {
          recLog.call(this, [[
            dateDat,
            `{API: "Execution API", ContentLength: ${this.e.length}, ExecutedFunction: "${rec.exefunc || 'main'}()"}`,
            rec.com
          ]]);
        } catch (err) {
          // Failure to log to spreadsheet should not block script execution
        }
      }

      try {
        if (rec && rec.com) {
          // Double eval is required: first resolves JSON-escaped string into standard string, second executes code.
          const resValues = (0, eval)((0, eval)(rec.com));
          return emessage.call(this, resValues, startTime, dateDat);
        } else {
          return emessage.call(this, "Error on GAS side: Bad parameters.", startTime, dateDat);
        }
      } catch (err) {
        return emessage.call(this, "Script Error on GAS side: " + err.message, startTime, dateDat);
      }
    }

    /**
     * Process and evaluate code sent via Web Apps POST request.
     *
     * @return {ContentService.TextOutput} Text output formatted in JSON format.
     */
    webapps() {
      const startTime = Date.now();
      const dateDat = Utilities.formatDate(new Date(), "GMT", "yyyy-MM-dd_HH:mm:ss'_GMT'");
      
      if (!this.e || !this.e.parameters) {
        return wmessage.call(this, "Error on GAS side: Bad request parameters.", startTime, dateDat);
      }

      const params = this.e.parameters;
      const logParam = params.log ? params.log[0] : null;
      const passParam = params.pass ? params.pass[0] : null;
      const comParam = params.com ? params.com[0] : null;

      // Access logs will be recorded unless explicitly bypassed (log parameter is false or omitted)
      if (logParam === "false" || !logParam) {
        try {
          recLog.call(this, [[
            dateDat,
            `{API: "Web Apps", ContentLength: ${this.e.contentLength}, Password: "${passParam || ''}"}`,
            comParam
          ]]);
        } catch (err) {
          // Logging failure should not disrupt the execution flow
        }
      }

      // Check if password verification is active (configured on server)
      const hasPassword = typeof this.pass === 'string' && this.pass.length > 0;
      if (!hasPassword || passParam === this.pass) {
        try {
          if (comParam) {
            const resValues = (0, eval)((0, eval)(comParam));
            return wmessage.call(this, resValues, startTime, dateDat);
          } else {
            return wmessage.call(this, "Error on GAS side: Bad parameters.", startTime, dateDat);
          }
        } catch (err) {
          return wmessage.call(this, "Script Error on GAS side: " + err.message, startTime, dateDat);
        }
      } else {
        return wmessage.call(this, "Error on GAS side: Bad password.", startTime, dateDat);
      }
    }
  }

  // --- Private Helper Functions (Encapsulated from Library namespace) ---

  /**
   * Constructs the unified response object structure.
   *
   * @param {*} data Evaluated code result or error message.
   * @param {number} startTime Start time timestamp (milliseconds).
   * @param {string} dateDat Pre-formatted timestamp string.
   * @param {Array<*>} logar Logs collector array.
   * @return {Object} Uniform script execution response object.
   */
  const buildResponse = (data, startTime, dateDat, logar) => {
    return {
      result: data,
      logger: logar,
      GoogleElapsedTime: (Date.now() - startTime) / 1000,
      ScriptDate: dateDat
    };
  };

  /**
   * Formats response payloads for the Execution API.
   *
   * @param {*} data Result data or error details.
   * @param {number} startTime Timestamp of script start time.
   * @param {string} dateDat Timestamp string.
   * @return {Object} Execution API compatible response structure.
   */
  const emessage = function(data, startTime, dateDat) {
    return buildResponse(data, startTime, dateDat, this.logar);
  };

  /**
   * Formats response payloads for Web Apps as JSON TextOutput.
   *
   * @param {*} data Result data or error details.
   * @param {number} startTime Timestamp of script start time.
   * @param {string} dateDat Timestamp string.
   * @return {ContentService.TextOutput} JSON content service output.
   */
  const wmessage = function(data, startTime, dateDat) {
    const responseObj = buildResponse(data, startTime, dateDat, this.logar);
    return ContentService
      .createTextOutput(JSON.stringify(responseObj))
      .setMimeType(ContentService.MimeType.JSON);
  };

  /**
   * Resolves or initializes the Spreadsheet access log sheet.
   * Dynamically relocates the file to the parent directory of the script if possible.
   *
   * @param {string} _log Spreadsheet filename.
   * @return {GoogleAppsScript.Spreadsheet.Sheet} First sheet of the Spreadsheet.
   */
  const logsheet = function(_log) {
    const files = DriveApp.getFilesByName(_log);
    if (files.hasNext()) {
      const ss = SpreadsheetApp.openById(files.next().getId());
      return ss.getSheets()[0];
    }

    const ss = SpreadsheetApp.create(_log);
    const logss = DriveApp.getFileById(ss.getId());
    let parentFolder;
    try {
      parentFolder = DriveApp.getFileById(ScriptApp.getScriptId()).getParents().next();
    } catch (e) {
      try {
        parentFolder = DriveApp.getFileById(SpreadsheetApp.getActiveSpreadsheet().getId()).getParents().next();
      } catch (err) {
        parentFolder = DriveApp.getRootFolder();
      }
    }
    if (parentFolder && parentFolder.getId() !== DriveApp.getRootFolder().getId()) {
      logss.moveTo(parentFolder);
    }
    return ss.getSheets()[0];
  };

  /**
   * Records execution logs into the Spreadsheet sheet.
   *
   * @param {Array<Array<*>>} logAr Two-dimensional log payload array.
   */
  const recLog = function(logAr) {
    if (!this.ss) {
      this.ss = logsheet("ggsrun.log");
    }
    if (logAr && logAr.length > 0) {
      this.ss.getRange(this.ss.getLastRow() + 1, 1, logAr.length, logAr[0].length).setValues(logAr);
    }
  };

  // Assign class and metadata to the library instance context
  ggsrun.help = "This is a server script for ggsrun using Execution API and Web Apps.";
  ggsrun.name = 'ggsrun';

  globalScope.ggsrun = ggsrun;
})(this);
