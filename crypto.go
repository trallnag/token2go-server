package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"strings"
)

// GenRandBytes returns securely generated random bytes. It will return
// an error if the system's secure random number generator fails.
func GenRandBytes(n int) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

type AESKeySizeError struct {
	length int
}

func (e *AESKeySizeError) Error() string {
	return fmt.Sprintf("key must be 32 bytes, not %v bytes", e.length)
}

// EncryptWithAES encrypts plaintext with key using AES-GCM. Returns ciphertext
// and nonce. AESKeyLengthError is returned if key is not 32 bytes. All other
// errors are bubbled up without wrapping.
func EncryptWithAES(
	key []byte,
	plaintext []byte,
) (ciphertext []byte, nonce []byte, err error) {
	keyLength := len(key)
	if keyLength != 32 {
		return nil, nil, &AESKeySizeError{keyLength}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	nonce = make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	return aesgcm.Seal(nil, nonce, plaintext, nil), nonce, nil
}

type PublicKeyParseError struct {
	Err error
}

func (e *PublicKeyParseError) Error() string {
	return fmt.Sprintf("error parsing public key: %v", e.Err)
}

type RSAOAEPEncryptionError struct {
	Err error
}

func (e *RSAOAEPEncryptionError) Error() string {
	return fmt.Sprintf("error encrypting with RSA-OAEP: %v", e.Err)
}

var ErrPEMDecode = errors.New("failed to decode PEM formatted block")

var ErrNotPublicKey = errors.New("decoded PEM block not a public key")

var ErrNotRSAPublicKey = errors.New("parsed key is not of type RSA")

var ErrForbiddenKeySize = errors.New("size of given key is forbidden")

// EncryptWithRSA encrypts the given plaintext with the given publicKey. The
// resulting ciphertext is returned. Ecyrption is done with RSA-OAEP.
//
// The public key must be PEM encoded.
//
// For the public key the forms RFC5280 (X.509) and RFC8017 (PKCS #1) are
// supported. In PEM encoded blocks these can be identified with the
// strings "PUBLIC KEY" and "RSA PUBLIC KEY".
//
// Sentinel errors: ErrPEMDecode, ErrNotPublicKey, ErrNotRSAPublicKey, ErrForbiddenKeySize.
//
// Custom error types: PublicKeyParseError and RSAOAEPEncryptionError.
//
// No other errors are bubbled up.
func EncryptWithRSA(publicKey []byte, plaintext []byte) (ciphertext []byte, err error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, ErrPEMDecode
	}
	if !strings.Contains(block.Type, "PUBLIC KEY") {
		return nil, ErrNotPublicKey
	}

	pub, parseErr := x509.ParsePKIXPublicKey(block.Bytes)
	if parseErr != nil {
		pub, parseErr = x509.ParsePKCS1PublicKey(block.Bytes)
	}
	if parseErr != nil {
		return nil, &PublicKeyParseError{err}
	}

	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, ErrNotRSAPublicKey
	}

	if rsaKey.Size() != 256 {
		return nil, ErrForbiddenKeySize
	}

	ciphertext, err = rsa.EncryptOAEP(
		sha256.New(), rand.Reader, rsaKey, plaintext, nil,
	)
	if err != nil {
		return nil, &RSAOAEPEncryptionError{err}
	}

	return ciphertext, nil
}
