package main

import (
	"log"
	"time"
)

func main() {
	log.Println("Orchion Node Agent starting...")

	for {
		// TODO: send heartbeat to orchestrator
		// TODO: report capabilities
		time.Sleep(5 * time.Second)
	}
}
