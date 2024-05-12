package grace

import (
	"os"
	"os/signal"
	"syscall"
)

var interrupt = make(chan os.Signal, 1)

// Prepare preparing grace stop
// should be called firstly
func Prepare() {
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
}

func Sig() *chan os.Signal {
	return &interrupt
}
