var program = require("commander");

program
  .version("0.1.0")
  .option("-u, --user []", "User id", "user1")
  .option("--name, --channel []", "A channel", "mychannel")
  .option("--chaincode, --chaincode []", "A chaincode", "origincert")
  .option("--host, --host []", "Host", "peer0.org1-f-1:7051")
  .option("--ehost, --event-host []", "Host", "peer0.org1-f-1:7053")
  .option("--ohost, --orderer-host []", "Host", "orderer0.orgorderer-f-1:7050")
  .option("-m, --method []", "A method", "getCreator")
  .option(
    "-a, --arguments [value]",
    "A repeatable value",
    (val, memo) => memo.push(val) && memo,
    []
  )
  .parse(process.argv);

var controller = require("./controller")(program.channel, program.host);

var request = {
  //targets: let default to the peer assigned to the client
  chaincodeId: program.chaincode,
  fcn: program.method,
  args: program.arguments,
  eventAddress: program.eventHost,
  ordererAddress: program.ordererHost
};

// each method require different certificate of user
controller
  .invoke(program.user, request)
  .then(results => {
    console.log(
      "Send transaction promise and event listener promise have completed",
      results
    );
  })
  .catch(err => {
    console.error(err);
  });
