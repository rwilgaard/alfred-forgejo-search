package cmd

import (
	"os"

	"github.com/rwilgaard/go-alfredutils/v2/alfred/clipboard"
	alfredui "github.com/rwilgaard/go-alfredutils/v2/alfred/ui"
	"github.com/spf13/cobra"
)

// envErrorMsg is set by the error trigger and read by workflow/error.alfredui.
const envErrorMsg = "error_msg"

// Component ids in workflow/error.alfredui/ui.json.
const (
	uiMessage  = "message"
	uiRetryBtn = "retry"
)

// uiRetryCmd refreshes the cache again from the error window. It calls the
// refresh directly, not the cache command: that one reports through
// wf.ReportError, which would open a second error window.
var uiRetryCmd = &cobra.Command{
	Use:    "ui-retry",
	Short:  "retry the cache refresh from the error window",
	Hidden: true,
	Run: func(_ *cobra.Command, _ []string) {
		emit(uiRetry(fetchAndCacheRepos))
	},
}

var uiCopyErrorCmd = &cobra.Command{
	Use:    "ui-copy-error",
	Short:  "copy the error message from the error window",
	Hidden: true,
	Run: func(_ *cobra.Command, _ []string) {
		// An error message isn't a secret, so a plain copy.
		if _, err := clipboard.Copy(os.Getenv(envErrorMsg)); err != nil {
			emit(status(err.Error(), alfredui.ColourError))
			return
		}
		emit(status("✓ Copied error", alfredui.ColourSuccess))
	},
}

func uiRetry(refresh func() error) alfredui.Actions {
	err := refresh()
	if err == nil {
		return alfredui.Actions{Finish: &alfredui.Finish{}}
	}
	a := alfredui.NewUpdate().
		Set(uiMessage, alfredui.Props{"text": err.Error()}).
		Enable(true, uiRetryBtn).
		Set(uiRetryBtn, alfredui.Props{"text": "Retry  ↩"}).
		Status(uiStatus, "Still failing", alfredui.ColourError).
		SetVar(envErrorMsg, err.Error()) // so ⌘C copies the new error
	a.Update.Relayout = true // the new message may need more lines
	return a
}

func init() {
	rootCmd.AddCommand(uiRetryCmd, uiCopyErrorCmd)
}
