// main package of server side of app.
package main

import (
	"fmt"

	"github.com/Melikhov-p/goph-keeper/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("work is stopped")
		fmt.Println(err.Error())
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	fmt.Println(cfg.RPC.Address)
	return nil
}
