package main

import (
	"fmt"
	"net/http"
)

type apiHandler struct{}

func (apiHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("Hello Lahcen"))
}

func main() {
	var serv http.Server
	serv.Addr = ":8080"
	mux := http.NewServeMux()
	mux.Handle("/", apiHandler{})

	fmt.Println("Listening on port:8080")
	if err := serv.ListenAndServe(); err != nil {
		fmt.Println("Error 500: %w", err)
	}
}
