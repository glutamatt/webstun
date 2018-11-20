package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/glutamatt/webstun/client"
	"github.com/glutamatt/webstun/server"
)

var appVersion string

func main() {
	run := flag.String("run", "", "client|server")
	ws := flag.String("ws", "", "websocket url")
	port := flag.Int("port", 0, "server port")
	back := flag.String("back", "", "backend url")
	insecure := flag.Bool("insecure", false, "Skip TLS verifications")
	flag.Parse()

	if *run == "server" {
		if *port == 0 {
			crash(fmt.Errorf("server port is not set"))
		}
		log.Printf("Let's go server http://0.0.0.0:%d !\n", *port)
		if err := server.ListenAndServe(fmt.Sprintf(":%d", *port)); err != nil {
			crash(fmt.Errorf("server.ListenAndServe error: %v", err))
		}
		os.Exit(0)
	}

	if *run == "client" {
		if err := client.ConnectWSAndServe(*ws, *back, *insecure); err != nil {
			crash(fmt.Errorf("client error: %v", err))
		}
		os.Exit(0)
	}

	crash(fmt.Errorf("No run"))
}

func crash(err error) {
	flag.PrintDefaults()
	log.Fatalf("Error: %v", err)
}
