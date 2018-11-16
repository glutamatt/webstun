package main

import (
	"flag"
	"log"
	"os"

	"github.com/glutamatt/webstun/client"
	"github.com/glutamatt/webstun/server"
)

func main() {
	run := flag.String("run", "", "client|server")
	flag.Parse()

	if *run == "server" {
		port := ":3001"
		log.Printf("Let's go server http://0.0.0.0%s !\n", port)
		if err := server.ListenAndServe(port); err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	if *run == "client" {
		log.Println("Client to implement")
		client.Start("ws://0.0.0.0:3000/_ws", "https://grafana.deez.re")
		os.Exit(0)
	}

	flag.PrintDefaults()
	panic("no run")
}
