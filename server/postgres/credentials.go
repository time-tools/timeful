package postgres

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	encryptionKeyEnvironment  = "ENCRYPTION_KEY"
	encryptionKeyLength       = 32
	credentialEnvelopeVersion = "v1:"
)

// ErrEncryptionKeyUnavailable reports a missing or wrong-length ENCRYPTION_KEY.
// The codec refuses to derive a key, so provider credentials fail closed rather
// than being written or read under an unexpected construction.
var ErrEncryptionKeyUnavailable = errors.New("ENCRYPTION_KEY must be exactly 32 raw bytes")

// credentialCodec encrypts provider secrets with authenticated AES-256-GCM. The
// stored envelope is the version prefix followed by the base64 encoding of
// nonce || ciphertext || tag, so a future key or algorithm change is
// distinguishable without ambiguity.
type credentialCodec struct {
	aead cipher.AEAD
}

func newCredentialCodec(key []byte) (*credentialCodec, error) {
	if len(key) != encryptionKeyLength {
		return nil, ErrEncryptionKeyUnavailable
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("initialize credential cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("initialize credential AEAD: %w", err)
	}
	return &credentialCodec{aead: aead}, nil
}

func credentialCodecFromEnvironment() (*credentialCodec, error) {
	return newCredentialCodec([]byte(os.Getenv(encryptionKeyEnvironment)))
}

// ValidateCredentialEncryptionKey reports whether ENCRYPTION_KEY is usable by
// the credential codec. Startup wiring calls it so a missing or wrong-length
// key refuses to serve rather than silently degrading credential storage.
func ValidateCredentialEncryptionKey() error {
	_, err := credentialCodecFromEnvironment()
	return err
}

func (c *credentialCodec) encrypt(plaintext string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate credential nonce: %w", err)
	}
	sealed := c.aead.Seal(nil, nonce, []byte(plaintext), nil)
	payload := make([]byte, 0, len(nonce)+len(sealed))
	payload = append(payload, nonce...)
	payload = append(payload, sealed...)
	return credentialEnvelopeVersion + base64.StdEncoding.EncodeToString(payload), nil
}

// decrypt rejects an unrecognized envelope, a truncated payload, and any
// authentication failure instead of returning a partial or empty credential.
func (c *credentialCodec) decrypt(envelope string) (string, error) {
	if !strings.HasPrefix(envelope, credentialEnvelopeVersion) {
		return "", errors.New("credential envelope version is not supported")
	}
	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(envelope, credentialEnvelopeVersion))
	if err != nil {
		return "", fmt.Errorf("decode credential envelope: %w", err)
	}
	nonceSize := c.aead.NonceSize()
	if len(payload) < nonceSize {
		return "", errors.New("credential envelope is truncated")
	}
	plaintext, err := c.aead.Open(nil, payload[:nonceSize], payload[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("decrypt credential: %w", err)
	}
	return string(plaintext), nil
}
