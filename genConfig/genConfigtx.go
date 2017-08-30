package main

import (
  "time"
  "strconv"
)

// TopLevel consists of the structs used by the configtxgen tool.
type TopLevel struct {
	Profiles      map[string]*Profile `yaml:"Profiles,omitempty"`
	Organizations []*Organization     `yaml:"Organizations,omitempty"`
	Application   *Application        `yaml:"Application,omitempty"`
	Orderer       *Orderer            `yaml:"Orderer,omitempty"`
}

// Profile encodes orderer/application configuration combinations for the configtxgen tool.
type Profile struct {
	Consortium  string                 `yaml:"Consortium,omitempty"`
	Application *Application           `yaml:"Application,omitempty"`
	Orderer     *Orderer               `yaml:"Orderer,omitempty"`
	Consortiums map[string]*Consortium `yaml:"Consortiums,omitempty"`
}

// Consortium represents a group of organizations which may create channels with eachother
type Consortium struct {
	Organizations []*Organization      `yaml:"Organizations,omitempty"`
}

// Application encodes the application-level configuration needed in config transactions.
type Application struct {
	Organizations []*Organization       `yaml:"Organizations,omitempty"`
}

// Organization encodes the organization-level configuration needed in config transactions.
type Organization struct {
	Name           string               `yaml:"Name,omitempty"`
	ID             string               `yaml:"ID,omitempty"`
	MSPDir         string               `yaml:"MSPDir,omitempty"`
	AdminPrincipal string               `yaml:"AdminPrincipal,omitempty"`

	// Note: Viper deserialization does not seem to care for
	// embedding of types, so we use one organization struct
	// for both orderers and applications.
	AnchorPeers []*AnchorPeer           `yaml:"AnchorPeers,omitempty"`
}

// AnchorPeer encodes the necessary fields to identify an anchor peer.
type AnchorPeer struct {
	Host string                         `yaml:"Host,omitempty"`
	Port int                            `yaml:"Port,omitempty"`
}

// ApplicationOrganization ...
// TODO This should probably be removed
type ApplicationOrganization struct {
	Organization `yaml:"Organization,omitempty"`
}

// Orderer contains configuration which is used for the
// bootstrapping of an orderer by the provisional bootstrapper.
type Orderer struct {
	OrdererType   string          `yaml:"OrdererType,omitempty"`
	Addresses     []string        `yaml:"Addresses,omitempty"`
	BatchTimeout  time.Duration   `yaml:"BatchTimeout,omitempty"`
	BatchSize     BatchSize       `yaml:"BatchSize,omitempty"`
	Kafka         Kafka           `yaml:"Kafka,omitempty"`
	Organizations []*Organization `yaml:"Organizations,omitempty"`
	MaxChannels   uint64          `yaml:"MaxChannels,omitempty"`
}

// BatchSize contains configuration affecting the size of batches.
type BatchSize struct {
	MaxMessageCount   uint32 `yaml:"MaxMessageCount,omitempty"`
	AbsoluteMaxBytes  uint32 `yaml:"AbsoluteMaxBytes,omitempty"`
	PreferredMaxBytes uint32 `yaml:"PreferredMaxBytes,omitempty"`
}

// Kafka contains configuration for the Kafka-based orderer.
type Kafka struct {
	Brokers []string `yaml:"Brokers"`
}

func GenConfigtx(domainName string, numOrgs int, numOrderer int, numKafka int) (TopLevel, error){

  var kafka Kafka
  kafka, _ = GenKafka(numKafka, domainName)

  var orderer Orderer
  orderer, _ = GenOrderer(numOrderer, domainName, kafka)

  var org []*Organization
  for i := 1; i <= numOrgs; i++ {
    temporg, _ := GenOrg(i, domainName)
    org = append(org, &temporg)
  }

  conList := make(map[string]*Consortium,1)
  conList["SampleConsortium"] = &Consortium{
    Organizations:  org,
  }

  profGenesis := Profile{
    Orderer:    &orderer,
    Consortiums: conList,
  }

  profChannel := Profile{
    Consortium:   "SampleConsortium",
    Application:  &Application{
      Organizations:  org,
    },
  }

  topProfile := make(map[string]*Profile,2)
  topProfile["TwoOrgsOrdererGenesis"] = &profGenesis
  topProfile["TwoOrgsChannel"] = &profChannel

  topOrg := make([]*Organization,numOrgs + 1)
  topOrg = append([]*Organization{ GenOrdererOrg(domainName) }, org...)

  topOrderer := &orderer

  topLevel := TopLevel{
    Profiles:       topProfile,
    Organizations:  topOrg,
    Orderer:        topOrderer,
  }

  return topLevel, nil
}

func GenOrg(orgId int, domainName string) (Organization, error) {
  orgIdStr := strconv.Itoa(orgId)
  anchor := AnchorPeer{
    Host:   "peer0.org" + orgIdStr + "." + domainName,
    Port:   7051,
  }

  org := Organization{
    Name:   "Org" + orgIdStr + "MSP",
    ID:     "Org" + orgIdStr + "MSP",
    MSPDir: "crypto-config/peerOrganizations/org" + orgIdStr + "." + domainName + "/msp",
    AnchorPeers:  []*AnchorPeer{&anchor},
  }

  return org, nil
}

func GenOrdererOrg(domainName string) (*Organization){
  ordererOrg := Organization{
    Name:   "OrdererOrg",
    ID:     "OrdererMSP",
    MSPDir: "crypto-config/ordererOrganizations/" + domainName + "/msp",
  }
  //orderer.Organizations[0] = &ordererOrg

  return &ordererOrg
}


func GenOrderer(numOrderer int, domainName string, kafka Kafka) (Orderer, error) {
  var address_list []string


  var orderer Orderer
  if numOrderer == 1 {
    address_list = append(address_list, "orderer0." + domainName + ":7050")
    orderer = Orderer{
      OrdererType:  "solo",
      Addresses:    address_list,
      BatchTimeout: 2 * time.Second,
      BatchSize:    BatchSize{
        MaxMessageCount:  10,
        AbsoluteMaxBytes: 99 * 1024 * 1024, // 99 MB
        PreferredMaxBytes:  512 * 1024, // 512 KB
      },
      Organizations:  make([]*Organization,1),
    }
  } else {
    for i := 0; i < numOrderer; i++ {
      address_list = append(address_list, "orderer" + strconv.Itoa(i) + "." + domainName + ":7050")
    }
    orderer = Orderer{
      OrdererType:  "kafka",
      Addresses:    address_list,
      BatchTimeout: 2 * time.Second,
      BatchSize:    BatchSize{
        MaxMessageCount:  10,
        AbsoluteMaxBytes: 99 * 1024 * 1024, // 99 MB
        PreferredMaxBytes:  512 * 1024, // 512 KB
      },
      Kafka:        kafka,
      Organizations:  make([]*Organization,1),
    }
  }

  orderer.Organizations[0] = GenOrdererOrg(domainName)
  return orderer, nil
}

func GenKafka(numKafka int, domainName string) (Kafka, error) {
  var kafka_list []string
  for i := 0; i < numKafka; i++ {
    kafka_list = append(kafka_list, "kafka" + strconv.Itoa(i) + "." + domainName + ":9092")
  }

  var kafka = Kafka{
    Brokers: kafka_list,
  }

  return kafka, nil
}
