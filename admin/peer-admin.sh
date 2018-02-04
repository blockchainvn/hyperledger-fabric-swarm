
ORG=$1
if [[ -z $ORG ]];then
  echo "Please enter organization"
fi

MSP_PATH=../crypto-config/peerOrganizations/$ORG/users/Admin@${ORG}/msp
  
# create Peer Admin
PRIVATE_KEY=$(ls $MSP_PATH/keystore/*_sk | head -1)
CERTIFICATE=$(cat $MSP_PATH/signcerts/Admin@${ORG}-cert.pem | sed 's/$/\\r\\n/' | tr -d '\n')
PRIVATE_KEY_NAME=`basename $PRIVATE_KEY | sed 's/_sk//'`

cat << EOF > hfc-key-store/PeerAdmin
{
  "name": "PeerAdmin",
  "mspid": "Org1MSP",
  "roles": null,
  "affiliation": "",
  "enrollmentSecret": "",
  "enrollment": {
    "signingIdentity": "$PRIVATE_KEY_NAME",
    "identity": {
      "certificate": "$CERTIFICATE"
    }
  }
}
EOF

cp $PRIVATE_KEY hfc-key-store/${PRIVATE_KEY_NAME}-priv

echo "created PeerAdmin successfully ..."  
