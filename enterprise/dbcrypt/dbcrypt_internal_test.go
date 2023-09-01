package dbcrypt

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/coder/v2/coderd/database"
	"github.com/coder/coder/v2/coderd/database/dbgen"
	"github.com/coder/coder/v2/coderd/database/dbtestutil"
)

func TestUserLinks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("InsertUserLink", func(t *testing.T) {
		t.Parallel()
		db, crypt, ciphers := setup(t)
		user := dbgen.User(t, crypt, database.User{})
		link := dbgen.UserLink(t, crypt, database.UserLink{
			UserID:            user.ID,
			OAuthAccessToken:  "access",
			OAuthRefreshToken: "refresh",
		})
		require.Equal(t, link.OAuthAccessToken, "access")
		require.Equal(t, link.OAuthRefreshToken, "refresh")

		link, err := db.GetUserLinkByLinkedID(ctx, link.LinkedID)
		require.NoError(t, err)
		requireEncryptedEquals(t, ciphers[0], link.OAuthAccessToken, "access")
		requireEncryptedEquals(t, ciphers[0], link.OAuthRefreshToken, "refresh")
	})

	t.Run("UpdateUserLink", func(t *testing.T) {
		t.Parallel()
		db, crypt, ciphers := setup(t)
		user := dbgen.User(t, crypt, database.User{})
		link := dbgen.UserLink(t, crypt, database.UserLink{
			UserID: user.ID,
		})
		updated, err := crypt.UpdateUserLink(ctx, database.UpdateUserLinkParams{
			OAuthAccessToken:  "access",
			OAuthRefreshToken: "refresh",
			UserID:            link.UserID,
			LoginType:         link.LoginType,
		})
		require.NoError(t, err)
		require.Equal(t, updated.OAuthAccessToken, "access")
		require.Equal(t, updated.OAuthRefreshToken, "refresh")

		link, err = db.GetUserLinkByLinkedID(ctx, link.LinkedID)
		require.NoError(t, err)
		requireEncryptedEquals(t, ciphers[0], link.OAuthAccessToken, "access")
		requireEncryptedEquals(t, ciphers[0], link.OAuthRefreshToken, "refresh")
	})

	t.Run("GetUserLinkByLinkedID", func(t *testing.T) {
		t.Parallel()
		db, crypt, ciphers := setup(t)
		user := dbgen.User(t, crypt, database.User{})
		link := dbgen.UserLink(t, crypt, database.UserLink{
			UserID:            user.ID,
			OAuthAccessToken:  "access",
			OAuthRefreshToken: "refresh",
		})
		link, err := db.GetUserLinkByLinkedID(ctx, link.LinkedID)
		require.NoError(t, err)
		requireEncryptedEquals(t, ciphers[0], link.OAuthAccessToken, "access")
		requireEncryptedEquals(t, ciphers[0], link.OAuthRefreshToken, "refresh")
	})

	t.Run("GetUserLinkByUserIDLoginType", func(t *testing.T) {
		t.Parallel()
		db, crypt, ciphers := setup(t)
		user := dbgen.User(t, crypt, database.User{})
		link := dbgen.UserLink(t, crypt, database.UserLink{
			UserID:            user.ID,
			OAuthAccessToken:  "access",
			OAuthRefreshToken: "refresh",
		})
		link, err := db.GetUserLinkByUserIDLoginType(ctx, database.GetUserLinkByUserIDLoginTypeParams{
			UserID:    link.UserID,
			LoginType: link.LoginType,
		})
		require.NoError(t, err)
		requireEncryptedEquals(t, ciphers[0], link.OAuthAccessToken, "access")
		requireEncryptedEquals(t, ciphers[0], link.OAuthRefreshToken, "refresh")
	})
}

func TestGitAuthLinks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("InsertGitAuthLink", func(t *testing.T) {
		t.Parallel()
		db, crypt, ciphers := setup(t)
		link := dbgen.GitAuthLink(t, crypt, database.GitAuthLink{
			OAuthAccessToken:  "access",
			OAuthRefreshToken: "refresh",
		})
		require.Equal(t, link.OAuthAccessToken, "access")
		require.Equal(t, link.OAuthRefreshToken, "refresh")

		link, err := db.GetGitAuthLink(ctx, database.GetGitAuthLinkParams{
			ProviderID: link.ProviderID,
			UserID:     link.UserID,
		})
		require.NoError(t, err)
		requireEncryptedEquals(t, ciphers[0], link.OAuthAccessToken, "access")
		requireEncryptedEquals(t, ciphers[0], link.OAuthRefreshToken, "refresh")
	})

	t.Run("UpdateGitAuthLink", func(t *testing.T) {
		t.Parallel()
		db, crypt, ciphers := setup(t)
		link := dbgen.GitAuthLink(t, crypt, database.GitAuthLink{})
		updated, err := crypt.UpdateGitAuthLink(ctx, database.UpdateGitAuthLinkParams{
			ProviderID:        link.ProviderID,
			UserID:            link.UserID,
			OAuthAccessToken:  "access",
			OAuthRefreshToken: "refresh",
		})
		require.NoError(t, err)
		require.Equal(t, updated.OAuthAccessToken, "access")
		require.Equal(t, updated.OAuthRefreshToken, "refresh")

		link, err = db.GetGitAuthLink(ctx, database.GetGitAuthLinkParams{
			ProviderID: link.ProviderID,
			UserID:     link.UserID,
		})
		require.NoError(t, err)
		requireEncryptedEquals(t, ciphers[0], link.OAuthAccessToken, "access")
		requireEncryptedEquals(t, ciphers[0], link.OAuthRefreshToken, "refresh")
	})

	t.Run("GetGitAuthLink", func(t *testing.T) {
		t.Parallel()
		db, crypt, ciphers := setup(t)
		link := dbgen.GitAuthLink(t, crypt, database.GitAuthLink{
			OAuthAccessToken:  "access",
			OAuthRefreshToken: "refresh",
		})
		link, err := db.GetGitAuthLink(ctx, database.GetGitAuthLinkParams{
			UserID:     link.UserID,
			ProviderID: link.ProviderID,
		})
		require.NoError(t, err)
		requireEncryptedEquals(t, ciphers[0], link.OAuthAccessToken, "access")
		requireEncryptedEquals(t, ciphers[0], link.OAuthRefreshToken, "refresh")
	})
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()
		// Given: a cipher is loaded
		cipher := initCipher(t)
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		rawDB, _ := dbtestutil.NewDB(t)

		// Before: no keys should be present
		keys, err := rawDB.GetActiveDBCryptKeys(ctx)
		require.Empty(t, keys, "no keys should be present")
		require.ErrorIs(t, err, sql.ErrNoRows)

		// When: we init the crypt db
		_, err = New(ctx, rawDB, cipher)
		require.NoError(t, err)

		// Then: a new key is inserted
		keys, err = rawDB.GetActiveDBCryptKeys(ctx)
		require.NoError(t, err)
		require.Len(t, keys, 1, "one key should be present")
		require.Equal(t, cipher.HexDigest(), keys[0].ActiveKeyDigest, "key digest mismatch")
		require.Equal(t, "coder", keys[0].Test, "test value mismatch")
	})

	t.Run("NoCipher", func(t *testing.T) {
		t.Parallel()
		// Given: no cipher is loaded
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		rawDB, _ := dbtestutil.NewDB(t)

		keys, err := rawDB.GetActiveDBCryptKeys(ctx)
		require.Empty(t, keys, "no keys should be present")
		require.ErrorIs(t, err, sql.ErrNoRows)

		// When: we init the crypt db with no ciphers
		cs := make([]Cipher, 0)
		_, err = New(ctx, rawDB, cs...)

		// Then: an error is returned
		require.ErrorContains(t, err, "no ciphers configured")

		// Assert invariant: no keys are inserted
		require.Empty(t, keys)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("CipherChanged", func(t *testing.T) {
		t.Parallel()
		// Given: no cipher is loaded
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		rawDB, _ := dbtestutil.NewDB(t)

		// Given: a key with a different cipher is inserted
		// Simulation: this will just be random data
		bs := make([]byte, 16)
		_, err := io.ReadFull(rand.Reader, bs)
		require.NoError(t, err)
		b64encrypted := base64.StdEncoding.EncodeToString(bs)
		require.NoError(t, rawDB.InsertDBCryptKey(ctx, database.InsertDBCryptKeyParams{
			Number:          1,
			ActiveKeyDigest: "deadbeef",
			Test:            b64encrypted,
		}))

		// When: we init the crypt db with no access to the old cipher
		cipher1 := initCipher(t)
		_, err = New(ctx, rawDB, cipher1)
		// Then: a special error is returned
		var derr *DecryptFailedError
		require.ErrorAs(t, err, &derr)

		// And the sentinel value should remain unchanged. For now.
		keys, err := rawDB.GetActiveDBCryptKeys(ctx)
		require.NoError(t, err)
		require.Len(t, keys, 1, "one key should be present")
		require.Equal(t, "deadbeef", keys[0].ActiveKeyDigest.String, "key digest mismatch")
	})
}

func requireEncryptedEquals(t *testing.T, c Cipher, value, expected string) {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(value)
	require.NoError(t, err, "invalid base64")
	got, err := c.Decrypt(data)
	require.NoError(t, err, "failed to decrypt data")
	require.Equal(t, expected, string(got), "decrypted data does not match")
}

func initCipher(t *testing.T) *aes256 {
	t.Helper()
	key := make([]byte, 32) // AES-256 key size is 32 bytes
	_, err := io.ReadFull(rand.Reader, key)
	require.NoError(t, err)
	c, err := cipherAES256(key)
	require.NoError(t, err)
	return c
}

func setup(t *testing.T) (db, cryptodb database.Store, cs []Cipher) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	rawDB, _ := dbtestutil.NewDB(t)

	cs = append(cs, initCipher(t))
	cryptDB, err := New(ctx, rawDB, cs...)
	require.NoError(t, err)

	return rawDB, cryptDB, cs
}
