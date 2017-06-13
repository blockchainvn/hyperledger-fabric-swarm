package main

import (
	"fmt"
  "time"
	"log"
	"gopkg.in/yaml.v2"

	"strconv"
	"strings"
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
  Networks      map[string]*ServNet   `yaml:"networks"`
  Environment   []string              `yaml:"environment"`
	WorkingDir 		string 								`yaml:"working_dir"`
	Command 			string 								`yaml:"command"`
	Volumes 			[]string 							`yaml:"volumes"`
}

type ServNet struct {
	Aliases 			[]string 							`yaml:"aliases"`
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
var TAG = `:x86_64-1.0.0-beta`

func main() {
	////// Create Zookeepr /////////////////
  t := &TopLevel{}

  //networks := make(map[string]*Network, 1)
  //networks["hyp-ov"], _ = GenNetwork("hyp-ov")
  //err := yaml.Unmarshal([]byte(net), &t)git@github.com:ChoiSD/hyperledger_on_swarm.git

  //t.Networks = networks
  t.Version = "3"

	err := GenNetwork(t, "hyp-ov")
	check(err)

  err = GenService(t, "sdchoi.com", "zookeeper", "hyp-ov", 3)
	check(err)
	//////////////////////////////////////////////////////////

	////////// Create Kafka ////////////////////////////
	tk := &TopLevel{}
	tk.Version = "3"
	err = GenNetwork(tk, "hyp-ov")
	check(err)

	err = GenService(tk, "sdchoi.com", "kafka", "hyp-ov", 3)
	check(err)
	/////////////////////////////////////////////////////////


	///////// Create Orderer ////////////////////////////////
	to := &TopLevel{}
	to.Version = "3"
	err = GenNetwork(to, "hyp-ov")
	check(err)

	err = GenService(to, "sdchoi.com", "orderer", "hyp-ov", 3)
	check(err)
	/////////////////////////////////////////////////////////

	////////////// Create CA ////////////////////////////////
	tc := &TopLevel{}
	tc.Version = "3"
	err = GenNetwork(tc, "hyp-ov")
	check(err)

	err = GenService(tc, "sdchoi.com", "ca", "hyp-ov", 3)
	check(err)
	err = GenService(tc, "sdchoi.com", "couchdb", "hyp-ov", 4)
	check(err)
	err = GenService(tc, "sdchoi.com", "peer", "hyp-ov", 2, 3)
	check(err)
	///////////////////////////////////////////////////////////
  fmt.Printf("--- t:\n%#v\n\n", *t)
  d, err := yaml.Marshal(tc)
  check(err)
	fmt.Printf("%v\n", string(d))
}

func GenDeploy(service *Service) (error) {
	deploy := &Deploy{
		Replicas:		1,
		RestartPolicy:	&RestartPolicy{
			Condition:		"on-failure",
			Delay:				5 * time.Second,
			MaxAttempts:	3,
		},
	}
	service.Deploy = deploy

	return nil
}

func GenService(topLevel *TopLevel, domainName string, serviceName string, networkName string, num ...int) (error) {
	var total int
	if len(num) > 1 {
		total = num[0] * num[1]
	} else {
		total = num[0]
	}

	topLevel.Services = make(map[string]*Service, total)

	for i := 0; i < total; i++ {
		err := GenDeploy(service)
		check(err)
		
		switch serviceName {
		case "zookeeper":
			serviceHost := serviceName + strconv.Itoa(i)
			service := &Service{
				Hostname:	serviceHost + "." + domainName,
			}
			service.Networks = make(map[string]*ServNet, 1)
			service.Networks[networkName] = &ServNet{
				Aliases: []string{serviceHost + "." + domainName},
			}
			service.Image = "hyperledger/fabric-zookeeper" + TAG
			var zookeeperArray []string
			for j := 0; j < total; j++ {
				zookeeperArray = append(zookeeperArray, "server." + strconv.Itoa(j + 1) + "=zookeeper" + strconv.Itoa(j) + ":2888:3888")
			}
			zookeeperList := arrayToString(zookeeperArray, " ")
			service.Environment = make([]string, 3)
			service.Environment[0] = "CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=" + networkName
			service.Environment[1] = "ZOO_MY_ID=" + strconv.Itoa(i + 1)
			service.Environment[2] = "ZOO_SERVERS=" + zookeeperList

		case "kafka":
			serviceHost := serviceName + strconv.Itoa(i)
			service := &Service{
				Hostname:	serviceHost + "." + domainName,
			}
			service.Networks = make(map[string]*ServNet, 1)
			service.Networks[networkName] = &ServNet{
				Aliases: []string{serviceHost + "." + domainName},
			}
			service.Image = "hyperledger/fabric-kafka" + TAG
			var zookeeperArray []string
			for j := 0; j < 3; j++ { // 3 is number of zookeeper nodes
				zookeeperArray = append(zookeeperArray, "zookeeper" + strconv.Itoa(j) + ":2181")
			}
			zookeeperString := arrayToString(zookeeperArray, ",")
			service.Environment = make([]string, 8)
			service.Environment[0] = "CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=" + networkName
			service.Environment[1] = "KAFKA_MESSAGE_MAX_BYTES=103809024" // 99 MB
			service.Environment[2] = "KAFAK_REPLICA_FETCH_MAX_BYTES=103809024" // 99 MB
			service.Environment[3] = "KAFKA_UNCLEAN_LEADER_ELECTION_ENABLE=false"
			service.Environment[4] = "KAFKA_DEFAULT_REPLICATION_FACTOR=3"
			service.Environment[5] = "KAFKA.MIN_INSYNC_REPLICAS=2"
			service.Environment[6] = "KAFKA_ZOOKEEPER_CONNECT=" + zookeeperString
			service.Environment[7] = "KAFKA_BROKER_ID=" +	strconv.Itoa(i)

		case "orderer":
			serviceHost := serviceName + strconv.Itoa(i)
			service := &Service{
				Hostname:	serviceHost + "." + domainName,
			}
			service.Networks = make(map[string]*ServNet, 1)
			service.Networks[networkName] = &ServNet{
				Aliases: []string{serviceHost + "." + domainName},
			}
			service.Image = "hyperledger/fabric-orderer" + TAG
			service.Environment = make([]string, 14)
			service.Environment[0]  = "CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=" + networkName
			service.Environment[1]  = "ORDERER_GENERAL_LOGLEVEL=debug"
			service.Environment[2]  = "ORDERER_GENERAL_LISTENADDRESS=0.0.0.0"
			service.Environment[3]  = "ORDERER_GENERAL_GENESISMETHOD=file"
			service.Environment[4]  = "ORDERER_GENERAL_GENESISFILE=/var/hyperledger/orderer/orderer.genesis.block"
			service.Environment[5]  = "ORDERER_GENERAL_LOCALMSPID=OrdererMSP"
			service.Environment[6]  = "ORDERER_GENERAL_LOCALMSPDIR=/var/hyperledger/orderer/msp"
			service.Environment[7]  = "ORDERER_GENERAL_TLS_ENABLED=true"
			service.Environment[8]  = "ORDERER_GENERAL_TLS_PRIVATEKEY=/var/hyperledger/orderer/tls/server.key"
			service.Environment[9]  = "ORDERER_GENERAL_TLS_CERTIFICATE=/var/hyperledger/orderer/tls/server.crt"
			service.Environment[10] = "ORDERER_GENERAL_TLS_ROOTCAS=[/var/hyperledger/orderer/tls/ca.crt]"
			service.Environment[11] = "ORDERER_KAFKA_RETRY_SHORTINTERVAL=1s"
			service.Environment[12] = "ORDERER_KAFAK_RETRY_SHORTTOTAL=30s"
			service.Environment[13] = "ORDERER_KAFKA_VERBOSE=true"

			service.WorkingDir = "/opt/gopath/src/github.com/hyperledger/fabric"
			service.Command = "orderer"

			service.Volumes = make([]string, 3)
			service.Volumes[0] = "./channel-artifacts/genesis.block:/var/hyperledger/orderer/orderer.genesis.block"
			service.Volumes[1] = "./crypto-config/ordererOrganizations/" + domainName + "/orderers/" + serviceHost + "." + domainName + "/msp:/var/hyperledger/orderer/msp"
			service.Volumes[2] = "./crypto-config/ordererOrganizations/" + domainName + "/orderers/" + serviceHost + "." + domainName + "/tls/:/var/hyperledger/orderer/tls"

		case "ca":
			serviceHost := serviceName + strconv.Itoa(i)
			service := &Service{
				Hostname:	serviceHost + "." + domainName,
			}
			orgId := strconv.Itoa(i + 1)
			service.Networks = make(map[string]*ServNet, 1)
			service.Networks[networkName] = &ServNet{
				Aliases: []string{serviceName + "_peerOrg" + orgId},
			}
			service.Image = "hyperledger/fabric-ca" + TAG
			service.Environment = make([]string, 5)
			service.Environment[0] = "FABRIC_CA_HOME=/etc/hyperledger/fabric-ca-server"
      service.Environment[1] = "FABRIC_CA_SERVER_CA_NAME=ca-org2"
      service.Environment[2] = "FABRIC_CA_SERVER_TLS_ENABLED=true"
      service.Environment[3] = "FABRIC_CA_SERVER_TLS_CERTFILE=/etc/hyperledger/fabric-ca-server-config/ca.org" + orgId + "." + domainName + "-cert.pem"
      service.Environment[4] = "FABRIC_CA_SERVER_TLS_KEYFILE=/etc/hyperledger/fabric-ca-server-config/CA2_PRIVATE_KEY"

			service.Volumes = make([]string, 1)
			service.Volumes[0] = "./crypto-config/peerOrganizations/org" + orgId + "." + domainName + "/ca/:/etc/hyperledger/fabric-ca-server-config"

		case "couchdb":
			serviceHost := serviceName + strconv.Itoa(i)
			service := &Service{
				Hostname:	serviceHost + "." + domainName,
			}
			service.Image = "hyperledger/fabric-couchdb" + TAG
			service.Networks = make(map[string]*ServNet, 1)
			service.Networks[networkName] = &ServNet{}

		case "peer":
			//numPeer := num[0]
			//numOrgs := num[1]
			peerNum := strconv.Itoa(i % num[0])
			orgNum := strconv.Itoa((i / num[0]) + 1)
			hostName := "peer" + peerNum + ".org" + orgNum + "." + domainName
			service := &Service{
				Hostname:	hostName,
			}
			service.Networks = make(map[string]*ServNet, 1)
			service.Networks[networkName] = &ServNet{
				Aliases: []string{hostName},
			}
			service.Environment = make([]string, 19)
			service.Environment[0]  = "CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock"
      service.Environment[1]  = "CORE_LOGGING_LEVEL=DEBUG"
      service.Environment[2]  = "CORE_PEER_TLS_ENABLED=true"
      service.Environment[3]  = "CORE_PEER_ENDORSER_ENABLED=true"
      service.Environment[4]  = "CORE_PEER_GOSSIP_USELEADERELECTION=true"
      service.Environment[5]  = "CORE_PEER_GOSSIP_ORGLEADER=false"
      service.Environment[6]  = "CORE_PEER_GOSSIP_SKIPHANDSHAKE=true"
      service.Environment[7]  = "CORE_PEER_PROFILE_ENABLED=true"
      service.Environment[8]  = "CORE_PEER_TLS_CERT_FILE=/etc/hyperledger/fabric/tls/server.crt"
      service.Environment[9]  = "CORE_PEER_TLS_KEY_FILE=/etc/hyperledger/fabric/tls/server.key"
      service.Environment[10] = "CORE_PEER_TLS_ROOTCERT_FILE=/etc/hyperledger/fabric/tls/ca.crt"
      service.Environment[11] = "CORE_PEER_ID=" + hostName
      service.Environment[12] = "CORE_PEER_ADDRESS=" + hostName + ":7051"
      service.Environment[13] = "CORE_PEER_GOSSIP_EXTERNALENDPOINT=" + hostName + ":7051"
      service.Environment[14] = "CORE_PEER_LOCALMSPID=Org" + orgNum +"MSP"
      service.Environment[15] = "CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=" + networkName
      service.Environment[16] = "CORE_LEDGER_STATE_STATEDATABASE=CouchDB"
      service.Environment[17] = "CORE_LEDGER_STATE_COUCHDBCONFIG_COUCHDBADDRESS=couchdb"+ strconv.Itoa(i) + ":5984"
			service.Environment[18] = "CORE_PEER_GOSSIP_BOOTSTRAP=peer0.org" + orgNum + "." + domainName + ":7051"
			service.WorkingDir = "/opt/gopath/src/github.com/hyperledger/fabric/peer"
			service.Command = "peer node start"
			service.Volumes = make([]string, 3)
			service.Volumes[0] = "/var/run/:/host/var/run/"
			service.Volumes[1] = "./crypto-config/peerOrganizations/org" + orgNum + "." + domainName + "/peers/" + hostName + "/msp:/etc/hyperledger/fabric/msp"
			service.Volumes[2] = "./crypto-config/peerOrganizations/org" + orgNum + "." + domainName + "/peers/" + hostName + "/tls:/etc/hyperledger/fabric/tls"
		}
		topLevel.Services[serviceHost] = service
	}
	return nil
}

func GenNetwork(topLevel *TopLevel, networkName string) (error){
	network := &Network{
    External:  &External{
      Name:   networkName,
    },
  }

	topLevel.Networks = make(map[string]*Network, 1)
	topLevel.Networks[networkName] = network

  return nil
}

func arrayToString(array []string, delim string) (string) {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(array)), delim), "[]")
}

func check(err error) {
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
}
