package cmd

import (
	"github.com/secretli/secretli/internal"
	"github.com/spf13/cobra"
)

func createDeleteCmd() *cobra.Command {
	var password bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a shared secret",
		Long: `Delete a shared secret

Delete a given secret from the server by entering a deletion token.
The share secret is never sent to the server!
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			baseUrl, _ := cmd.Flags().GetString("base-url")
			store := internal.NewHTTPRemoteStore(baseUrl)

			pwd := ""
			if password {
				pwd = internal.GetPasswordFromTerminalOrDie()
			}

			keySet, err := internal.KeySetFromString(args[0], pwd)
			if err != nil {
				return err
			}

			return store.Delete(keySet, args[1])
		},
	}

	cmd.Flags().BoolVarP(&password, "password", "p", false, "ask for password")

	return cmd
}
