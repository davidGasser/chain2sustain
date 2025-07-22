/**
@module blockchainService
@description A module that provides functions for interacting with a Hyperledger Fabric network through a gateway.
@exports {function} manufactureParts - Manufactures parts using the provided inputs.
@exports {function} changeContract - Changes the contract based on the inputs.
@exports {function} changeGateway - Changes the gateway configuration based on the inputs.
*/

const grpc = require('@grpc/grpc-js');
const { connect, signers } = require('@hyperledger/fabric-gateway');
const fs = require('fs').promises;
const crypto = require('crypto');
const { exec } = require("child_process");

const utf8Decoder = new TextDecoder();

// Client configuration
let gatewayAddress = 'localhost:7041'; // Default address of Org1 Peer 0
let membershipID = 'Org1'; // Default membership ID of Org1
let pathToCertificate = 'fablo-target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp/signcerts/peer0.org1.example.com-cert.pem'; // Default path of Org1 Peer 0 certificate
let pathToPrivateKey = 'fablo-target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp/keystore/priv-key.pem'; // TODO: Replace with actual path to private key
let tlsRootCertPath = '/Users/machineone/Documents/DLT4PI/sustainable-supply-chain/fablo-target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt';

// Channel/Chaincode constants
let channelName = 'my-channel1';//'supply-chain-channel';
let emissionsChaincodeName = 'channel1'//'emissionsAudit';//'emissions-cc';
let transferChaincodeName = 'transfer-cc';

// Gateway configuration
var contract;
var gateway;

/**
 * Retrieves the contract instance from the gateway for the specified channel and chaincode.
 * @param {string} channelName - The name of the channel.
 * @param {string} chaincodeName - The name of the chaincode.
 * @returns {object} The contract instance.
 */
function getContract(channelName, chaincodeName) {
  return gateway.getNetwork(channelName).getContract(chaincodeName);
}

/**
 * Submits a transaction to the contract with the specified name and data.
 * @param {object} contract - The contract instance.
 * @param {string} transactionName - The name of the transaction.
 * @param {Array} transactionData - The data to be passed to the transaction.
 * @returns {Promise<any>} Resolves with the result of the transaction.
 */
async function submitTransactionDocker(contract, transactionName, transactionData) {
  console.log('Submitting transaction...');
  transactionData.unshift(transactionName);
  let command = "docker exec cli."+ gateway.getIdentity().mspId.toString().toLowerCase() +".example.com peer chaincode invoke -C " + contract.channelName + " -n "+ contract.chaincodeName +" --peerAddresses peer0."+ gateway.getIdentity().mspId.toString().toLowerCase() +".example.com:"+gatewayAddress.slice(-4)+" -c '{\"Args\":"+ JSON.stringify(transactionData)+"}'"
  console.log(command) 
  //const result = await contract.submitTransaction(...transactionData);
  exec(command, (error, stdout, stderr) => {
    if (error) {
        console.log(`error: ${error.message}`);
        return;
    }
    if (stderr) {
        console.log(`stderr: ${stderr}`);
        return;
    }
    console.log(`stdout: ${stdout}`);
    return stdout;
  });
}

/**
 * Queries the ledger using the specified transaction name and data.
 * @param {object} contract - The contract instance.
 * @param {string} transactionName - The name of the transaction.
 * @param {Array} transactionData - The data to be passed to the transaction.
 * @returns {Promise<any>} Resolves with the result of the query.
 */
async function queryLedgerDocker(contract, transactionName, transactionData, channelName, chaincodeName) {
  console.log('Submitting transaction...');
  transactionData.unshift(transactionName);
  let command = "docker exec cli."+ gateway.getIdentity().mspId.toString().toLowerCase() +".example.com peer chaincode invoke -C " + channelName + " -n "+ chaincodeName +" --peerAddresses peer0."+ gateway.getIdentity().mspId.toString().toLowerCase() +".example.com:"+gatewayAddress.slice(-4)+" -c '{\"Args\":"+JSON.stringify(transactionData)+"}'"
  console.log(command) 
  //const result = await contract.submitTransaction(...transactionData);
  exec(command, (error, stdout, stderr) => {
    if (error) {
        console.log(`error: ${error.message}`);
        return;
    }
    if (stderr) {
        console.log(`stderr: ${stderr}`);
        return;
    }
    console.log(`stdout: ${stdout}`);
    console.log(utf8Decoder.decode(stdout));
    return stdout;
  });
}

/**
 * Submits a transaction to the contract with the specified name and data.
 * @param {object} contract - The contract instance.
 * @param {string} transactionName - The name of the transaction.
 * @param {Array} transactionData - The data to be passed to the transaction.
 * @returns {Promise<any>} Resolves with the result of the transaction.
 */
async function submitTransaction(contract, transactionName, transactionData) {
  console.log('Submitting transaction...');
  transactionData.unshift(transactionName);
  const result = await contract.submitTransaction(...transactionData);
  return result;
}

/**
 * Queries the ledger using the specified transaction name and data.
 * @param {object} contract - The contract instance.
 * @param {string} transactionName - The name of the transaction.
 * @param {Array} transactionData - The data to be passed to the transaction.
 * @returns {Promise<any>} Resolves with the result of the query.
 */
async function queryLedger(contract, transactionName, transactionData) {
  console.log('Submitting transaction...');
  transactionData.unshift(transactionName);
  const result = await contract.evaluateTransaction(...transactionData);
  return result;
}

/**
 * Create Products using the provided inputs.
 * @param {string} recipeID - The ID of the recipe used to create the product.
 * @param {string} assetName - The name of the product.
 * @param {number} emissionsTokens - List of the emissions token IDs used to create the product.
 * @param {Array} assets - List of the asset IDs used to create the product.
 */
async function createProduct(recipeID, assetName, emissionsTokens, assets) {
  // Set contract
  contract = getContract(channelName, transferChaincodeName);

  // Check if given inputs are valid
  // Input IDs need to exist `somewhere`
  // TODO: Input sanitation/checking

  // Query all assets to find suitable id
  const assetRecords = await queryLedger(contract, ['GetAllAssets']);
  const usedIDs = assetRecords.map(asset => Number(asset.assetID));
  // Generate unique ID
  if (usedIDs.length > 0) {
    var id = Math.max(...usedIDs) + 1;
  } else {
    var id = 1;
  }
  id = id.toString();

  const transientData = {
    'assetID': id,
    'recipeID': recipeID,
    'assetName': assetName,
    'emissionsTokens': emissionsTokens,
    'assets': assets
  };
  contract.setTransient(transientData)
  return submitTransaction(contract, 'ManufactureAsset', []);
}

/**
 * Creates a new shipping record using the provided inputs.
 * @param {string} id - The ID of the shipping record.
 * @param {number} quantity - The quantity of the asset being shipped.
 * @param {string} list_id - 
 * @param {number} emissionsTokens - List of the emissions token IDs used to create the product.
 * @param {string} assetName - The name of the asset being shipped.
 */
async function createTransfer(id, quantity, list_id, emissionsTokens, assetName) {
  // Set contract
  contract = getContract(channelName, transferChaincodeName);

  const transientData = {
    'shippingID': id,
    'quantity': quantity,
    'list_ID': list_id,
    'assetName': assetName,
    'date': new Date().toISOString(),
    'emissionsTokens': emissionsTokens
  };
  contract.setTransient(transientData)
  return submitTransaction(contract, 'CreateShipping', []);
}

/**
 * Confirms the transfer of the given shipping record.
 * @param {string} id - The ID of the shipping record.
 * @param {number} quantity - The quantity of the asset being shipped.
 * @param {string} list_id -
 * @param {number} emissionsTokens - List of the emissions token IDs used to create the product.
 * @param {string} assetName - The name of the asset being shipped.
 */
async function confirmTransfer(id, quantity, list_id, emissionsTokens, assetName) {
  // Set contract
  contract = getContract(channelName, transferChaincodeName);

  const transientData = {
    'shippingID': id,
    'quantity': quantity,
    'list_ID': list_id,
    'assetName': assetName,
    'date': new Date().toISOString(),
    'emissionsTokens': emissionsTokens
  };
  contract.setTransient(transientData)
  return submitTransaction(contract, 'ClaimShipping', []);
}

/**
 * Uploads the given emissions data to the ledger.
 * @param {number} ghgNumber - The amount of CO2 emissions in kg.
 * @param {string} additionalInfo - Additional information about the emissions.
 */
function recordEmissions(ghgNumber, additionalInfo) {
  contract = getContract(channelName, emissionsChaincodeName);

  return queryLedger(contract, 'GetAllEmissionsRecords', [])
    .then(emissionRecords => {
      console.log(emissionRecords);
      const usedIDs = emissionRecords.map(struct => Number(struct.ID));
      if (usedIDs.length > 0) {
        var id = Math.max(...usedIDs) + 1;
      } else {
        var id = 1;
      }
      id = id.toString();

      contract.setTransient({"ownerID": membershipID});
      return queryLedger(contract, 'GetEmissionsRecordsOfOwner', []);
    })
    .then(prevEmissionIDs => {
      var transactionData = [
        id, 
        prevEmissionIDs,
        ghgNumber,
        additionalInfo
      ];
      return submitTransaction(contract, 'AuditEmissions', transactionData);
    });
}


/**
 * Configures the gateway configuration based on the inputs.
 * Updates the gateway address, membership ID, certificate path, and private key path.
 * If a gateway connection exists, creates a new connection with the updated configuration.
 * @param {object} inputs - The inputs for changing the gateway configuration.
 * @returns {Promise<boolean>} Resolves with true if the gateway connection was successfully established.
 */
async function configureGateway(inputs) {
  console.log(process.cwd());
  // Store inputs
  gatewayAddress = inputs.gatewayAddress;
  membershipID = inputs.organizationID;
  pathToCertificate = inputs.certificatePath;
  pathToPrivateKey = inputs.privateKeyPath;
  tlsRootCertPath = inputs.tlsRootCertPath;
  
  // Create Identity
  const credentials = await fs.readFile(pathToCertificate);
  const identity = { mspId: membershipID, credentials };

  // Create Signer
  const privateKeyPem = await fs.readFile(pathToPrivateKey);
  const privateKey = crypto.createPrivateKey(privateKeyPem);
  const signer = signers.newPrivateKeySigner(privateKey);

  // TLS Configuration
  const tlsRootCert = await fs.readFile(tlsRootCertPath);
  const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);

  // Create GRPC client
  console.log('Credentiols: ' + credentials);
  const client = new grpc.Client(gatewayAddress, tlsCredentials);

  // Create Gateway Connection
  gateway = connect({ identity, signer, client });

  if (!gateway) {
    throw new Error('Failed to connect to gateway');
  }
  console.log('Gateway connection established');
  return true;
}

// Exports
module.exports = {
  recordEmissions: recordEmissions,
  createProduct: createProduct,
  createTransfer: createTransfer,
  confirmTransfer: confirmTransfer,
  configureGateway: configureGateway,
};
