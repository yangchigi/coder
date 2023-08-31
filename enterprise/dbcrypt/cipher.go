package dbcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/xerrors"
)

// cipherAES256GCM is the name of the AES-256 cipher.
// This is used to identify the cipher used to encrypt a value.
// It is added to the digest to ensure that if, in the future,
// we add a new cipher type, and a key is re-used, we don't
// accidentally decrypt the wrong values.
// When adding a new cipher type, add a new constant here
// and ensure to add the cipher name to the digest of the new
// cipher type.
const (
	cipherAES256GCM = "aes256gcm"
)

type Cipher interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
	HexDigest() string
}

// cipherAES256 returns a new AES-256 cipher.
func cipherAES256(key []byte) (*aes256, error) {
	if len(key) != 32 {
		return nil, xerrors.Errorf("key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	// We add the cipher name to the digest to ensure that if, in the future,
	// we add a new cipher type, and a key is re-used, we don't accidentally
	// decrypt the wrong values.
	toDigest := []byte(cipherAES256GCM)
	toDigest = append(toDigest, key...)
	digest := fmt.Sprintf("%x", sha256.Sum256(toDigest))[:7]
	return &aes256{aead: aead, digest: digest}, nil
}

type aes256 struct {
	aead cipher.AEAD
	// digest is the first 7 bytes of the hex-encoded SHA-256 digest of aead.
	digest string
}

func (a *aes256) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, a.aead.NonceSize())
	_, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}
	dst := make([]byte, len(nonce))
	copy(dst, nonce)
	return a.aead.Seal(dst, nonce, plaintext, nil), nil
}

func (a *aes256) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < a.aead.NonceSize() {
		return nil, xerrors.Errorf("ciphertext too short")
	}
	decrypted, err := a.aead.Open(nil, ciphertext[:a.aead.NonceSize()], ciphertext[a.aead.NonceSize():], nil)
	if err != nil {
		return nil, &DecryptFailedError{Inner: err}
	}
	return decrypted, nil
}

func (a *aes256) HexDigest() string {
	return a.digest
}

type (
	CipherDigest string
	Ciphers      struct {
		primary string
		m       map[string]Cipher
	}
)

// NewCiphers returns a new Ciphers instance with the given keys.
// The first key in the list is the primary cipher. Any keys after the
// first are considered secondary ciphers and are only used for decryption.
func NewCiphers(keys ...[]byte) (*Ciphers, error) {
	var ks []Cipher
	for _, k := range keys {
		c, err := cipherAES256(k)
		if err != nil {
			return nil, err
		}
		ks = append(ks, c)
	}
	return ciphers(ks...), nil
}

func ciphers(cs ...Cipher) *Ciphers {
	var primary string
	m := make(map[string]Cipher)
	for idx, c := range cs {
		if _, ok := c.(*Ciphers); ok {
			panic("developer error: do not nest Ciphers")
		}
		m[c.HexDigest()] = c
		if idx == 0 {
			primary = c.HexDigest()
		}
	}
	return &Ciphers{primary: primary, m: m}
}

// Encrypt encrypts the given plaintext using the primary cipher and returns the
// ciphertext. The ciphertext is prefixed with the primary cipher's digest.
func (cs Ciphers) Encrypt(plaintext []byte) ([]byte, error) {
	c, ok := cs.m[cs.primary]
	if !ok {
		return nil, xerrors.Errorf("no ciphers configured")
	}
	prefix := []byte(c.HexDigest() + "-")
	encrypted, err := c.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	return append(prefix, encrypted...), nil
}

// Decrypt decrypts the given ciphertext using the cipher indicated by the
// ciphertext's prefix. The prefix is the first 7 bytes of the hex-encoded
// SHA-256 digest of the cipher name and key. Decryption will fail if the
// prefix does not match any of the configured ciphers.
func (cs Ciphers) Decrypt(ciphertext []byte) ([]byte, error) {
	requiredPrefix := string(ciphertext[:7])
	c, ok := cs.m[requiredPrefix]
	if !ok {
		return nil, xerrors.Errorf("missing required decryption cipher %s", requiredPrefix)
	}
	return c.Decrypt(ciphertext[8:])
}

// HexDigest returns the digest of the primary cipher. This consists of the
// first 7 bytes of the hex-encoded SHA-256 digest of the cipher name and key.
func (cs Ciphers) HexDigest() string {
	return cs.primary
}
