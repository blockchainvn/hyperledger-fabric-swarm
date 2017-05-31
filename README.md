# hyperledger_on_swarm

This repository is for deploying Hyperledger Fabric on Swarm cluster easily.

## Limitation
* This works WITHOUT TLS only.
  - Whenever enable TLS, grpc error code 14 occurs.
* CA, CouchDB, Kafka, Zookeeper has not been tested.

### Pre-reqs
- 2 or more machines with Linux
- Install Docker >= 1.13

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
* Generate crypto keys and Create orderer genesis block, transaction artifacts
  ```
  cd release/linux-amd64
  ./generateArtifacts.sh <channel-ID>
  ```
  
### Create Docker Swarm cluster
* Create one or more master hosts and other worker hosts
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
     On worker hosts,
    ```
     docker swarm join \
    --token SWMTKN-1-49nj1cmql0jkz5s954yi3oex3nedyz0fb0xx14ie39trti4wxv-8vxv8rssmk743ojnwacrr2e7c \
    192.168.99.100:2377
    ```
### Create overlay network
* Create overlay network which will be used as path between hyperledger nodes
  - on Master host,
    ```
    $ docker network create --attachable --driver overlay --subnet=10.200.1.0/24 hyperledger-ov
    ```
### Deploy Hyperledger nodes
* Get 
