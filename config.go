package main

import (
	"os"
	"strings"
)

// Config represents central configuration of this app. Should only be used in
// the app entrypoint and not handed down to individual components / functions.
//
// To instantiate a Config use the NewConfig function.
type Config struct {
	// Core configuration.
	serverPort string

	// Token extraction.
	tokenHeaderNames    []string
	addTokenHeaderNames []string
	fallbackToken       string

	// User interface.
	uiTarget string
	uiTitle  string
	uiDesc1  string
	uiDesc2  string
	uiMisc   string
}

// NewConfig inits config struct. Values are retrieved from environments
// variables. Includes internal defaults.
func NewConfig() Config {
	c := Config{}

	// Core configuration.
	c.serverPort = GetEnv("SERVER_PORT", "8080")

	// Token extraction.
	c.tokenHeaderNames = SplitToSlice(GetEnv("TOKEN_HEADER_NAMES",
		strings.Join([]string{
			"Access-Token",
			"Authorization",
			"Token",
			"X-Auth-Request-Access-Token",
			"X-Forwarded-Access-Token",
		}, ","),
	))
	c.addTokenHeaderNames = SplitToSlice(GetEnv("ADD_TOKEN_HEADER_NAMES", ""))
	c.fallbackToken = GetEnv("FALLBACK_TOKEN", "")

	// User interface.
	c.uiTarget = GetEnv("UI_TARGET", "")
	c.uiTitle = GetEnv("UI_TITLE", "")
	c.uiDesc1 = GetEnv("UI_DESC1", "")
	c.uiDesc2 = GetEnv("UI_DESC2", "")
	c.uiMisc = GetEnv("UI_MISC", "")

	return c
}

// GetEnv gets environment variable value after prefixing the key. Default value
// in case of absence must be provided.
//
// Used for retrieving configuration provided by environment variables and
// enforcing a common prefix.
func GetEnv(key, def string) string {
	v := os.Getenv("T2G_" + key)

	if v == "" {
		return def
	}

	return v
}

// SplitToSlice splits string by commas into a slice. Resulting items are space
// trimmed. Empty string items are removed. Finally, the slice is returned.
func SplitToSlice(str string) []string {
	var r []string

	for _, e := range strings.Split(str, ",") {
		trimmed := strings.TrimSpace(e)
		if len(trimmed) > 0 {
			r = append(r, strings.TrimSpace(e))
		}
	}

	return r
}
