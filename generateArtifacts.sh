#!/bin/bash +x
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#


#set -e

export FABRIC_ROOT=${PWD}
export FABRIC_CFG_PATH=${PWD}
echo

# Print the usage message
function printHelp () {
  echo "Usage: "
	echo "  generateArtifacts.sh [-c <channel name>] [-d <domain name>] [-o <number of orgs]"
  echo "  generateArtifacts.sh -h|--help (print this message)"
  echo "    -c <channel name> - channel name to use (defaults to \"mychannel\")"
  echo "    -d <domain name> - domain name to use (defaults to \"example.com\")"
	echo "    -o <number of orgs> - number of organizations to use (defaults to \"2\")"
  echo
  echo "Taking all defaults:"
  echo "	generateArtifacts.sh"
}

OS_ARCH=$(echo "$(uname -s|tr '[:upper:]' '[:lower:]'|sed 's/mingw64_nt.*/windows/')-$(uname -m | sed 's/x86_64/amd64/g')" | awk '{print tolower($0)}')

## Using docker-compose template replace private key file names with constants
function replacePrivateKey () {
	ARCH=`uname -s | grep Darwin`
	if [ "$ARCH" == "Darwin" ]; then
		OPTS="-it"
	else
		OPTS="-i"
	fi

	#cp docker-compose-e2e-template.yaml docker-compose-e2e.yaml
  #cp hyperledger-swarm-template.yaml hyperledger-swarm.yaml
	i=1
	while [ "$i" -le "$NUM_ORGS" ]; do
		CURRENT_DIR=$PWD
  	cd crypto-config/peerOrganizations/org${i}.${DOMAIN_NAME}/ca/
  	PRIV_KEY=$(ls *_sk)
  	cd $CURRENT_DIR
  	#sed $OPTS "s/CA1_PRIVATE_KEY/${PRIV_KEY}/g" docker-compose-e2e.yaml
		sed $OPTS "s/CA${i}_PRIVATE_KEY/${PRIV_KEY}/g" hyperledger-ca.yaml
		i=$(($i + 1))
	done
}

## Generates Org certs using cryptogen tool
function generateCerts (){
	CRYPTOGEN=$FABRIC_ROOT/bin/cryptogen
  which $CRYPTOGEN
	if [ "$?" -ne 0 ]; then
    echo "cryptogen tool not found. exiting"
    exit 1
  fi

  rm -rf ./crypto-config

	echo
	echo "##########################################################"
	echo "##### Generate certificates using cryptogen tool #########"
	echo "##########################################################"
	$CRYPTOGEN generate --config=./crypto-config.yaml
	if [ "$?" -ne 0 ]; then
    echo "Failed to generate certificates..."
    exit 1
  fi
	echo
}

## Generate orderer genesis block , channel configuration transaction and anchor peer update transactions
function generateChannelArtifacts() {

	CONFIGTXGEN=$FABRIC_ROOT/bin/configtxgen
	which $CONFIGTXGEN
	if [ "$?" -ne 0 ]; then
    echo "configtxgen tool not found. exiting"
    exit 1
  fi

  # rm -rf ./channel-artifacts

	echo "##########################################################"
	echo "#########  Generating Orderer Genesis block ##############"
	echo "##########################################################"
	# Note: For some unknown reason (at least for now) the block file can't be
	# named orderer.genesis.block or the orderer will fail to launch!
	$CONFIGTXGEN -profile TwoOrgsOrdererGenesis -outputBlock ./channel-artifacts/genesis.block
	if [ "$?" -ne 0 ]; then
    echo "Failed to generate orderer genesis block..."
    exit 1
  fi
	echo
	echo "#################################################################"
	echo "### Generating channel configuration transaction 'channel.tx' ###"
	echo "#################################################################"
	$CONFIGTXGEN -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID $CHANNEL_NAME
	if [ "$?" -ne 0 ]; then
    echo "Failed to generate channel configuration transaction..."
    exit 1
  fi

	i=1
	while [ "$i" -le "$NUM_ORGS" ]; do
		echo
		echo "#################################################################"
		echo "#######    Generating anchor peer update for Org${i}MSP   ##########"
		echo "#################################################################"
		$CONFIGTXGEN -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org${i}MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org${i}MSP
		if [ "$?" -ne 0 ]; then
	    echo "Failed to generate anchor peer update for Org${i}MSP..."
	    exit 1
	  fi
		i=$(($i + 1))
	done
}

CHANNEL_NAME="mychannel"
DOMAIN_NAME="example.com"
NUM_ORGS=2

# Parse commandline args
while getopts "h?c:d:o:" opt; do
  case "$opt" in
    h|\?)
      printHelp
      exit 0
    ;;
    c)  CHANNEL_NAME=$OPTARG
    ;;
    d)  DOMAIN_NAME=$OPTARG
    ;;
		o)  NUM_ORGS=$OPTARG
		;;
  esac
done

generateCerts
replacePrivateKey
generateChannelArtifacts
