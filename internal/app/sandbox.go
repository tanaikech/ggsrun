package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

//go:embed for_sandbox_gas.js
var sandboxGuardTemplate string

type SandboxConfig struct {
	AllowedFileIds     []string `json:"allowedFileIds"`
	AllowedFolderIds   []string `json:"allowedFolderIds"`
	AllowedCalendarIds []string `json:"allowedCalendarIds"`
	AllowedEventIds    []string `json:"allowedEventIds"`
	AllowedEmails      []string `json:"allowedEmails"`
	AllowedUrls        []string `json:"allowedUrls"`
	BlockedUrls        []string `json:"blockedUrls"`
}

func ensureSlices(config *SandboxConfig) {
	if config.AllowedFileIds == nil {
		config.AllowedFileIds = []string{}
	}
	if config.AllowedFolderIds == nil {
		config.AllowedFolderIds = []string{}
	}
	if config.AllowedCalendarIds == nil {
		config.AllowedCalendarIds = []string{}
	}
	if config.AllowedEventIds == nil {
		config.AllowedEventIds = []string{}
	}
	if config.AllowedEmails == nil {
		config.AllowedEmails = []string{}
	}
	if config.AllowedUrls == nil {
		config.AllowedUrls = []string{}
	}
	if config.BlockedUrls == nil {
		config.BlockedUrls = []string{}
	}
}

func InjectSandbox(rawScript string, configPath string) (string, error) {
	if configPath == "bypass" || configPath == "none" {
		return rawScript, nil
	}

	var config SandboxConfig
	if configPath == "" {
		// Use default empty (strictly blocked) whitelists
	} else {
		content, err := os.ReadFile(configPath)
		if err != nil {
			return "", fmt.Errorf("failed to read sandbox config file: %w", err)
		}

		if err := json.Unmarshal(content, &config); err != nil {
			return "", fmt.Errorf("failed to parse sandbox config JSON: %w", err)
		}
	}

	ensureSlices(&config)

	allowedFileIdsJSON, _ := json.Marshal(config.AllowedFileIds)
	allowedFolderIdsJSON, _ := json.Marshal(config.AllowedFolderIds)
	allowedCalendarIdsJSON, _ := json.Marshal(config.AllowedCalendarIds)
	allowedEventIdsJSON, _ := json.Marshal(config.AllowedEventIds)
	allowedEmailsJSON, _ := json.Marshal(config.AllowedEmails)
	allowedUrlsJSON, _ := json.Marshal(config.AllowedUrls)
	blockedUrlsJSON, _ := json.Marshal(config.BlockedUrls)

	guardCode := sandboxGuardTemplate
	guardCode = strings.ReplaceAll(guardCode, "__ALLOWED_FILE_IDS__", string(allowedFileIdsJSON))
	guardCode = strings.ReplaceAll(guardCode, "__ALLOWED_FOLDER_IDS__", string(allowedFolderIdsJSON))
	guardCode = strings.ReplaceAll(guardCode, "__ALLOWED_CALENDAR_IDS__", string(allowedCalendarIdsJSON))
	guardCode = strings.ReplaceAll(guardCode, "__ALLOWED_EVENT_IDS__", string(allowedEventIdsJSON))
	guardCode = strings.ReplaceAll(guardCode, "__ALLOWED_EMAILS__", string(allowedEmailsJSON))
	guardCode = strings.ReplaceAll(guardCode, "__ALLOWED_URLS__", string(allowedUrlsJSON))
	guardCode = strings.ReplaceAll(guardCode, "__BLOCKED_URLS__", string(blockedUrlsJSON))

	// Static token replacement for sandboxed execution
	replacements := map[string]string{
		"DriveApp":       "_wrappedDriveApp",
		"GmailApp":       "_wrappedGmailApp",
		"MailApp":        "_wrappedMailApp",
		"SpreadsheetApp": "_wrappedSpreadsheetApp",
		"DocumentApp":    "_wrappedDocumentApp",
		"SlidesApp":      "_wrappedSlidesApp",
		"CalendarApp":    "_wrappedCalendarApp",
		"FormApp":        "_wrappedFormApp",
		"UrlFetchApp":    "_wrappedUrlFetchApp",
		"Drive":          "_wrappedDrive",
		"Sheets":         "_wrappedSheets",
		"Docs":           "_wrappedDocs",
		"Slides":         "_wrappedSlides",
		"Gmail":          "_wrappedGmail",
		"Calendar":       "_wrappedCalendar",
	}

	for orig, repl := range replacements {
		re := regexp.MustCompile(`\b` + orig + `\b`)
		rawScript = re.ReplaceAllString(rawScript, repl)
	}

	return guardCode + "\n" + rawScript, nil
}
