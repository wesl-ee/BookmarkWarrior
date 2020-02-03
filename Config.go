package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	MinimumPasswordLength int
	Web WebSettings
	Database DBSettings
	PayPal PayPalSettings
	Templates []TemplateSettings }

type DBSettings struct {
	ConnectionString string }

type TemplateSettings struct {
	Name string
	Dependencies []string }

type PayPalSettings struct {
	OAuthAPI string
	OrderAPI string
	Client string
	Secret string
	OneTimeCost float64
}

type WebSettings struct {
	Canon string
	SessionCookie string
	Host string }

const SYSDEFAULT_CONFIG string = "Config.toml"

func ReadDefaultConfig(c *Config) (error) {
	return ReadConfig(c, SYSDEFAULT_CONFIG); }

func ReadConfig(c *Config, fpath string) (err error) {
	_, err = toml.DecodeFile(fpath, c)
	return
}
