package main

import (
	"log"
	"net/http"

	"github.com/grandcat/zeroconf"
)

// HTTP Web Server Service
func startHTTPService() {
	http.Handle("/", http.FileServer(http.Dir("../")))
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Testing API"))
	})
	log.Println("Starting HTTP service on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Start out http service
	go startHTTPService()

	// Metadata information about the service
	metadata := []string{
		"version = 0.1",
		"developer = Shachindra",
	}

	mDNSService, err := zeroconf.Register(
		"Shachindra",     // service instance name
		"_istellar._tcp", // service type and protocol
		"local.",         // service domain
		8080,             // service port
		metadata,         // service metadata
		nil,              // register on all network interfaces
	)

	if err != nil {
		log.Fatal(err)
	}

	defer mDNSService.Shutdown()

	// Sleep forever
	select {}
}
