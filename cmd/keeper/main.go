package main

import "fmt"

func main() {
	if err := run(); err != nil {
		fmt.Println("work is stopped")
	}
}

func run() error {
	return nil
}
