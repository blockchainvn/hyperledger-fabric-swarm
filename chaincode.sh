
ORG=$1
if [[ -z $ORG ]];then
  echo "Please enter organization"
fi

cat <<EOF > hyperledger-chaincode.yaml
version: "3"
networks:
  hyperledger-ov:
    external:
      name: hyperledger-ov
services:
  chaincode:
    deploy:
      replicas: 1
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
    image: hyperledger/fabric-ccenv:x86_64-1.0.2
    networks:
      hyperledger-ov:
        aliases:
        - chaincode
    environment:
    - GOPATH=/opt/gopath
      - CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock
      - CORE_LOGGING_LEVEL=DEBUG
      - CORE_PEER_ID=chaincode
      - CORE_PEER_ADDRESS=peer0.${ORG}:7051
    
    working_dir: /opt/gopath/src/chaincode
    command: sleep 3600
    volumes:
    - /var/run/:/host/var/run/
    - ./chaincode:/opt/gopath/src/chaincode    
EOF


docker stack deploy -c hyperledger-chaincode.yaml hyperledger-admin-api