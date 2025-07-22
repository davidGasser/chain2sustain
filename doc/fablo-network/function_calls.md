# Some function calls for EmmisionAudit Smart Contract while fablo network is up.
- CreateEmissionsRecord
```sh
docker exec cli.org1.example.com peer chaincode invoke -C my-channel1 -n channel1 --peerAddresses peer0.org1.example.com:7041 -c '{"Args":["CreateEmissionsRecord","record5", "100"]}'
```
- CreateEmissionsRecordPrivateDetails
```sh
docker exec cli.org1.example.com peer chaincode invoke -C my-channel1 -n channel1 --peerAddresses peer0.org1.example.com:7041 --transient "{\"ownerID\": \"YmFzZTY0IGVuY29kZWQgc3RyaW5n\"}" -c '{"Args":["CreateEmissionsRecordPrivateDetails","record1"]}'
```

- GetEmissionsRecord
```sh
docker exec cli.org1.example.com peer chaincode invoke -C my-channel1 -n channel1 --peerAddresses peer0.org1.example.com:7041 -c '{"Args":["GetEmissionsRecord","record1"]}'
```

- GetEmissionsRecordsList
```sh
docker exec cli.org1.example.com peer chaincode invoke -C my-channel1 -n channel1 --peerAddresses peer0.org1.example.com:7041 -c '{"Args":["GetEmissionsRecordsList", "[\"record1\"]"]}'
```

- GetAllEmissionsRecords
```sh
docker exec cli.org1.example.com peer chaincode invoke -C my-channel1 -n channel1 --peerAddresses peer0.org1.example.com:7041 -c '{"Args":["GetAllEmissionsRecords"]}'
```

- Get All Emissions Records note we should swap network to CauchDb to do this operation.
```sh
docker exec cli.org1.example.com peer chaincode invoke -C my-channel1 -n channel1 --peerAddresses peer0.org1.example.com:7041 -c '{"Args":["GetAllEmissionsRecords"]}'
```


- Get Emissions Record Private Details
```sh
docker exec cli.org1.example.com peer chaincode invoke -C my-channel1 -n channel1 --peerAddresses peer0.org1.example.com:7041 -c '{"Args":["GetEmissionsRecordPrivateDetails","record1"]}'
```

- Get All Emissions Records Private Details
```sh
docker exec cli.org1.example.com peer chaincode invoke -C my-channel1 -n channel1 --peerAddresses peer0.org1.example.com:7041 -c '{"Args":["GetAllEmissionsRecordsPrivateDetails"]}'
```

- Emissions Record Exists
```sh
docker exec cli.org1.example.com peer chaincode invoke -C my-channel1 -n channel1 --peerAddresses peer0.org1.example.com:7041 -c '{"Args":["EmissionsRecordExists","record1"]}'
```

- Audit Emissions
```sh
docker exec cli.org1.example.com peer chaincode invoke -C my-channel1 -n channel1 --peerAddresses peer0.org1.example.com:7041 --transient "{\"ownerID\": \"YmFzZTY0IGVuY29kZWQgc3RyaW5n\"}" -c '{"Args":["AuditEmissions","record53", "[]", "100", "info"]}'
```
