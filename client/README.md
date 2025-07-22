# Web Client for the Chain2Sustain Blockchain Application

- Node.js Application

## Installation
- run `npm install` to install all dependencies
- run `npm start` to start the application
- creates a web server on port 3000 (alternatively, you can run `npm run start-web`)
- creates a websocket server on port 3001 (for communication with the API) (alternatively, you can run `npm run start-api`)

## Limitations
In the current state, the client is not able to connect to a peer of the supply chain network. 
Furthermore, some functionalities are not yet implemented, such as the ability to create receips. 
The API endpoints are created, however, the client is not yet able process the requests.

# API Documentation

## Introduction
This API allows you to interact with the main application by providing endpoints for various functionalities, including manufacturing parts, transferring products, recording emissions, and changing settings.

## Base URL
The base URL for all endpoints is: `http://localhost:3001`

## Endpoints

### 1. Manufacture Parts

**Endpoint:** `/manufacture`  
**Method:** POST  
**Description:** Manufacture new parts and create a product.

**Request Body:**
- `ids` (string, optional): Comma-separated list of part IDs.
- `data` (string, optional): Product data.
- `additionalInfo` (string, optional): Additional information.
- `file` (file, optional): File attachment.

**Response:**
- `message` (string): Success message.

### 2. Transfer Products

**Endpoint:** `/transfer`  
**Method:** POST  
**Description:** Transfer products from one entity to another.

**Request Body:**
- `productID` (string): ID of the product being transferred.
- `supplierID` (string): ID of the supplier.
- `consumerID` (string): ID of the consumer.

**Response:**
- `message` (string): Success message.

### 3. Record Emissions

**Endpoint:** `/emissions`  
**Method:** POST  
**Description:** Record greenhouse gas emissions for a product.

**Request Body:**
- `ghgEmissions` (string): Greenhouse gas emissions data.
- `additionalInfo` (string, optional): Additional information.
- `file` (file, optional): File attachment.

**Response:**
- `message` (string): Success message.

### 4. Change Gateway Settings

**Endpoint:** `/settings/gateway`  
**Method:** POST  
**Description:** Change gateway settings.

**Request Body:**
- `gatewayAddress` (string): Gateway address.
- `organizationID` (string): Organization ID.
- `certificatePath` (string): Path to the certificate file.
- `privateKeyPath` (string): Path to the private key file.

**Response:**
- `message` (string): Success message.

### 5. Change Channel Settings

**Endpoint:** `/settings/channel`  
**Method:** POST  
**Description:** Change channel settings.

**Request Body:**
- `channel` (string): Channel information.
- `chaincode` (string): Chaincode information.

**Response:**
- `message` (string): Success message.

### 6. Query Product Overview

**Endpoint:** `/overview`  
**Method:** POST  
**Description:** Query product information.

**Request Body:**
- `productID` (string): ID of the product to query.

**Response:**
- `message` (string): Success message.
- `result` (string): Placeholder for the result of the query.

## Error Handling
In case of errors, the API will return a JSON response with an `error` field containing a descriptive error message. Ensure to handle errors appropriately in your client applications.

## Rate Limiting
The API does not currently implement rate limiting. Consider implementing rate limiting on the client-side to avoid abuse and improve security.

## Authentication
The API does not currently implement authentication mechanisms. If required, consider adding authentication, such as API keys or OAuth, for more secure access.

## File Uploads
When sending a request that involves file uploads, ensure that the file is properly included in the request body with the correct field name (`file` in this case).

---

Please note that this documentation assumes you are running the API on your local machine with the default port (3001). Adjust the base URL accordingly if you are running the API on a different server or port. Additionally, make sure to provide the necessary data in the request bodies when making API calls.
