package main

import (
	"errors"
	"io"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestIsRequiredQueryParamSet(t *testing.T) {
	for _, tc := range []struct {
		name           string
		params         []string
		requiredParams []string
		expectedResult bool
	}{{
		name:           "1_single",
		params:         []string{"fooBar"},
		requiredParams: []string{"fooBar"},
		expectedResult: true,
	}, {
		name:           "2_duplicated",
		params:         []string{"fooBar", "fooBar"},
		requiredParams: []string{"fooBar"},
		expectedResult: true,
	}, {
		name:           "3_case_sensitive",
		params:         []string{"FOOBAR"},
		requiredParams: []string{"fooBar"},
		expectedResult: false,
	}, {
		name:           "4_not_required",
		params:         []string{"foo", "tux", "bar"},
		requiredParams: []string{"foo", "bar"},
		expectedResult: true,
	}, {
		name:           "5_no_required",
		params:         []string{"foo", "tux", "bar"},
		requiredParams: []string{},
		expectedResult: true,
	}, {
		name:           "6_missing",
		params:         []string{"foo", "tux", "bar"},
		requiredParams: []string{"foo", "tux", "dream"},
		expectedResult: false,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			params := url.Values{}
			for _, param := range tc.params {
				params.Add(param, "x")
			}

			rr := httptest.NewRecorder()
			result := IsRequiredQueryParamSet(rr, params, tc.requiredParams...)
			if result != tc.expectedResult {
				t.Errorf("Wrong result: got %v, want %v", result, tc.expectedResult)
			}
			if tc.expectedResult == true && rr.Code != 200 {
				t.Errorf("Wrong code: got %v, want %v", rr.Code, 200)
			}
			if tc.expectedResult == false && rr.Code != 400 {
				t.Errorf("Wrong code: got %v, want %v", rr.Code, 400)
			}
		})
	}
}

func TestIsQueryParamValueAllowed(t *testing.T) {
	for _, tc := range []struct {
		name           string
		value          string
		allowedValues  []string
		expectedResult bool
	}{{
		name:           "1",
		value:          "x",
		allowedValues:  []string{"x"},
		expectedResult: true,
	}, {
		name:           "2_multiple",
		value:          "x",
		allowedValues:  []string{"fefe", "x"},
		expectedResult: true,
	}, {
		name:           "3_case_sensitive",
		value:          "x",
		allowedValues:  []string{"X"},
		expectedResult: false,
	}, {
		name:           "4_nothing_allowed",
		value:          "x",
		allowedValues:  []string{},
		expectedResult: false,
	}, {
		name:           "5_empty_string",
		value:          "x",
		allowedValues:  []string{""},
		expectedResult: false,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			result := IsQueryParamValueAllowed(rr, "X", tc.value, tc.allowedValues...)
			if result != tc.expectedResult {
				t.Errorf("Wrong result: got %v, want %v", result, tc.expectedResult)
			}
			if tc.expectedResult == true && rr.Code != 200 {
				t.Errorf("Wrong code: got %v, want %v", rr.Code, 200)
			}
			if tc.expectedResult == false && rr.Code != 400 {
				t.Errorf("Wrong code: got %v, want %v", rr.Code, 400)
			}
		})
	}
}

func TestIsSucceededEncryptWithRSA(t *testing.T) {
	var vPublicKeyParseError *PublicKeyParseError
	var vRSAOAEPEncryptionError *RSAOAEPEncryptionError

	for _, tc := range []struct {
		name           string
		substr         string
		err            error
		expectedCode   int
		expectedResult bool
	}{{
		name:           "1_no_error",
		substr:         "",
		err:            nil,
		expectedCode:   200,
		expectedResult: true,
	}, {
		name:           "2_ErrPEMDecode",
		substr:         "ErrPEMDecode",
		err:            ErrPEMDecode,
		expectedCode:   400,
		expectedResult: false,
	}, {
		name:           "3_ErrNotPublicKey",
		substr:         "ErrNotPublicKey",
		err:            ErrNotPublicKey,
		expectedCode:   400,
		expectedResult: false,
	}, {
		name:           "4_PublicKeyParseError",
		substr:         "PublicKeyParseError",
		err:            vPublicKeyParseError,
		expectedCode:   400,
		expectedResult: false,
	}, {
		name:           "5_ErrNotRSAPublicKey",
		substr:         "ErrNotRSAPublicKey",
		err:            ErrNotRSAPublicKey,
		expectedCode:   400,
		expectedResult: false,
	}, {
		name:           "6_ErrForbiddenKeySize",
		substr:         "ErrForbiddenKeySize",
		err:            ErrForbiddenKeySize,
		expectedCode:   400,
		expectedResult: false,
	}, {
		name:           "7_RSAOAEPEncryptionError",
		substr:         "RSAOAEPEncryptionError",
		err:            vRSAOAEPEncryptionError,
		expectedCode:   500,
		expectedResult: false,
	}, {
		name:           "8_unknown_error",
		substr:         "",
		err:            errors.New("foobar"),
		expectedCode:   500,
		expectedResult: false,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			result := IsSucceededEncryptWithRSA(rr, tc.err)
			rrr := rr.Result()
			defer rrr.Body.Close()

			if result != tc.expectedResult {
				t.Errorf("Wrong result: got %v, want %v", result, tc.expectedResult)
			}
			if tc.expectedResult == false && rrr.StatusCode != tc.expectedCode {
				t.Errorf("Wrong code: got %v, want %v", rrr.StatusCode, tc.expectedCode)
			}

			b, err := io.ReadAll(rrr.Body)
			if err != nil {
				t.Fatalf("Unexpected error while reading body: %v", err)
			}
			if !strings.Contains(string(b), tc.substr) {
				t.Errorf(
					"Didn't find substr in body: got %q, want %q",
					string(b),
					tc.substr,
				)
			}
		})
	}
}
