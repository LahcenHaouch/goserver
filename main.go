package main

import (
	"fmt"
	"net/http"
)

func main() {
	const port string = "8080"
	mux := http.NewServeMux()
	serv := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.Handle("/", http.FileServer(http.Dir(".")))

	fmt.Println("Listening on port:8080")
	if err := serv.ListenAndServe(); err != nil {
		fmt.Println("Error 500: %w", err)
	}
}
