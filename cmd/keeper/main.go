// main package of server side of app.
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	application "github.com/Melikhov-p/goph-keeper/internal/app"
	"github.com/Melikhov-p/goph-keeper/internal/config"
	"golang.org/x/sync/errgroup"
)

const (
	timeoutServerShutdown = time.Second * 5
	timeoutShutdown       = time.Second * 10
)

func main() {
	if err := run(); err != nil {
		fmt.Println("work is stopped")
		fmt.Println(err.Error())
	}
}

func run() error {
	var (
		app *application.App
		cfg *config.Config
		eg  *errgroup.Group
		ctx context.Context
		err error
	)

	cfg, err = config.Load()
	if err != nil {
		return fmt.Errorf("failed to get config %w", err)
	}

	rootCtx, cancelCtx := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer cancelCtx()

	eg, ctx = errgroup.WithContext(rootCtx)
	// нештатное завершение программы по таймауту
	// происходит, если после завершения контекста
	// приложение не смогло завершиться за отведенный промежуток времени
	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	})

	app, err = application.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get app %w", err)
	}

	eg.Go(func() error {
		err = app.RunGRPC()
		if err != nil {
			return fmt.Errorf("error in gRPC server %w", err)
		}

		return nil
	})

	log.Println("server podnyalsya")

	eg.Go(func() error {
		<-ctx.Done()

		app.StopGRPC()

		return nil
	})

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("errgroup error: %w", err)
	}

	return nil
}
