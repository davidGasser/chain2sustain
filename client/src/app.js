//=======================================================
// Client Main Application
//=======================================================

// Import required modules
const express = require('express');
const multer = require('multer');
const path = require('path');
const fs = require('fs').promises;  // Use promises-based fs module
const session = require('express-session');
const api = require('./services/api/api.js');

const {
  recordEmissions,
  createProduct,
  createTransfer,
  confirmTransfer,
  configureGateway,
} = require('./services/blockchainService');

// Create Express app
const app = express();
const port = 3000;

// Serve static files from the "public" directory
app.use(express.static('public'));

// Configure EJS as the view engine
app.set('view engine', 'ejs');
app.set('views', path.join(__dirname, 'views'));  // Set views directly

// Configure session middleware
app.use(session({
  secret: process.env.SESSION_SECRET || 'secret',  // Use environment variable or default value
  resave: true,
  saveUninitialized: true
}));

// Configure Multer for file upload
const storage = multer.diskStorage({
  destination: './uploads',
  filename: (req, file, cb) => {
    cb(null, file.fieldname + '-' + Date.now() + path.extname(file.originalname));
  }
});
const upload = multer({ storage });

// Home page route
app.get('/', (req, res) => {
  res.render('index');
});

// Input page route
app.get('/input', (req, res) => {
  const { successMsg, errorMsg } = req.session;
  delete req.session.successMsg;
  delete req.session.errorMsg;

  const { inputData } = req.session;
  delete req.session.inputData;

  res.render('input', { successMsg, errorMsg, inputData });
});

// Submit form route
app.post('/submit', upload.single('file'), async (req, res, next) => {
  const { recipeID, assetName, consumedAssetIDs, emissionsTokens, additionalInfo } = req.body;
  const file = req.file;

  const parsedAssetIDs = consumedAssetIDs ? consumedAssetIDs.split(',').map((id) => id.trim()) : [];
  const parsedEmissionsTokens = emissionsTokens ? emissionsTokens.split(',').map((token) => token.trim()) : [];

  if (file) {
    const filePath = `uploads/${file.filename}`;
    try {
      await fs.rename(file.path, filePath);
      console.log('File uploaded:', filePath);
    } catch (error) {
      console.error('Error uploading the file:', error);
    }
  }

  console.log('Received data:');
  console.log('Recipe ID:', recipeID);
  console.log('Asset Name:', assetName);
  console.log('Consumed AssetIDs:', parsedAssetIDs);
  console.log('emissions Tokens:', parsedEmissionsTokens);
  console.log('Additional Information:', additionalInfo);

  const successMsg = 'Product successfully created!';
  const errorMsg = 'Error creating product!';
  const inputData = {
    recipeID: recipeID || '',
    consumedAssetIDs: parsedAssetIDs || '',
    emissionsTokens: parsedEmissionsTokens || '',
    additionalInfo: additionalInfo || '',
  };

  // Invoke chaincode 
  createProduct(recipeID, assetName, parsedEmissionsTokens, parsedAssetIDs ).then((result) => {
    req.session.successMsg = successMsg + '\n' + result;
  }).catch((error) => {
    req.session.errorMsg = errorMsg;
  }).finally(() => {  
    req.session.inputData = inputData;
    res.redirect('/input');
  });  
});


// Transfer products route
app.get('/transfer', (req, res) => {
  const { successMsg, errorMsg } = req.session;
  delete req.session.successMsg;
  delete req.session.errorMsg;

  const { transferData } = req.session;
  delete req.session.transferData;

  res.render('transfer', { successMsg, errorMsg, transferData });
});

// Submit transfer products route
app.post('/submit_transfer', upload.none(), (req, res) => {
  const { shippingID, quantity, list_ID, assetName, emissionsTokens} = req.body;

  console.log('Received transfer:');
  console.log('Shipping ID: ', shippingID);
  console.log('Quantity: ', quantity);
  console.log('List ID: ', list_ID);
  console.log('Asset Name: ', assetName);
  console.log('Emissions Tokens: ', emissionsTokens);

  const successMsg = 'Transfer successfull!';
  const errorMsg = 'Error while transfering asset!';
  const transferData = {
    shippingID: shippingID || '',
    quantity: quantity || '',
    list_ID: list_ID || '',
    assetName: assetName || '',
    emissionsTokens: emissionsTokens || '',
  };

  // Invoke chaincode
  createTransfer(shippingID, quantity, list_ID, assetName, emissionsTokens).then((result) => {
    req.session.successMsg = successMsg + '\n' + result;
  }).catch((error) => {
    req.session.errorMsg = errorMsg;
  }).finally(() => { 
    req.session.transferData = transferData;
    res.redirect('/transfer');
  });
});

// Submit transfer products route
app.post('/submit_transfer_confirm', upload.none(), (req, res) => {
  const { shippingID, quantity, list_ID, assetName, emissionsTokens} = req.body;

  console.log('Received transfer:');
  console.log('Shipping ID: ', shippingID);
  console.log('Quantity: ', quantity);
  console.log('List ID: ', list_ID);
  console.log('Asset Name: ', assetName);
  console.log('Emissions Tokens: ', emissionsTokens);

  const successMsg = 'Transfer successfully confirmed!';
  const errorMsg = 'Error while confirming transfer!';
  const transferData = {
    shippingID: shippingID || '',
    quantity: quantity || '',
    list_ID: list_ID || '',
    assetName: assetName || '',
    emissionsTokens: emissionsTokens || '',
  };

  // Invoke chaincode
  confirmTransfer(shippingID, quantity, list_ID, assetName, emissionsTokens).then((result) => {
    req.session.successMsg = successMsg + '\n' + result;
  }).catch((error) => {
    req.session.errorMsg = errorMsg;
  }).finally(() => { 
    req.session.transferData = transferData;
    res.redirect('/transfer');
  }); 
});

// Emissions recording page route
app.get('/emissions', (req, res) => {
  const { successMsg, errorMsg } = req.session;
  delete req.session.successMsg;
  delete req.session.errorMsg;

  const { inputData } = req.session;
  delete req.session.inputData;

  res.render('emissions', { successMsg, errorMsg, inputData });
});

// Submit Emissions Record form route
app.post('/submit_emissions', upload.single('file'), async (req, res, next) => {
  const { ghgEmissions, additionalInfo } = req.body;
  const file = req.file;

  // Store file on Filesystem
  if (file) {
    const filePath = `uploads/${file.filename}`;
    try {
      await fs.rename(file.path, filePath);
      console.log('File uploaded:', filePath);
    } catch (error) {
      console.error('Error uploading the file:', error);
    }
  }

  console.log('Received Emissions Record:');
  console.log('GHG Emissions:', ghgEmissions);
  console.log('Additional Information:', additionalInfo);

  // Create response notification 
  const successMsg = 'Emissions successfully recorded!';
  const errorMsg = 'Error recording emissions!';
  const inputData = {
    ghgEmissions: ghgEmissions || '',
    additionalInfo: additionalInfo || '',
  };

  // Invoke chaincode
  recordEmissions(ghgEmissions, additionalInfo).then((result) => {
    console.log('Emissions recorded successfully!');
    console.log(result);
    req.session.successMsg = successMsg + '\n' + result;
  }).catch((error) => {
    console.log('Error recording emissions!');
    console.log(error);
    req.session.errorMsg = errorMsg;
  }).finally(() => {
    console.log('Emissions recording finished!');
    req.session.inputData = inputData;
    res.redirect('/input');
  });
});

// Settings page route
app.get('/settings', (req, res) => {
  const { successMsg, errorMsg } = req.session;
  delete req.session.successMsg;
  delete req.session.errorMsg;

  const { settingsData } = req.session;
  delete req.session.settingsData;

  res.render('settings', { successMsg, errorMsg, settingsData });
});

// Submit gateway settings route
app.post('/submit_settings', upload.none(), (req, res) => {
  const { preconfiguredSettings, gatewayAddress, organizationID, certificatePath, privateKeyPath, tlsRootCertPath } = req.body;

  const gatewayConfig = { gatewayAddress, organizationID, certificatePath, privateKeyPath, tlsRootCertPath };

  console.log('Received data:');
  console.log('Preconfigured Settings:', preconfiguredSettings);
  console.log('Gateway Address:', gatewayAddress);
  console.log('Organization ID:', organizationID);
  console.log('Certificate Path:', certificatePath);
  console.log('Private Key Path:', privateKeyPath);
  console.log('TLS Root Certificate Path:', tlsRootCertPath);

  if (preconfiguredSettings !== '') {
    switch (preconfiguredSettings) {
      case 'Org1':
        gatewayConfig.gatewayAddress = 'localhost:7041';
        gatewayConfig.organizationID = 'Org1';
        gatewayConfig.certificatePath = '/Users/machineone/Documents/DLT4PI/sustainable-supply-chain/fablo-target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp/signcerts/peer0.org1.example.com-cert.pem';
        gatewayConfig.privateKeyPath = '/Users/machineone/Documents/DLT4PI/sustainable-supply-chain/fablo-target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp/keystore/priv-key.pem';
        gatewayConfig.tlsRootCertPath = '/Users/machineone/Documents/DLT4PI/sustainable-supply-chain/fablo-target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt';
        break;
      case 'Org2':
        gatewayConfig.gatewayAddress = 'localhost:7061';
        gatewayConfig.organizationID = 'Org2';
        gatewayConfig.certificatePath = '/Users/machineone/Documents/DLT4PI/sustainable-supply-chain/fablo-target/fabric-config/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/msp/signcerts/peer0.org2.example.com-cert.pem';
        gatewayConfig.privateKeyPath = '/Users/machineone/Documents/DLT4PI/sustainable-supply-chain/fablo-target/fabric-config/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/msp/keystore/priv-key.pem';
        gatewayConfig.tlsRootCertPath = '/Users/machineone/Documents/DLT4PI/sustainable-supply-chain/fablo-target/fabric-config/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt';
        break;
      case 'Org3':
        gatewayConfig.gatewayAddress = 'localhost:7081';
        gatewayConfig.organizationID = 'Org3';
        gatewayConfig.certificatePath = '/Users/machineone/Documents/DLT4PI/sustainable-supply-chain/fablo-target/fabric-config/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/msp/signcerts/peer0.org3.example.com-cert.pem';
        gatewayConfig.privateKeyPath = '/Users/machineone/Documents/DLT4PI/sustainable-supply-chain/fablo-target/fabric-config/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/msp/keystore/priv-key.pem';
        gatewayConfig.tlsRootCertPath = '/Users/machineone/Documents/DLT4PI/sustainable-supply-chain/fablo-target/fabric-config/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt';
        break;
      default:
        console.error('Invalid preconfigured settings!');
        break;
    }
  }

  const successMsg = 'Settings successfully changed!';
  const errorMsg = 'Error while changing the settings!';
  const settingsData = {
    preconfiguredSettings: preconfiguredSettings || '',
    gatewayAddress: gatewayAddress || '',
    organizationID: organizationID || '',
    certificatePath: certificatePath || '',
    privateKeyPath: privateKeyPath || '',
    tlsRootCertPath: tlsRootCertPath || ''
  };

  configureGateway(gatewayConfig).then(result => {
    if (result) {
      console.log('Gateway successfully configured!');
      req.session.successMsg = successMsg;
    } else {
      console.error('Error while configuring the gateway!');
      req.session.errorMsg = errorMsg;
    }
  }).catch(error => {
    console.error('Error while configuring the gateway:', error);
    req.session.errorMsg = errorMsg;
  }).finally(() => {
    req.session.settingsData = settingsData;
    res.redirect('/settings');
  });  
});

// Overview page route
app.get('/overview', (req, res) => {
  const { successMsg, errorMsg } = req.session;
  delete req.session.successMsg;
  delete req.session.errorMsg;

  const { inputData, result } = req.session;
  delete req.session.inputData;

  res.render('overview', { successMsg, errorMsg, inputData, result });
});

// Submit gateway settings route
app.post('/submit_overview', upload.none(), (req, res) => {
  const { productID } = req.body;

  console.log('Received data:');
  console.log('ProductID:', productID);

  // TODO: Query product data from blockchain

  const isError = false;

  const successMsg = 'ProductID successfully querried!';
  const errorMsg = 'Error retrieving product info!';
  const inputData = {
    id: productID || '',
  };

  if (isError) {
    req.session.errorMsg = errorMsg;
  } else {
    req.session.successMsg = successMsg;
  }
  req.session.inputData = inputData;
  req.session.result = "Placeholder for result of query";
  res.redirect('/overview');
});

// Error handling middleware
app.use((err, req, res, next) => {
  console.error(err);
  res.status(500).send('Internal Server Error');
});

// Start the server
app.listen(port, () => {
  console.log(`Server listening at http://localhost:${port}`);
});
