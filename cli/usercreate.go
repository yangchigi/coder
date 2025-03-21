package cli

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"golang.org/x/xerrors"

	"github.com/coder/pretty"

	"github.com/coder/coder/v2/cli/clibase"
	"github.com/coder/coder/v2/cli/cliui"
	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/coder/v2/cryptorand"
)

func (r *RootCmd) userCreate() *clibase.Cmd {
	var (
		email        string
		username     string
		password     string
		disableLogin bool
		loginType    string
	)
	client := new(codersdk.Client)
	cmd := &clibase.Cmd{
		Use: "create",
		Middleware: clibase.Chain(
			clibase.RequireNArgs(0),
			r.InitClient(client),
		),
		Handler: func(inv *clibase.Invocation) error {
			organization, err := CurrentOrganization(r, inv, client)
			if err != nil {
				return err
			}
			if username == "" {
				username, err = cliui.Prompt(inv, cliui.PromptOptions{
					Text: "Username:",
				})
				if err != nil {
					return err
				}
			}
			if email == "" {
				email, err = cliui.Prompt(inv, cliui.PromptOptions{
					Text: "Email:",
					Validate: func(s string) error {
						err := validator.New().Var(s, "email")
						if err != nil {
							return xerrors.New("That's not a valid email address!")
						}
						return err
					},
				})
				if err != nil {
					return err
				}
			}
			userLoginType := codersdk.LoginTypePassword
			if disableLogin && loginType != "" {
				return xerrors.New("You cannot specify both --disable-login and --login-type")
			}
			if disableLogin {
				userLoginType = codersdk.LoginTypeNone
			} else if loginType != "" {
				userLoginType = codersdk.LoginType(loginType)
			}

			if password == "" && userLoginType == codersdk.LoginTypePassword {
				// Generate a random password
				password, err = cryptorand.StringCharset(cryptorand.Human, 20)
				if err != nil {
					return err
				}
			}

			_, err = client.CreateUser(inv.Context(), codersdk.CreateUserRequest{
				Email:          email,
				Username:       username,
				Password:       password,
				OrganizationID: organization.ID,
				UserLoginType:  userLoginType,
			})
			if err != nil {
				return err
			}

			authenticationMethod := ""
			switch codersdk.LoginType(strings.ToLower(string(userLoginType))) {
			case codersdk.LoginTypePassword:
				authenticationMethod = `Your password is: ` + pretty.Sprint(cliui.DefaultStyles.Field, password)
			case codersdk.LoginTypeNone:
				authenticationMethod = "Login has been disabled for this user. Contact your administrator to authenticate."
			case codersdk.LoginTypeGithub:
				authenticationMethod = `Login is authenticated through GitHub.`
			case codersdk.LoginTypeOIDC:
				authenticationMethod = `Login is authenticated through the configured OIDC provider.`
			}

			_, _ = fmt.Fprintln(inv.Stderr, `A new user has been created!
Share the instructions below to get them started.
`+pretty.Sprint(cliui.DefaultStyles.Placeholder, "—————————————————————————————————————————————————")+`
Download the Coder command line for your operating system:
https://github.com/coder/coder/releases

Run `+pretty.Sprint(cliui.DefaultStyles.Code, "coder login "+client.URL.String())+` to authenticate.

Your email is: `+pretty.Sprint(cliui.DefaultStyles.Field, email)+`
`+authenticationMethod+`

Create a workspace  `+pretty.Sprint(cliui.DefaultStyles.Code, "coder create")+`!`)
			return nil
		},
	}
	cmd.Options = clibase.OptionSet{
		{
			Flag:          "email",
			FlagShorthand: "e",
			Description:   "Specifies an email address for the new user.",
			Value:         clibase.StringOf(&email),
		},
		{
			Flag:          "username",
			FlagShorthand: "u",
			Description:   "Specifies a username for the new user.",
			Value:         clibase.StringOf(&username),
		},
		{
			Flag:          "password",
			FlagShorthand: "p",
			Description:   "Specifies a password for the new user.",
			Value:         clibase.StringOf(&password),
		},
		{
			Flag:   "disable-login",
			Hidden: true,
			Description: "Deprecated: Use '--login-type=none'. \nDisabling login for a user prevents the user from authenticating via password or IdP login. Authentication requires an API key/token generated by an admin. " +
				"Be careful when using this flag as it can lock the user out of their account.",
			Value: clibase.BoolOf(&disableLogin),
		},
		{
			Flag: "login-type",
			Description: fmt.Sprintf("Optionally specify the login type for the user. Valid values are: %s. "+
				"Using 'none' prevents the user from authenticating and requires an API key/token to be generated by an admin.",
				strings.Join([]string{
					string(codersdk.LoginTypePassword), string(codersdk.LoginTypeNone), string(codersdk.LoginTypeGithub), string(codersdk.LoginTypeOIDC),
				}, ", ",
				)),
			Value: clibase.StringOf(&loginType),
		},
	}
	return cmd
}
