package cmd

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tasks := []func() error{
		setLogLevel,
		printStartMessage,
		//setPostgreSQLConnection,
		setHjClient,
		chkRestart,
		readDev,
		autoUpload,
	}
	for _, t := range tasks {
		if err := t(); err != nil {
			log.Fatal(err)
		}
	}
	sigChan := make(chan os.Signal)
	exitChan := make(chan struct{})
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	log.WithField("signal", <-sigChan).Info("signal received")
	go func() {
		log.Warning("stopping fkhj212 process")
		exitChan <- struct{}{}
	}()
	select {
	case <-exitChan:
	case s := <-sigChan:
		log.WithField("signal", s).Info("signal received, stopping immediately")
	}

	return nil
}
