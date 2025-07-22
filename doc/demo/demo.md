

# Demo function calls for presentation


The following scripts provide an exemplary demonstration of the various calls used in our supply chain project. These calls effectively showcase the key functionalities and features of our project, offering a comprehensive understanding of its capabilities.



### Making Org1 Mine
```sh
echo ASSET_PROPERTIES=$(echo -n "{\"ID\":\"RIGHTS\",\"Role\":\"Mine\",\"Collection\":\"Org2MSPPrivateCollection\"}" | base64 | tr -d \\n)
```
```sh
docker exec cli.org1.example.com peer chaincode invoke -C my-channel1 -n transferAssets --peerAddresses peer0.org1.example.com:7041 -c '{"Args":["GiveRights"]}' --transient "{\"asset_properties\": \"eyJJRCI6IlJJR0hUUyIsIlJvbGUiOiJNaW5lIiwiQ29sbGVjdGlvbiI6Ik9yZzJNU1BQcml2YXRlQ29sbGVjdGlvbiJ9\"}"
```

## Create recipe from org2 -------------------- CreateRecipe
```sh
echo ASSET_PROPERTIES=$(echo -n "{\"recipeID\":\"R1\",\"Product\":\"product1\",\"Ingredients\":[\"battery1\",\"battery2\"],\"Quantity\":[1,1],\"Collection\":\"Org2MSPPrivateCollection\"}" | base64 | tr -d \\n)
```

```sh
docker exec cli.org2.example.com peer chaincode invoke -C my-channel1 -n transferAssets --peerAddresses peer0.org2.example.com:7061 -c '{"Args":["CreateRecipe"]}' --transient "{\"asset_properties\": \"eyJyZWNpcGVJRCI6IlIxIiwiUHJvZHVjdCI6InByb2R1Y3QxIiwiSW5ncmVkaWVudHMiOlsiYmF0dGVyeTEiLCJiYXR0ZXJ5MiJdLCJRdWFudGl0eSI6WzEsMV0sIkNvbGxlY3Rpb24iOiJPcmcyTVNQUHJpdmF0ZUNvbGxlY3Rpb24ifQ==\"}"
```

## Create asset from org2 -------------------- CreateAssetIn
```sh
echo ASSET_PROPERTIES=$(echo -n "{\"assetName\":\"battery1\",\"assetID\":\"A0001\",\"emissionsIDs\":[\"Emission1\",\"Emission2\"]}" | base64 | tr -d \\n)
```
```sh
docker exec cli.org2.example.com peer chaincode invoke -C my-channel1 -n transferAssets --peerAddresses peer0.org2.example.com:7061 -c '{"Args":["CreateAssetIn"]}' --transient "{\"asset_properties\": \"eyJhc3NldE5hbWUiOiJiYXR0ZXJ5MSIsImFzc2V0SUQiOiJBMDAwMSIsImVtaXNzaW9uc0lEcyI6WyJFbWlzc2lvbjEiLCJFbWlzc2lvbjIiXX0=\"}"
```

```sh
echo ASSET_PROPERTIES=$(echo -n "{\"assetName\":\"battery2\",\"assetID\":\"A0002\",\"emissionsIDs\":[\"Emission1\",\"Emission2\"]}" | base64 | tr -d \\n)
```


```sh
docker exec cli.org2.example.com peer chaincode invoke -C my-channel1 -n transferAssets --peerAddresses peer0.org2.example.com:7061 -c '{"Args":["CreateAssetIn"]}' --transient "{\"asset_properties\": \"eyJhc3NldE5hbWUiOiJiYXR0ZXJ5MiIsImFzc2V0SUQiOiJBMDAwMiIsImVtaXNzaW9uc0lEcyI6WyJFbWlzc2lvbjEiLCJFbWlzc2lvbjIiXX0=\"}"

```

## Read created asset 
```sh
docker exec cli.org1.example.com peer chaincode query -C my-channel1 -n transferAssets -c '{"Args":["ReadPrivateAsset","Org2MSPPrivateCollection","A0001"]}'
```

```sh
echo ASSET_PROPERTIES=$(echo -n "{\"recipeID\":\"R1\",\"assetName\":\"product1\",\"assetID\":\"A0003\",\"emissionsIDs\":[\"Emission1\",\"Emission2\"],\"assets\":[\"A0001\",\"A0002\"]}" | base64 | tr -d \\n)
```

```sh
docker exec cli.org2.example.com peer chaincode invoke -C my-channel1 -n transferAssets --peerAddresses peer0.org2.example.com:7061 -c '{"Args":["CreateAssetIn"]}' --transient "{\"asset_properties\": \"eyJyZWNpcGVJRCI6IlIxIiwiYXNzZXROYW1lIjoicHJvZHVjdDEiLCJhc3NldElEIjoiQTAwMDMiLCJlbWlzc2lvbnNJRHMiOlsiRW1pc3Npb24xIiwiRW1pc3Npb24yIl0sImFzc2V0cyI6WyJBMDAwMSIsIkEwMDAyIl19\"}"
```


### Create shiping for org2 CreateShipping
```sh
 echo ASSET_PROPERTIES=$(echo -n "{\"shippingID\":\"S1\",\"quantity\":2,\"list_ID\":[\"Emission1\",\"Emission2\"],\"assetName\":\"product1\",\"date\":\"11-07-2023\",\"shipEmissionsIDs\":[\"Emission3\",\"Emission4\"]}" | base64 | tr -d \\n)
 ```


```sh
 docker exec cli.org2.example.com peer chaincode invoke -C my-channel1 -n transferAssets --peerAddresses peer0.org2.example.com:7061 -c '{"Args":["CreateShipping"]}' --transient "{\"asset_properties\":\"eyJzaGlwcGluZ0lEIjoiUzEiLCJxdWFudGl0eSI6MiwibGlzdF9JRCI6WyJFbWlzc2lvbjEiLCJFbWlzc2lvbjIiXSwiYXNzZXROYW1lIjoicHJvZHVjdDEiLCJkYXRlIjoiMTEtMDctMjAyMyIsInNoaXBFbWlzc2lvbnNJRHMiOlsiRW1pc3Npb24zIiwiRW1pc3Npb240Il19\"}"
 ```