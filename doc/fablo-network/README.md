

# Hyperledger Fablo Usage

In order to use Hyperledger Fablo, we will need to download the necessary scripts, initialize the setup, and perform various operations. The following steps will guide you through the process:

## Step 1: Download and Setup

**All Commands need to be excuted in the top level directory of this repository!**

First, we need to download and setup Fablo. Use the following command to download the script from the GitHub repository and make it executable 

```bash
curl -Lf https://github.com/hyperledger-labs/fablo/releases/download/1.1.0/fablo.sh -o ./fablo
chmod +x ./fablo
```

## Step 2: Initialization

Start Docker

After setup, initialize the Fablo nodes:

```bash
./fablo init 
```
The init function autmatically creates a new `fablo_config.json` file in the execution directory. However, since we provide our own config file, the newly created on is not needed anymore.  

## Step 3: Starting Fablo

To start Fablo using our preset network configuration with 3 Organisations, use the following command:

```bash
./fablo up fablo_config.yaml
```
This deploys a range of docker containers, installes the chaincode and joins organisations to the channel.

## Check deployment
To check if the network is deployed correctly, use the following command to list all installed channels.

```bash
./fablo channel list org1 peer0
```
The responde should be `my-channel1`

## Step 4: Invoking Chaincode:
To invoke chaincode directly the peer srcipt can be used
```bash
docker exec cli.org1.example.com peer chaincode invoke -C my-channel1 -n transferAssets --peerAddresses peer0.org1.example.com:7041 -c '{"Args":["GetAllAssets","Org1MSPPrivateCollection"]}'
```

## Stop the network
To stop the network use the following command:
```bash
./fablo down
```

To remove all automatically created config and crypto files use the following command:
```bash
./fablo prune
```

## Troubleshoting
If issues with MacOS occure the Security & Privacy Settings for docker might need to be adapted.
Therefor go to 
- >`System Preferences`
- >`Security & Privacy` Settings, 
- >`Privacy Tab` 
- enable Docker and Terminal for Full Disk Access

Open Docker Desktop and Click the tree dots in Features in the Development tab. Click "Experimental Features" disable to access  experimental features. 