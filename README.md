# hyperledger_on_swarm

This repository is for deploying Hyperledger Fabric on Swarm cluster easily.
- Supported version : 1.0.0 GA and above (1.0.1 currently)
* You can specify number of each component. Supported components are:
  - Number of Organizations (CA will be one per a Organization)
  - Number of Peer per a Organization
  - Number of Orderer
  - Number of Kafka
  - Number of Zookeeper
  - Domain name
  
## Instructions

### Pre-reqs
- 1 or more machines with Linux
- Install Docker >= 1.13

### [Create Docker Swarm cluster](https://docs.docker.com/engine/swarm/swarm-tutorial/)
* Create one or more master hosts and other worker hosts
  - Open ports for Swarm. On ALL hosts, (eg, CentOS)
  ```
  firewall-cmd --permanent --zone=public --add-port=2377/tcp --add-port=7946/tcp --add-port=7946/udp --add-port=4789/udp
  firewall-cmd --reload
  ```
    - I think opening swarm ports only is sufficient because all nodes communicates thru overlay network.

  - on master host,
  ```
  docker swarm init
  ```

  - You can see like below,
  ```
  Swarm initialized: current node (dxn1zf6l61qsb1josjja83ngz) is now a manager.

  To add a worker to this swarm, run the following command:

    docker swarm join \
    --token SWMTKN-1-49nj1cmql0jkz5s954yi3oex3nedyz0fb0xx14ie39trti4wxv-8vxv8rssmk743ojnwacrr2e7c \
    192.168.99.100:2377

  To add a manager to this swarm, run 'docker swarm join-token manager' and follow the instructions.
  ```
   - Use last command to join worker host to Swarm cluster
     eg, on worker hosts,
    ```
    docker swarm join \
      --token SWMTKN-1-49nj1cmql0jkz5s954yi3oex3nedyz0fb0xx14ie39trti4wxv-8vxv8rssmk743ojnwacrr2e7c \
      192.168.99.100:2377
    ```
### Create overlay network
* Create overlay network which will be used as path between hyperledger nodes
  - on Master host,
    ```
    docker network create --attachable --driver overlay --subnet=10.200.1.0/24 hyperledger-ov
    ```
### Get Hyperledger Fabric artifacts and binaries
* As encryption keys and artifacts share same path, you need to put them on same location.
   - In this case, the location is '/nfs-share' (I personally use NFS)
   - For example,
   ```
    cd /nfs-share
    git clone https://github.com/hyperledger/fabric-samples.git
    cd fabric-samples
    curl -sSL https://goo.gl/iX9dek | bash
    ```
* I recommend you get ccenv image on all host. Fabric does not pull the latest ccenv image for you.
    so on ALL host,
    ```
    docker pull hyperledger/fabric-ccenv:x86_64-1.0.0
    ```

### Generate the artifacts
* clone this git
  ```
  cd first-network
  git clone https://github.com/ChoiSD/hyperledger_on_swarm.git
  ```
* copy config generator & crypto-config generating script
  ```
  cp hyperledger_on_swarm/generateArtifacts.sh $PWD
  cp hyperledger_on_swarm/genConfig/genConfig $PWD
  ```
* generate config file for your own configuration. For example,
  ```
  ./genConfig -domain sdchoi.com -Kafka 3 -Orderer 5 -Zookeeper 3 -Orgs 4 -Peer 4
  ```
  For information,
  ```
  ./genConfig -h
  ```
* generate artifacts
  ```
  ./generateArtifacts.sh -c <CHANNEL-NAME> -d <DOMAIN-NAME> -o <NUM-ORGS>
  ```
  For example,
  ```
  ./generateArtifacts.sh -c sdchannel -d sdchoi.com -o 4
  ```
* share two directories (channel-artifacts, crypto-config) among all hosts with same path

### Deploy Hyperledger Fabric nodes
* On Master host,
  - If you choose to configure with multiple orderers,  
  ```
  docker stack deploy -c hyperledger-zookeeper.yaml hyperledger-zk
  docker stack deploy -c hyperledger-kafka.yaml hyperledger-kafka
  docker stack deploy -c hyperledger-orderer.yaml hyperledger-orderer
  docker stack deploy -c hyperledger-couchdb.yaml hyperledger-couchdb
  docker stack deploy -c hyperledger-peer.yaml hyperledger-peer
  docker stack deploy -c hyperledger-ca.yaml hyperledger-ca
  docker stack deploy -c hyperledger-cli.yaml hyperledger-cli
  ```
  - If you choose to configure in 'solo' mode
  ```
  docker stack deploy -c hyperledger-orderer.yaml hyperledger-orderer
  docker stack deploy -c hyperledger-couchdb.yaml hyperledger-couchdb
  docker stack deploy -c hyperledger-peer.yaml hyperledger-peer
  docker stack deploy -c hyperledger-ca.yaml hyperledger-ca
  docker stack deploy -c hyperledger-cli.yaml hyperledger-cli
  ```
  - I recommend you to check services' status after every deploy.
  ```
  docker service ls
  ```
  - If something is wrong, check it, remove the stack, correct it and re-deploy.
  - In case you want to check if network is OK or ports are opened on containers, you can use busybox. busybox has ping, telnet, etc.
  ```
  docker stack deploy -c busybox.yaml busybox
  ```

### Do the rest of [Getting Started of Hyperledger Fabric Documentation](https://hyperledger-fabric.readthedocs.io/en/latest/getting_started.html)
* Create & Join channel, Install, Instantiate & Query chaincodes
  - NOTE: environment variables & orderer's name should be corrected to work properly.

#### Examples
* Generate Hyperledger Fabric config
  ```
  ./genConfig -domain sdchoi.com -Kafka 3 -Orderer 2 -Zookeeper 3 -Orgs 2 -Peer 2
  ```
* Generate artifacts
  ```
  ./generateArtifacts.sh -c sdchannel -d sdchoi.com -o 2
  ```
* Deploy Hyperledger Fabric
  ```
  docker stack deploy -c hyperledger-zookeeper.yaml hyperledger-zk
  docker stack deploy -c hyperledger-kafka.yaml hyperledger-kafka
  docker stack deploy -c hyperledger-orderer.yaml hyperledger-orderer
  docker stack deploy -c hyperledger-couchdb.yaml hyperledger-couchdb
  docker stack deploy -c hyperledger-peer.yaml hyperledger-peer
  docker stack deploy -c hyperledger-ca.yaml hyperledger-ca
  docker stack deploy -c hyperledger-cli.yaml hyperledger-cli
  ```
* Attach to cli container
  ```
  docker exec -it $(docker ps --format "table {{.ID}}" -f "label=com.docker.stack.namespace=hyperledger-cli" | tail -1) bash
  ```
* Create a channel
  ```
  export CHANNEL_NAME=sdchannel
  peer channel create -o orderer0.sdchoi.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --tls $CORE_PEER_TLS_ENABLED --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sdchoi.com/orderers/orderer0.sdchoi.com/msp/tlscacerts/tlsca.sdchoi.com-cert.pem
  ```
* Join a channel
  ```
  peer channel join -b sdchannel.block
  peer channel list
  ```
* Install & Initiate Chaincode
  ```
  peer chaincode install -n mycc -v 1.0 -p github.com/hyperledger/fabric/examples/chaincode/go/chaincode_example02
  peer chaincode instantiate -o orderer0.sdchoi.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sdchoi.com/orderers/orderer0.sdchoi.com/msp/tlscacerts/tlsca.sdchoi.com-cert.pem -C $CHANNEL_NAME -n mycc -v 1.0 -c '{"Args":["init","a", "100", "b","200"]}' -P "OR ('Org1MSP.member','Org2MSP.member')"
  ```
* Query & Invoke Chaincode
  ```
  peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["query","a"]}'
  peer chaincode invoke -o orderer1.sdchoi.com:7050  --tls $CORE_PEER_TLS_ENABLED --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sdchoi.com/orderers/orderer1.sdchoi.com/msp/tlscacerts/tlsca.sdchoi.com-cert.pem  -C $CHANNEL_NAME -n mycc -c '{"Args":["invoke","a","b","10"]}'
  peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["query","a"]}'
  peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["query","b"]}'
  ```
