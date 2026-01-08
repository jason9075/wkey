package clipboard

import (
	"fmt"
	"os/exec"
	"strings"
)

// CopyToClipboard writes the text to the Wayland clipboard using wl-copy
func CopyToClipboard(text string) error {
	cmd := exec.Command("wl-copy")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run wl-copy: %w", err)
	}
	return nil
}

// Paste triggers a paste action using wtype (Ctrl+V)
func Paste() error {
	// wtype -M ctrl -k v -m ctrl
	// -M ctrl: press modifier
	// -k v: press key v
	// -m ctrl: release modifier (optional if -M implies hold, but wtype usually needs explicit release or just -M for modifier)
	// Actually wtype syntax: wtype -M ctrl -k v -m ctrl (release)
	// Or simpler: wtype -M ctrl -k v
	// Let's try: wtype -M ctrl -k v -m ctrl
	
	// Check if wtype is available
	_, err := exec.LookPath("wtype")
	if err != nil {
		return fmt.Errorf("wtype not found: %w", err)
	}

	cmd := exec.Command("wtype", "-M", "ctrl", "-k", "v", "-m", "ctrl")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run wtype: %w", err)
	}
	return nil
}
