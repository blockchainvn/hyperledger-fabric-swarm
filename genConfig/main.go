package main

import (
  "flag"
  "fmt"
  "os"
  "gopkg.in/yaml.v2"
  "path/filepath"
  "io/ioutil"
)


func main() {
  var domain string
  var numOrgs, numPeer, numOrderer, numKafka int

	flag.StringVar(&domain, "domain", "example.com", "Generate config file for a particular doamin")
	flag.IntVar(&numOrgs, "numOrgs", 2, "Choose number of Organizations except Orderer's Organization. CA will be created per each organization")
	flag.IntVar(&numPeer, "numPeer", 2, "Choose number of peers per organizations")
	flag.IntVar(&numOrderer, "numOrderer", 1, "Choose number of orderers (if set, need to specify number of Kafka nodes)")
	flag.IntVar(&numKafka, "numKafka", 3, "Choose number of kafka nodes")

	flag.Parse()

	crypto, err := GenCrypto(domain, numOrgs, numPeer, numOrderer)
  fmt.Println("Generating YAML file from crypto config....")
  cryptoYAML, err := yaml.Marshal(&crypto)
  check(err)

  configtx, err := GenConfigtx(domain, numOrgs, numOrderer, numKafka)
  check(err)
  fmt.Println("Generating YAML file from configtx config....")
  configtxYAML, err := yaml.Marshal(&configtx)
  check(err)

	pwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	check(err)
  err = ioutil.WriteFile(pwd + "/crypto-config.yaml", []byte(cryptoYAML), 0644)
	check(err)
  err = ioutil.WriteFile(pwd + "/configtx.yaml", []byte(configtxYAML), 0644)
  check(err)
	fmt.Println("The output YAML file is located on " + pwd)
}
