package main

import (
	"net/http"
	"io/ioutil"
	"strings"
)

func ShortTitle(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err }
	defer resp.Body.Close()

	dataInBytes, err := ioutil.ReadAll(resp.Body)
	pageContent := string(dataInBytes)

	titleStartIndex := strings.Index(pageContent, "<title>")
	if titleStartIndex < 0 {
		return "", nil }
	titleStartIndex += 7

	titleEndIndex := strings.Index(pageContent, "</title>")
	if titleEndIndex < 0 {
		return "", nil }
	pageTitle := string([]byte(pageContent[titleStartIndex:titleEndIndex]))

	return pageTitle, nil
}
