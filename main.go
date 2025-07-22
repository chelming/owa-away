package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "regexp"
    "strings"
)

var (
    url        string
    calName    string
    eventTypes string
    re         = regexp.MustCompile(`(BEGIN:VEVENT.*?END:VEVENT%%)`)
    keywords   = []string{"SUMMARY:Away", "SUMMARY:Tentative", "SUMMARY:Free"}
)

func fetchCalendar(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    bodyBytes, err := ioutil.ReadAll(resp.Body)
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
		sentenceCasePart := strings.ToUpper(trimmedPart[0]) + strings.ToLower(trimmedPart[1:])
		result = append(result, "SUMMARY:"+sentenceCasePart)
	}
	return result
}

func processCalendar(bodyString string) string {
    bodyString = strings.Replace(bodyString, "\r\n", "%%", -1) // join to make regex easier
    matches := re.FindAllStringSubmatch(bodyString, -1)
    for _, match := range matches {
        fullMatch := match[0]
        for _, keyword := range keywords {
            if strings.Contains(fullMatch, keyword) {
                bodyString = strings.Replace(bodyString, fullMatch, "", -1)
                break
            }
        }
    }
    bodyString = strings.Replace(bodyString, "%%", "\r\n", -1) // split back to separate lines
    return bodyString
}

func handler(w http.ResponseWriter, r *http.Request) {
    bodyString, err := fetchCalendar(url)
    if err != nil {
        log.Println("Error fetching calendar:", err)
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
        log.Fatal("URL is not set.")
    }

    calName = os.Getenv("DISPLAY_NAME")
    if calName == "" {
        calName = "My Calendar"
    }

    eventTypes = os.Getenv("EVENT_TYPES")
    if eventTypes != "" {
        keywords = processEventTypes(eventTypes)
    }

    fmt.Fprint(os.Stdout, "Value of URL: \n", url)

    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
