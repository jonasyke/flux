//go:build !windows

package protocol

import "fmt"

func RegisterNXM() error {
	return fmt.Errorf("nxm protocol registration is only implemented on Windows")
}

func UnregisterNXM() error {
	return fmt.Errorf("nxm protocol registration is only implemented on Windows")
}
