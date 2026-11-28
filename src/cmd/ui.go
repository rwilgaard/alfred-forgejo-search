package cmd

import (
	"log"

	alfredui "github.com/rwilgaard/go-alfredutils/v2/alfred/ui"
)

// The ui-* commands are run by Alfred UI script actions (workflow/*.alfredui,
// through workflow/scripts/run.sh). Their stdout must be JSON actions only, so
// they use Run (not RunE: a returned error would make run() call
// wf.FatalError) and never send Script Filter output.

// Component id of the status label every window has.
const uiStatus = "status"

func emit(a alfredui.Actions) {
	if err := alfredui.Emit(a); err != nil {
		log.Printf("[ui] emit failed: %v", err)
	}
}

// status sets the status label. It relays out the window: a label keeps the
// width of its first layout, which for the status label is a single space.
func status(msg, colour string) alfredui.Actions {
	a := alfredui.NewUpdate().Status(uiStatus, msg, colour)
	a.Update.Relayout = true
	return a
}

// orSpace keeps a label's height when its text is empty.
func orSpace(s string) string {
	if s == "" {
		return " "
	}
	return s
}
