//go:build !slim

package cli

import (
	"github.com/coder/coder/v2/cli/clibase"
	"github.com/coder/coder/v2/codersdk"

	"golang.org/x/xerrors"
)

func (*RootCmd) dbcryptRotate() *clibase.Cmd {
	var (
		vals = new(codersdk.DeploymentValues)
		opts = vals.Options()
	)
	cmd := &clibase.Cmd{
		Use:   "dbcrypt-rotate --postgres-url <postgres_url> --external-token-encryption-keys <new-key>,<old-key>",
		Short: "Rotate database encryption keys",
		Options: clibase.OptionSet{
			*opts.ByName("Postgres Connection URL"),
			*opts.ByName("External Token Encryption Keys"),
		},
		Middleware: clibase.Chain(
			clibase.RequireNArgs(0),
		),
		Handler: func(inv *clibase.Invocation) error {
			return xerrors.Errorf("needs reimplementation")
			// ctx, cancel := context.WithCancel(inv.Context())
			// defer cancel()
			// logger := slog.Make(sloghuman.Sink(inv.Stdout))
			//
			// if vals.PostgresURL == "" {
			// return xerrors.Errorf("no database configured")
			// }
			//
			// if vals.ExternalTokenEncryptionKeys == nil || len(vals.ExternalTokenEncryptionKeys) < 2 {
			// return xerrors.Errorf("dbcrypt-rotate requires at least two external token encryption keys")
			// }
			//
			// newKey, err := base64.StdEncoding.DecodeString(vals.ExternalTokenEncryptionKeys[0])
			// if err != nil {
			// return xerrors.Errorf("new key must be base64-encoded")
			// }
			// oldKey, err := base64.StdEncoding.DecodeString(vals.ExternalTokenEncryptionKeys[1])
			// if err != nil {
			// return xerrors.Errorf("old key must be base64-encoded")
			// }
			// if bytes.Equal(newKey, oldKey) {
			// return xerrors.Errorf("old and new keys must be different")
			// }
			//
			// keys := make([][]byte, 0, len(vals.ExternalTokenEncryptionKeys))
			// for _, ek := range vals.ExternalTokenEncryptionKeys {
			// dk, err := base64.StdEncoding.DecodeString(ek)
			// if err != nil {
			// return xerrors.Errorf("key must be base64-encoded")
			// }
			// keys = append(keys, dk)
			// }
			//
			// ciphers, err := dbcrypt.NewCiphers(keys...)
			// if err != nil {
			// return xerrors.Errorf("create ciphers: %w", err)
			// }
			//
			// sqlDB, err := cli.ConnectToPostgres(inv.Context(), logger, "postgres", vals.PostgresURL.Value())
			// if err != nil {
			// return xerrors.Errorf("connect to postgres: %w", err)
			// }
			// defer func() {
			// _ = sqlDB.Close()
			// }()
			// logger.Info(ctx, "connected to postgres")
			//
			// db := database.New(sqlDB)
			//
			// cryptDB, err := dbcrypt.New(ctx, db, ciphers)
			// if err != nil {
			// return xerrors.Errorf("create cryptdb: %w", err)
			// }
			//
			// users, err := cryptDB.GetUsers(ctx, database.GetUsersParams{})
			// if err != nil {
			// return xerrors.Errorf("get users: %w", err)
			// }
			// for idx, usr := range users {
			// userLinks, err := cryptDB.GetUserLinksByUserID(ctx, usr.ID)
			// if err != nil {
			// return xerrors.Errorf("get user links for user: %w", err)
			// }
			// for _, userLink := range userLinks {
			// if _, err := cryptDB.UpdateUserLink(ctx, database.UpdateUserLinkParams{
			// OAuthAccessToken:  userLink.OAuthAccessToken,
			// OAuthRefreshToken: userLink.OAuthRefreshToken,
			// OAuthExpiry:       userLink.OAuthExpiry,
			// UserID:            usr.ID,
			// LoginType:         usr.LoginType,
			// }); err != nil {
			// return xerrors.Errorf("update user link: %w", err)
			// }
			// }
			// gitAuthLinks, err := cryptDB.GetGitAuthLinksByUserID(ctx, usr.ID)
			// if err != nil {
			// return xerrors.Errorf("get git auth links for user: %w", err)
			// }
			// for _, gitAuthLink := range gitAuthLinks {
			// if _, err := cryptDB.UpdateGitAuthLink(ctx, database.UpdateGitAuthLinkParams{
			// ProviderID:        gitAuthLink.ProviderID,
			// UserID:            usr.ID,
			// UpdatedAt:         gitAuthLink.UpdatedAt,
			// OAuthAccessToken:  gitAuthLink.OAuthAccessToken,
			// OAuthRefreshToken: gitAuthLink.OAuthRefreshToken,
			// OAuthExpiry:       gitAuthLink.OAuthExpiry,
			// }); err != nil {
			// return xerrors.Errorf("update git auth link: %w", err)
			// }
			// }
			// logger.Info(ctx, "encrypted user tokens", slog.F("current", idx+1), slog.F("of", len(users)))
			// }
			// logger.Info(ctx, "operation completed successfully")
			// return nil
		},
	}
	return cmd
}
