package cmd

import (
	"fmt"
	"strings"

	"github.com/ncruces/zenity"
	forgejo "github.com/rwilgaard/alfred-forgejo-search/src/internal/forgejo"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "authenticate with a Forgejo API token",
	RunE: func(_ *cobra.Command, _ []string) error {
		_, token, err := zenity.Password(zenity.Title("Enter Forgejo API Token"))
		if err != nil {
			return err
		}
		token = strings.TrimSpace(token)

		if cfg.URL == "" {
			return fmt.Errorf("forgejo_url is not configured in workflow settings")
		}

		client, err := forgejo.NewClient(cfg.URL, token)
		if err != nil {
			return err
		}
		if err := client.TestAuthentication(); err != nil {
			zerr := zenity.Error(
				fmt.Sprintf("Authentication failed: %s", err),
				zenity.ErrorIcon,
			)
			if zerr != nil {
				return err
			}
			return err
		}

		if err := wf.Keychain.Set(keychainAccount, token); err != nil {
			return err
		}
		fmt.Println("Successfully authenticated with Forgejo")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
