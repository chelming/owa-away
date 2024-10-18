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
    bodyString = strings.Replace(bodyString, "\n", "~", -1)
    re := regexp.MustCompile("(BEGIN:VEVENT.*?END:VEVENT~)")
    matches := re.FindAllStringSubmatch(bodyString, -1)
    for _, match := range matches {
	    if strings.Contains(match[0], "SUMMARY:Away") {
               bodyString = strings.Replace(bodyString, match[0], "", -1)
	    } 
    }
    fmt.Fprint(w, strings.Replace(bodyString, "~", "\n", -1))
}

func main() {
    envVar := os.Getenv("URL")

    if envVar != "" {
        fmt.Fprint(os.Stdout, "Value of URL: ", envVar)
    } else {
        fmt.Fprint(os.Stderr, "URL is not set.")	
    }

    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
