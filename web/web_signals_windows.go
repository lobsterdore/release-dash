// +build windows

package web

import (
	"os"
	"os/signal"
)

func notifySignals(runChan chan os.Signal) chan os.Signal {
	signal.Notify(runChan, os.Interrupt)
	return runChan
}
