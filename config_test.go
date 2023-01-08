package main

import (
	"os"
	"strings"
	"testing"
)

func TestNewConfig_Default(t *testing.T) {
	os.Unsetenv("T2G_FALLBACK_TOKEN")
	os.Unsetenv("T2G_SERVER_PORT")
	os.Unsetenv("T2G_TOKEN_HEADER_NAMES")
	os.Unsetenv("T2G_ADD_TOKEN_HEADER_NAMES")
	os.Unsetenv("T2G_UI_TARGET")
	os.Unsetenv("T2G_UI_TITLE")
	os.Unsetenv("T2G_UI_DESC1")
	os.Unsetenv("T2G_UI_DESC2")
	os.Unsetenv("T2G_UI_MISC")

	c := NewConfig()

	var got string
	var want string

	eq := func(n string, g string, w string) {
		if got != want {
			t.Errorf("Unexpected %s: got %q want %q", n, g, w)
		}
	}

	eq("fallbackToken", c.fallbackToken, "")
	eq("serverPort", c.serverPort, "8080")
	eq(
		"tokenHeaderNames",
		strings.Join(c.tokenHeaderNames, ","),
		"Access-Token,Authorization,Token,X-Auth-Request-Access-Token,X-Forwarded-Access-Token",
	)
	eq("addTokenHeaderNames", strings.Join(c.addTokenHeaderNames, ","), "")
	eq("uiTarget", c.uiTarget, "")
	eq("uiTitle", c.uiTitle, "")
	eq("uiDesc1", c.uiDesc1, "")
	eq("uiDesc2", c.uiDesc2, "")
	eq("uiMisc", c.uiMisc, "")
}

func TestNewConfig_Custom(t *testing.T) {
	t.Setenv("T2G_FALLBACK_TOKEN", "x")
	t.Setenv("T2G_SERVER_PORT", "x")
	t.Setenv("T2G_TOKEN_HEADER_NAMES", "x")
	t.Setenv("T2G_ADD_TOKEN_HEADER_NAMES", "x")
	t.Setenv("T2G_UI_TARGET", "x")
	t.Setenv("T2G_UI_TITLE", "x")
	t.Setenv("T2G_UI_DESC1", "x")
	t.Setenv("T2G_UI_DESC2", "x")
	t.Setenv("T2G_UI_MISC", "x")

	c := NewConfig()

	var got string
	var want string

	eq := func(n string, g string, w string) {
		if got != want {
			t.Errorf("Unexpected %s: got %q want %q", n, g, w)
		}
	}

	eq("fallbackToken", c.fallbackToken, "x")
	eq("serverPort", c.serverPort, "x")
	eq("tokenHeaderNames", strings.Join(c.tokenHeaderNames, ","), "x")
	eq("addTokenHeaderNames", strings.Join(c.addTokenHeaderNames, ","), "x")
	eq("uiTarget", c.uiTarget, "x")
	eq("uiTitle", c.uiTitle, "x")
	eq("uiDesc1", c.uiDesc1, "x")
	eq("uiDesc2", c.uiDesc2, "x")
	eq("uiMisc", c.uiMisc, "x")
}

func TestGetEnv(t *testing.T) {
	t.Setenv("T2G_FOO", "bar")

	got := GetEnv("FOO", "alice")
	want := "bar"
	if got != want {
		t.Errorf("Unexpected result: got %q want %q", got, want)
	}

	os.Unsetenv("T2G_DOES_NOT_EXIST")
	got = GetEnv("DOES_NOT_EXIST", "alice")
	want = "alice"
	if got != want {
		t.Errorf("Unexpected result: got %q want %q", got, want)
	}
}

func TestSplitToSlice(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		want  []string
	}{{
		name:  "1_single",
		input: "d",
		want:  []string{"d"},
	}, {
		name:  "2_commas",
		input: "Foo,bar,,,lol , Dummy?, Test,",
		want:  []string{"Foo", "bar", "lol", "Dummy?", "Test"},
	}, {
		name:  "3_empty",
		input: "",
		want:  []string{""},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			got := strings.Join(SplitToSlice(tc.input), ",")
			want := strings.Join(tc.want, ",")
			if got != want {
				t.Errorf("Wrong result: got %q, want %q", got, want)
			}
		})
	}
}
