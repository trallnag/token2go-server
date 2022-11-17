package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed all:static
//go:embed all:swagger-ui
//go:embed all:template
var content embed.FS

func main() {
	c := NewConfig()

	err := http.ListenAndServe(
		fmt.Sprintf(":"+c.serverPort),
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
		),
	)
	if err != nil {
		panic(err)
	}
}

func initRouter(
	fallbackToken string,
	tokenHeaderNames []string,
	addTokenHeaderNames []string,
	itd IndexTmplData,
) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	ServeTmpl(ServeTmplArgs{
		router:   r,
		patterns: []string{"/", "/index.html"},
		file:     "index.html",
		data:     itd,
	})

	ServeSwaggerUI(r)

	ServeStatic(r)

	r.Group(func(r chi.Router) {
		r.Use(middleware.NoCache)
		r.Get("/echo", GetEchoHandler)
		r.Get("/health", GetHealthHandler)
		r.Get("/token", MakeGetTokenHandler(
			append(tokenHeaderNames, addTokenHeaderNames...),
			fallbackToken,
		))
		r.Get("/flow/redirect/token", MakeGetTokenRedirectFlowHandler(
			append(tokenHeaderNames, addTokenHeaderNames...),
			fallbackToken,
		))
	})

	return r
}

// Echo is the representation of the GetEchoHandler's body.
type Echo struct {
	Parameters url.Values  `json:"parameters"`
	Headers    http.Header `json:"headers"`
	RemoteAddr string      `json:"remoteAddr"`
}

// GetEchoHandler writes a response with all headers, parameters, and other data
// from from the request encoded as non-pretty JSON in the body.
func GetEchoHandler(w http.ResponseWriter, r *http.Request) {
	jsonEncoder := json.NewEncoder(w)

	if r.URL.Query().Has("pretty") {
		jsonEncoder.SetIndent("", "  ")
	}

	w.Header().Set("Content-Type", "application/json")
	err := jsonEncoder.Encode(Echo{
		r.URL.Query(),
		r.Header,
		r.RemoteAddr,
	})

	if err != nil {
		panic(err)
	}
}

// GetHealthHandler informs about the health of the Token2go server. Currently
// this handler always writes a JSON response and status code 200.
func GetHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := fmt.Fprintln(w, `{"status": "OK."}`)
	if err != nil {
		panic(err)
	}
}

// MakeGetTokenHandler returns a handler that extracts a token from the request
// and returns the token including metadata encoded as non-pretty JSON.
//
// Handler will only look for given token header names. If the fallback token
// is an empty string and no token has been found, a client error response
// will be written.
func MakeGetTokenHandler(
	tokenHeaderNames []string,
	fallbackToken string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := ExtractToken(r.Header, tokenHeaderNames, fallbackToken)
		if err != nil {
			msg := "Token not found. Looking for: "
			http.Error(w, msg+strings.Join(tokenHeaderNames, ", "), 444)
			return
		}

		tokenJSON, err := json.Marshal(token)
		if err != nil {
			msg := "Internal Server Error. Marshalling failed."
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(tokenJSON)
	}
}

// MakeGetTokenRedirectFlowHandler returns a handler for the token redirect
// flow. This handler extracts the token from the request and attaches it to
// the redirect URL as an encrypted payload.
func MakeGetTokenRedirectFlowHandler(
	tokenHeaderNames []string,
	fallbackToken string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()

		// Extract query parameters.
		target := queryParams.Get("target")
		state := queryParams.Get("state")
		publicKeyType := queryParams.Get("publicKeyType")
		publicKey := []byte(queryParams.Get("publicKey"))

		// Ensure required query parameters are set.
		if !IsRequiredQueryParamSet(w, queryParams,
			"target", "state", "publicKeyType", "publicKey",
		) {
			return
		}

		// Ensure query parameter values are allowed.
		if !IsQueryParamValueAllowed(w, "publicKeyType", publicKeyType,
			"rsa2048-rfc5280-x509-pem", "rsa2048-rfc8017-pksc1-pem",
		) {
			return
		}

		// Generate key for AES encryption of payload.
		payloadKey, err := GenRandBytes(32)
		if err != nil {
			msg := "Internal Server Error. Secure random number generator failure."
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		// Encrypt payload key with public key.
		encryptedPayloadKey, err := EncryptWithRSA(publicKey, payloadKey)
		if !IsSucceededEncryptWithRSA(w, err) {
			return
		}

		// Build JSON payload containing token.
		token, err := ExtractToken(r.Header, tokenHeaderNames, fallbackToken)
		if err != nil {
			msg := "Token not found. Looking for: "
			http.Error(w, msg+strings.Join(tokenHeaderNames, ", "), 444)
			return
		}
		payload, err := json.Marshal(token)
		if err != nil {
			msg := "Internal Server Error. Marshalling failed."
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		// Encrypt payload with AES-GCM.
		encryptedPayload, nonce, err := EncryptWithAES(payloadKey, payload)
		if err != nil {
			msg := "Internal Server Error. Payload encryption failed."
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		// Perform permanent redirect.
		redirectUrl := fmt.Sprintf("%v?%v", target, url.Values{
			"payload": {base64.StdEncoding.EncodeToString(encryptedPayload)},
			"key":     {base64.StdEncoding.EncodeToString(encryptedPayloadKey)},
			"nonce":   {base64.StdEncoding.EncodeToString(nonce)},
			"state":   {state},
		}.Encode())
		http.Redirect(w, r, redirectUrl, http.StatusMovedPermanently)
	}
}

func ServeSwaggerUI(router chi.Router) {
	swaggerContent, err := fs.Sub(content, "swagger-ui")
	if err != nil {
		panic(err)
	}

	fs := http.FileServer(http.FS(swaggerContent))
	router.Handle("/swagger-ui/*", http.StripPrefix("/swagger-ui/", fs))
}

// ServeStatic adds a handler to the given router that serves static content
// from the hardcoded and embedded "static" directory using a fileserver.
func ServeStatic(router chi.Router) {
	staticContent, err := fs.Sub(content, "static")
	if err != nil {
		panic(err)
	}

	router.Handle("/*", http.FileServer(http.FS(staticContent)))
}

// ServeTmplArgs represents the arguments for the ServeTmpl function.
type ServeTmplArgs struct {
	router   chi.Router
	patterns []string
	file     string
	data     any
}

func ServeTmpl(a ServeTmplArgs) {
	tmplContent, err := fs.Sub(content, "template")
	if err != nil {
		panic(err)
	}

	tmpl, err := template.ParseFS(tmplContent, a.file)
	if err != nil {
		panic(err)
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, a.data)
	if err != nil {
		panic(err)
	}

	bytes := buffer.Bytes()

	for _, pattern := range a.patterns {
		a.router.Get(pattern, func(w http.ResponseWriter, r *http.Request) {
			w.Write(bytes)
		})
	}
}

// IndexTmplData is the input data for the index.html template.
type IndexTmplData struct {
	Title string
	Desc1 template.HTML
	Desc2 template.HTML
	Misc  template.HTML
}

// NewIndexTmplData constructs indexTmplData after juggling around the input
// parameters. Removes a bit of logic from the actual index.html template.
func NewIndexTmplData(
	uiTarget string,
	uiTitle string,
	uiDesc1 string,
	uiDesc2 string,
	uiMisc string,
) IndexTmplData {
	d := IndexTmplData{}

	if len(uiTitle) > 0 {
		d.Title = uiTitle
	} else if len(uiTarget) > 0 {
		d.Title = "Token2go | " + uiTarget
	} else {
		d.Title = "Token2go"
	}

	if len(uiTarget) > 0 {
		x := fmt.Sprintf("Get a token for %s with the Token2go service", uiTarget)
		d.Desc1 = template.HTML(x)
	} else {
		d.Desc1 = template.HTML("Go ahead and grab a token with the Token2go service")
	}

	d.Desc2 = template.HTML(uiDesc2)
	d.Misc = template.HTML(uiMisc)

	return d
}
