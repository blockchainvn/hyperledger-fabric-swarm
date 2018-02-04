//SPDX-License-Identifier: Apache-2.0

// nodejs server setup

// call the packages we need
const express = require("express"); // call express
const bodyParser = require("body-parser");
const controller = require("./controller")("mychannel", "peer0.org1-f-1:7051");
const app = express(); // define our app using express
// Load all of our middleware
// configure app to use bodyParser()
// this will let us get the data from a POST
// app.use(express.static(__dirname + '/client'));
app.use(bodyParser.urlencoded({ extended: true }));
app.use(bodyParser.json());

app.get("/query", function(req, res) {
  const request = {
    //targets : --- letting this default to the peers assigned to the channel
    chaincodeId: req.query.chaincode,
    fcn: req.query.method,
    args: req.query.arguments
  };

  // each method require different certificate of user
  controller
    .query(req.query.user, request)
    .then(ret => {
      res.send(ret.toString());
    })
    .catch(err => {
      res.status(500).send(err);
    });
});

app.get("/invoke", function(req, res) {
  const request = {
    chaincodeId: req.query.chaincode,
    fcn: req.query.method,
    args: req.query.arguments,
    eventAddress: req.query.eventHost || "peer0.org1-f-1:7053",
    ordererAddress: req.query.ordererHost || "orderer0.orgorderer-f-1:7050"
  };

  // each method require different certificate of user
  controller
    .invoke(req.query.user, request)
    .then(ret => {
      res.send(ret.toString());
    })
    .catch(err => {
      res.status(500).send(err);
    });
});

// Save our port
const port = process.env.PORT || 9000;

// Start the server and listen on port
app.listen(port, "0.0.0.0", () => {
  console.log("Live on port: " + port);
});
