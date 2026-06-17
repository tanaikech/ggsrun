package utl

import "os"

// TUIExitHandler can be set by the TUI package to handle application exits gracefully.
var TUIExitHandler func(int)

// Exit terminates the process, or delegates to TUIExitHandler if defined.
func Exit(code int) {
	if TUIExitHandler != nil {
		TUIExitHandler(code)
	}
	os.Exit(code)
}
