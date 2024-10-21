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

func handler(w http.ResponseWriter, r *http.Request) {
    resp, err := http.Get(os.Getenv("URL"))
    if err != nil {
        // handle error
        log.Fatal(err)
    }
    defer resp.Body.Close()
    bodyBytes, err := ioutil.ReadAll(resp.Body)
    bodyString := string(bodyBytes)
    bodyString = strings.Replace(bodyString, "\r\n", "%%", -1) // join to make regex easier
    re := regexp.MustCompile(`(BEGIN:VEVENT.*?END:VEVENT%%)`)
    matches := re.FindAllStringSubmatch(bodyString, -1)
    for _, match := range matches {
	   if (strings.Contains(match[0], "SUMMARY:Away") || strings.Contains(match[0], "SUMMARY:Tentative") || strings.Contains(match[0], "SUMMARY:Free")) {
               bodyString = strings.Replace(bodyString, match[0], "", -1)
	    } 
    }
    bodyString = strings.Replace(bodyString, "%%", "\r\n", -1) // split back to separate lines

    newCalName := fmt.Sprintf("%s%s", "X-WR-CALNAME:", os.Getenv("DISPLAY_NAME"))
    bodyString = strings.Replace(bodyString, "X-WR-CALNAME:Calendar", newCalName, -1) // rename the calendar

    fmt.Fprint(w, bodyString)
}

func main() {
    envVar := os.Getenv("URL")

    if envVar != "" {
        fmt.Fprint(os.Stdout, "Value of URL: \n", envVar)
    } else {
        fmt.Fprint(os.Stderr, "URL is not set.")	
    }

    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
