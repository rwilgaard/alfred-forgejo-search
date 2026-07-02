package util

import (
	"fmt"
	"os"

	aw "github.com/deanishe/awgo"
)

// GetIcon returns an icon from the workflow's icons/ directory, falling back
// to the workflow's own icon if the named icon doesn't exist.
func GetIcon(name string) *aw.Icon {
	iconPath := fmt.Sprintf("icons/%s.png", name)
	if _, err := os.Stat(iconPath); err == nil {
		return &aw.Icon{Value: iconPath}
	}
	return aw.IconWorkflow
}
