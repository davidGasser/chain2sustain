

# Hyperledger Fablo Usage
This is a general guide on how to interact with the fablo network. For specific deployment instructions of the supply chain network see the `README.md`

In order to use Hyperledger Fablo, we will need to download the necessary scripts, initialize the setup, and perform various operations. The following steps will guide you through the process:

## Prerequisites
In System Preferences in Macbook click Sequrity & Privacy Settings in Privacy Tab enable Docker and Terminal for Full Disk Access. 
Open Docker Desktop Click the tree dots in Features in Development tab click Experimental Features disable accessing experimental features. 

## Step 1: Download and Setup

First, we need to download and setup Fablo. Use the following command to download the script from the GitHub repository and make it executable 

```bash
curl -Lf https://github.com/hyperledger-labs/fablo/releases/download/1.1.0/fablo.sh -o ./fablo
chmod +x ./fablo
```

## Step 2: Initialization

Start Docker

After setup, initialize the Fablo nodes:

```bash
./fablo init rest
```

## Step 3: Starting Fablo

To start Fablo, use the following command:

```bash
./fablo up
```

Or to start Fablo with a specific configuration file, use:

```bash
./fablo up /path/to/fablo-config.json
```

- An example.yaml provided in this folder we can use this yaml forcreating network with specified configuration with: 
```bash
./fablo up /path/to/example.yaml
```

## Step 4: Install Chaincodes

If you need to install chaincodes, use the following command:

```bash
fablo chaincodes install
```

## Step 5: Install chain code to peers
```bash
fablo chaincode install chaincode1 0.0.1
```

## Step 6: call chaincode in a peer:
### Edit params and function name part when you want to call another method
```bash
docker exec cli.org1.example.com peer chaincode invoke -C my-channel1 -n chaincode1 --peerAddresses peer0.org1.example.com:7041 -c '{"Args":["KVContract:put","arg1","arg2"]}'
```


## Other Operations 
### Note these opperations are not required to start chaincode

To stop, start, or clean up the Fablo environment, you can use the following commands:

```bash
fablo down
fablo start
fablo stop
fablo prune
```

## List Channels
### Note these opperations are not required to start chaincode

To list all channels that the peer has joined:

```bash
fablo channel list org1 peer0
```

Please replace "org1" and "peer0" with the names of your organization and peer if they are different.

Please note, each command should be executed from the directory where the Fablo script is located.
