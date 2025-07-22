# Emissions Audit Chaincode

Chaincode implementation for auditing emissions and creating emission tokens.
Key aspects:
- Written as a GO module

## Design
- Writes an emissions record to the public ledger and the emissions record ID to the PCD of the invoking organization
- Performs a simple outlier detection on the emissions record. 

### Chaincode Functions
| Function | Description | Comment |
| --- | --- | --- |
| CreateEmissionsRecord(id string, kgCO2 int) |  Creates a new emissions record and stores it in the ledger. | *Should maybe not be public for production*
| CreateEmissionsRecordPrivateDetails(id string) | Adds new private emissions record details to the private data collection | Transient data: ownerID |
| GetEmissionsRecord(id string) | Returns the emissions record with the given ID. | 
| GetEmissionsRecordsList(ids []string) | Returns a list of emissions records with the given IDs. | |
| GetAllEmissionsRecords() | Returns all emissions records in the ledger. | *ONLY FOR TESTING*
| GetEmissionsRecordsOfOwner() | Returns all emissions records owned by the given owner. | Transient data: ownerID |
| GetEmissionsRecordPrivateDetails(recordID string) | Returns the private emissions record details of the given emissions record. |  | 
| GetAllEmissionsRecordsPrivateDetails() | Returns all private emissions record details in the pdc. | *ONLY FOR TESTING* |
| EmissionsRecordExists(id string) | Returns true if an emissions record with the given ID exists in the ledger. |
| AuditEmissions(id string, prevEmissionsIDs []string, kgCO2 int, info string) | Main function for auditing emissions. Checks inout emissions agains previous emissions of the particular owner and then creates a new emissions record and stores it in the ledger. 

### Chaincode Access Control
- EmissionsRecords can be created by any member of the channel due to an endorsement policy which only requires an endorsement from one channel-member. *Not Implemented yet(How?)*

## Development Instructions
Managing Dependencies:
To ensure that all modules and dependencies are properly installed (e.g. with `peer chaincode package` and  `peer chaincode install`) run the following command:
```bash
go mod tidy
go mod vendor
```
To deploy the chaincode to the test network run the following command:
**Update the path in the following commands depending on the location of your test-network**
TODO: The endorsement policy should be updated to allow more generic access to the chaincode (create an asset without receiving an endorsement from the other organization)
```bash
./network.sh up createChannel -s couchdb
```
```bash
./network.sh deployCC -ccn emissionsAudit -ccp ../../sustainable-supply-chain/chaincode/emissionsAudit -ccl go -ccep "OR('Org1MSP.peer','Org2MSP.peer')" -cccg ../../sustainable-supply-chain/chaincode/emissionsAudit/collections_config.json
```
```bash
peer chaincode invoke -o localhost:7050 -C mychannel -n emissionsAudit -c '{"function":"AuditEmissions","Args":["id1", "owner1", "99", "info string"]}' --transient "{\"ownerID\":\"$OWNER_ID\"}"
```


### Alternative with CA: 
```bash
./network.sh up createChannel -ca -s couchdb
```
```bash
./network.sh deployCC -ccn emissionsAudit -ccp ../../sustainable-supply-chain/chaincode/emissionsAudit -ccl go -ccep "OR('Org1MSP.peer','Org2MSP.peer')" -cccg ../../sustainable-supply-chain/chaincode/collections_config.json
```
Register Identities
```bash
export PATH=${PWD}/../bin:${PWD}:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export FABRIC_CA_CLIENT_HOME=${PWD}/organizations/peerOrganizations/org1.example.com/
```
```bash
fabric-ca-client register --caname ca-org1 --id.name owner --id.secret ownerpw --id.type client --tls.certfiles "${PWD}/organizations/fabric-ca/org1/tls-cert.pem"
```
```bash
fabric-ca-client enroll -u https://owner:ownerpw@localhost:7054 --caname ca-org1 -M "${PWD}/organizations/peerOrganizations/org1.example.com/users/owner@org1.example.com/msp" --tls.certfiles "${PWD}/organizations/fabric-ca/org1/tls-cert.pem"
```
```bash
cp "${PWD}/organizations/peerOrganizations/org1.example.com/msp/config.yaml" "${PWD}/organizations/peerOrganizations/org1.example.com/users/owner@org1.example.com/msp/config.yaml"
```
```bash
export FABRIC_CA_CLIENT_HOME=${PWD}/organizations/peerOrganizations/org2.example.com/
```
```bash
fabric-ca-client register --caname ca-org2 --id.name buyer --id.secret buyerpw --id.type client --tls.certfiles "${PWD}/organizations/fabric-ca/org2/tls-cert.pem"
```
```bash
fabric-ca-client enroll -u https://buyer:buyerpw@localhost:8054 --caname ca-org2 -M "${PWD}/organizations/peerOrganizations/org2.example.com/users/buyer@org2.example.com/msp" --tls.certfiles "${PWD}/organizations/fabric-ca/org2/tls-cert.pem"
```
```bash
cp "${PWD}/organizations/peerOrganizations/org2.example.com/msp/config.yaml" "${PWD}/organizations/peerOrganizations/org2.example.com/users/buyer@org2.example.com/msp/config.yaml"
```

### Audit emissions
```bash
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/owner@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
```
```bash
export OWNER_ID=$(echo -n "ownerID1" | base64 | tr -d \\n)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n emissionsAudit -c '{"function":"AuditEmissions","Args":["id1", "[]", "99", "info string"]}' --transient "{\"ownerID\":\"$OWNER_ID\"}"
```

### Query public ledger
```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n emissionsAudit -c '{"function":"GetAllEmissionsRecords","Args":[]}' 
```
### Query private data collection
Query private data collection by id
```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n emissionsAudit -c '{"function":"AllEmissionsRecordPrivateDetails","Args":["id1"]}'
```
Query private data collection by owner
```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n emissionsAudit -c '{"function":"GetEmissionsRecordsOfOwner","Args":[]}' --transient "{\"ownerID\":\"$OWNER_ID\"}"
```
Query entire private data collection
```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n emissionsAudit -c '{"function":"GetAllEmissionsRecordsPrivateDetails","Args":[]}'
```

Create private data element
```bash
 peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n emissionsAudit -c '{"function":"CreateEmissionsRecordPrivateDetails","Args":["id1"]}' --transient "{\"ownerID\":\"$OWNER_ID\"}"
 ```