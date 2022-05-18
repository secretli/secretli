package cmd

import (
	"github.com/secretli/secretli/internal"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	store, err := setupStore()
	if err != nil {
		log.Fatalln(err)
	}

	rootCmd := createRootCmd()
	rootCmd.AddCommand(
		createShareCmd(store),
		createRetrieveCmd(store),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func createRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "secretli",
		Short: "Share secrets easily and securely across the internet.",
	}
}

func setupStore() (*internal.HTTPRemoteStore, error) {
	client, err := internal.NewClient()
	if err != nil {
		return nil, err
	}
	return internal.NewHTTPRemoteStore(client), nil
}
