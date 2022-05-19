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
			store, err := internal.SetupStore(baseUrl)
			if err != nil {
				return err
			}

			var subkeys internal.KeySet
			if password {
				pwd := internal.GetPasswordFromTerminalOrDie()
				subkeys, err = internal.NewRandomKeySetWithPassword(pwd)
			} else {
				subkeys, err = internal.NewRandomKeySet()
			}
			if err != nil {
				return err
			}

			encrypted, err := subkeys.Encrypt(plaintext)
			if err != nil {
				return err
			}

			err = store.Store(subkeys, encrypted, expiration, burnAfterRead)
			if err != nil {
				return err
			}

			fmt.Println(subkeys.ShareSecret())
			return nil
		},
	}

	cmd.Flags().BoolVarP(&password, "password", "p", false, "ask for password")
	cmd.Flags().StringVarP(&expiration, "expiration", "e", "5m", "expiration time of secret")
	cmd.Flags().BoolVar(&burnAfterRead, "burn-after-read", false, "burn the secret after reading")

	return cmd
}
