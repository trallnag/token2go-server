package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// IsRequiredQueryParamSet ensures existence of query parameters. An HTTP error
// is written to w if at least one parameter is missing. Left for the function
// caller is to return if the function returns false.
func IsRequiredQueryParamSet(
	w http.ResponseWriter,
	params url.Values,
	requiredParams ...string,
) bool {
	var missingParams []string

	for _, requiredParam := range requiredParams {
		if !params.Has(requiredParam) {
			missingParams = append(missingParams, requiredParam)
		}
	}

	if len(missingParams) > 0 {
		msg := fmt.Sprintf(
			"Bad Request. Missing query parameters: %s",
			strings.Join(missingParams, ", "),
		)
		http.Error(w, msg, http.StatusBadRequest)
		return false
	} else {
		return true
	}
}

// IsQueryParamValueAllowed checks if given value is allowed. An HTTP error is written to w if
// the value is not allowed. Left for the function caller is to return if the
// function returns false.
func IsQueryParamValueAllowed(
	w http.ResponseWriter,
	name string,
	value string,
	allowedValues ...string,
) bool {
	for _, allowedValue := range allowedValues {
		if value == allowedValue {
			return true
		}
	}

	msg := fmt.Sprintf(
		"Bad Request. Value of query parameter %s forbidden. Allowed: %s",
		name, strings.Join(allowedValues, ","),
	)
	http.Error(w, msg, http.StatusBadRequest)
	return false
}

// IsSucceededEncryptWithRSA checks and handles errors coming from the
// EncryptWithRSA function. An HTTP error is written to w if given err not nil.
// Left for the function caller is to return if the function returns false.
func IsSucceededEncryptWithRSA(w http.ResponseWriter, err error) bool {
	if err == nil {
		return true
	}

	var publicKeyParseError *PublicKeyParseError
	var rsaoaepEncryptionError *RSAOAEPEncryptionError

	var msg string
	var code int

	if errors.Is(err, ErrPEMDecode) {
		msg = fmt.Sprintf("Bad Request. ErrPEMDecode: %v", err)
		code = http.StatusBadRequest
	} else if errors.Is(err, ErrNotPublicKey) {
		msg = fmt.Sprintf("Bad Request. ErrNotPublicKey: %v", err)
		code = http.StatusBadRequest
	} else if errors.As(err, &publicKeyParseError) {
		msg = fmt.Sprintf("Bad Request. PublicKeyParseError: %v", err)
		code = http.StatusBadRequest
	} else if errors.Is(err, ErrNotRSAPublicKey) {
		msg = fmt.Sprintf("Bad Request. ErrNotRSAPublicKey: %v", err)
		code = http.StatusBadRequest
	} else if errors.Is(err, ErrForbiddenKeySize) {
		msg = fmt.Sprintf("Bad Request. ErrForbiddenKeySize: %v", err)
		code = http.StatusBadRequest
	} else if errors.As(err, &rsaoaepEncryptionError) {
		msg = fmt.Sprintf("Internal Server Error. RSAOAEPEncryptionError: %v", err)
		code = http.StatusInternalServerError
	} else {
		msg = fmt.Sprintf("Internal Server Error: %v", err)
		code = http.StatusInternalServerError
	}

	http.Error(w, msg, code)

	return false
}
