# hyperledger_on_swarm

This repository is for deploying Hyperledger Fabric on Swarm cluster easily.

## Limitation
* ~~This works WITHOUT TLS only.~~
  - ~~Whenever enable TLS, grpc error code 14 occurs.~~ (seems it works with TLS..)
* ~~Kafka, Zookeeper has not been tested.~~ (Kafka & Zookeeper are also tested)


## Instructions
* There are two versions
  - solo : 2 CAs, 4 peers, 4 CouchDBs, 1 orderer
  - kafka : 2 CAs, 4 peers, 4 CouchDBs, 3 orderers, 3 kafkas, 3 zookeepers
  
### Pre-reqs
- 2 or more machines with Linux
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
* As all containers share same cryption keys and artifacts, you need to put them on same location.
   - In this case, the location is '/nfs-share' (I personally use NFS)
   - For example,
   ```
    cd /nfs-share
    curl -sSL https://goo.go/NIKLiU | bash
    ```
    - You need to change YAML file if you use other path.
    
### Generate the artifacts
* clone this git
  ```
  cd release/linux-amd64
  git clone https://github.com/ChoiSD/hyperledger_on_swarm.git
  cp hyperledger_on_swarm/generateArtifacts-swarm.sh $PWD
  ```
  - If you choose to deploy solo version,
    ```
    cp hyperledger_on_swarm/solo/* $PWD
    ```
  - If choose to deploy kafka version,
    ```
    cp hyperledger_on_swarm/kafka/* $PWD
    ```
* generate artifacts
  ```
  ./generateArtifacts-swarm.sh <CHANNEL-NAME>
  ```

### Deploy Hyperledger nodes
* On Master host,
  - If you choose to deploy solo version,
    ```
    docker stack deploy -c hyperledger-swarm.yaml hyperledger
    ```
  - If you choose to deploy kafka version,
    ```
    docker stack deploy -c hyperledger-zk.yaml hyperledger-zk
    docker stack deploy -c hyperledger-kafka.yaml hyperledger-kafka
    docker stack deploy -c hyperledger-orderer.yaml hyperledger-orderer
    docker stack deploy -c hyperledger-swarm.yaml hyperledger
    ```
  - In case you want to check if network is OK or ports are opened on containers, you can use busybox. busybox has ping, telnet, etc. 
    ```
    docker stack deploy -c busybox.yaml busybox
    ```

### Do the rest of [Getting Started of Hyperledger Fabric Documentation](https://hyperledger-fabric.readthedocs.io/en/latest/getting_started.html)
* Create & Join channel, Install, Instantiate & Query chaincodes
