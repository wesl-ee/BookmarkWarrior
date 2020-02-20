package main

import (
	"os"
	"log"
	"errors"
	"github.com/BurntSushi/toml"
)

type Config struct {
	MaxUsernameLength int
	MaxDisplaynameLength int
	MinimumPasswordLength int
	Web WebSettings
	Database DBSettings
	PayPal PayPalSettings
	Templates []TemplateSettings }

type DBSettings struct {
	ConnectionString string
	DatetimeFormat string }

type TemplateSettings struct {
	Name string
	Dependencies []string }

type PayPalSettings struct {
	OAuthAPI string
	OrderAPI string
	Client string
	Secret string
	OneTimeCost float64
	DomesticCurrency string
	DomesticCurrencySigil string
}

type WebSettings struct {
	Canon string
	SessionCookie string
	SessionExpiryDays int
	Host string
	DateFormat string }

var CONFIG_DEFAULT_LOCS = [...]string{
	"Config.toml" }

func ReadDefaultConfig(c *Config) (error) {
	for _, f := range CONFIG_DEFAULT_LOCS {
		if FExists(f) {
			return ReadConfig(c, f)
		}
	}

	return errors.New("Could not find config file!")
}

func ReadConfig(c *Config, fpath string) (err error) {
	log.Printf("Loading configuration file: %s\n", fpath)
	_, err = toml.DecodeFile(fpath, c)
	return
}

func FExists(fname string) bool {
	info, err := os.Stat(fname)
	if os.IsNotExist(err) { return false }
	return !info.IsDir()
}
