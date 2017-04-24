var VERSION = "1.0.0";
var IGGSRUN;

/**
 * This method is for ggsrun with Execution API.
 * "ggsrun" brings us more convenient development-environment of Google Apps Script
 * on local PC. "ggsrun" sends your script of Google Apps Script created at local PC
 * to Google and retrieves the results from Google. This library "ggsrunif" is used
 * for "ggsrun" made of golang as an InterFace. Namely, this is a server script.
 * WEB https://github.com/tanaikech
 * <h3>usage</h3>
 * <pre>
 * $ ggsrun e2 -s [Script File] options
 * </pre>
 * @return {string} Executed result of script.
 */
function ExecutionApi(e) {
  IGGSRUN = new ggsrun(e, null, []);
  return IGGSRUN.executionapi();
}

/**
 * This method is for ggsrun with Web Apps.
 * "ggsrun" brings us more convenient development-environment of Google Apps Script
 * on local PC. "ggsrun" sends your script of Google Apps Script created at local PC
 * to Google and retrieves the results from Google. This library "ggsrunif" is used
 * for "ggsrun" made of golang as an InterFace. Namely, this is a server script.
 * WEB https://github.com/tanaikech
 * <h3>usage</h3>
 * <pre>
 * function doPost(e) {
 *   var password = "password"; // Password for using Web Apps
 *   return ggsrunif.WebApps(e, password);
 * }
 * </pre>
 * @param {e} Script and parameters.
 * @param {password} Password for launching script sent from local PC.
 * @return {string} Executed result of script.
 */
function WebApps(e, password) {
  IGGSRUN = new ggsrun(e, password, []);
  return IGGSRUN.webapps();
//  return new ggsrun(e, password).webapps();
}

/**
 * This method is used to retrieve parameters as a log. Please use this in your script you send using ”ggsrun” instead of "Logger.log()".
 * @param {Object} channelId Group to leave
 * @return {Object} result
 */
function Log(l) {
  IGGSRUN.logg(l);
}

/**
 * This method is a beacon to confirm the existence of ggsrun's server.
 * WEB https://github.com/tanaikech
 * <h3>usage</h3>
 * <pre>
 * ggsrunif.Beacon()
 * </pre>
 * @return {string} Same string for sending string.
 */
function Beacon() {
  return new ggsrun(null, null, null).beacon();
}

// Server script
(function(x) {
  var ggsrun;

  ggsrun = (function() {
      var emessage, wmessage, logsheet, recLog;
      ggsrun.help = "This is a server script for ggsrun using Execution API and Web Apps.";
      ggsrun.name = 'ggsrun';
  
      function ggsrun(e, pass, logar) {
        this.e = e;
        this.pass = pass;
        this.ss = logsheet("ggsrun.log");
        this.logar = logar;
      }

      ggsrun.prototype.beacon = function() {
        return "This is a server for ggsrun. Version is " + VERSION + ". Autor is https://github.com/tanaikech .";
      };
    
      ggsrun.prototype.logg = function(val) {
        this.logar.push(val);
      };
    
      ggsrun.prototype.nodocsdownloader = function() {
        return (function(id){
          try {
            var file = DriveApp.getFileById(id);
            return {
              result: file.getBlob().getBytes(),
              name: file.getName(),
              mimeType: file.getBlob().getContentType()
            };
          } catch(err) {
            return {
              result: "Error"
            };          
          }
        })(this.e);
      };

      ggsrun.prototype.foldertree = function() {
          return (function(folder, folderSt, results){
              var ar = [];
              var folders = folder.getFolders();
              while(folders.hasNext()) ar.push(folders.next());
              folderSt += folder.getName() + "(" + folder.getId() + ")#_aabbccddee_#";
              var array_folderSt = folderSt.split("#_aabbccddee_#");
              array_folderSt.pop()
              results.push(array_folderSt);
              ar.length == 0 && (folderSt = "");
              for (var i in ar) arguments.callee(ar[i], folderSt, results);
              return results;
          })(DriveApp.getRootFolder(), "", []);
      };
    
      ggsrun.prototype.executionapi = function() {
          var startTime = Date.now();
          var dateDat = Utilities.formatDate(new Date(), "GMT", "yyyy-MM-dd_HH:mm:ss'_GMT'");
          var rec = JSON.parse(this.e);
          if (rec.log) {
              try {
                  recLog.call(this, [[
                      dateDat,
                      "{API: \"Execution API\", ContentLength: " + this.e.length + ", ExecutedFunction: \"" + rec.exefunc + "()\"}",
                      rec.com
                  ]]);
              } catch(err) {
                  var ss = err.message; // temporary
              }
          }
          var res = "";
          try {
              var resValues = (0,eval)((0,eval)(rec.com));
              res = emessage.call(this, rec.com ? resValues : "Error on GAS side: Bad parameters.", startTime, dateDat);
          } catch(err) {
              res = emessage.call(this, "Script Error on GAS side: " + err.message, startTime);
          }
          return res;
      };
  
      ggsrun.prototype.webapps = function() {
          var startTime = Date.now();
          var dateDat = Utilities.formatDate(new Date(), "GMT", "yyyy-MM-dd_HH:mm:ss'_GMT'");
          if (this.e.parameters.log == "false" || !this.e.parameters.log) {
              try {
                  recLog.call(this, [[
                      dateDat,
                      "{API: \"Web Apps\", ContentLength: " + this.e.contentLength + ", Password: \"" + this.e.parameters.pass + "\"}",
                      this.e.parameters.com
                  ]]);
              } catch(err) {
                  var ss = err.message; // temporary
              }
          }
          var res = "";
          if (this.e.parameters.pass == this.pass) {
              try {
                  var resValues = (0,eval)((0,eval)(this.e.parameters.com[0]));
                  res = wmessage.call(this, this.e.parameters.com ? resValues : "Error on GAS side: Bad parameters.", startTime, dateDat);
              } catch(err) {
                  res = wmessage.call(this, "Script Error on GAS side: " + err.message, startTime);
              }
          } else {
              res = wmessage.call(this, "Error on GAS side: Bad password.", startTime);
          }
          return res;
      };
  
      emessage = function(data, startTime, dateDat) {
          return {
              result: data,
              logger: this.logar,
              GoogleElapsedTime: ((Date.now() - startTime) / 1000),
              ScriptDate: dateDat
          };
      };
  
      wmessage = function(data, startTime, dateDat) {
          return ContentService
              .createTextOutput(JSON.stringify({
                  result: data,
                  logger: this.logar,
                  GoogleElapsedTime: ((Date.now() - startTime) / 1000),
                  ScriptDate: dateDat
              }))
              .setMimeType(ContentService.MimeType.JSON);
      };
  
      logsheet = function(_log) {
          var logit = DriveApp.getFilesByName(_log);
          var logar = [];
          while (logit.hasNext()) {
              logar.push(logit.next().getId());
          }
          if (logar.length == 0) {
              var ss = SpreadsheetApp.create(_log);
              var logss = DriveApp.getFileById(ss.getId());
              try {
                  DriveApp.getFileById(ScriptApp.getScriptId()).getParents().next().addFile(logss);
              } catch(e) {
                  DriveApp.getFileById(SpreadsheetApp.getActiveSpreadsheet().getId()).getParents().next().addFile(logss);
              }
              logss.getParents().next().removeFile(logss);
          } else {
              var ss = SpreadsheetApp.openById(logar[0]);
          }
          return ss.getSheets()[0];
      };
  
      recLog = function(logAr) {
          this.ss.getRange(this.ss.getLastRow() + 1, 1, logAr.length, logAr[0].length).setValues(logAr);
      };
    
      return ggsrun;
  })();
  return x.ggsrun = ggsrun;
})(this);
