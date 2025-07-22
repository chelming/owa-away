package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	url                   string
	calName               string
	re                    = regexp.MustCompile(`(BEGIN:VEVENT.*?END:VEVENT%%)`)
	additionalDeleteRegex = regexp.MustCompile(`DTSTART(?:;TZID=[^:]+)?:\d{8}T\d{2}0100%%`)
	userDeleteRegexes     []*regexp.Regexp
	keywords              = []string{"SUMMARY:Away", "SUMMARY:Tentative", "SUMMARY:Free"}
)

func fetchCalendar(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

func processEventTypes(types string) []string {
	parts := strings.Split(types, ",")
	var result []string
	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if len(trimmedPart) == 0 {
			continue
		}
		firstChar := strings.ToUpper(string(trimmedPart[0]))
		rest := ""
		if len(trimmedPart) > 1 {
			rest = strings.ToLower(trimmedPart[1:])
		}
		sentenceCasePart := firstChar + rest
		result = append(result, "SUMMARY:"+sentenceCasePart)
	}
	return result
}

func processCalendar(bodyString string) string {
	bodyString = strings.Replace(bodyString, "\r\n", "%%", -1) // join to make regex easier
	matches := re.FindAllStringSubmatch(bodyString, -1)
	for _, match := range matches {
		fullMatch := match[0]
		shouldDelete := false
		// Check keywords
		for _, keyword := range keywords {
			if strings.Contains(fullMatch, keyword) {
				shouldDelete = true
				break
			}
		}
		// Check built-in additional regex
		if !shouldDelete && additionalDeleteRegex.MatchString(fullMatch) {
			shouldDelete = true
		}
		// Check user-provided regexes
		if !shouldDelete {
			for _, userRe := range userDeleteRegexes {
				if userRe.MatchString(fullMatch) {
					shouldDelete = true
					break
				}
			}
		}
		if shouldDelete {
			bodyString = strings.Replace(bodyString, fullMatch, "", -1)
			continue
		}

		// If SUMMARY:Tentative, set STATUS:TENTATIVE
		if strings.Contains(fullMatch, "SUMMARY:Tentative") {
			// Check if STATUS already exists
			statusRe := regexp.MustCompile(`STATUS:[^%]*%%`)
			if statusRe.MatchString(fullMatch) {
				// Replace existing STATUS
				newMatch := statusRe.ReplaceAllString(fullMatch, "STATUS:TENTATIVE%%")
				bodyString = strings.Replace(bodyString, fullMatch, newMatch, 1)
			} else {
				// Insert STATUS:TENTATIVE after SUMMARY:Tentative
				summaryRe := regexp.MustCompile(`(SUMMARY:Tentative%%)`)
				newMatch := summaryRe.ReplaceAllString(fullMatch, "$1STATUS:TENTATIVE%%")
				bodyString = strings.Replace(bodyString, fullMatch, newMatch, 1)
			}
		}
	}
	bodyString = strings.Replace(bodyString, "%%", "\r\n", -1) // split back to separate lines
	return bodyString
}

func syslogLog(level, msg string) {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	log.Printf("%s %s %s", timestamp, level, msg)
}

func handler(w http.ResponseWriter, r *http.Request) {
	syslogLog("INFO", fmt.Sprintf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))

	bodyString, err := fetchCalendar(url)
	if err != nil {
		syslogLog("ERR", fmt.Sprintf("Error fetching calendar: %v", err))
		return
	}

	bodyString = processCalendar(bodyString)
	reCalName := regexp.MustCompile(`X-WR-CALNAME:Calendar`)
	bodyString = reCalName.ReplaceAllString(bodyString, fmt.Sprintf("X-WR-CALNAME:%s", calName)) // rename the calendar

	fmt.Fprint(w, bodyString)
}

func main() {
	url = os.Getenv("URL")
	if url == "" {
		syslogLog("CRIT", "URL is not set.")
		log.Fatal("URL is not set.")
	}

	calName = os.Getenv("DISPLAY_NAME")
	if calName == "" {
		calName = "My Calendar"
	}

	eventTypes, eventTypesSet := os.LookupEnv("EVENT_TYPES")
	if eventTypesSet {
		keywords = processEventTypes(eventTypes)
	}

	// Load user-provided delete regexes from environment variable
	deleteRegexEnv := os.Getenv("DELETE_REGEXES")
	if deleteRegexEnv != "" {
		regexStrings := strings.Split(deleteRegexEnv, ",")
		for _, regexStr := range regexStrings {
			trimmed := strings.TrimSpace(regexStr)
			if trimmed == "" {
				continue
			}
			re, err := regexp.Compile(trimmed)
			if err != nil {
				syslogLog("ERR", fmt.Sprintf("Invalid regex in DELETE_REGEXES: %s, error: %v", trimmed, err))
				continue
			}
			userDeleteRegexes = append(userDeleteRegexes, re)
		}
	}

	syslogLog("INFO", fmt.Sprintf("Value of URL: %s", url))

	http.HandleFunc("/", handler)
	syslogLog("INFO", "Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
