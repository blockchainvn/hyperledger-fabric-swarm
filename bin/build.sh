BASE_DIR=$PWD

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