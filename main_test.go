/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"errors"
	"os"
	"testing"
	"time"

	ecc "github.com/ernestio/ernest-config-client"
	"github.com/nats-io/nats"

	. "github.com/smartystreets/goconvey/convey"
)

func wait(ch chan bool) error {
	return waitTime(ch, 500*time.Millisecond)
}

func waitTime(ch chan bool, timeout time.Duration) error {
	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
	}
	return errors.New("timeout")
}

func TestBasicRedirections(t *testing.T) {
	n := ecc.NewConfig(os.Getenv("NATS_URI")).Nats()
	n.Subscribe("config.get.connectors", func(msg *nats.Msg) {
		n.Publish(msg.Reply, []byte(`{"networks":["fake","vcloud","aws","aws-fake","vcloud-fake"]}`))
	})
	setup()

	Convey("Given this service is fully set up", t, func() {
		chfak := make(chan bool)
		cherr := make(chan bool)
		chvcl := make(chan bool)
		chvfaws := make(chan bool)
		chvaws := make(chan bool)

		n.Subscribe("network.create.fake", func(msg *nats.Msg) {
			chfak <- true
		})
		n.Subscribe("network.create.error", func(msg *nats.Msg) {
			cherr <- true
		})
		n.Subscribe("network.create.vcloud", func(msg *nats.Msg) {
			chvcl <- true
		})
		n.Subscribe("network.create.fake-aws", func(msg *nats.Msg) {
			chvfaws <- true
		})
		n.Subscribe("network.create.aws", func(msg *nats.Msg) {
			chvaws <- true
		})

		n.Subscribe("network.create.aws", func(msg *nats.Msg) {
			ex := `{"_batch_id":"a","_type":"aws","_uuid":"c","datacenter_access_key":"k","datacenter_access_token":"t","datacenter_region":"r","datacenter_vpc_id":"n","network_subnet":"ns"}`
			if ex == string(msg.Data) {
				chvaws <- true
			}
		})

		Convey("When it receives an invalid fake message", func() {
			n.Publish("network.create", []byte(`{"service":"aaa"}`))
			Convey("Then it should redirect to network error creation", func() {
				So(wait(cherr), ShouldBeNil)
			})
		})

		Convey("When it receives a valid fake message", func() {
			n.Publish("network.create", []byte(`{"service":"aaa","router_type":"fake","range":"10.1.1.10/24"}`))
			Convey("Then it should redirect it to a fake connector", func() {
				So(wait(chfak), ShouldBeNil)
			})
		})

		Convey("When it receives a valid vcloud message", func() {
			n.Publish("network.create", []byte(`{"service":"aaa","router_type":"vcloud","range":"10.1.1.10/24"}`))
			Convey("Then it should redirect it to a fake connector", func() {
				So(wait(chvcl), ShouldBeNil)
			})
		})

		Convey("When it receives a valid aws message", func() {
			n.Publish("network.create", []byte(`{"_batch_id":"a","_uuid":"c","service":"aaa","router_type":"aws","range":"10.1.1.10/24","datacenter_region":"r","datacenter_access_token":"t","datacenter_access_key":"k","datacenter_name":"n","network_subnet":"ns"}`))
			Convey("Then it should redirect it to a fake connector", func() {
				So(wait(chvaws), ShouldBeNil)
			})
		})

		Convey("When it receives a valid fake-aws message", func() {
			n.Publish("network.create", []byte(`{"_batch_id":"a","_uuid":"c","service":"aaa","router_type":"fake-aws","range":"10.1.1.10/24","datacenter_region":"r","datacenter_access_token":"t","datacenter_access_key":"k","datacenter_name":"n","network_subnet":"ns"}`))
			Convey("Then it should redirect it to a fake connector", func() {
				So(wait(chvfaws), ShouldBeNil)
			})
		})
	})
}
