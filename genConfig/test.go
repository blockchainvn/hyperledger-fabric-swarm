package main

import (
	"fmt"
  "time"
)
type TopLevel struct {
  Version       string               `yaml:"version"`
  Networks      map[string]*Network  `yaml:"networks"`
  Services      map[string]*Service   `yaml:"services"`
}

type Network struct {
  External      *External             `yaml:"external"`
}

type External struct {
  Name          string                `yaml:"name"`
}

type Service struct {
  Deploy        *Deploy               `yaml:"deploy"`
  Hostname      string                `yaml:"hostname"`
  Image         string                `yaml:"image"`
  Networks      []string              `yaml:"networks"`
  Environment   []string              `yaml:"environment"`
}

type Deploy struct {
  Replicas      int                   `yaml:"replicas"`
  Placement     *Placement            `yaml:"placement"`
  RestartPolicy *RestartPolicy        `yaml:"restart_policy"`
}

type Placement struct {
  Constraint    []string              `yaml:"constraints"`
}

type RestartPolicy struct {
  Condition     string                `yaml:"condition"`
  Delay         time.Duration         `yaml:"delay"`
  MaxAttempts   int                   `yaml:"max_attempts"`
  Window        time.Duration         `yaml:"window"`
}

var net = `
version: '3'
#
networks:
  hyperledger-ov:
    # If network is created with deplyment, Chaincode container cannot connect to network
    external:
      name: hyperledger-ov

services:
  zookeeper0:
    deploy:
      replicas: 1
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
    hostname: zookeeper0.example.com
    image: hyperledger/fabric-zookeeper:x86_64-1.0.0-beta
    # Give network alias
    networks:
      - hyperledger-ov
    environment:
      - CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=hyperledger-ov
      - ZOO_MY_ID=1
      - ZOO_SERVERS=server.1=zookeeper0:2888:3888 server.2=zookeeper1:2888:3888 server.3=zookeeper2:2888:3888
`


func main() {
  t := TopLevel{
  }

  networks := make(map[string]*Network, 1)
  networks["hyp-ov"], _ = GenNetwork("hyp-ov")
  //err := yaml.Unmarshal([]byte(net), &t)

  t.Networks = networks

  
  fmt.Printf("--- t:\n%#v\n\n", t)

}

func GenNetwork(networkName string) (*Network, error){
  network := Network{
    External:  &External{
      Name:   networkName,
    },
  }

  return &network, nil
}
