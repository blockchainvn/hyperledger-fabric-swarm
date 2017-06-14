// This code generates crypto-config.yaml for each one's flavour.

package main

import (
  "strconv"
	"fmt"
)

type HostnameData struct {
	Prefix string
	Index  int
	Domain string
}

type SpecData struct {
	Hostname   string
	Domain     string
	CommonName string
}

type NodeTemplate struct {
	Count    int      `yaml:"Count,omitempty"`
	Start    int      `yaml:"Start,omitempty"`
	Hostname string   `yaml:"Hostname,omitempty"`
	SANS     []string `yaml:"SANS,omitempty"`
}

type NodeSpec struct {
	Hostname   string   `yaml:"Hostname,omitempty"`
	CommonName string   `yaml:"CommonName,omitempty"`
	SANS       []string `yaml:"SANS,omitempty"`
}

type UsersSpec struct {
	Count int `yaml:"Count,omitempty"`
}

type OrgSpec struct {
	Name     string       `yaml:"Name,omitempty"`
	Domain   string       `yaml:"Domain,omitempty"`
	CA       NodeSpec     `yaml:"CA,omitempty"`
	Template NodeTemplate `yaml:"Template,omitempty"`
	Specs    []NodeSpec   `yaml:"Specs,omitempty"`
	Users    UsersSpec    `yaml:"Users,omitempty"`
}

type Config struct {
	OrdererOrgs []OrgSpec `yaml:"OrdererOrgs,omitempty"`
	PeerOrgs    []OrgSpec `yaml:"PeerOrgs,omitempty"`
}

func GenCrypto(domainName string, numOrgs int, numPeer int, numOrderer int) (Config, error){
  fmt.Println("Generating Orderer's crypto config....")
  ordererData, err := GenOrdererConfig(domainName, numOrderer)
  check(err)
	fmt.Println("Generating Peer's crypto config....")
  peerData, err := GenPeerConfig(domainName, numOrgs, numPeer)
  check(err)

  config := Config{
    OrdererOrgs:  ordererData,
    PeerOrgs: peerData,
  }

  return config, nil
}

func GenOrdererConfig(domainName string, numOrderers int) ([]OrgSpec, error) {
  config := []OrgSpec{}
  tempconfig := OrgSpec{
    Name: "Orderer",
    Domain: domainName,
  }

  var hostname NodeSpec

  if numOrderers > 1 {
    for i := 0; i < numOrderers; i++ {
      hostname = NodeSpec{
        Hostname: "orderer" + strconv.Itoa(i),
      }
      tempconfig.Specs = append(tempconfig.Specs, hostname)
    }
  } else {
    hostname = NodeSpec{
      Hostname: "orderer0",
    }
    tempconfig.Specs = append(tempconfig.Specs, hostname)
  }
  config = append(config, tempconfig)

  return config, nil
}

func GenPeerConfig(domainName string, numOrgs int, numPeers int) ([]OrgSpec, error) {
  config := []OrgSpec{}

  for i := 0; i < numOrgs; i++ {
    nodeTemplate := NodeTemplate{
      Count:  numPeers,
    }
    users := UsersSpec{
      Count:  1,
    }
    tempconfig := OrgSpec{
      Name: "Org" + strconv.Itoa(i+1),
      Domain: "org" + strconv.Itoa(i+1) + "." + domainName,
      Template: nodeTemplate,
      Users:  users,
    }
    config = append(config, tempconfig)
  }
  return config, nil
}
