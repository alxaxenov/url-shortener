package main

import (
	"log"
	"os"
)

func main() {
	panic("test_panic") // want "call panic"
	log.Fatal("test_fatal")
	os.Exit(1)
}

func run() {
	panic("test_panic")     // want "call panic"
	log.Fatal("test_fatal") // want "call log.Fatal"
	os.Exit(1)              // want "call os.Exit"
}
