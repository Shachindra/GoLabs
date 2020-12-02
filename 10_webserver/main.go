package main

import (
	"net/http"
)

func sayTest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("You have encountered a test use case"))
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("../")))
	http.HandleFunc("/test", sayTest)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
