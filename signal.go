package go_graceful_daemon

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// ErrStop should be returned signal handler function
// for termination of handling signals.
var ErrStop = errors.New("stop serve signals")

// SignalHandlerFunc is the interface for signal handler functions.
type SignalHandlerFunc func(sig os.Signal) (err error)

// SetSigHandler sets handler for the given signals.
// SIGTERM has the default handler, he returns ErrStop.
func SetSigHandler(handler SignalHandlerFunc, signals ...os.Signal) {
	for _, sig := range signals {
		handlers[sig] = handler
	}
}

// ServeSignals calls handlers for system signals.
func ServeSignals() (err error) {
	signals := make([]os.Signal, 0, len(handlers))
	for sig := range handlers {
		signals = append(signals, sig)
	}

	ch := make(chan os.Signal, 8)
	signal.Notify(ch, signals...)

	for sig := range ch {
		err = handlers[sig](sig)
		if err != nil {
			break
		}
	}

	signal.Stop(ch)

	if err == ErrStop {
		err = nil
	}

	return
}

var handlers = make(map[os.Signal]SignalHandlerFunc)

func init() {
	handlers[syscall.SIGINT] = sigtermDefaultHandler
	handlers[syscall.SIGTERM] = sigtermDefaultHandler
	handlers[syscall.SIGHUP] = sighupDefaultHandler
}

func sigtermDefaultHandler(sig os.Signal) error {
	log.Printf("[%d] %s stop graceful", os.Getpid(), AppName)
	log.Printf("[%d] %s stopped.", os.Getpid(), AppName)
	os.Remove(PidFile)
	os.Exit(1)
	return ErrStop
}

func sighupDefaultHandler(sig os.Signal) error {
	//only deamon时不支持kill -HUP,因为可能监听地址会占用
	log.Printf("[%d] %s stopped.", os.Getpid(), AppName)
	os.Remove(PidFile)
	os.Exit(2)
	return ErrStop
}
