BASE_DIR=$PWD

if [[ -d $GOPATH ]];then
  mkdir -p /opt/gopath/src
  export GOPATH=/opt/gopath
  cd $GOPATH/src
  mkdir -p github.com/hyperledger
  cd github.com/hyperledger
  git clone https://github.com/hyperledger/fabric.git
fi

cd ${GOPATH}/src/github.com/hyperledger/fabric/
make configtxgen
make cryptogen  
# check combind of 2 results
echo "===================== Crypto tools built successfully ===================== "
echo 
echo "Copying to bin folder of network..."
echo
cp ./build/bin/configtxgen $BASE_DIR
cp ./build/bin/cryptogen $BASE_DIR