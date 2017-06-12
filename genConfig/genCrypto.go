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
	Count    int      `yaml:"Count"`
	Start    int      `yaml:"Start"`
	Hostname string   `yaml:"Hostname"`
	SANS     []string `yaml:"SANS"`
}

type NodeSpec struct {
	Hostname   string   `yaml:"Hostname"`
	CommonName string   `yaml:"CommonName"`
	SANS       []string `yaml:"SANS"`
}

type UsersSpec struct {
	Count int `yaml:"Count"`
}

type OrgSpec struct {
	Name     string       `yaml:"Name"`
	Domain   string       `yaml:"Domain"`
	CA       NodeSpec     `yaml:"CA"`
	Template NodeTemplate `yaml:"Template"`
	Specs    []NodeSpec   `yaml:"Specs"`
	Users    UsersSpec    `yaml:"Users"`
}

type Config struct {
	OrdererOrgs []OrgSpec `yaml:"OrdererOrgs"`
	PeerOrgs    []OrgSpec `yaml:"PeerOrgs"`
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
      Hostname: "orderer",
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

func check(e error) {
    if e != nil {
        panic(e)
    }
}
