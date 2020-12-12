package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/koron/go-ssdp"
)

func main() {
	st := flag.String("st", "istellar:device1", "ST: Device Type")
	usn := flag.String("usn", "uuid:istellar:device1:1XX000000000", "USN: UUID")
	loc := flag.String("loc", "http://127.0.0.1:8080", "LOCATION: location header")
	srv := flag.String("srv", "", "SERVER:  server header")
	maxAge := flag.Int("maxage", 1800, "cache control, max-age")
	ai := flag.Int("ai", 600, "alive interval")
	v := flag.Bool("v", true, "verbose mode")
	h := flag.Bool("h", false, "show help")
	flag.Parse()
	if *h {
		flag.Usage()
		return
	}
	if *v {
		ssdp.Logger = log.New(os.Stderr, "[SSDP] ", log.LstdFlags)
	}

	ad, err := ssdp.Advertise(*st, *usn, *loc, *srv, *maxAge)
	if err != nil {
		log.Fatal(err)
	}
	var aliveTick <-chan time.Time
	if *ai > 0 {
		aliveTick = time.Tick(time.Duration(*ai) * time.Second)
	} else {
		aliveTick = make(chan time.Time)
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

loop:
	for {
		select {
		case <-aliveTick:
			ad.Alive()
		case <-quit:
			break loop
		}
	}
	ad.Bye()
	ad.Close()
}
