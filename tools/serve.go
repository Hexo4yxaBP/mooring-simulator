package main

import (
	"log"
	"mime"
	"net/http"
)

func main() {
	// Ensure .wasm files are served with the correct MIME type.
	_ = mime.AddExtensionType(".wasm", "application/wasm")

	fs := http.FileServer(http.Dir("."))
	log.Println("Serving on http://localhost:8080")
	if err := http.ListenAndServe(":8080", fs); err != nil {
		log.Fatal(err)
	}
}
