package main

import (
	"net/http"
	"io/ioutil"
	"strings"
)

func ShortTitle(url string) (string) {
	resp, err := http.Get(url)
	if err != nil { return "" }
	defer resp.Body.Close()

	// Pages that are not OK don't matter
	if resp.StatusCode != http.StatusOK { return "" }

	if resp.ContentLength > 10*1024*1024 { return "" }

	dataInBytes, err := ioutil.ReadAll(resp.Body)
	pageContent := string(dataInBytes)

	titleStartIndex := strings.Index(pageContent, "<title>")
	if titleStartIndex < 0 { return "" }
	titleStartIndex += 7

	titleEndIndex := strings.Index(pageContent, "</title>")
	if titleEndIndex < 0 { return "" }
	pageTitle := string([]byte(pageContent[titleStartIndex:titleEndIndex]))

	return pageTitle
}
