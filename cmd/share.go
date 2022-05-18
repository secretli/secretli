package cmd

import (
	"fmt"
	"github.com/secretli/secretli/internal"
	"github.com/spf13/cobra"
)

func createShareCmd(store *internal.HTTPRemoteStore) *cobra.Command {
	description := `Share a secret securely

Share a given secret and provide a user with a share secret.
This share secret allows someone else to retrieve this secret.

The share secret is never sent to the server!
`

	return &cobra.Command{
		Use:   "share [plaintext]",
		Short: "Share a secret securely",
		Long:  description,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			subkeys, err := internal.NewRandomKeySet()
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
}
