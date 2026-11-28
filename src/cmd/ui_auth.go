package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	forgejo "github.com/rwilgaard/alfred-forgejo-search/src/internal/forgejo"
	alfredui "github.com/rwilgaard/go-alfredutils/v2/alfred/ui"
	"github.com/spf13/cobra"
)

// envAuthToken is set by the Save action in workflow/login.alfredui. The
// token arrives in the environment, never in argv.
const envAuthToken = "auth_token"

// Component ids in workflow/login.alfredui/ui.json.
const (
	uiToken = "token"
	uiSave  = "save"
)

var uiAuthCmd = &cobra.Command{
	Use:    "ui-auth",
	Short:  "store the token from the login form",
	Hidden: true,
	Run: func(_ *cobra.Command, _ []string) {
		emit(uiAuth(os.Getenv(envAuthToken), checkToken))
	},
}

// checkToken verifies a token against the configured Forgejo instance.
func checkToken(token string) error {
	if cfg.URL == "" {
		return errors.New("forgejo_url is not configured in workflow settings")
	}
	client, err := forgejo.NewClient(cfg.URL, token, nil)
	if err != nil {
		return err
	}
	if err := client.TestAuthentication(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return nil
}

func uiAuth(token string, check func(string) error) alfredui.Actions {
	// Undo the busy state ui.json sets before running this.
	idle := alfredui.NewUpdate().
		Enable(true, uiSave).
		Set(uiSave, alfredui.Props{"text": "Save"}).
		Focus(uiToken)
	idle.Update.Relayout = true // the status label starts as a single space

	token = strings.TrimSpace(token)
	if token == "" {
		return idle.Status(uiStatus, "Enter an access token.", alfredui.ColourError)
	}
	if err := check(token); err != nil {
		return idle.Status(uiStatus, err.Error(), alfredui.ColourError)
	}
	if err := wf.SetAuth(keychainAccount, token); err != nil {
		return idle.Status(uiStatus, err.Error(), alfredui.ColourError)
	}

	// The finish args go to the notification after the Alfred UI object.
	return alfredui.Actions{Finish: &alfredui.Finish{Args: "Successfully authenticated with Forgejo"}}
}

func init() {
	rootCmd.AddCommand(uiAuthCmd)
}
