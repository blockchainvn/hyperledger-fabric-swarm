package main

import (
	"fmt"
	"log"
	"time"

	"strconv"
	"strings"
)

type DockerCompose struct {
	Version  string              `yaml:"version,omitempty"`
	Networks map[string]*Network `yaml:"networks,omitempty"`
	Services map[string]*Service `yaml:"services,omitempty"`
}

type Network struct {
	External *External `yaml:"external,omitempty"`
}

type External struct {
	Name string `yaml:"name,omitempty"`
}

type Service struct {
	Deploy      *Deploy             `yaml:"deploy,omitempty"`
	Hostname    string              `yaml:"hostname,omitempty"`
	Image       string              `yaml:"image,omitempty"`
	Networks    map[string]*ServNet `yaml:"networks,omitempty"`
	Environment []string            `yaml:"environment,omitempty"`
	WorkingDir  string              `yaml:"working_dir,omitempty"`
	Command     string              `yaml:"command,omitempty"`
	Volumes     []string            `yaml:"volumes,omitempty"`
}

type ServNet struct {
	Aliases []string `yaml:"aliases,omitempty"`
}

// Placement will be added future
type Deploy struct {
	Replicas      int            `yaml:"replicas,omitempty"`
	Placement     *Placement     `yaml:"placement,omitempty"`
	RestartPolicy *RestartPolicy `yaml:"restart_policy,omitempty"`
}

type Placement struct {
	Constraint []string `yaml:"constraints,omitempty"`
}

type RestartPolicy struct {
	Condition   string        `yaml:"condition,omitempty"`
	Delay       time.Duration `yaml:"delay,omitempty"`
	MaxAttempts int           `yaml:"max_attempts,omitempty"`
	Window      time.Duration `yaml:"window,omitempty"`
}

var TAG = `:x86_64-1.0.2`

// var TAG = `:latest`

func GenDockerCompose(serviceName string, domainName string, networkName string, num ...int) (*DockerCompose, error) {
	var dockerCompose = &DockerCompose{}
	dockerCompose.Version = "3"

	err := GenNetwork(dockerCompose, networkName)
	check(err)

	switch serviceName {
	case "peer", "couchdb":
		err = GenService(dockerCompose, domainName, serviceName, networkName, num[0], num[1])
	default:
		err = GenService(dockerCompose, domainName, serviceName, networkName, num[0])
	}

	return dockerCompose, nil
}

func GenDeploy(service *Service) error {
	deploy := &Deploy{
		Replicas: 1,
		RestartPolicy: &RestartPolicy{
			Condition:   "on-failure",
			Delay:       5 * time.Second,
			MaxAttempts: 3,
		},
	}
	service.Deploy = deploy

	return nil
}

func GenService(dockerCompose *DockerCompose, domainName string, serviceName string, networkName string, num ...int) error {
	var total int
	if len(num) > 1 {
		total = num[0] * num[1]
	} else {
		total = num[0]
	}

	dockerCompose.Services = make(map[string]*Service, total)

	for i := 0; i < total; i++ {
		var serviceHost string
		var service *Service

		switch serviceName {
		case "zookeeper":
			serviceHost = "zookeeper" + strconv.Itoa(i)
			service = &Service{
				Hostname: serviceHost,
			}
			service.Networks = make(map[string]*ServNet, 1)
			service.Networks[networkName] = &ServNet{
				Aliases: []string{serviceHost + "." + domainName},
			}
			service.Image = "hyperledger/fabric-zookeeper" + TAG
			var zookeeperArray []string
			for j := 0; j < total; j++ {
				zookeeperArray = append(zookeeperArray, "server."+strconv.Itoa(j+1)+"=zookeeper"+strconv.Itoa(j)+":2888:3888")
			}
			zookeeperList := arrayToString(zookeeperArray, " ")
			service.Environment = make([]string, 3)
			service.Environment[0] = "CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=" + networkName
			service.Environment[1] = "ZOO_MY_ID=" + strconv.Itoa(i+1)
			service.Environment[2] = "ZOO_SERVERS=" + zookeeperList
			err := GenDeploy(service)
			check(err)

		case "kafka":
			serviceHost = "kafka" + strconv.Itoa(i)
			service = &Service{
				Hostname: serviceHost + "." + domainName,
			}
			service.Networks = make(map[string]*ServNet, 1)
			service.Networks[networkName] = &ServNet{
				Aliases: []string{serviceHost + "." + domainName},
			}
			service.Image = "hyperledger/fabric-kafka" + TAG
			var zookeeperArray []string
			for j := 0; j < 3; j++ { // 3 is number of zookeeper nodes
				zookeeperArray = append(zookeeperArray, "zookeeper"+strconv.Itoa(j)+":2181")
			}
			zookeeperString := arrayToString(zookeeperArray, ",")
			service.Environment = make([]string, 8)
			service.Environment[0] = "CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=" + networkName
			service.Environment[1] = "KAFKA_MESSAGE_MAX_BYTES=103809024"       // 99 MB
			service.Environment[2] = "KAFKA_REPLICA_FETCH_MAX_BYTES=103809024" // 99 MB
			service.Environment[3] = "KAFKA_UNCLEAN_LEADER_ELECTION_ENABLE=false"
			service.Environment[4] = "KAFKA_DEFAULT_REPLICATION_FACTOR=3"
			service.Environment[5] = "KAFKA_MIN_INSYNC_REPLICAS=2"
			service.Environment[6] = "KAFKA_ZOOKEEPER_CONNECT=" + zookeeperString
			service.Environment[7] = "KAFKA_BROKER_ID=" + strconv.Itoa(i)
			err := GenDeploy(service)
			check(err)

		case "orderer":
			serviceHost = "orderer" + strconv.Itoa(i)
			service = &Service{
				Hostname: serviceHost + "." + domainName,
			}
			service.Networks = make(map[string]*ServNet, 1)
			service.Networks[networkName] = &ServNet{
				Aliases: []string{serviceHost + "." + domainName},
			}
			service.Image = "hyperledger/fabric-orderer" + TAG
			service.Environment = make([]string, 14)
			service.Environment[0] = "CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=" + networkName
			service.Environment[1] = "ORDERER_GENERAL_LOGLEVEL=debug"
			service.Environment[2] = "ORDERER_GENERAL_LISTENADDRESS=0.0.0.0"
			service.Environment[3] = "ORDERER_GENERAL_GENESISMETHOD=file"
			service.Environment[4] = "ORDERER_GENERAL_GENESISFILE=/var/hyperledger/orderer/orderer.genesis.block"
			service.Environment[5] = "ORDERER_GENERAL_LOCALMSPID=OrdererMSP"
			service.Environment[6] = "ORDERER_GENERAL_LOCALMSPDIR=/var/hyperledger/orderer/msp"
			service.Environment[7] = "ORDERER_GENERAL_TLS_ENABLED=true"
			service.Environment[8] = "ORDERER_GENERAL_TLS_PRIVATEKEY=/var/hyperledger/orderer/tls/server.key"
			service.Environment[9] = "ORDERER_GENERAL_TLS_CERTIFICATE=/var/hyperledger/orderer/tls/server.crt"
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
			err := GenDeploy(service)
			check(err)

		case "ca":
			serviceHost = "ca" + strconv.Itoa(i)
			service = &Service{
				Hostname: serviceHost + "." + domainName,
			}
			orgId := strconv.Itoa(i + 1)
			service.Networks = make(map[string]*ServNet, 1)
			service.Networks[networkName] = &ServNet{
				Aliases: []string{serviceName + "_peerOrg" + orgId},
			}
			service.Image = "hyperledger/fabric-ca" + TAG
			service.Environment = make([]string, 5)
			service.Environment[0] = "FABRIC_CA_HOME=/etc/hyperledger/fabric-ca-server"
			service.Environment[1] = "FABRIC_CA_SERVER_CA_NAME=ca-org" + orgId
			service.Environment[2] = "FABRIC_CA_SERVER_TLS_ENABLED=true"
			service.Environment[3] = "FABRIC_CA_SERVER_TLS_CERTFILE=/etc/hyperledger/fabric-ca-server-config/ca.org" + orgId + "." + domainName + "-cert.pem"
			service.Environment[4] = "FABRIC_CA_SERVER_TLS_KEYFILE=/etc/hyperledger/fabric-ca-server-config/CA" + orgId + "_PRIVATE_KEY"
			service.Command = "sh -c 'fabric-ca-server start --ca.certfile /etc/hyperledger/fabric-ca-server-config/ca.org" + orgId + "." + domainName + "-cert.pem --ca.keyfile /etc/hyperledger/fabric-ca-server-config/CA" + orgId + "_PRIVATE_KEY -b admin:adminpw -d'"
			service.Volumes = make([]string, 1)
			service.Volumes[0] = "./crypto-config/peerOrganizations/org" + orgId + "." + domainName + "/ca/:/etc/hyperledger/fabric-ca-server-config"
			err := GenDeploy(service)
			check(err)

		case "couchdb":
			serviceHost = serviceName + strconv.Itoa(i)
			service = &Service{
				Hostname: serviceHost + "." + domainName,
			}
			service.Image = "hyperledger/fabric-couchdb" + TAG
			service.Networks = make(map[string]*ServNet, 1)
			service.Networks[networkName] = &ServNet{
				Aliases: []string{serviceHost},
			}
			err := GenDeploy(service)
			check(err)

		case "peer":
			peerNum := strconv.Itoa(i % num[0])
			orgNum := strconv.Itoa((i / num[0]) + 1)
			serviceHost = "peer" + peerNum + "_org" + orgNum
			hostName := "peer" + peerNum + ".org" + orgNum + "." + domainName
			service = &Service{
				Hostname: hostName,
			}
			service.Image = "hyperledger/fabric-peer" + TAG
			service.Networks = make(map[string]*ServNet, 1)
			service.Networks[networkName] = &ServNet{
				Aliases: []string{hostName},
			}
			service.Environment = make([]string, 17)
			service.Environment[0] = "CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock"
			service.Environment[1] = "CORE_LOGGING_LEVEL=DEBUG"
			service.Environment[2] = "CORE_PEER_TLS_ENABLED=true"
			service.Environment[3] = "CORE_PEER_GOSSIP_USELEADERELECTION=true"
			service.Environment[4] = "CORE_PEER_GOSSIP_ORGLEADER=false"
			service.Environment[5] = "CORE_PEER_PROFILE_ENABLED=true"
			service.Environment[6] = "CORE_PEER_TLS_CERT_FILE=/etc/hyperledger/fabric/tls/server.crt"
			service.Environment[7] = "CORE_PEER_TLS_KEY_FILE=/etc/hyperledger/fabric/tls/server.key"
			service.Environment[8] = "CORE_PEER_TLS_ROOTCERT_FILE=/etc/hyperledger/fabric/tls/ca.crt"
			service.Environment[9] = "CORE_PEER_ID=" + hostName
			service.Environment[10] = "CORE_PEER_ADDRESS=" + hostName + ":7051"
			service.Environment[11] = "CORE_PEER_GOSSIP_EXTERNALENDPOINT=" + hostName + ":7051"
			service.Environment[12] = "CORE_PEER_LOCALMSPID=Org" + orgNum + "MSP"
			service.Environment[13] = "CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=" + networkName
			service.Environment[14] = "CORE_LEDGER_STATE_STATEDATABASE=CouchDB"
			service.Environment[15] = "CORE_LEDGER_STATE_COUCHDBCONFIG_COUCHDBADDRESS=couchdb" + strconv.Itoa(i) + ":5984"
			service.Environment[16] = "CORE_PEER_GOSSIP_BOOTSTRAP=peer0.org" + orgNum + "." + domainName + ":7051"
			//service.Environment[3]  = "CORE_PEER_ENDORSER_ENABLED=true"
			//service.Environment[6]  = "CORE_PEER_GOSSIP_SKIPHANDSHAKE=true"
			service.WorkingDir = "/opt/gopath/src/github.com/hyperledger/fabric/peer"
			service.Command = "peer node start"
			service.Volumes = make([]string, 3)
			service.Volumes[0] = "/var/run/:/host/var/run/"
			service.Volumes[1] = "./crypto-config/peerOrganizations/org" + orgNum + "." + domainName + "/peers/" + hostName + "/msp:/etc/hyperledger/fabric/msp"
			service.Volumes[2] = "./crypto-config/peerOrganizations/org" + orgNum + "." + domainName + "/peers/" + hostName + "/tls:/etc/hyperledger/fabric/tls"
			err := GenDeploy(service)
			check(err)

		case "cli":
			serviceHost = "cli"
			service = &Service{}
			service.Image = "hyperledger/fabric-tools" + TAG
			service.Networks = make(map[string]*ServNet, 1)
			service.Networks[networkName] = &ServNet{
				Aliases: []string{"cli"},
			}
			service.Environment = make([]string, 12)
			service.Environment[0] = "CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=" + networkName
			service.Environment[1] = "GOPATH=/opt/gopath"
			service.Environment[2] = "CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock"
			service.Environment[3] = "CORE_LOGGING_LEVEL=DEBUG"
			service.Environment[4] = "CORE_PEER_ID=cli"
			service.Environment[5] = "CORE_PEER_ADDRESS=peer0.org1." + domainName + ":7051"
			service.Environment[6] = "CORE_PEER_LOCALMSPID=Org1MSP"
			service.Environment[7] = "CORE_PEER_TLS_ENABLED=true"
			service.Environment[8] = "CORE_PEER_TLS_CERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1." + domainName + "/peers/peer0.org1." + domainName + "/tls/server.crt"
			service.Environment[9] = "CORE_PEER_TLS_KEY_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1." + domainName + "/peers/peer0.org1." + domainName + "/tls/server.key"
			service.Environment[10] = "CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1." + domainName + "/peers/peer0.org1." + domainName + "/tls/ca.crt"
			service.Environment[11] = "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1." + domainName + "/users/Admin@org1." + domainName + "/msp"
			service.WorkingDir = "/opt/gopath/src/github.com/hyperledger/fabric/peer"
			service.Command = "sleep 3600"
			service.Volumes = make([]string, 5)
			service.Volumes[0] = "/var/run/:/host/var/run/"
			service.Volumes[1] = "./chaincode/:/opt/gopath/src/github.com/hyperledger/fabric/examples/chaincode/go"
			service.Volumes[2] = "./crypto-config:/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/"
			service.Volumes[3] = "./scripts:/opt/gopath/src/github.com/hyperledger/fabric/peer/scripts/"
			service.Volumes[4] = "./channel-artifacts:/opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts"
			err := GenDeploy(service)
			check(err)

		default:
			log.Fatalf("You didn't specify service name!!..\n")
		}
		dockerCompose.Services[serviceHost] = service
	}
	return nil
}

func GenNetwork(dockerCompose *DockerCompose, networkName string) error {
	network := &Network{
		External: &External{
			Name: networkName,
		},
	}

	dockerCompose.Networks = make(map[string]*Network, 1)
	dockerCompose.Networks[networkName] = network

	return nil
}

func arrayToString(array []string, delim string) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(array)), delim), "[]")
}
