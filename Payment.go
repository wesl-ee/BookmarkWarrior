package main

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"bytes"
	"net/http"
	"encoding/json"
	"strconv"
	"strings"
)

type PayPalOrder struct {
	ID string `json:"id"`
	PurchaseUnits []PayPalPurchaseUnit `json:"purchase_units"`
	Status string `json:status`
}

type PayPalAuth struct {
	AccessToken string `json:"access_token"`
}

type PayPalPurchaseUnit struct {
	Amount PayPalAmount `json:amount`
}

type PayPalAmount struct {
	CurrencyCode string `json:"currency_code"`
	Value string `json:value`
}

func VerifyPayment(orderID string) (bool, error) {
	client := http.Client{}

	accessToken := PayPalOAuthToken(client)

	endpoint := Settings.PayPal.OrderAPI + orderID
	headers := map[string]string{
		"Accept": "application/json",
		"Authorization": "Bearer " + accessToken }

	var order PayPalOrder
	res, err := PayPalGet(client, endpoint, headers, nil)
	if err != nil { panic(err) }
	json.Unmarshal(res, &order)

	if IsAdequatePayment(order) {
		return true, nil
	} else { return false, nil }
}

func PayPalOAuthToken(client http.Client) (string) {
	rBody := bytes.NewBuffer([]byte("grant_type=client_credentials"))
	endpoint := Settings.PayPal.OAuthAPI
	headers := map[string]string{
		"Accept": "application/json",
		"Authorization": "Basic " + PayPalBasicAuth() }

	var auth PayPalAuth
	res, err := PayPalPost(client, endpoint, headers, rBody)
	if err != nil { panic(err) }
	json.Unmarshal(res, &auth)
	accessToken := auth.AccessToken
	return accessToken
}

func IsAdequatePayment(order PayPalOrder) (bool) {
	var ought = Settings.PayPal.OneTimeCost

	var total float64
	for _, unit := range order.PurchaseUnits {
		val, _ := strconv.ParseFloat(unit.Amount.Value, 64)
		total += val
	}
	return total >= ought
}

func PayPalGet(client http.Client,
	endpoint string,
	headers map[string]string,
	body io.Reader) ([]byte, error) {
	return DoPayPalRequest(client,
		http.MethodGet,
		endpoint,
		headers,
		body)
}

func PayPalPost(client http.Client,
	endpoint string,
	headers map[string]string,
	body io.Reader) ([]byte, error) {
	return DoPayPalRequest(client,
		http.MethodPost,
		endpoint,
		headers,
		body)
}

func DoPayPalRequest(client http.Client,
	method string,
	endpoint string,
	headers map[string]string,
	body io.Reader) ([]byte, error) {

	if body == nil { body = http.NoBody }
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil { panic(err) }
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	resp, err := client.Do(req)
	if err != nil { panic(err) }
	defer resp.Body.Close()

	// json.NewDecoder(resp.Body).Decode(&r)

	resBody, err := ioutil.ReadAll(resp.Body)
	return resBody, err
}



func PayPalBasicAuth() (string) {
	auth := []byte(strings.Join([]string{
		Settings.PayPal.Client,
		Settings.PayPal.Secret}, ":"))

	return string(base64.StdEncoding.EncodeToString(auth))
}
