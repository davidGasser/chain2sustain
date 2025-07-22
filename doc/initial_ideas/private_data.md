# Ideas on how to handle private data

## Potential starting points and resources
- Medium Article - Channel vs. private data in Fabric [link](https://kctheservant.medium.com/data-privacy-among-organizations-channel-and-private-data-in-hyperledger-fabric-ee268cd44916) 
- Fabric Doc on [private data](https://hyperledger-fabric.readthedocs.io/en/release-1.4/private-data/private-data.html)
- Fabric Tutorial on [private data](https://hyperledger-fabric.readthedocs.io/en/release-1.4/private_data_tutorial.html)

### A example supply chain based utilizing private channels
- [Introduction](https://medium.com/@abhinav.garg_90821/hyperledger-chapter-5-tuna-fishing-supplychain-context-553d4a0be5a) to the use-case (Tuna supply chain) and a system design overview
- Explenation and tutorial of the [chaincode](https://medium.com/@abhinav.garg_90821/hyperledger-chapter-9-chaincode-in-tunafish-scenario-using-hyperledger-fabric-1fdd2a87cb96) used in the tuna supply chain example
- [Tutorial](https://medium.com/@abhinav.garg_90821/hyperledger-chapter-10-blockchain-application-on-hyperledger-fabric-6e40de190512) on how to build and deploy an application for the tuna supply chain network example

## Pros and Cons
Using Channels:
- good when privacy is at chaincode/application level
| Pro | Con |
-------------
| greater flexibility for chaincode  | cross channel chaincode invocation complex   |
|   |    |
- Idea: Inheritance between different chaincode levels
- Cross-chaincode invocation requires both chaincodes to be installed on the same peer
  - Querrying possible from other channel
  - Updating only possible from within the same channel
  - [Medium Article](https://kctheservant.medium.com/cross-chaincode-invoking-in-hyperledger-fabric-8b8df1183c04) 
  - 
Using Private Collections:
- good when privacy is at data level
- adding a organisation to a collection requires upgrade of the chaincode
| Pro | Con |
-------------
| scalability  | requires one-fits-all chaincode  |
|   |   |