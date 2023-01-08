package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestEmbeddedContent(t *testing.T) {
	for _, expected := range []string{
		"static/apple-touch-icon.png",
		"static/css/main.css",
		"static/css/modern-normalize@1.1.0",
		"static/css/mvp@1.12.0",
		"static/css/toastify@1.12.0",
		"static/favicon-16x16.png",
		"static/favicon-32x32.png",
		"static/favicon.ico",
		"static/favicon.svg",
		"static/js/index.js",
		"static/js/toastify@1.12.0",
		"template/index.html",
	} {
		_, err := fs.Stat(content, expected)
		if err != nil {
			t.Errorf("node does not exist in embedded content: %s", expected)
		}
	}
}

func TestUsageOfExLibs(t *testing.T) {
	expectedStrings := []string{
		"modern-normalize@1.1.0/modern-normalize.min.css",
		"mvp@1.12.0/mvp.min.css",
		"toastify@1.12.0/toastify.min.css",
		"toastify@1.12.0/toastify.min.js",
	}

	var sb strings.Builder

	err := fs.WalkDir(content, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".html") {
			fileContent, err := fs.ReadFile(content, path)
			if err != nil {
				return fmt.Errorf("failed reading file: %w", err)
			}

			sb.Write(fileContent)
			sb.WriteString("\n")
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	searchSpace := sb.String()

	for _, expectedString := range expectedStrings {
		if !strings.Contains(searchSpace, expectedString) {
			t.Errorf("Not found in search space: %v", expectedString)
		}
	}
}

func TestGetEchoHandler(t *testing.T) {
	handler := http.HandlerFunc(GetEchoHandler)

	request, err := http.NewRequestWithContext(
		context.TODO(),
		"GET",
		"/echo?lol=lol",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("X-Tutu", "x")
	request.Header.Set("x-foobar", "x")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, request)
	rrr := rr.Result()
	defer rrr.Body.Close()

	if rrr.StatusCode != 200 {
		t.Errorf("Wrong status code: got %v, want 200", rrr.StatusCode)
	}

	b, err := io.ReadAll(rrr.Body)
	if err != nil {
		t.Fatalf("Unexpected error while reading body: %v", err)
	}
	body := string(b)

	want := `"X-Tutu":["x"]`
	if !strings.Contains(body, want) {
		t.Errorf("Did not find '%v' in '%v'", want, body)
	}

	want = `"lol":["lol"]`
	if !strings.Contains(body, want) {
		t.Errorf("Did not find '%v' in '%v'", want, body)
	}

	want = `"remoteAddr":""`
	if !strings.Contains(body, want) {
		t.Errorf("Did not find '%v' in '%v'", want, body)
	}
}

func TestGetHealthHandler(t *testing.T) {
	handler := http.HandlerFunc(GetHealthHandler)

	request, err := http.NewRequestWithContext(context.TODO(), "GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, request)
	rrr := rr.Result()
	defer rrr.Body.Close()

	if status := rrr.StatusCode; status != 200 {
		t.Errorf("Wrong status code: got %v want %v", status, 200)
	}

	gotContentType := rrr.Header.Get("Content-Type")
	wantContentType := "application/json"
	if gotContentType != wantContentType {
		t.Errorf("Wrong content type: got %q want %q", gotContentType, wantContentType)
	}
}

func TestMakeGetTokenHandler(t *testing.T) {
	for _, tc := range []struct {
		name             string
		headers          http.Header
		tokenHeaderNames []string
		fallbackToken    string
		expectedCode     int
		expectedSecret   string
	}{{
		name:             "1_simple",
		headers:          http.Header{"Foo": []string{"x"}},
		tokenHeaderNames: []string{"Foo"},
		fallbackToken:    "",
		expectedCode:     200,
		expectedSecret:   "x",
	}, {
		name:             "2_order",
		headers:          http.Header{"Foo": []string{"f"}, "Bar": []string{"b"}},
		tokenHeaderNames: []string{"Bar", "Foo"},
		fallbackToken:    "",
		expectedCode:     200,
		expectedSecret:   "b",
	}, {
		name:             "3_missing",
		headers:          http.Header{"Foo": []string{"f"}, "Bar": []string{"b"}},
		tokenHeaderNames: []string{"fefefe"},
		fallbackToken:    "",
		expectedCode:     444,
		expectedSecret:   "",
	}, {
		name:             "4_fallback",
		headers:          http.Header{"Foo": []string{"f"}, "Bar": []string{"b"}},
		tokenHeaderNames: []string{"fefefe"},
		fallbackToken:    "lol",
		expectedCode:     200,
		expectedSecret:   "lol",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			handler := MakeGetTokenHandler(tc.tokenHeaderNames, tc.fallbackToken)

			request, err := http.NewRequestWithContext(
				context.TODO(),
				"GET",
				"/health",
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			request.Header = tc.headers

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, request)
			rrr := rr.Result()
			defer rrr.Body.Close()

			if rrr.StatusCode != tc.expectedCode {
				t.Errorf(
					"Wrong status code: got %v, want %v",
					rrr.StatusCode, tc.expectedCode,
				)
			}

			if rrr.StatusCode != 200 {
				return
			}

			b, err := io.ReadAll(rrr.Body)
			if err != nil {
				t.Fatal(err)
			}
			body := string(b)

			if !strings.Contains(body, tc.expectedSecret) {
				t.Errorf("Did not find '%v' in '%v'", tc.expectedSecret, body)
			}
		})
	}
}

func TestMakeGetTokenRedirectFlowHandler(t *testing.T) {
	aPublic1, err := os.ReadFile("testdata/a-public-key-rsa2048-rfc5280-x509.pem")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name             string
		queryParams      url.Values
		headers          http.Header
		tokenHeaderNames []string
		fallbackToken    string
		expectedCode     int
	}{{
		name: "1_success_x509",
		queryParams: url.Values{
			"target":        {"https://example.com"},
			"state":         {"state"},
			"publicKeyType": {"rsa2048-rfc5280-x509-pem"},
			"publicKey":     {string(aPublic1)},
		},
		headers:          http.Header{"Foo": []string{"x"}},
		tokenHeaderNames: []string{"Foo"},
		fallbackToken:    "",
		expectedCode:     301,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			handler := MakeGetTokenRedirectFlowHandler(
				tc.tokenHeaderNames, tc.fallbackToken,
			)

			request, err := http.NewRequestWithContext(context.TODO(),
				"GET", "/flows/redirect/token?"+tc.queryParams.Encode(), nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			request.Header = tc.headers

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, request)
			rrr := rr.Result()
			defer rrr.Body.Close()

			if rrr.StatusCode != tc.expectedCode {
				t.Errorf(
					"Wrong status code: got %v, want %v",
					rrr.StatusCode, tc.expectedCode,
				)
			}

			gotTarget := strings.Split(rrr.Header.Get("Location"), "?")[0]
			wantTarget := tc.queryParams.Get("target")
			if gotTarget != wantTarget {
				t.Errorf("Wrong target: got %v, want %v", gotTarget, wantTarget)
			}
		})
	}
}

func TestServeStatic(t *testing.T) {
	router := chi.NewRouter()
	ServeStatic(router)
	server := httptest.NewServer(router)
	defer server.Close()
	client := server.Client()

	m := map[string]string{
		"/favicon.ico":          "image/vnd.microsoft.icon",
		"/apple-touch-icon.png": "image/png",
		"/favicon-16x16.png":    "image/png",
		"/favicon-32x32.png":    "image/png",
		"/favicon.svg":          "image/svg+xml",
		"/js/index.js":          "text/javascript; charset=utf-8",
	}

	for k, v := range m {
		resp, err := client.Get(server.URL + k) //nolint
		if err != nil {
			t.Fatalf("Unexpected error performing GET from server: %v", err)
		} else {
			defer resp.Body.Close()
		}
		if status := resp.StatusCode; status != http.StatusOK {
			t.Errorf("Returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
		if contentType := resp.Header.Get("Content-Type"); contentType != v {
			t.Errorf("Returned wrong content type: got %v want %v",
				contentType, v)
		}
	}
}

func TestIndexTmplData(t *testing.T) {
	d := IndexTmplData{"TITLE", "DESC1", "DESC2", "<p>MISC</p>"}

	tmplContent, err := fs.Sub(content, "template")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	tmpl, err := template.ParseFS(tmplContent, "index.html")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, d)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	renderedStr := buffer.String()
	for _, e := range []string{
		d.Title, string(d.Desc1), string(d.Desc2), string(d.Misc),
	} {
		if !strings.Contains(renderedStr, e) {
			t.Errorf("Expected render to contain %q", e)
		}
	}
}

func TestNewIndexTmplData_Title(t *testing.T) {
	got := NewIndexTmplData("", "", "", "", "").Title
	want := "Token2go"
	if got != want {
		t.Errorf("Wrong Title: got %q, want %q", got, want)
	}

	got = NewIndexTmplData("MyApp", "", "", "", "").Title
	want = "Token2go | MyApp"
	if got != want {
		t.Errorf("Wrong Title: got %q, want %q", got, want)
	}

	got = NewIndexTmplData("MyApp", "Custom Title", "", "", "").Title
	want = "Custom Title"
	if got != want {
		t.Errorf("Wrong Title: got %q, want %q", got, want)
	}
}

func TestNewIndexTmplData_Desc1(t *testing.T) {
	got := string(NewIndexTmplData("MyApp", "", "", "", "").Desc1)
	want := "Get a token for MyApp with the Token2go service"
	if got != want {
		t.Errorf("Wrong Desc1: got %q, want %q", got, want)
	}

	got = string(NewIndexTmplData("", "", "", "", "").Desc1)
	want = "Go ahead and grab a token with the Token2go service"
	if got != want {
		t.Errorf("Wrong Desc1: got %q, want %q", got, want)
	}
}

func TestNewIndexTmplData_Desc2(t *testing.T) {
	x := "<a href=\"https://www.google.com\">link text</a>"
	got := string(NewIndexTmplData("", "", "", x, "").Desc2)
	want := "<a href=\"https://www.google.com\">link text</a>"
	if got != want {
		t.Errorf("Wrong Desc2: got %q, want %q", got, want)
	}
}

func TestNewIndexTmplData_Misc(t *testing.T) {
	got := string(NewIndexTmplData("", "", "", "", "<p>Foobar</p>").Misc)
	want := "<p>Foobar</p>"
	if got != want {
		t.Errorf("Wrong Misc: got %q, want %q", got, want)
	}
}

func TestInitRouter(t *testing.T) {
	c := NewConfig()
	initRouter(
		c.fallbackToken,
		c.tokenHeaderNames,
		c.addTokenHeaderNames,
		NewIndexTmplData(
			c.uiTarget,
			c.uiTitle,
			c.uiDesc1,
			c.uiDesc2,
			c.uiMisc,
		),
	)
}
