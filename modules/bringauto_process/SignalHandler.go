// Functions for global signal handling
//
// The SignalHandlerRegisterSignal() function must be called to start handling signals with this
// module. Afterward, SignalHandlerAddHandler() can be used to add handlers that will execute when
// a registered signal is received. The method returns a function which should be deffered or
// called later in program to execute handler and remove it from handling by this module. When a
// registered signal is received, all added (and not yet removed) handlers will be executed in
// reverse order and then the program exits with status code 1.
//
// Note: Do not use with concurrent programming. Can behave unexpectedly!

package bringauto_process

import (
	"bringauto/modules/bringauto_log"
	"os"
	"sync"
	"os/signal"
)


var lock sync.Mutex
var handlers []func() error

// SignalHandlerRegisterSignal
// Registers handling of specified signals to bringauto_process package
func SignalHandlerRegisterSignal(sig ...os.Signal) {
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

// SignalHandlerAddHandler
// Adds handler for execution after signal is received by bringauto_process package. Returns
// function, which executes handler and removes it from handling by bringauto_process module.
// The returned function should be deferred by caller. It should be used as this:
// handlerRemover := SignalHandlerAddHandler(my_handler)
// defer handlerRemover()
func SignalHandlerAddHandler(handler func() error) func() {
	lock.Lock()
	defer lock.Unlock()
	handlers = append(handlers, handler)
	return func() {
		lock.Lock()
		defer lock.Unlock()
		err := handler()
		if err != nil {
			bringauto_log.GetLogger().Error("Handler returned error - %s", err)
		}
		removeLastHandler()
	}
}

func removeLastHandler() {
	if len(handlers) == 0 {
		return
	}
	handlers = handlers[:len(handlers) - 1]
}

func executeAllHandlers() {
	for i := len(handlers)-1; i >= 0; i-- {
		err := handlers[i]()
		if err != nil {
			bringauto_log.GetLogger().Error("Handler returned error - %s", err)
		}
	}
}
