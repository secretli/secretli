package cmd

import (
	"errors"
	"fmt"
	"github.com/secretli/secretli/internal"
	"github.com/spf13/cobra"
)

func createShareCmd() *cobra.Command {
	var password bool
	var expiration string
	var burnAfterRead bool

	cmd := &cobra.Command{
		Use:   "share",
		Short: "Share a secret securely",
		Long: `Share a secret securely

Share a given secret and provide a user with a share secret.
This share secret allows someone else to retrieve this secret.

The secret is read from stdin.
The share secret is never sent to the server!

Expiration Time:
A user can select the following expiration times (default: 5 minutes).

 - Minutes: 5m, 10m, 15m
 - Hours: 1h, 4h, 12h
 - Days: 1d, 3d, 7d

`,
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch expiration {
			case "5m", "10m", "15m", "1h", "4h", "12h", "1d", "3d", "7d":
				return nil
			default:
				return fmt.Errorf("invalid expiraton time selected: %s", expiration)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			plaintext, err := internal.ReadFromStdin()
			if err != nil {
				return fmt.Errorf("error reading secret: %w", err)
			}

			if len(plaintext) > 5000 {
				return errors.New("secret is too large (> 5000)")
			}

			baseUrl, _ := cmd.Flags().GetString("base-url")
			store := internal.NewHTTPRemoteStore(baseUrl)

			pwd := ""
			if password {
				pwd = internal.GetPasswordFromTerminalOrDie()
			}

			keySet, err := internal.NewKeySet(pwd)
			if err != nil {
				return err
			}

			encrypted, err := keySet.Encrypt(plaintext)
			if err != nil {
				return err
			}

			err = store.Store(keySet, encrypted, expiration, burnAfterRead)
			if err != nil {
				return err
			}

			baseUrlInOutput := ""
			if baseUrl != "" {
				baseUrlInOutput = fmt.Sprintf("--base-url '%s' ", baseUrl)
			}

			pwdFlagInOutput := ""
			if password {
				pwdFlagInOutput = "-p "
			}

			fmt.Println("Success!")
			fmt.Println()
			fmt.Println("Want to retrieve your secret?")
			fmt.Printf("$ secretli %sretrieve %s'%s'\n", baseUrlInOutput, pwdFlagInOutput, keySet.ShareSecret)
			fmt.Println()
			fmt.Println("Have to delete your secret?")
			fmt.Printf("$ secretli %sdelete %s'%s' '%s'\n", baseUrlInOutput, pwdFlagInOutput, keySet.ShareSecret, keySet.DeletionToken)
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().BoolVarP(&password, "password", "p", false, "ask for password")
	cmd.Flags().StringVarP(&expiration, "expiration", "e", "5m", "expiration time of secret")
	cmd.Flags().BoolVar(&burnAfterRead, "burn-after-read", false, "burn the secret after reading")

	return cmd
}
