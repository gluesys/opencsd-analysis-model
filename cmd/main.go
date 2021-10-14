package main

import (
	"analysis-model/pkg/rest"
	"log"
)

func main() {
	log.SetFlags(log.Lshortfile)
	log.Println("50500 Server Start")
	rest.Run()
}
