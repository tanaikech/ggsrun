// === SANDBOX SECURITY GUARD INJECTED ===
const DriveApp = (function() {
  const original = this.DriveApp;
  if (!original) return null;
  const allowedFileIds = __ALLOWED_FILE_IDS__;
  const allowedFolderIds = __ALLOWED_FOLDER_IDS__;

  function wrapIterator(iter) {
    return {
      hasNext: function() { return iter.hasNext(); },
      next: function() {
        const item = iter.next();
        const id = item.getId();
        if (!allowedFileIds.includes(id) && !allowedFolderIds.includes(id)) {
          throw new Error("Sandbox Runtime Blocked: Accessed resource ID '" + id + "' is not whitelisted.");
        }
        return item;
      }
    };
  }

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'getFileById') {
        return function(id) {
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: File ID '" + id + "' is not in the whitelist.");
          }
          return original.getFileById(id);
        };
      }
      if (prop === 'getFolderById') {
        return function(id) {
          if (!allowedFolderIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Folder ID '" + id + "' is not in the whitelist.");
          }
          return original.getFolderById(id);
        };
      }
      if (prop === 'getRootFolder') {
        return function() {
          throw new Error("Sandbox Runtime Blocked: Direct access to Root Folder is prohibited.");
        };
      }
      if (prop === 'getFiles' || prop === 'searchFiles' || prop === 'getFilesByName') {
        return function(...args) {
          const rawIter = original[prop].apply(original, args);
          return wrapIterator(rawIter);
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'DriveApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const GmailApp = (function() {
  const original = this.GmailApp;
  if (!original) return null;
  const allowedEmails = __ALLOWED_EMAILS__;

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'sendEmail' || prop === 'createDraft') {
        return function(recipient, ...args) {
          if (!allowedEmails.includes(recipient)) {
            throw new Error("Sandbox Runtime Blocked: Recipient address '" + recipient + "' is not whitelisted.");
          }
          return original[prop].apply(original, [recipient, ...args]);
        };
      }
      if (prop === 'getInboxThreads' || prop === 'getSpamThreads' || prop === 'getTrashThreads' ||
          prop === 'search' || prop === 'getChatThreads' || prop === 'getStarredThreads') {
        return function() {
          throw new Error("Sandbox Runtime Blocked: Inbox scanning is prohibited.");
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'GmailApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const MailApp = (function() {
  const original = this.MailApp;
  if (!original) return null;
  const allowedEmails = __ALLOWED_EMAILS__;

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'sendEmail') {
        return function(recipient, ...args) {
          if (!allowedEmails.includes(recipient)) {
            throw new Error("Sandbox Runtime Blocked: Recipient address '" + recipient + "' is not whitelisted.");
          }
          return original.sendEmail(recipient, ...args);
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'MailApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const SpreadsheetApp = (function() {
  const original = this.SpreadsheetApp;
  if (!original) return null;
  const allowedFileIds = __ALLOWED_FILE_IDS__;

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'openById') {
        return function(id) {
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Spreadsheet ID '" + id + "' is not whitelisted.");
          }
          return original.openById(id);
        };
      }
      if (prop === 'openByUrl') {
        return function(url) {
          const match = url.match(/\/d\/([^\/]+)/);
          const id = match ? match[1] : url;
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Spreadsheet URL '" + url + "' is not whitelisted.");
          }
          return original.openByUrl(url);
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'SpreadsheetApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const DocumentApp = (function() {
  const original = this.DocumentApp;
  if (!original) return null;
  const allowedFileIds = __ALLOWED_FILE_IDS__;

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'openById') {
        return function(id) {
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Document ID '" + id + "' is not whitelisted.");
          }
          return original.openById(id);
        };
      }
      if (prop === 'openByUrl') {
        return function(url) {
          const match = url.match(/\/d\/([^\/]+)/);
          const id = match ? match[1] : url;
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Document URL '" + url + "' is not whitelisted.");
          }
          return original.openByUrl(url);
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'DocumentApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const SlidesApp = (function() {
  const original = this.SlidesApp;
  if (!original) return null;
  const allowedFileIds = __ALLOWED_FILE_IDS__;

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'openById') {
        return function(id) {
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Presentation ID '" + id + "' is not whitelisted.");
          }
          return original.openById(id);
        };
      }
      if (prop === 'openByUrl') {
        return function(url) {
          const match = url.match(/\/d\/([^\/]+)/);
          const id = match ? match[1] : url;
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Presentation URL '" + url + "' is not whitelisted.");
          }
          return original.openByUrl(url);
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'SlidesApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const CalendarApp = (function() {
  const original = this.CalendarApp;
  if (!original) return null;
  const allowedCalendarIds = __ALLOWED_CALENDAR_IDS__;
  const allowedEventIds = __ALLOWED_EVENT_IDS__;

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'getCalendarById') {
        return function(id) {
          if (!allowedCalendarIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Calendar ID '" + id + "' is not whitelisted.");
          }
          return original.getCalendarById(id);
        };
      }
      if (prop === 'getDefaultCalendar') {
        return function() {
          if (!allowedCalendarIds.includes('primary')) {
            throw new Error("Sandbox Runtime Blocked: Default Calendar is not whitelisted.");
          }
          return original.getDefaultCalendar();
        };
      }
      if (prop === 'getEventById') {
        return function(id) {
          if (!allowedEventIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Event ID '" + id + "' is not whitelisted.");
          }
          return original.getEventById(id);
        };
      }
      if (prop === 'getEventSeriesById') {
        return function(id) {
          if (!allowedEventIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Event Series ID '" + id + "' is not whitelisted.");
          }
          return original.getEventSeriesById(id);
        };
      }
      if (prop === 'getOwnedCalendars' || prop === 'getSharedCalendars' || prop === 'getCalendarsByName') {
        return function(...args) {
          const rawArr = original[prop].apply(original, args);
          return rawArr.filter(cal => allowedCalendarIds.includes(cal.getId()));
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'CalendarApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const FormApp = (function() {
  const original = this.FormApp;
  if (!original) return null;
  const allowedFileIds = __ALLOWED_FILE_IDS__;

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'openById') {
        return function(id) {
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Form ID '" + id + "' is not whitelisted.");
          }
          return original.openById(id);
        };
      }
      if (prop === 'openByUrl') {
        return function(url) {
          const match = url.match(/\/d\/([^\/]+)/);
          const id = match ? match[1] : url;
          if (!allowedFileIds.includes(id)) {
            throw new Error("Sandbox Runtime Blocked: Form URL '" + url + "' is not whitelisted.");
          }
          return original.openByUrl(url);
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'FormApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

const UrlFetchApp = (function() {
  const original = this.UrlFetchApp;
  if (!original) return null;
  const allowedUrls = __ALLOWED_URLS__;
  const blockedUrls = __BLOCKED_URLS__;

  function matchPattern(url, pattern) {
    const escaped = pattern.replace(/[-\/\\^$+?.()|[\]{}]/g, '\\$&');
    const regexStr = '^' + escaped.replace(/\*/g, '.*') + '$';
    const regex = new RegExp(regexStr, 'i');
    return regex.test(url);
  }

  function checkUrl(url) {
    const isBlocked = blockedUrls.some(pattern => matchPattern(url, pattern));
    if (isBlocked) {
      throw new Error("Sandbox Runtime Blocked: URL '" + url + "' is explicitly blacklisted.");
    }
    const isAllowed = allowedUrls.some(pattern => matchPattern(url, pattern));
    if (!isAllowed) {
      throw new Error("Sandbox Runtime Blocked: URL '" + url + "' is not whitelisted. Default policy is BLOCK ALL.");
    }
  }

  const proxy = new Proxy(original, {
    get(target, prop) {
      if (prop === 'fetch') {
        return function(url, ...args) {
          checkUrl(url);
          return original.fetch(url, ...args);
        };
      }
      if (prop === 'fetchAll') {
        return function(requests, ...args) {
          if (Array.isArray(requests)) {
            requests.forEach(req => {
              if (typeof req === 'string') {
                checkUrl(req);
              } else if (req && typeof req.url === 'string') {
                checkUrl(req.url);
              }
            });
          }
          return original.fetchAll(requests, ...args);
        };
      }
      const value = target[prop];
      return typeof value === 'function' ? value.bind(target) : value;
    }
  });
  try {
    Object.defineProperty(this, 'UrlFetchApp', { value: proxy, configurable: true, writable: true });
  } catch(e) {}
  return proxy;
})();

// === ADVANCED GOOGLE SERVICES SECURITY PROXIES ===
(function(global) {
  const allowedFileIds = __ALLOWED_FILE_IDS__;
  const allowedFolderIds = __ALLOWED_FOLDER_IDS__;
  const allowedCalendarIds = __ALLOWED_CALENDAR_IDS__;
  const allowedEmails = __ALLOWED_EMAILS__;

  const allAllowedIds = [].concat(allowedFileIds, allowedFolderIds, allowedCalendarIds, allowedEmails);

  function createAdvancedServiceProxy(originalService, serviceName) {
    if (!originalService) return null;

    function wrap(obj, path) {
      return new Proxy(obj, {
        get(target, prop) {
          const value = target[prop];
          if (typeof value === 'function') {
            return function(...args) {
              for (const arg of args) {
                if (typeof arg === 'string') {
                  const isPotentialId = (arg.length >= 20 && /^[a-zA-Z0-9-_]+$/.test(arg)) || 
                                        arg.includes('@') || 
                                        arg === 'primary';
                  
                  if (isPotentialId) {
                    const isWhitelisted = allAllowedIds.some(id => id === arg);
                    if (!isWhitelisted) {
                      throw new Error("Sandbox Runtime Blocked: Advanced Service '" + serviceName + "' accessed non-whitelisted ID: '" + arg + "'");
                    }
                  }
                }
              }
              return value.apply(target, args);
            };
          } else if (value !== null && typeof value === 'object') {
            return wrap(value, path + '.' + String(prop));
          }
          return value;
        }
      });
    }

    return wrap(originalService, serviceName);
  }

  const advancedServices = ['Drive', 'Sheets', 'Docs', 'Slides', 'Gmail', 'Calendar'];
  for (const serviceName of advancedServices) {
    if (global[serviceName]) {
      const proxy = createAdvancedServiceProxy(global[serviceName], serviceName);
      try {
        Object.defineProperty(global, serviceName, { value: proxy, configurable: true, writable: true });
      } catch(e) {}
    }
  }
})(typeof globalThis !== 'undefined' ? globalThis : this);
// === END OF SANDBOX SECURITY GUARD ===
