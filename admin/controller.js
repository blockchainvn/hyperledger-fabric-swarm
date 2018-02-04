/*
* Copyright IBM Corp All Rights Reserved
*
* SPDX-License-Identifier: Apache-2.0
*/
/*
 * Chaincode query
 */
// var fs = require("fs-extra");
var x509 = require("x509");
var Fabric_Client = require("fabric-client");
var path = require("path");
var util = require("util");

module.exports = function(channelName, address) {
  var fabric_client = new Fabric_Client();
  // const tlsCACertPEM = fs.readFileSync(
  //   "./crypto-config/peerOrganizations/org" +
  //     program.org +
  //     ".example.com/peers/peer0.org" +
  //     program.org +
  //     ".example.com/tls/ca.crt"
  // );

  // setup the fabric network
  // var channel = fabric_client.newChannel(channelName);
  // var peer = fabric_client.newPeer(
  //   "grpcs://localhost:" + (program.org == 1 ? 7051 : 8051),
  //   {
  //     pem: tlsCACertPEM.toString(),
  //     "ssl-target-name-override": "peer0.org" + program.org + ".example.com"
  //   }
  // );
  var store_path = path.join(__dirname, "hfc-key-store");
  console.log("Store path:" + store_path);

  return {
    get_member_user(user) {
      // create the key value store as defined in the fabric-client/config/default.json 'key-value-store' setting
      return Fabric_Client.newDefaultKeyValueStore({
        path: store_path
      })
        .then(state_store => {
          // assign the store to the fabric client
          fabric_client.setStateStore(state_store);
          var crypto_suite = Fabric_Client.newCryptoSuite();
          // use the same location for the state store (where the users' certificate are kept)
          // and the crypto store (where the users' keys are kept)
          var crypto_store = Fabric_Client.newCryptoKeyStore({
            path: store_path
          });
          crypto_suite.setCryptoKeyStore(crypto_store);
          fabric_client.setCryptoSuite(crypto_suite);

          // get the enrolled user from persistence, this user will sign all requests
          return fabric_client.getUserContext(user, true);
        })
        .then(user_from_store => {
          if (user_from_store && user_from_store.isEnrolled()) {
            console.log("Successfully loaded " + user + " from persistence");
            return user_from_store;
          } else {
            throw new Error(
              "Failed to get " + user + ".... run node register.js -u " + user
            );
          }
        });
    },

    getEventTxPromise(eventAdress, transaction_id_string) {
      return new Promise((resolve, reject) => {
        let event_hub = fabric_client.newEventHub();
        event_hub.setPeerAddr("grpc://" + eventAdress);
        console.log("eventhub: grpc://" + eventAdress);
        event_hub.connect();

        let handle = setTimeout(() => {
          event_hub.disconnect();
          resolve({ event_status: "TIMEOUT" }); //we could use reject(new Error('Trnasaction did not complete within 30 seconds'));
        }, 3000);

        event_hub.registerTxEvent(
          transaction_id_string,
          (tx, code) => {
            // this is the callback for transaction event status
            // first some clean up of event listener
            clearTimeout(handle);
            event_hub.unregisterTxEvent(transaction_id_string);
            event_hub.disconnect();

            // now let the application know what happened
            var return_status = {
              event_status: code,
              tx_id: transaction_id_string
            };
            if (code !== "VALID") {
              console.error("The transaction was invalid, code = " + code);
              resolve(return_status); // we could use reject(new Error('Problem with the tranaction, event status ::'+code));
            } else {
              console.log(
                "The transaction has been committed on peer " +
                  event_hub._ep._endpoint.addr
              );
              resolve(return_status);
            }
          },
          err => {
            //this is the callback if something goes wrong with the event registration or processing
            reject(new Error("There was a problem with the eventhub ::" + err));
          }
        );
      });
    },

    query(user, request) {
      var channel = fabric_client.newChannel(channelName);
      var peer = fabric_client.newPeer("grpc://" + address);
      channel.addPeer(peer);
      console.log("Peer: " + "grpc://" + address);

      return this.get_member_user(user)
        .then(user_from_store => {
          return channel.queryByChaincode(request);
        })
        .then(query_responses => {
          console.log(
            "Query has completed on channel [" +
              channelName +
              "], checking results"
          );
          // query_responses could have more than one  results if there multiple peers were used as targets
          if (query_responses && query_responses.length == 1) {
            if (query_responses[0] instanceof Error) {
              console.error("error from query = ", query_responses[0]);
            } else {
              // const response = query_responses[0];
              return query_responses[0];
              // console.log("Response is \n", response);
            }
          } else {
            console.log("No payloads were returned from query");
            return null;
          }
        });
    },

    invoke(user, invokeRequest) {
      var tx_id;
      var channel = fabric_client.newChannel(channelName);
      var peer = fabric_client.newPeer("grpc://" + address);
      channel.addPeer(peer);
      console.log("Peer: " + "grpc://" + address);
      var orderer = fabric_client.newOrderer(
        "grpc://" + invokeRequest.ordererAddress
      );
      channel.addOrderer(orderer);

      return this.get_member_user(user)
        .then(user_from_store => {
          tx_id = fabric_client.newTransactionID();

          return channel.sendTransactionProposal({
            chaincodeId: invokeRequest.chaincodeId,
            fcn: invokeRequest.fcn,
            args: invokeRequest.args,
            chainId: channelName,
            txId: tx_id
          });
        })
        .then(results => {
          var proposalResponses = results[0];
          var proposal = results[1];
          let isProposalGood = false;
          if (
            proposalResponses &&
            proposalResponses[0].response &&
            proposalResponses[0].response.status === 200
          ) {
            isProposalGood = true;
            console.log("Transaction proposal was good");
          } else {
            console.error("Transaction proposal was bad");
          }

          if (isProposalGood) {
            console.log(
              util.format(
                'Successfully sent Proposal and received ProposalResponse: Status - %s, message - "%s"',
                proposalResponses[0].response.status,
                proposalResponses[0].response.message
              )
            );

            var txPromise = this.getEventTxPromise(
              invokeRequest.eventAddress,
              tx_id.getTransactionID()
            );

            var sendPromise = channel.sendTransaction({
              proposalResponses: proposalResponses,
              proposal: proposal
            });

            return Promise.all([sendPromise, txPromise]);
          } else {
            throw new Error(
              "Failed to send Proposal or receive valid response. Response null or status is not 200. exiting..."
            );
          }
        });
    }
  };
};
