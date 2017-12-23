package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rickbau5/anomaly-tracker-server/cmd/internal/tracker"
	"golang.org/x/net/unix"
)

func main() {
	conf := tracker.InitConfig()

	server := &http.Server{
		Addr:         conf.ListenAddr,
		IdleTimeout:  conf.IdleTimeout,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	}

	go runSever(server)

	awaitShutdown(server)
}

func awaitShutdown(server *http.Server) error {
	c := make(chan os.Signal)
	signal.Notify(c,
		os.Interrupt,
		unix.SIGTERM,
		unix.SIGINT,
		unix.SIGABRT,
		unix.SIGHUP,
		unix.SIGKILL,
		unix.SIGQUIT,
		unix.SIGKILL,
	)
	stop := <-c
	fmt.Println("Got stop signal:", stop.String())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	// Wait the rest of the time if context hasn't expired yet
	select {
	case _, ok := <-ctx.Done():
		if !ok {
			fmt.Println("Context closed.")
		}
	}

	return err
}

func runSever(server *http.Server) {
	fmt.Println("Starting server.")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println("Server encountered an unexpected error:", err.Error())
		return
	}
	fmt.Println("Server stopping cleanly.")
}
