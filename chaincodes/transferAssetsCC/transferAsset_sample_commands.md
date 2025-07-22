
./network.sh down
./network.sh up createChannel -ca -s couchdb
./network.sh deployCC -ccn private -ccp ../asset-transfer-private-data/chaincode-go/ -ccl go -ccep "OR('Org1MSP.peer','Org2MSP.peer')" -cccg ../asset-transfer-private-data/chaincode-go/collections_config.json
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name private --version 1.0 --collections-config ../asset-transfer-private-data/chaincode-go/collections_config.json --signature-policy "OR('Org1MSP.member','Org2MSP.member')" --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile $ORDERER_CA
peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name private --version 1.0 --sequence 1 --collections-config ../asset-transfer-private-data/chaincode-go/collections_config.json --signature-policy "OR('Org1MSP.member','Org2MSP.member')" --tls --cafile $ORDERER_CA --peerAddresses localhost:7051 --tlsRootCertFiles $ORG1_CA --peerAddresses localhost:9051 --tlsRootCertFiles $ORG2_CA
export PATH=${PWD}/../bin:${PWD}:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export FABRIC_CA_CLIENT_HOME=${PWD}/organizations/peerOrganizations/org1.example.com/
fabric-ca-client register --caname ca-org1 --id.name owner --id.secret ownerpw --id.type client --tls.certfiles "${PWD}/organizations/fabric-ca/org1/tls-cert.pem"
fabric-ca-client enroll -u https://owner:ownerpw@localhost:7054 --caname ca-org1 -M "${PWD}/organizations/peerOrganizations/org1.example.com/users/owner@org1.example.com/msp" --tls.certfiles "${PWD}/organizations/fabric-ca/org1/tls-cert.pem"
cp "${PWD}/organizations/peerOrganizations/org1.example.com/msp/config.yaml" "${PWD}/organizations/peerOrganizations/org1.example.com/users/owner@org1.example.com/msp/config.yaml"
export FABRIC_CA_CLIENT_HOME=${PWD}/organizations/peerOrganizations/org2.example.com/
fabric-ca-client register --caname ca-org2 --id.name buyer --id.secret buyerpw --id.type client --tls.certfiles "${PWD}/organizations/fabric-ca/org2/tls-cert.pem"
fabric-ca-client enroll -u https://buyer:buyerpw@localhost:8054 --caname ca-org2 -M "${PWD}/organizations/peerOrganizations/org2.example.com/users/buyer@org2.example.com/msp" --tls.certfiles "${PWD}/organizations/fabric-ca/org2/tls-cert.pem"
cp "${PWD}/organizations/peerOrganizations/org2.example.com/msp/config.yaml" "${PWD}/organizations/peerOrganizations/org2.example.com/users/buyer@org2.example.com/msp/config.yaml"

### switch to Org1
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/owner@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051


### GiveRights
export ASSET_PROPERTIES=$(echo -n "{\"ID\":\"RIGHTS\",\"Role\":\"Mine\",\"Collection\":\"Org1MSPPrivateCollection\"}" | base64 | tr -d \\n)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n private -c '{"function":"GiveRights","Args":[]}' --transient "{\"asset_properties\":\"$ASSET_PROPERTIES\"}"
export ASSET_PROPERTIES=$(echo -n "{\"ID\":\"RIGHTS\",\"Role\":\"OEM\",\"Collection\":\"Org2MSPPrivateCollection\"}" | base64 | tr -d \\n)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n private -c '{"function":"GiveRights","Args":[]}' --transient "{\"asset_properties\":\"$ASSET_PROPERTIES\"}"

### Create Recipe
export ASSET_PROPERTIES=$(echo -n "{\"recipeID\":\"R1\",\"Product\":\"product1\",\"Ingredients\":[\"battery1\",\"battery2\"],\"Quantity\":[1,1],\"Collection\":\"Org1MSPPrivateCollection\"}" | base64 | tr -d \\n)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n private -c '{"function":"CreateRecipe","Args":[]}' --transient "{\"asset_properties\":\"$ASSET_PROPERTIES\"}"

### Create Asset
export ASSET_PROPERTIES=$(echo -n "{\"assetName\":\"battery1\",\"assetID\":\"A0001\",\"GHG\":20}" | base64 | tr -d \\n)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n private -c '{"function":"CreateAssetIn","Args":[]}' --transient "{\"asset_properties\":\"$ASSET_PROPERTIES\"}"
export ASSET_PROPERTIES=$(echo -n "{\"assetName\":\"battery2\",\"assetID\":\"A0002\",\"GHG\":20}" | base64 | tr -d \\n)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n private -c '{"function":"CreateAssetIn","Args":[]}' --transient "{\"asset_properties\":\"$ASSET_PROPERTIES\"}"

### Manufacture new Asset
export ASSET_PROPERTIES=$(echo -n "{\"RecipeID\":\"R1\",\"assetName\":\"product1\",\"assetID\":\"A0003\",\"prodGHG\":50,\"Assets\":[\"A0001\",\"A0002\"]}" | base64 | tr -d \\n)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n private -c '{"function":"ManufactureAsset","Args":[]}' --transient "{\"asset_properties\":\"$ASSET_PROPERTIES\"}"

### Create new Shipping
export ASSET_PROPERTIES=$(echo -n "{\"shippingID\":\"S0001\",\"quantity\":1,\"list_ID\":[\"A0003\"],\"assetName\":\"product1\",\"date\":\"11-07-2023\",\"shipGHG\":20}" | base64 | tr -d \\n)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n private -c '{"function":"CreateShipping","Args":[]}' --transient "{\"asset_properties\":\"$ASSET_PROPERTIES\"}"

### Claim Shipping
export ASSET_PROPERTIES=$(echo -n "{\"shippingID\":\"S0001\",\"quantity\":1,\"list_ID\":[\"A0003\"],\"assetName\":\"product1\",\"date\":\"09-07-2023\",\"GHG\":110}" | base64 | tr -d \\n)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n private -c '{"function":"ClaimShipping","Args":[]}' --transient "{\"asset_properties\":\"$ASSET_PROPERTIES\"}"

### Create Recipe for Org2
export ASSET_PROPERTIES=$(echo -n "{\"recipeID\":\"R1\",\"Product\":\"FinalProduct\",\"Ingredients\":[\"product1\"],\"Quantity\":[2],\"Collection\":\"Org2MSPPrivateCollection\"}" | base64 | tr -d \\n)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n private -c '{"function":"CreateRecipe","Args":[]}' --transient "{\"asset_properties\":\"$ASSET_PROPERTIES\"}"

### Manufacture FinalProduct
export ASSET_PROPERTIES=$(echo -n "{\"RecipeID\":\"R1\",\"assetName\":\"product1\",\"assetID\":\"A0003\",\"prodGHG\":50,\"Assets\":[\"A0001\",\"A0002\"]}" | base64 | tr -d \\n)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n private -c '{"function":"FinalProduct","Args":[]}' --transient "{\"asset_properties\":\"$ASSET_PROPERTIES\"}"

### Read Assets
peer chaincode query -C mychannel -n private -c '{"function":"ReadRight","Args":["Org1MSPPrivateCollection"]}'
peer chaincode query -C mychannel -n private -c '{"function":"ReadPublicShipping","Args":["S0001"]}'
peer chaincode query -C mychannel -n private -c '{"function":"ReadPrivateAsset","Args":["Org1MSPPrivateCollection","A0001"]}'
peer chaincode query -C mychannel -n private -c '{"function":"ReadPrivateShipping","Args":["Org1MSPPrivateCollection","S0001"]}'
peer chaincode query -C mychannel -n private -c '{"function":"ReadRecipe","Args":["Org1MSPPrivateCollection","R1"]}'
peer chaincode query -C mychannel -n private -c '{"function":"ReadDelList","Args":[]}'

peer chaincode query -C mychannel -n private -c '{"function":"GetAllFlags","Args":["Org1MSPPrivateCollection"]}'
peer chaincode query -C mychannel -n private -c '{"function":"GetAllFlags","Args":["Org2MSPPrivateCollection"]}'
peer chaincode query -C mychannel -n private -c '{"function":"GetAllAssets","Args":["Org1MSPPrivateCollection"]}'
peer chaincode query -C mychannel -n private -c '{"function":"GetAllAssets","Args":["Org2MSPPrivateCollection"]}'
peer chaincode query -C mychannel -n private -c '{"function":"GetAllPrivateShippings","Args":["Org1MSPPrivateCollection"]}'
peer chaincode query -C mychannel -n private -c '{"function":"GetAllRecipes","Args":["Org1MSPPrivateCollection"]}'
peer chaincode query -C mychannel -n private -c '{"function":"GetAllRecipes","Args":["Org2MSPPrivateCollection"]}'
peer chaincode query -C mychannel -n private -c '{"function":"GetAllPublicShippings","Args":["shippingCollection"]}'

### Switch to Org2
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/buyer@org2.example.com/msp
export CORE_PEER_ADDRESS=localhost:9051

### Read Assets Org2
peer chaincode query -C mychannel -n private -c '{"function":"ReadPublicShipping","Args":["S0001"]}'
peer chaincode query -C mychannel -n private -c '{"function":"ReadPrivateAsset","Args":["Org2MSPPrivateCollection","A0001"]}'
peer chaincode query -C mychannel -n private -c '{"function":"ReadPrivateShipping","Args":["Org2MSPPrivateCollection","S0001"]}'
peer chaincode query -C mychannel -n private -c '{"function":"ReadRecipe","Args":["Org1MSPPrivateCollection","R1"]}'

export ASSET_PROPERTIES=$(echo -n "{\"shippingID\":\"A0002\",\"collection\":\"Org1MSPPrivateCollection\"}" | base64 | tr -d \\n)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n private -c '{"function":"DeleteAs","Args":[]}' --transient "{\"asset_properties\":\"$ASSET_PROPERTIES\"}"


