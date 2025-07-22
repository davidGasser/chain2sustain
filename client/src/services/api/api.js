// Import required modules
const express = require('express');
const multer = require('multer');
const path = require('path');
const fs = require('fs').promises;  // Use promises-based fs module

const {
  manufactureParts,
  changeContract,
  changeGateway
} = require('../blockchainService');

// Create Express app for the API
const app = express();
const port = 3001; // Use a different port for the API

// Configure Multer for file upload
const storage = multer.diskStorage({
  destination: './uploads',
  filename: (req, file, cb) => {
    cb(null, file.fieldname + '-' + Date.now() + path.extname(file.originalname));
  }
});
const upload = multer({ storage });

// Define API endpoints
app.post('/manufacture', upload.single('file'), async (req, res) => {
    // Process manufacture data and handle file upload here
  res.json({ message: 'Product successfully created!' });
});

app.post('/transfer', upload.none(), (req, res) => {
  // Process transfer data here
  // ...

  res.json({ message: 'Transfer successfully submitted!' });
});

app.post('/emissions', upload.single('file'), async (req, res) => {
  // Process emissions data and handle file upload here
  // ...

  res.json({ message: 'Emissions successfully recorded!' });
});

app.post('/settings/gateway', upload.none(), (req, res) => {
  // Process gateway settings data here
  // ...

  res.json({ message: 'Settings successfully changed!' });
});

app.post('/settings/channel', upload.none(), (req, res) => {
  // Process channel settings data here
  // ...

  res.json({ message: 'Settings successfully changed!' });
});

app.post('/overview', upload.none(), (req, res) => {
  // Process overview data here
  // ...

  res.json({ message: 'ProductID successfully queried!', result: 'Placeholder for result of query' });
});

// Error handling middleware
app.use((err, req, res, next) => {
  console.error(err);
  res.status(500).json({ error: 'Internal Server Error' });
});

// Start the API server
app.listen(port, () => {
  console.log(`API server listening at http://localhost:${port}`);
});
