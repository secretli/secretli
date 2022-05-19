package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := createRootCmd()
	rootCmd.AddCommand(
		createShareCmd(),
		createRetrieveCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func createRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secretli",
		Short: "Share secrets easily and securely across the internet.",
	}

	cmd.PersistentFlags().String("base-url", "", "use different server")

	return cmd
}
