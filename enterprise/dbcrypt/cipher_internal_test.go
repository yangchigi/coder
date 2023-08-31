package dbcrypt

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCipherAES256(t *testing.T) {
	t.Parallel()

	t.Run("ValidInput", func(t *testing.T) {
		t.Parallel()
		key := bytes.Repeat([]byte{'a'}, 32)
		cipher, err := cipherAES256(key)
		require.NoError(t, err)

		output, err := cipher.Encrypt([]byte("hello world"))
		require.NoError(t, err)

		response, err := cipher.Decrypt(output)
		require.NoError(t, err)
		require.Equal(t, "hello world", string(response))
	})

	t.Run("InvalidInput", func(t *testing.T) {
		t.Parallel()
		key := bytes.Repeat([]byte{'a'}, 32)
		cipher, err := cipherAES256(key)
		require.NoError(t, err)
		_, err = cipher.Decrypt(bytes.Repeat([]byte{'a'}, 100))
		var decryptErr *DecryptFailedError
		require.ErrorAs(t, err, &decryptErr)
	})

	t.Run("InvalidKeySize", func(t *testing.T) {
		t.Parallel()

		_, err := cipherAES256(bytes.Repeat([]byte{'a'}, 31))
		require.ErrorContains(t, err, "key must be 32 bytes")
	})

	t.Run("TestNonce", func(t *testing.T) {
		t.Parallel()
		key := bytes.Repeat([]byte{'a'}, 32)
		cipher, err := cipherAES256(key)
		require.NoError(t, err)
		require.Equal(t, "864f702", cipher.HexDigest())

		encrypted1, err := cipher.Encrypt([]byte("hello world"))
		require.NoError(t, err)
		encrypted2, err := cipher.Encrypt([]byte("hello world"))
		require.NoError(t, err)
		require.NotEqual(t, encrypted1, encrypted2, "nonce should be different for each encryption")

		munged := make([]byte, len(encrypted1))
		copy(munged, encrypted1)
		munged[0] = munged[0] ^ 0xff
		_, err = cipher.Decrypt(munged)
		var decryptErr *DecryptFailedError
		require.ErrorAs(t, err, &decryptErr, "munging the first byte of the encrypted data should cause decryption to fail")
	})
}

func TestCiphers(t *testing.T) {
	t.Parallel()

	// Given: two ciphers
	key1 := bytes.Repeat([]byte{'a'}, 32)
	key2 := bytes.Repeat([]byte{'b'}, 32)
	cipher1, err := cipherAES256(key1)
	require.NoError(t, err)
	cipher2, err := cipherAES256(key2)
	require.NoError(t, err)

	cs := ciphers(cipher1, cipher2)
	require.NoError(t, err)

	// Then: it should encrypt with the cipher1
	output, err := cs.Encrypt([]byte("hello world"))
	require.NoError(t, err)
	// The first 7 bytes of the output should be the hex digest of cipher1
	require.Equal(t, cipher1.HexDigest(), string(output[:7]))

	// And: it should decrypt successfully
	decrypted, err := cs.Decrypt(output)
	require.NoError(t, err)
	require.Equal(t, "hello world", string(decrypted))

	// Decryption of the above should fail with cipher2
	_, err = cipher2.Decrypt(output)
	var decryptErr *DecryptFailedError
	require.ErrorAs(t, err, &decryptErr)

	// Decryption of data encrypted with cipher2 should succeed
	output2, err := cipher2.Encrypt([]byte("hello world"))
	require.NoError(t, err)
	decrypted2, err := cs.Decrypt(bytes.Join([][]byte{[]byte(cipher2.HexDigest()), output2}, []byte{'-'}))
	require.NoError(t, err)
	require.Equal(t, "hello world", string(decrypted2))

	// Decryption of data encrypted with cipher1 should succeed
	output1, err := cipher1.Encrypt([]byte("hello world"))
	require.NoError(t, err)
	decrypted1, err := cs.Decrypt(bytes.Join([][]byte{[]byte(cipher1.HexDigest()), output1}, []byte{'-'}))
	require.NoError(t, err)
	require.Equal(t, "hello world", string(decrypted1))

	// Wrapping a Ciphers with itself should panic.
	require.PanicsWithValue(t, "developer error: do not nest Ciphers", func() {
		_ = ciphers(cs)
	})
}

// This test ensures backwards compatibility. If it breaks, something is very wrong.
func TestCiphersBackwardCompatibility(t *testing.T) {
	t.Parallel()
	var (
		msg = "hello world"
		key = bytes.Repeat([]byte{'a'}, 32)
		//nolint: gosec // The below is the base64-encoded result of encrypting the above message with the above key.
		encoded = `ODY0ZjcwMi0Rrj5Q0mm+UDcKb0Ge4kU1HX2XsQ6yV7za/2lQrdVb/ZbaUjLBhuU=`
	)

	cipher, err := cipherAES256(key)
	require.NoError(t, err)
	require.Equal(t, "864f702", cipher.HexDigest())
	cs := ciphers(cipher)

	// This is the code that was used to generate the above.
	// Note that the output of this code will change every time it is run.
	// encrypted, err := cs.Encrypt([]byte(msg))
	// require.NoError(t, err)
	// t.Logf("encoded: %q", base64.StdEncoding.EncodeToString(encrypted))

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err, "the encoded string should be valid base64")
	decrypted, err := cs.Decrypt(decoded)
	require.NoError(t, err, "decryption should succeed")
	require.Equal(t, msg, string(decrypted), "decrypted message should match original message")
}
