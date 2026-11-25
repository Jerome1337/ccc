//go:build darwin || linux

package term

import (
	"os"
	"syscall"
	"unsafe"
)

// isTerminal asks the kernel for the terminal settings, like isatty(3). Checking for a character
// device isn't enough: /dev/null is one too, and `ccc update </dev/null` must not prompt.
func isTerminal(file *os.File) bool {
	var settings syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), ioctlReadTermios, uintptr(unsafe.Pointer(&settings)))

	return errno == 0
}
