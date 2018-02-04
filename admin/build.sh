IMAGE_NAME='hyperledger/admin-api'
NAMESPACE=$1 
PORT=$2
IMAGE_CHECK=$(docker images | grep $IMAGE_NAME)
WORKING_PATH=$PWD

if [[ -z $IMAGE_CHECK ]];then
  echo "Building image $IMAGE_NAME ..."
  echo
  docker build -t $IMAGE_NAME --target build-env .
fi

: ${NAMESPACE:="default"}
: ${PORT:="8888"}
  

# create template then you can run it normally
# policy is IfNotPresent because we might create service while image being created
# if set policy to never, we should wait for image ready
cat <<EOF > api-server.yaml
version: '3'
networks:
  hyperledger-ov:
    external:
      name: hyperledger-ov

services:
  admin:
    deploy:
      replicas: 1
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
    image: $IMAGE_NAME
    networks:
      hyperledger-ov:
        aliases:
        - admin
    
    volumes:       
      - $WORKING_PATH:/home
    ports:
      - $PORT    
    environment:
      - NODE_ENV=development    
    restart: always
    working_dir: /home
    command: /bin/bash -c 'yarn && yarn start'      

EOF


echo "Created api-server.yaml"