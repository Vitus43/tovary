package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Vitus43/tovary/reporter"
)

func main() {
	r, err := reporter.NewReporter()
	if err != nil {
		panic(err)
	}

	r.WaitForTovary()

	// Create a channel to receive OS signals.
	sigChan := make(chan os.Signal, 1)

	// Notify the channel on Interrupt (Ctrl+C) or SIGTERM.
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Waiting for interrupt signal...")

	// Block until a signal is received.
	sig := <-sigChan
	fmt.Println("\nReceived signal:", sig)

	// Do cleanup if needed
	fmt.Println("Shutting down gracefully...")
}
