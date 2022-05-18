package cmd

import (
	"fmt"
	"github.com/secretli/secretli/internal"
	"github.com/spf13/cobra"
)

func createRetrieveCmd(store *internal.HTTPRemoteStore) *cobra.Command {
	var password bool

	cmd := &cobra.Command{
		Use:   "retrieve [share secret]",
		Short: "Retrieve a shared secret",
		Long: `Retrieve a shared secret

Use a share secret to retrieve a secret from the remote store.
The secret is decrypted on your computer.

The share secret is never sent to the server!
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var subkeys internal.KeySet

			if password {
				pwd := internal.GetPasswordFromTerminalOrDie()
				subkeys, err = internal.KeySetWithPasswordFromString(args[0], pwd)
			} else {
				subkeys, err = internal.KeySetFromString(args[0])
			}
			if err != nil {
				return err
			}

			encryptedData, err := store.Load(subkeys)
			if err != nil {
				return err
			}

			plaintext, err := subkeys.Decrypt(encryptedData)
			if err != nil {
				return err
			}

			fmt.Println(plaintext)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&password, "password", "p", false, "ask for password")

	return cmd
}
