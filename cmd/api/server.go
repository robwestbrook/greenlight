package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// Declare an HTTP server struct.
	// Parameters:
	//	1.	Addr: server port
	//	2.	Handler: use the httprouter in the routes file
	//	3.	ErrorLog: a new Go logger instance with the
	//								custom logger. The "" and 0 indicate
	//								no logger prefix or any flags.
	//	4.	IdleTimeout
	//	5.	ReadTimeout
	//	6.	WriteTimeout
	srv := &http.Server{
		Addr:					fmt.Sprintf(":%d", app.config.port),
		Handler:			app.routes(),
		IdleTimeout: 	time.Minute,
		ReadTimeout: 	10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Create a shutdownError channel. This is used to
	// recieve any errors returned by the graceful
	// Shutdown() function.
	shutdownError := make(chan error)

	// Start a background goroutine.
	go func() {
		// Create a quit channel which carries
		// the Signal values. The channel is a buffered
		// channel with size 1.
		quit := make(chan os.Signal, 1)

		// Use signal.Notify() to listen for incoming
		// SIGINT or SIGTERM signals and relay them to the
		// quit channel. Any other signals will not be
		// caught by signal.Notify() and will retain their
		// default value.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel. This code
		// will block until a signal is received.
		s := <-quit

		// Log a message that signal has been caught. Call
		// the String() method to get the signal name and
		// include it in the log entry properties.
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		// Create a context with a 5 second timeout.
		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

		defer cancel()

		// Call Shutdown() on the server, passing in the
		// context. Shutdown() will return nil if the 
		// graceful shutdown is successful, or an error.
		// This is relayed to the shutdownError channel.
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		// Log a message for waiting for any background
		// goroutines to complete their tasks.
		app.logger.PrintInfo(
			"completing background tasks",
			map[string]string{
				"addr": srv.Addr,
			},
		)

		// Call Wait() to block until the WaitGroup counter
		// is zero. Then return nil on the shutdowmError
		// channel, to indicate the shutdown completed
		// without any issues.
		app.wg.Wait()
		shutdownError <- nil
	}()

	// Log a "starting server" message.
	app.logger.PrintInfo("starting server", map[string]string {
		"addr": srv.Addr,
		"env": app.config.env,
	})

	// Start the server. Calling Shutdown() on the server
	// will cause ListenAndServe() to immediately return
	// a http.ErrServerClosed error. The error returns
	// only if it is NOT http.ErrServerClosed.
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// If there is an error returned from Shutdown() on
	// the shutdownError channel, there is a problem with
	// the graceful shutdown and the error is returned.
	err = <- shutdownError
	if err != nil {
		return err
	}

	// If this position is reached, the graeful shutdown
	// completed successfully. Log a "stopped server"
	// message.
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})
		
	return nil
	
}