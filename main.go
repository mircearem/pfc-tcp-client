package main

import (
	"log"
	"time"

	"github.com/mircearem/pfc-tcp-client/client"
)

const (
	MAX_ATTEMPTS    = 25
	REDIAL_INTERVAL = 15 * time.Second
	REMOTE_ADDR     = ":3000"
)

func main() {
	config, err := client.NewClientConfig(REMOTE_ADDR, REDIAL_INTERVAL, MAX_ATTEMPTS)
	if err != nil {
		log.Fatalln(err)

	}

	c := client.NewClient(config)
	if err := c.Run(); err != nil {
		log.Fatalln(err)
	}
}
