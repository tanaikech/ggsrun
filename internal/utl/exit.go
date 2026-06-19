package utl

import "os"

// TUIExitHandler can be set by the TUI package to handle application exits gracefully.
var TUIExitHandler func(int)

// CleanUpHandler can be set by execution packages to run rollback tasks before process exits.
var CleanUpHandler func()

// Exit terminates the process, or delegates to TUIExitHandler if defined.
func Exit(code int) {
	if CleanUpHandler != nil {
		CleanUpHandler()
	}
	if TUIExitHandler != nil {
		TUIExitHandler(code)
	}
	os.Exit(code)
}
