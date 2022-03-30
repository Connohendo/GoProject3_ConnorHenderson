package main

import (
	"log"
	"net/http"
)

// https://itnext.io/webassembley-with-golang-by-scratch-e05ec5230558

const (
	AddSrv       = ":8080"
	TemplatesDir = "."
)

func main() {
	log.Printf("listening in %q...", AddSrv)
	fileSrv := http.FileServer(http.Dir(TemplatesDir))
	if err := http.ListenAndServe(AddSrv, fileSrv); err != nil {
		log.Fatal(err)
	}
}
