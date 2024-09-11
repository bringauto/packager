// Functions for global signal handling
//
// The Register() must be called to start handling signal by this module. Then AddHandler() can be
// called to add handlers to execute when the registered signal is received. The method returns
// function which should be deferred after AddHandler() call by caller. This returned function
// basically removes handler from execution after signal is received. When registered signal is
// received, all added handlers (and not removed yet) will be executed in reverse order and then
// the program exits with status code 1.
//
// Note: Do not use with concurrent programming. Can behave unexpectedly!

package bringauto_process

import (
	"bringauto/modules/bringauto_log"
	"fmt"
	"os"
	"sync"
	"os/signal"
)


var lock sync.Mutex
var handlers []func() error

// Registers handling of specified signals to bringauto_process package
func RegisterSignal(sig ...os.Signal) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, sig...)
	go func() {
		_ = <-sigs
		lock.Lock()
		defer lock.Unlock()
		logger := bringauto_log.GetLogger()
		logger.Info("SIGINT received - %d handlers to execute", len(handlers))
		executeAllHandlers()
		os.Exit(1)
	}()
}

// Adds handler for execution after signal is received by bringauto_process package. Returns handler
// remover which should be deferred by caller.
func AddHandler(handler func() error) func() {
	lock.Lock()
	defer lock.Unlock()
	handlers = append(handlers, handler)
	return func() {
		lock.Lock()
		defer lock.Unlock()
		removeLastHandler()
	}
}

func removeLastHandler() error {
	bringauto_log.GetLogger().Info("Removing last handler, new size: %d", len(handlers) - 1)
	if len(handlers) == 0 {
		return fmt.Errorf("no handler to remove")
	}
	handlers = handlers[:len(handlers) - 1]
	return nil
}

func executeAllHandlers() {
	for i := len(handlers)-1; i >= 0; i-- {
   		err := handlers[i]()
     	if err != nil {
      		bringauto_log.GetLogger().Error("Handler returned error - %s", err)
      	}
	}
}
