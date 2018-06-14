BASE_DIR=$PWD

if [ ! -d "$GOPATH" ];then

  ARCH=`uname -s | grep Darwin`
  if [ "$ARCH" == "Darwin" ]; then
    if [ ! `command -v go` ]; then
      brew install go
    fi
  else
    if [ ! `command -v go` ]; then        
      apt install golang-go -y
    fi  
    apt install libtool libltdl-dev -y
  fi  

  apt install libtool libltdl-dev
  mkdir -p /opt/gopath/src
  export GOPATH=/opt/gopath  
fi

if [ ! -d "$GOPATH/src/github.com/hyperledger" ];then
  cd $GOPATH/src
  mkdir -p github.com/hyperledger
  cd github.com/hyperledger
  git clone -b release-1.1 https://github.com/hyperledger/fabric.git
fi

cd ${GOPATH}/src/github.com/hyperledger/fabric/
make configtxgen
make cryptogen  
# check combind of 2 results
echo "===================== Crypto tools built successfully ===================== "
echo 
echo "Copying to bin folder of network..."
echo

# make sure folder exists
mkdir -p $BASE_DIR/bin
cp ./build/bin/configtxgen $BASE_DIR/bin
cp ./build/bin/cryptogen $BASE_DIR/bin