package main

import (
	"log"
	"os"
)

func main() {
	panic("test_panic") // want "call panic"
	log.Fatal("test_fatal")
	log.Fatalf("test_fatalf")
	log.Fatalln("test_fatalln")
	os.Exit(1)
}

func run() {
	panic("test_panic")         // want "call panic"
	log.Fatal("test_fatal")     // want "call log.Fatal"
	log.Fatalf("test_fatalf")   // want "call log.Fatalf"
	log.Fatalln("test_fatalln") // want "call log.Fatalln"
	os.Exit(1)                  // want "call os.Exit"
}
