//go:build windows

package tui

import (
	"os"
	"syscall"
	"unsafe"
)

// Windows console mode flag constants (wincon.h).
const (
	// enableVirtualTerminalProcessing enables ANSI/VT escape-sequence output.
	enableVirtualTerminalProcessing uint32 = 0x0004
	// disableNewlineAutoReturn prevents automatic CR+LF on LF output.
	disableNewlineAutoReturn uint32 = 0x0008
)

var (
	modKernel32         = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode  = modKernel32.NewProc("GetConsoleMode")
	procSetConsoleMode  = modKernel32.NewProc("SetConsoleMode")
	procSetConsoleCP    = modKernel32.NewProc("SetConsoleCP")
	procSetConsoleOutCP = modKernel32.NewProc("SetConsoleOutputCP")
)

func init() {
	initWindowsConsole()
}

// initWindowsConsole switches the Windows console to UTF-8 (code-page 65001)
// and enables ANSI/VT virtual-terminal processing so that Unicode characters
// and colour escape sequences display correctly in CMD.exe and PowerShell.
// This runs automatically at program start via Go's init() mechanism.
func initWindowsConsole() {
	// Switch console to UTF-8 so multi-byte characters encode correctly.
	procSetConsoleCP.Call(65001)
	procSetConsoleOutCP.Call(65001)

	// Enable VT/ANSI rendering for stdout and stderr.
	for _, fd := range []uintptr{os.Stdout.Fd(), os.Stderr.Fd()} {
		var mode uint32
		r, _, _ := procGetConsoleMode.Call(fd, uintptr(unsafe.Pointer(&mode)))
		if r != 0 {
			procSetConsoleMode.Call(fd,
				uintptr(mode|enableVirtualTerminalProcessing|disableNewlineAutoReturn))
		}
	}
}
