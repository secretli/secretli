package cmd

import (
	"fmt"
	"github.com/secretli/secretli/internal"
	"github.com/spf13/cobra"
)

func createShareCmd(store *internal.HTTPRemoteStore) *cobra.Command {
	var password bool

	cmd := &cobra.Command{
		Use:   "share [plaintext]",
		Short: "Share a secret securely",
		Long: `Share a secret securely

Share a given secret and provide a user with a share secret.
This share secret allows someone else to retrieve this secret.

The share secret is never sent to the server!
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
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

			encrypted, err := subkeys.Encrypt(args[0])
			if err != nil {
				return err
			}

			err = store.Store(subkeys, encrypted)
			if err != nil {
				return err
			}

			fmt.Println(subkeys.ShareSecret())
			return nil
		},
	}

	cmd.Flags().BoolVarP(&password, "password", "p", false, "ask for password")

	return cmd
}
