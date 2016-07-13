package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
)

type builderEvent struct {
	Uuid               string   `json:"_uuid"`
	BatchID            string   `json:"_batch_id"`
	Type               string   `json:"type"`
	Name               string   `json:"name"`
	Range              string   `json:"range"`
	Subnet             string   `json:"subnet,omitempty"`
	Netmask            string   `json:"netmask,omitempty"`
	StartAddress       string   `json:"start_address,omitempty"`
	EndAddress         string   `json:"end_address,omitempty"`
	Gateway            string   `json:"gateway,omitempty"`
	DNS                []string `json:"dns"`
	Router             string   `json:"router"`
	RouterType         string   `json:"router_type"`
	RouterName         string   `json:"router_name,omitempty"`
	ClientName         string   `json:"client_name"`
	DatacenterType     string   `json:"datacenter_type,omitempty"`
	DatacenterName     string   `json:"datacenter_name,omitempty"`
	DatacenterUsername string   `json:"datacenter_username,omitempty"`
	DatacenterPassword string   `json:"datacenter_password,omitempty"`
	VCloudURL          string   `json:"vcloud_url"`
	Status             string   `json:"status"`
	ErrorCode          string   `json:"error_code"`
	ErrorMessage       string   `json:"error_message"`
}

type vcloudEvent struct {
	Uuid                string   `json:"_uuid"`
	BatchID             string   `json:"_batch_id"`
	Type                string   `json:"type"`
	Service             string   `json:"service"`
	NetworkType         string   `json:"network_type"`
	NetworkName         string   `json:"network_name"`
	NetworkNetmask      string   `json:"network_netmask"`
	NetworkStartAddress string   `json:"network_start_address"`
	NetworkEndAddress   string   `json:"network_end_address"`
	NetworkGateway      string   `json:"network_gateway"`
	DNS                 []string `json:"network_dns"`
	RouterName          string   `json:"router_name"`
	RouterType          string   `json:"router_type,omitempty"`
	RouterIP            string   `json:"router_ip"`
	ClientName          string   `json:"client_name,omitempty"`
	DatacenterType      string   `json:"datacenter_type,omitempty"`
	DatacenterName      string   `json:"datacenter_name,omitempty"`
	DatacenterUsername  string   `json:"datacenter_username,omitempty"`
	DatacenterPassword  string   `json:"datacenter_password,omitempty"`
	DatacenterRegion    string   `json:"datacenter_region,omitempty"`
	VCloudURL           string   `json:"vcloud_url"`
}

type Translator struct{}

func (t Translator) BuilderToConnector(j []byte) []byte {
	var input builderEvent
	var output vcloudEvent

	json.Unmarshal(j, &input)
	octets := getIPOctets(input.Range)

	output.Uuid = input.Uuid
	output.BatchID = input.BatchID
	output.Type = input.Type
	output.RouterName = input.RouterName
	output.RouterType = input.RouterType
	output.NetworkType = input.RouterType
	output.NetworkName = input.Name
	output.ClientName = input.ClientName
	output.DatacenterName = input.DatacenterName
	output.DatacenterUsername = input.DatacenterUsername
	output.DatacenterPassword = input.DatacenterPassword
	output.DatacenterType = input.DatacenterType
	output.VCloudURL = input.VCloudURL
	output.DNS = input.DNS
	output.NetworkNetmask = ParseNetmask(input.Range)
	output.NetworkStartAddress = octets + ".5"
	output.NetworkEndAddress = octets + ".250"
	output.NetworkGateway = octets + ".1"

	body, _ := json.Marshal(output)
	return body
}

func (t Translator) ConnectorToBuilder(j []byte) []byte {
	var input vcloudEvent
	var output builderEvent

	json.Unmarshal(j, &input)

	output.Uuid = input.Uuid
	output.BatchID = input.BatchID
	output.Type = input.Type
	output.RouterName = input.RouterName
	output.RouterType = input.RouterType
	output.Name = input.NetworkName
	output.ClientName = input.ClientName
	output.DatacenterName = input.DatacenterName
	output.DatacenterUsername = input.DatacenterUsername
	output.DatacenterPassword = input.DatacenterPassword
	output.DatacenterType = input.DatacenterType
	output.VCloudURL = input.VCloudURL
	output.DNS = input.DNS
	output.Netmask = input.NetworkNetmask
	output.StartAddress = input.NetworkStartAddress
	output.EndAddress = input.NetworkEndAddress
	output.Gateway = input.NetworkGateway

	body, _ := json.Marshal(output)
	return body
}

func getIPOctets(rng string) string {
	// Splits the network range and returns the first three octets
	ip, _, err := net.ParseCIDR(rng)
	if err != nil {
		log.Println(err)
	}
	octets := strings.Split(ip.String(), ".")
	octets = append(octets[:3], octets[3+1:]...)
	octetString := strings.Join(octets, ".")
	return octetString
}

func ParseNetmask(r string) string {
	// Convert netmask hex to string, generated from network range CIDR
	_, nw, _ := net.ParseCIDR(r)
	hx, _ := hex.DecodeString(nw.Mask.String())
	netmask := fmt.Sprintf("%v.%v.%v.%v", hx[0], hx[1], hx[2], hx[3])
	return netmask
}
