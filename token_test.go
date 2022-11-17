package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestTokenMarshalToJSON(t *testing.T) {
	b, err := json.Marshal(Token{"x", "x", "x"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	str := string(b)
	for _, substr := range []string{`"timestamp"`, `"fingerprint"`, `"secret"`} {
		if !strings.Contains(str, substr) {
			t.Errorf("Failed to find in marshalled JSON: str %q, substr %q", str, substr)
		}
	}
}

func TestNewToken(t *testing.T) {
	token := NewToken("mysecret")
	if len(token.Fingerprint) == 0 {
		t.Error(`Field "Fingerprint" must be set.`)
	}
	if len(token.Secret) == 0 {
		t.Error(`Field "Secret" must be set.`)
	}
	if len(token.Timestamp) == 0 {
		t.Error(`Field "Timestamp" must be set.`)
	}
}

func TestExtractToken(t *testing.T) {
	for _, tc := range []struct {
		name             string
		headers          http.Header
		tokenHeaderNames []string
		fallbackToken    string
		expectedSecret   string
		expectedError    bool
	}{{
		name: "1_case_sensitive",
		headers: http.Header{
			"Baz":   {"c"},
			"Token": {"d"}},
		tokenHeaderNames: []string{"foo", "bar", "baz"},
		fallbackToken:    "",
		expectedSecret:   "",
		expectedError:    true,
	}, {
		name: "2_token_header_empty",
		headers: http.Header{
			"Baz":   {"c"},
			"token": {"d"},
			"Token": {}},
		tokenHeaderNames: []string{"Token"},
		fallbackToken:    "",
		expectedSecret:   "",
		expectedError:    true,
	}, {
		name: "3_header_order",
		headers: http.Header{
			"A": {"a"},
			"C": {"c"},
			"B": {"b"}},
		tokenHeaderNames: []string{"X", "B", "C"},
		fallbackToken:    "",
		expectedSecret:   "b",
		expectedError:    false,
	}, {
		name: "3_fallback",
		headers: http.Header{
			"x": {"a"},
			"y": {},
			"z": {"c"}},
		tokenHeaderNames: []string{"X", "Y", "Z"},
		fallbackToken:    "Foobar",
		expectedSecret:   "Foobar",
		expectedError:    false,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			token, err := ExtractToken(
				tc.headers, tc.tokenHeaderNames, tc.fallbackToken,
			)
			if tc.expectedError && err == nil {
				t.Errorf("Unexpected success: got %q, want error", token)
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tc.expectedSecret != token.Secret {
				t.Errorf(
					"Wrong secret extracted: got %q, want %q",
					token.Secret, tc.expectedSecret,
				)
			}
		})
	}
}
