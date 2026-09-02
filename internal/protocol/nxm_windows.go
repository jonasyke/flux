//go:build windows

package protocol

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

func RegisterNXM() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find executable: %w", err)
	}

	key, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`SOFTWARE\Classes\nxm`,
		registry.SET_VALUE,
	)
	if err != nil {
		return err
	}
	defer key.Close()

	if err := key.SetStringValue("", "URL:Nexus Mods Protocol"); err != nil {
		return err
	}
	if err := key.SetStringValue("URL Protocol", ""); err != nil {
		return err
	}

	cmdKey, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`SOFTWARE\Classes\nxm\shell\open\command`,
		registry.SET_VALUE,
	)
	if err != nil {
		return err
	}
	defer cmdKey.Close()

	cmd := fmt.Sprintf(`"%s" "%s"`, exe, "%1")
	return cmdKey.SetStringValue("", cmd)
}

func UnregisterNXM() error {
	// Best-effort cleanup of the whole tree
	_ = registry.DeleteKey(registry.CURRENT_USER, `SOFTWARE\Classes\nxm\shell\open\command`)
	_ = registry.DeleteKey(registry.CURRENT_USER, `SOFTWARE\Classes\nxm\shell\open`)
	_ = registry.DeleteKey(registry.CURRENT_USER, `SOFTWARE\Classes\nxm\shell`)
	return registry.DeleteKey(registry.CURRENT_USER, `SOFTWARE\Classes\nxm`)
}
