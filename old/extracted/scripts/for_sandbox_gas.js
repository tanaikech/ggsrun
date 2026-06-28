// === SANDBOX SECURITY GUARD INJECTED ===
function createSafeWrapper(original, overrides) {
  if (!original) return null;
  var wrapper = {
    _isProxied: true
  };

  var proto = original;
  while (proto && proto !== Object.prototype) {
    var props = Object.getOwnPropertyNames(proto);
    for (var i = 0; i < props.length; i++) {
      (function(prop) {
        if (prop === 'constructor' || wrapper.hasOwnProperty(prop)) return;
        try {
          var desc = Object.getOwnPropertyDescriptor(proto, prop);
          if (desc && (desc.get || desc.set)) {
            Object.defineProperty(wrapper, prop, {
              get: function() { return original[prop]; },
              set: function(val) { original[prop] = val; },
              configurable: true,
              enumerable: true
            });
            return;
          }
          var value = original[prop];
          if (typeof value === 'function') {
            wrapper[prop] = function(...args) {
              return value.apply(original, args);
            };
          } else {
            Object.defineProperty(wrapper, prop, {
              get: function() { return original[prop]; },
              set: function(val) { original[prop] = val; },
              configurable: true,
              enumerable: true
            });
          }
        } catch(e) {}
      })(props[i]);
    }
    proto = Object.getPrototypeOf(proto);
  }

  for (var key in overrides) {
    if (overrides.hasOwnProperty(key)) {
      wrapper[key] = overrides[key];
    }
  }

  return wrapper;
}

var _wrappedDriveApp = (function(global) {
  var allowedFileIds = __ALLOWED_FILE_IDS__;
  var allowedFolderIds = __ALLOWED_FOLDER_IDS__;

  function wrapIterator(iter) {
    return {
      hasNext: function() { return iter.hasNext(); },
      next: function() {
        var item = iter.next();
        var id = item.getId();
        if (!allowedFileIds.includes(id) && !allowedFolderIds.includes(id)) {
          throw new Error("Sandbox Runtime Blocked: Accessed resource ID '" + id + "' is not whitelisted.");
        }
        return item;
      }
    };
  }

  return createSafeWrapper(DriveApp, {
    getFileById: function(id) {
      if (!allowedFileIds.includes(id)) {
        throw new Error("Sandbox Runtime Blocked: File ID '" + id + "' is not in the whitelist.");
      }
      return DriveApp.getFileById(id);
    },
    getFolderById: function(id) {
      if (!allowedFolderIds.includes(id)) {
        throw new Error("Sandbox Runtime Blocked: Folder ID '" + id + "' is not in the whitelist.");
      }
      return DriveApp.getFolderById(id);
    },
    getRootFolder: function() {
      throw new Error("Sandbox Runtime Blocked: Direct access to Root Folder is prohibited.");
    },
    getFiles: function() { return wrapIterator(DriveApp.getFiles()); },
    getFilesByName: function(name) { return wrapIterator(DriveApp.getFilesByName(name)); },
    searchFiles: function(query) { return wrapIterator(DriveApp.searchFiles(query)); }
  });
})(typeof globalThis !== 'undefined' ? globalThis : this);

var _wrappedGmailApp = (function(global) {
  var allowedEmails = __ALLOWED_EMAILS__;

  return createSafeWrapper(GmailApp, {
    sendEmail: function(recipient, ...args) {
      if (!allowedEmails.includes(recipient)) {
        throw new Error("Sandbox Runtime Blocked: Recipient address '" + recipient + "' is not whitelisted.");
      }
      return GmailApp.sendEmail.apply(GmailApp, [recipient, ...args]);
    },
    createDraft: function(recipient, ...args) {
      if (!allowedEmails.includes(recipient)) {
        throw new Error("Sandbox Runtime Blocked: Recipient address '" + recipient + "' is not whitelisted.");
      }
      return GmailApp.createDraft.apply(GmailApp, [recipient, ...args]);
    },
    getInboxThreads: function() { throw new Error("Sandbox Runtime Blocked: Inbox scanning is prohibited."); },
    getSpamThreads: function() { throw new Error("Sandbox Runtime Blocked: Inbox scanning is prohibited."); },
    getTrashThreads: function() { throw new Error("Sandbox Runtime Blocked: Inbox scanning is prohibited."); },
    search: function() { throw new Error("Sandbox Runtime Blocked: Inbox scanning is prohibited."); },
    getChatThreads: function() { throw new Error("Sandbox Runtime Blocked: Inbox scanning is prohibited."); },
    getStarredThreads: function() { throw new Error("Sandbox Runtime Blocked: Inbox scanning is prohibited."); }
  });
})(typeof globalThis !== 'undefined' ? globalThis : this);

var _wrappedMailApp = (function(global) {
  var allowedEmails = __ALLOWED_EMAILS__;

  return createSafeWrapper(MailApp, {
    sendEmail: function(recipient, ...args) {
      if (!allowedEmails.includes(recipient)) {
        throw new Error("Sandbox Runtime Blocked: Recipient address '" + recipient + "' is not whitelisted.");
      }
      return MailApp.sendEmail.apply(MailApp, [recipient, ...args]);
    }
  });
})(typeof globalThis !== 'undefined' ? globalThis : this);

var _wrappedSpreadsheetApp = (function(global) {
  var allowedFileIds = __ALLOWED_FILE_IDS__;

  return createSafeWrapper(SpreadsheetApp, {
    openById: function(id) {
      if (!allowedFileIds.includes(id)) {
        throw new Error("Sandbox Runtime Blocked: Spreadsheet ID '" + id + "' is not whitelisted.");
      }
      return SpreadsheetApp.openById(id);
    },
    openByUrl: function(url) {
      var match = url.match(/\/d\/([^\/]+)/);
      var id = match ? match[1] : url;
      if (!allowedFileIds.includes(id)) {
        throw new Error("Sandbox Runtime Blocked: Spreadsheet URL '" + url + "' is not whitelisted.");
      }
      return SpreadsheetApp.openByUrl(url);
    }
  });
})(typeof globalThis !== 'undefined' ? globalThis : this);

var _wrappedDocumentApp = (function(global) {
  var allowedFileIds = __ALLOWED_FILE_IDS__;

  return createSafeWrapper(DocumentApp, {
    openById: function(id) {
      if (!allowedFileIds.includes(id)) {
        throw new Error("Sandbox Runtime Blocked: Document ID '" + id + "' is not whitelisted.");
      }
      return DocumentApp.openById(id);
    },
    openByUrl: function(url) {
      var match = url.match(/\/d\/([^\/]+)/);
      var id = match ? match[1] : url;
      if (!allowedFileIds.includes(id)) {
        throw new Error("Sandbox Runtime Blocked: Document URL '" + url + "' is not whitelisted.");
      }
      return DocumentApp.openByUrl(url);
    }
  });
})(typeof globalThis !== 'undefined' ? globalThis : this);

var _wrappedSlidesApp = (function(global) {
  var allowedFileIds = __ALLOWED_FILE_IDS__;

  return createSafeWrapper(SlidesApp, {
    openById: function(id) {
      if (!allowedFileIds.includes(id)) {
        throw new Error("Sandbox Runtime Blocked: Presentation ID '" + id + "' is not whitelisted.");
      }
      return SlidesApp.openById(id);
    },
    openByUrl: function(url) {
      var match = url.match(/\/d\/([^\/]+)/);
      var id = match ? match[1] : url;
      if (!allowedFileIds.includes(id)) {
        throw new Error("Sandbox Runtime Blocked: Presentation URL '" + url + "' is not whitelisted.");
      }
      return SlidesApp.openByUrl(url);
    }
  });
})(typeof globalThis !== 'undefined' ? globalThis : this);

var _wrappedCalendarApp = (function(global) {
  var allowedCalendarIds = __ALLOWED_CALENDAR_IDS__;
  var allowedEventIds = __ALLOWED_EVENT_IDS__;

  return createSafeWrapper(CalendarApp, {
    getCalendarById: function(id) {
      if (!allowedCalendarIds.includes(id)) {
        throw new Error("Sandbox Runtime Blocked: Calendar ID '" + id + "' is not whitelisted.");
      }
      return CalendarApp.getCalendarById(id);
    },
    getDefaultCalendar: function() {
      if (!allowedCalendarIds.includes('primary')) {
        throw new Error("Sandbox Runtime Blocked: Default Calendar is not whitelisted.");
      }
      return CalendarApp.getDefaultCalendar();
    },
    getEventById: function(id) {
      if (!allowedEventIds.includes(id)) {
        throw new Error("Sandbox Runtime Blocked: Event ID '" + id + "' is not whitelisted.");
      }
      return CalendarApp.getEventById(id);
    },
    getEventSeriesById: function(id) {
      if (!allowedEventIds.includes(id)) {
        throw new Error("Sandbox Runtime Blocked: Event Series ID '" + id + "' is not whitelisted.");
      }
      return CalendarApp.getEventSeriesById(id);
    },
    getOwnedCalendars: function() {
      return CalendarApp.getOwnedCalendars().filter(cal => allowedCalendarIds.includes(cal.getId()));
    },
    getSharedCalendars: function() {
      return CalendarApp.getSharedCalendars().filter(cal => allowedCalendarIds.includes(cal.getId()));
    },
    getCalendarsByName: function(name) {
      return CalendarApp.getCalendarsByName(name).filter(cal => allowedCalendarIds.includes(cal.getId()));
    }
  });
})(typeof globalThis !== 'undefined' ? globalThis : this);

var _wrappedFormApp = (function(global) {
  var allowedFileIds = __ALLOWED_FILE_IDS__;

  return createSafeWrapper(FormApp, {
    openById: function(id) {
      if (!allowedFileIds.includes(id)) {
        throw new Error("Sandbox Runtime Blocked: Form ID '" + id + "' is not whitelisted.");
      }
      return FormApp.openById(id);
    },
    openByUrl: function(url) {
      var match = url.match(/\/d\/([^\/]+)/);
      var id = match ? match[1] : url;
      if (!allowedFileIds.includes(id)) {
        throw new Error("Sandbox Runtime Blocked: Form URL '" + url + "' is not whitelisted.");
      }
      return FormApp.openByUrl(url);
    }
  });
})(typeof globalThis !== 'undefined' ? globalThis : this);

var _wrappedUrlFetchApp = (function(global) {
  var allowedUrls = __ALLOWED_URLS__;
  var blockedUrls = __BLOCKED_URLS__;

  function matchPattern(url, pattern) {
    var escaped = pattern.replace(/[-\/\\^$+?.()|[\]{}]/g, '\\$&');
    var regexStr = '^' + escaped.replace(/\*/g, '.*') + '$';
    var regex = new RegExp(regexStr, 'i');
    return regex.test(url);
  }

  function checkUrl(url) {
    var isBlocked = blockedUrls.some(pattern => matchPattern(url, pattern));
    if (isBlocked) {
      throw new Error("Sandbox Runtime Blocked: URL '" + url + "' is explicitly blacklisted.");
    }
    var isAllowed = allowedUrls.some(pattern => matchPattern(url, pattern));
    if (!isAllowed) {
      throw new Error("Sandbox Runtime Blocked: URL '" + url + "' is not whitelisted. Default policy is BLOCK ALL.");
    }
  }

  return createSafeWrapper(UrlFetchApp, {
    fetch: function(url, ...args) {
      checkUrl(url);
      return UrlFetchApp.fetch.apply(UrlFetchApp, [url, ...args]);
    },
    fetchAll: function(requests, ...args) {
      if (Array.isArray(requests)) {
        requests.forEach(req => {
          if (typeof req === 'string') {
            checkUrl(req);
          } else if (req && typeof req.url === 'string') {
            checkUrl(req.url);
          }
        });
      }
      return UrlFetchApp.fetchAll.apply(UrlFetchApp, [requests, ...args]);
    }
  });
})(typeof globalThis !== 'undefined' ? globalThis : this);

// === ADVANCED GOOGLE SERVICES SECURITY WRAPPERS ===
var _wrappedDrive, _wrappedSheets, _wrappedDocs, _wrappedSlides, _wrappedGmail, _wrappedCalendar;
(function(global) {
  var allowedFileIds = __ALLOWED_FILE_IDS__;
  var allowedFolderIds = __ALLOWED_FOLDER_IDS__;
  var allowedCalendarIds = __ALLOWED_CALENDAR_IDS__;
  var allowedEmails = __ALLOWED_EMAILS__;

  var allAllowedIds = [].concat(allowedFileIds, allowedFolderIds, allowedCalendarIds, allowedEmails);

  function wrapAdvancedService(originalService, serviceName) {
    if (!originalService) return null;

    var overrides = {};
    
    var proto = originalService;
    while (proto && proto !== Object.prototype) {
      var props = Object.getOwnPropertyNames(proto);
      for (var i = 0; i < props.length; i++) {
        (function(prop) {
          if (prop === 'constructor') return;
          try {
            var value = originalService[prop];
            if (typeof value === 'function') {
              overrides[prop] = function(...args) {
                for (var arg of args) {
                  if (typeof arg === 'string') {
                    var isPotentialId = (arg.length >= 20 && /^[a-zA-Z0-9-_]+$/.test(arg)) || 
                                          arg.includes('@') || 
                                          arg === 'primary';
                    
                    if (isPotentialId) {
                      var isWhitelisted = allAllowedIds.some(id => id === arg);
                      if (!isWhitelisted) {
                        throw new Error("Sandbox Runtime Blocked: Advanced Service '" + serviceName + "' accessed non-whitelisted ID: '" + arg + "'");
                      }
                    }
                  }
                }
                return value.apply(originalService, args);
              };
            }
          } catch(e) {}
        })(props[i]);
      }
      proto = Object.getPrototypeOf(proto);
    }

    return createSafeWrapper(originalService, overrides);
  }

  try { _wrappedDrive = wrapAdvancedService(Drive, 'Drive'); } catch(e) {}
  try { _wrappedSheets = wrapAdvancedService(Sheets, 'Sheets'); } catch(e) {}
  try { _wrappedDocs = wrapAdvancedService(Docs, 'Docs'); } catch(e) {}
  try { _wrappedSlides = wrapAdvancedService(Slides, 'Slides'); } catch(e) {}
  try { _wrappedGmail = wrapAdvancedService(Gmail, 'Gmail'); } catch(e) {}
  try { _wrappedCalendar = wrapAdvancedService(Calendar, 'Calendar'); } catch(e) {}
})(typeof globalThis !== 'undefined' ? globalThis : this);
// === END OF SANDBOX SECURITY GUARD ===
