//go:build windows

package tui

func platformInit(cap *Capabilities) {
	// console_windows.go init() already enables UTF-8 + VT processing.
	if asciiMode() {
		cap.Unicode = false
		cap.Rounded = false
	}
}
