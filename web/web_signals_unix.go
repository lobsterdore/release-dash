// +build linux bsd darwin

package web

import (
	"os"
	"os/signal"
	"syscall"
)

func notifySignals(runChan chan os.Signal) chan os.Signal {
	signal.Notify(runChan, os.Interrupt, syscall.SIGTSTP)
	return runChan
}
