package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestGenRandBytes_Length(t *testing.T) {
	want := 10

	b, err := GenRandBytes(10)
	if err != nil {
		t.Errorf("Unexpected generator failure: %s", err)
	}

	got := len(b)
	if got != want {
		t.Errorf("Wrong amount of bytes: got %v, want %v", got, want)
	}
}

func TestGenRandBytes_Unique(t *testing.T) {
	b1, err := GenRandBytes(10)
	if err != nil {
		t.Errorf("Unexpected generator failure: %s", err)
	}

	b2, err := GenRandBytes(10)
	if err != nil {
		t.Errorf("Unexpected generator failure: %s", err)
	}

	if bytes.Equal(b1, b2) {
		t.Errorf("Should not be equal: %x and %x", b1, b2)
	}
}

func TestEncryptWithAES(t *testing.T) {
	for _, tc := range []struct {
		name      string
		key       []byte
		plaintext []byte
	}{{
		name:      "1_simple",
		key:       []byte("abcdefghijklmnopabcdefghijklmnop"),
		plaintext: []byte("hallo wie geht es dir heute mein guter?"),
	}, {
		name:      "2_long",
		key:       []byte("abcdefghijklmnopabcdefghijklmnop"),
		plaintext: []byte(strings.Repeat("ewkfpoew fewfioew fewfopk", 1000)),
	}, {
		name:      "2_empty",
		key:       []byte("12345678123456781234567812345678"),
		plaintext: []byte(""),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, nonce, err := EncryptWithAES(tc.key, tc.plaintext)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			block, err := aes.NewCipher(tc.key)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			aesgcm, err := cipher.NewGCM(block)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			got := hex.EncodeToString(plaintext)
			want := hex.EncodeToString(tc.plaintext)
			if got != want {
				t.Errorf(
					"Decrypted encoded plaintext does not match input: got %q, want %q",
					got,
					want,
				)
			}
		})
	}
}

func TestEncryptWithAES_KeySize(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  string
	}{
		{"1_too_short", "abcdefghijklmnop"},
		{"2_too_long", "abcdefghijklmnopabcdefghijklmnopabcdefghijklmnop"},
		{"3_empty", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := EncryptWithAES([]byte(tc.key), []byte("Does not matter"))

			if err == nil {
				t.Errorf("Unexpected success. Expected error.")
			}

			var e *AESKeySizeError
			if !errors.As(err, &e) {
				t.Errorf("Unexpected error type: got %v, want %v", err, &e)
			}
		})
	}
}

func TestEncryptWithRSA(t *testing.T) {
	for _, tc := range []struct {
		name       string
		publicKey  string
		privateKey string
	}{{
		name:       "1_pub_key_rfc5280_x509",
		publicKey:  "testdata/a-public-key-rsa2048-rfc5280-x509.pem",
		privateKey: "testdata/a-private-key-rsa2048-rfc5958-pksc8.pem",
	}, {
		name:       "2_pub_key_rfc8017_pksc1",
		publicKey:  "testdata/a-public-key-rsa2048-rfc8017-pksc1.pem",
		privateKey: "testdata/a-private-key-rsa2048-rfc5958-pksc8.pem",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			publicKeyBytes, err := os.ReadFile(tc.publicKey)
			if err != nil {
				t.Fatal(err)
			}

			privateKeyBytes, err := os.ReadFile(tc.privateKey)
			if err != nil {
				t.Fatal(err)
			}

			plaintext := []byte("12345678123456781234567812345678")

			ciphertext, err := EncryptWithRSA(publicKeyBytes, plaintext)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			block, _ := pem.Decode(privateKeyBytes)
			if block == nil {
				t.Fatal("failed to decode PEM formatted block")
			}

			privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				t.Fatal(err)
			}

			privateKeyRSA, ok := privateKey.(*rsa.PrivateKey)
			if !ok {
				t.Fatal("unexpected; not an RSA key")
			}

			decryptedCipherText, err := rsa.DecryptOAEP(
				sha256.New(), rand.Reader, privateKeyRSA, ciphertext, nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			got := hex.EncodeToString(decryptedCipherText)
			want := hex.EncodeToString(plaintext)

			if got != want {
				t.Fatalf("Output does not match input: got %v, want %v", got, want)
			}
		})
	}
}

func TestEncryptWithRSA_ErrPEMDecode(t *testing.T) {
	_, err := EncryptWithRSA([]byte("foo"), []byte("bar"))
	if !errors.Is(err, ErrPEMDecode) {
		t.Errorf("Missing ErrPEMDecode: got %q, want %q", err, ErrPEMDecode)
	}
}

func TestEncryptWithRSA_ErrNotPublicKey(t *testing.T) {
	privateKey, err := os.ReadFile("testdata/a-private-key-rsa2048-rfc5958-pksc8.pem")
	if err != nil {
		t.Fatal(err)
	}

	_, err = EncryptWithRSA(privateKey, []byte("bar"))
	if !errors.Is(err, ErrNotPublicKey) {
		t.Errorf("Missing ErrNotPublicKey: got %q, want %q", err, ErrNotPublicKey)
	}
}

func TestEncryptWithRSA_PublicKeyParseError(t *testing.T) {
	_, err := EncryptWithRSA(
		[]byte("-----BEGIN PUBLIC KEY-----\nxxxx\n-----END PUBLIC KEY-----"),
		[]byte("bar"),
	)
	var want *PublicKeyParseError
	if !errors.As(err, &want) {
		t.Errorf("Missing PublicKeyParseError: got %q", err)
	}
}

func TestEncryptWithRSA_ErrNotRSAPublicKey(t *testing.T) {
	publicKey, err := os.ReadFile(
		"testdata/b-public-key-ecdsa-prime256v1-rfc5280-x509.pem",
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = EncryptWithRSA(publicKey, []byte("foo"))
	if !errors.Is(err, ErrNotRSAPublicKey) {
		t.Errorf("Missing ErrNotRSAPublicKey: got %q, want %q", err, ErrNotRSAPublicKey)
	}
}

func TestEncryptWithRSA_ErrForbiddenKeySize(t *testing.T) {
	for _, tc := range []struct {
		name            string
		publicKey       string
		successExpected bool
	}{{
		name:            "1_pub_key_rfc5280_x509_1024_bits",
		publicKey:       "testdata/c-public-key-rsa1024-rfc5280-x509.pem",
		successExpected: false,
	}, {
		name:            "2_pub_key_rfc8017_pksc1_1024_bits",
		publicKey:       "testdata/c-public-key-rsa1024-rfc8017-pksc1.pem",
		successExpected: false,
	}, {
		name:            "3_pub_key_rfc5280_x509_2048_bits",
		publicKey:       "testdata/a-public-key-rsa2048-rfc5280-x509.pem",
		successExpected: true,
	}, {
		name:            "4_pub_key_rfc8017_pksc1_2048_bits",
		publicKey:       "testdata/a-public-key-rsa2048-rfc8017-pksc1.pem",
		successExpected: true,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			publicKey, err := os.ReadFile(tc.publicKey)
			if err != nil {
				t.Fatal(err)
			}

			_, err = EncryptWithRSA(publicKey, []byte("foo"))
			if !tc.successExpected && !errors.Is(err, ErrForbiddenKeySize) {
				t.Errorf(
					"Missing ErrForbiddenKeySize: got %q, want %q",
					err, ErrForbiddenKeySize,
				)
			} else if tc.successExpected && err != nil {
				t.Errorf("Expected success, but got error: %v", err)
			}
		})
	}
}
