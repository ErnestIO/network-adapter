/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/nats-io/nats"
	o "github.com/r3labs/otomo"
)

var nc *nats.Conn
var natsErr error

func getConnectorTypes(ctype string) []string {
	var connectors map[string][]string

	resp, err := nc.Request("config.get.connectors", nil, time.Second)
	if err != nil {
		log.Println("could not get config for connectors")
		log.Fatal(err)
	}

	err = json.Unmarshal(resp.Data, &connectors)
	if err != nil {
		log.Println("could not read config response")
		log.Fatal(err)
	}

	if connectors[ctype] == nil {
		log.Fatal("connector type not found")
	}

	return connectors[ctype]
}

func main() {
	natsURI := os.Getenv("NATS_URI")
	if natsURI == "" {
		natsURI = nats.DefaultURL
	}

	nc, natsErr = nats.Connect(natsURI)
	if natsErr != nil {
		log.Fatal(natsErr)
	}

	c := o.Config{
		Client:     nc,
		ValidTypes: getConnectorTypes("networks"),
	}

	log.Println("Setting up networks")
	o.StandardSubscription(&c, "network.create", "router_type")
	o.StandardSubscription(&c, "network.delete", "router_type")

	runtime.Goexit()
}
