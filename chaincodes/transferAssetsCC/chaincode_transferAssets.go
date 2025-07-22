package main

import (
	//"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"encoding/hex"
	"crypto/sha256"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	//"reflect"
)

const shippingCollection = "shippingCollection"
const transferAgreementObjectType = "transferAgreement"

// SmartContract of this fabric sample
type SmartContract struct {
	contractapi.Contract
}

type ShippingPublic struct {
	ID 			string `json:"shippingID"`
	SellerID 	string `json:"sellerID"`
	Name 		string `json:"assetName"`
}

type ShippingPrivate struct {
	ID 			string `json:"shippingID"`
	Quantity 	int `json:"quantity"`
	List_ID 	[]string `json:"list_ID"`
	Name		string `json:"assetName"`
	Date 		string `json:"date"`
	EmissionsIDs [][]string `json:"emissionsIDs"`
}

type Asset struct {
	Name 			string `json:"assetName"`
	ID 				string `json:"assetID"`
	EmissionsIDs 	[]string `json:"emissionsIDs"`
	Dir 			string `json:"Direction"`
}

type PublicAsset struct {
	ID 				string `json:"assetID"`
	EmissionsIDs 	[]string `json:"emissionsIDs"`
	BasedOn			[]string `json:"BasedOn"`
}

type FinalAsset struct {
	ID 				string `json:"assetID"`
	EmissionsIDs 	[]string `json:"emissionsIDs"`
	GHG			 	int `json:"GHG"`
	BasedOn			[]string `json:"BasedOn"`
}

type Recipe struct {
	ID 			string `json:"recipeID"`
	Product		string `json:"product"`
	Ingredients []string `json:"ingredients"`
	Quantity 	[]int `json:"quantity"`
}

type DeletionShippingList struct {
	ID 			string `json:"ID"`
	Del_List	[]string `json:"Del_List"`
}

type Flag struct {
	ID 		string `json:"ID"`
	Date	string `json:"Date"`
	Mesg	string `json:"Mesg"`
}

type Rights struct {
	ID 		string `json:"ID"`
	Role 	string `json:"Role"`
}



func (s *SmartContract) GiveRights (ctx contractapi.TransactionContextInterface) error {

	err := deleteClaimedShippments(ctx)
	if err != nil{
		return fmt.Errorf("Failure with deleting claimed shipment: %v", err)
	}

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		// log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}

	type rightsTransient struct {
		ID 			string `json:"ID"`
		Role 		string `json:"Role"`
		Collection	string `json:"Collection"`
	}

	var rightsInput rightsTransient
	err = json.Unmarshal(transientAssetJSON, &rightsInput)
	//check the user's input
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	if rightsInput.ID != "RIGHTS" {
		return fmt.Errorf("ID must be adjusted to 'RIGHTS'")
	}
	if len(rightsInput.Role) == 0 {
		return fmt.Errorf("Role field must be a non-empty string")
	}
	if len(rightsInput.Collection) == 0 {
		return fmt.Errorf("Collection field must be a non-empty string")
	}


	// Get ID of submitting client identity
	clientID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if clientID != "0"{clientID= "0"}

	// Verify that the client is submitting request to peer in their organization
	// This is to ensure that a client from another org doesn't attempt to read or
	// write private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateAsset cannot be performed: Error %v", err)
	}

	//get clientMSPID
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get verified MSPID: %v", err)
	}
	//if it isn't Org1 create a flag
	if clientMSPID != "Org1MSP"{
		
		//get Owner Private Collection
		orgCollection, err := getCollectionName(ctx) // get owner collection from caller identity
		if err != nil {
			return fmt.Errorf("failed to infer private collection name for the org: %v", err)
		}
		//Find ID for the flag
		count:= 0
		for{
			flagAsBytes, err := ctx.GetStub().GetPrivateData(orgCollection, fmt.Sprintf("%s%d","F",count))
			if err != nil {
				return fmt.Errorf("failed to get flag: %v", err)
			} else if flagAsBytes == nil {
				break
			}
			count += 1
		}

		//get current Time and Date
		time_now := time.Now()

		flag := Flag{
			ID: 	fmt.Sprintf("%s%d","F",count),
			Date: 	time_now.String(),
			Mesg: 	"Unauthorized attempt at invoking GiveRights chaincode",
		}

		flagJSONasBytes, err := json.Marshal(flag)
		if err != nil {
			return fmt.Errorf("failed to marshal asset into JSON: %v", err)
		}
		log.Printf("Unauthorized attempt of access: function GiveRights")
		err = ctx.GetStub().PutPrivateData(orgCollection, flag.ID, flagJSONasBytes)
		if err != nil {
			return fmt.Errorf("failed to put flag into private data collecton: %v", err)
		}
		return nil
	}

	//Check that the Role is correct
	if rightsInput.Role != "Supplier" && rightsInput.Role != "Mine" && rightsInput.Role != "OEM"{
		return fmt.Errorf("Role does not fit any of the predefined ones [Mine,Supplier,OEM]")
	}

	//create the rights struct
	right := Rights{
		ID:    		rightsInput.ID,
		Role: 		rightsInput.Role,
	}

	rightJSONasBytes, err := json.Marshal(right)
	if err != nil {
		return fmt.Errorf("failed to marshal right into JSON: %v", err)
	}

	log.Printf("GiveRights Put: collection %v, ID %v", rightsInput.Collection, rightsInput.ID)

	err = ctx.GetStub().PutPrivateData(rightsInput.Collection, rightsInput.ID, rightJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}
	return nil
}


func (s *SmartContract) CreateAssetIn(ctx contractapi.TransactionContextInterface) error {

	err := deleteClaimedShippments(ctx)
	if err != nil{
		return fmt.Errorf("Failure with deleting claimed shipment: %v", err)
	}
	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		// log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}

	type assetTransient struct {
		Name 	string `json:"assetName"`
		ID 		string `json:"assetID"`
		EmissionsIDs []string `json:"emissionsIDs"`
	}

	var assetInput assetTransient
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetInput.Name) == 0 {
		return fmt.Errorf("Name field must be a non-empty string")
	}
	if len(assetInput.ID) == 0 {
		return fmt.Errorf("assetID field must be a non-empty string")
	}
	if len(assetInput.EmissionsIDs) <= 0 {
		return fmt.Errorf("EmissionsIDs field must be a non-empty, list of IDs")
	}


	// Get ID of submitting client identity
	clientID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if clientID != "0"{clientID= "0"}

	// Verify that the client is submitting request to peer in their organization
	// This is to ensure that a client from another org doesn't attempt to read or
	// write private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateAsset cannot be performed: Error %v", err)
	}

	//get Owner Private Collection
	orgCollection, err := getCollectionName(ctx) // get owner collection from caller identity
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	//check if the org has the right to create an asset
	rightsDetailsJSON, err := ctx.GetStub().GetPrivateData(orgCollection, "RIGHTS")
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if rightsDetailsJSON == nil {
		return fmt.Errorf("Need to receive rights to use this function")
	}

	var right *Rights
	err = json.Unmarshal(rightsDetailsJSON, &right)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	// if the Org is not a Mine create flag
	if right.Role != "Mine" || assetInput.ID =="RIGHTS"{
		//Find ID for the flag
		count:= 0
		for{
			flagAsBytes, err := ctx.GetStub().GetPrivateData(orgCollection, fmt.Sprintf("%s%d","F",count))
			if err != nil {
				return fmt.Errorf("failed to get flag: %v", err)
			} else if flagAsBytes == nil {
				break
			}
			count += 1
		}

		//get current Time and Date
		time_now := time.Now()

		if assetInput.ID =="RIGHTS"{
			flag := Flag{
				ID: 	fmt.Sprintf("%s%d","F",count),
				Date: 	time_now.String(),
				Mesg: 	"GRAVE: tried to set ID to RIGHTS",
			}
	
			flagJSONasBytes, err := json.Marshal(flag)
			if err != nil {
				return fmt.Errorf("failed to marshal flag into JSON: %v", err)
			}
			log.Printf("GRAVE: not allowed to set ID to RIGHTS")
			err = ctx.GetStub().PutPrivateData(orgCollection, flag.ID, flagJSONasBytes)
			if err != nil {
				return fmt.Errorf("failed to put flag into private data collecton: %v", err)
			}
			return nil
		} else{
			flag := Flag{
				ID: 	fmt.Sprintf("%s%d","F",count),
				Date: 	time_now.String(),
				Mesg: 	"Unauthorized attempt at invoking CreateAssetIn chaincode",
			}

			flagJSONasBytes, err := json.Marshal(flag)
			if err != nil {
				return fmt.Errorf("failed to marshal flag into JSON: %v", err)
			}
			log.Printf("Unauthorized attempt of access: function CreateAssetIn")
			err = ctx.GetStub().PutPrivateData(orgCollection, flag.ID, flagJSONasBytes)
			if err != nil {
				return fmt.Errorf("failed to put flag into private data collecton: %v", err)
			}
			return nil
		}
	}

	// Check if asset already exists
	assetAsBytes, err := ctx.GetStub().GetState(assetInput.ID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if assetAsBytes != nil {
		fmt.Println("Asset already exists: " + assetInput.ID)
		return fmt.Errorf("this asset already exists: " + assetInput.ID)
	}

	//create the public asset for tracking
	publicAsset := PublicAsset{
		ID: 			assetInput.ID,
		EmissionsIDs: 	assetInput.EmissionsIDs,
		BasedOn: 		[]string{"nil"},

	}
	publicAssetJSONasBytes, err := json.Marshal(publicAsset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}
	log.Printf("CreateAsset Put Public Asset: ID %v, EmissionsIDs %v", assetInput.ID, assetInput.EmissionsIDs)

	err = ctx.GetStub().PutState(assetInput.ID, publicAssetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}
	
	// Mark private Asset as incoming
	asset := Asset{
		Name:  	assetInput.Name,
		ID:    	assetInput.ID,
		EmissionsIDs: 	assetInput.EmissionsIDs,
		Dir: 	"in",
	}
	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}


	log.Printf("CreateAsset Put: collection %v, ID %v, EmissionsIDs %v", orgCollection, assetInput.ID, assetInput.EmissionsIDs)

	err = ctx.GetStub().PutPrivateData(orgCollection, assetInput.ID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}
	return nil
}


//CreateRecipe defines the recipe for manufacturing and safes it in the Private Collection
func (s *SmartContract) CreateRecipe(ctx contractapi.TransactionContextInterface) error {

	err := deleteClaimedShippments(ctx)
	if err != nil{
		return fmt.Errorf("Failure with deleting claimed shipment: %v", err)
	}

	// Get new recipe from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Recipe properties are private, therefore they get passed in transient field, instead of func args
	transientRecipeJSON, ok := transientMap["asset_properties"]
	if !ok {
		// log error to stdout
		return fmt.Errorf("recipe not found in the transient map input")
	}

	type recipeTransient struct {
		ID 			string 	`json:"recipeID"`
		Product		string 	`json:"Product"`
		Ingredients []string `json:"Ingredients"`
		Quantity 	[]int `json:"Quantity"`
		Collection 	string `json:"Collection"`
	}

	var recipeInput recipeTransient
	err = json.Unmarshal(transientRecipeJSON, &recipeInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(recipeInput.ID) == 0 {
		return fmt.Errorf("Name field must be a non-empty string")
	}
	if len(recipeInput.Product) == 0 {
		return fmt.Errorf("Product field must be a non-empty string")
	}
	if len(recipeInput.Ingredients) == 0 {
		return fmt.Errorf("Name field must be a non-empty string")
	}
	if len(recipeInput.Quantity) == 0 {
		return fmt.Errorf("recipeID field must be a non-empty string")
	}
	if len(recipeInput.Quantity) != len(recipeInput.Ingredients){
		return fmt.Errorf("Lists parameters must have the same length")
	}

	// Get ID of submitting client identity
	clientID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if clientID != "0"{clientID= "0"}

	// Verify that the client is submitting request to peer in their organization
	// This is to ensure that a client from another org doesn't attempt to read or
	// write private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateRecipe cannot be performed: Error %v", err)
	}

	//check if client is allowed to access this function. If not create flag
	//get clientMSPID
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get verified MSPID: %v", err)
	}
	//if it isn't Org1 create a flag
	if clientMSPID != "Org1MSP"{
		
		//get Owner Private Collection
		orgCollection, err := getCollectionName(ctx) // get owner collection from caller identity
		if err != nil {
			return fmt.Errorf("failed to infer private collection name for the org: %v", err)
		}
		//Find ID for the flag
		count:= 0
		for{
			flagAsBytes, err := ctx.GetStub().GetPrivateData(orgCollection, fmt.Sprintf("%s%d","F",count))
			if err != nil {
				return fmt.Errorf("failed to get flag: %v", err)
			} else if flagAsBytes == nil {
				break
			}
			count += 1
		}

		//get current Time and Date
		time_now := time.Now()

		flag := Flag{
			ID: 	fmt.Sprintf("%s%d","F",count),
			Date: 	time_now.String(),
			Mesg: 	"Unauthorized attempt at invoking CreateRecipe chaincode",
		}

		flagJSONasBytes, err := json.Marshal(flag)
		if err != nil {
			return fmt.Errorf("failed to marshal asset into JSON: %v", err)
		}
		log.Printf("Unauthorized attempt of access: function GiveRights")
		err = ctx.GetStub().PutPrivateData(orgCollection, flag.ID, flagJSONasBytes)
		if err != nil {
			return fmt.Errorf("failed to put flag into private data collecton: %v", err)
		}
		return nil
	}

	// Check if recipe already exists
	recipeAsBytes, err := ctx.GetStub().GetPrivateData(recipeInput.Collection, recipeInput.ID)
	if err != nil {
		return fmt.Errorf("failed to get recipe: %v", err)
	} else if recipeAsBytes != nil {
		fmt.Println("Recipe already exists: " + recipeInput.ID)
		return fmt.Errorf("this recipe already exists: " + recipeInput.ID)
	}


	recipe := Recipe{
		ID:    			recipeInput.ID,
		Product:		recipeInput.Product,
		Ingredients: 	recipeInput.Ingredients,
		Quantity: 		recipeInput.Quantity,
	}
	recipeJSONasBytes, err := json.Marshal(recipe)
	if err != nil {
		return fmt.Errorf("failed to marshal recipe into JSON: %v", err)
	}


	log.Printf("CreateRecipe Put: collection %v, ID %v", recipeInput.Collection, recipeInput.ID)

	err = ctx.GetStub().PutPrivateData(recipeInput.Collection, recipeInput.ID, recipeJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put recipe into private data collecton: %v", err)
	}
	return nil
}


func (s *SmartContract) ManufactureAsset(ctx contractapi.TransactionContextInterface) error {

	err := deleteClaimedShippments(ctx)
	if err != nil{
		return fmt.Errorf("Failure with deleting claimed shipment: %v", err)
	}

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		// log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}

	type dataTransient struct {
		RecipeID 	string `json:"recipeID"`
		Name 		string `json:"assetName"`
		ID 			string `json:"assetID"`
		EmissionsIDs []string `json:"emissionsIDs"`
		Assets  	[]string `json:"assets"`
	}

	//get data and check it 
	var dataInput dataTransient
	err = json.Unmarshal(transientAssetJSON, &dataInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	if len(dataInput.RecipeID) == 0 {
		return fmt.Errorf("RecipeID field must be a non-empty string")
	}
	if len(dataInput.Name) == 0 {
		return fmt.Errorf("Name field must be a non-empty string")
	}
	if len(dataInput.ID) == 0 {
		return fmt.Errorf("assetID field must be a non-empty string")
	}
	if len(dataInput.EmissionsIDs) <= 0 {
		return fmt.Errorf("Length of List of EmissionsIDs field must be greater than 0 %v",dataInput.EmissionsIDs)
	}
	if len(dataInput.Assets) == 0 {
		return fmt.Errorf("Assets slice must be a non-empty")
	}


	// Get ID of submitting client identity
	clientID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if clientID != "0"{clientID= "0"}

	// Verify that the client is submitting request to peer in their organization
	// This is to ensure that a client from another org doesn't attempt to read or
	// write private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateAsset cannot be performed: Error %v", err)
	}

	//get Owner Private Collection
	orgCollection, err := getCollectionName(ctx) // get owner collection from caller identity
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	//make sure that there is no attempt to set ID to "RIGHTS"
	if dataInput.ID =="RIGHTS"{
		//Find ID for the flag
		count:= 0
		for{
			flagAsBytes, err := ctx.GetStub().GetPrivateData(orgCollection, fmt.Sprintf("%s%d","F",count))
			if err != nil {
				return fmt.Errorf("failed to get flag: %v", err)
			} else if flagAsBytes == nil {
				break
			}
			count += 1
		}

		//get current Time and Date
		time_now := time.Now()

		flag := Flag{
			ID: 	fmt.Sprintf("%s%d","F",count),
			Date: 	time_now.String(),
			Mesg: 	"GRAVE: tried to set ID to RIGHTS",
		}

		flagJSONasBytes, err := json.Marshal(flag)
		if err != nil {
			return fmt.Errorf("failed to marshal flag into JSON: %v", err)
		}
		log.Printf("GRAVE: not allowed to set ID to RIGHTS")
		err = ctx.GetStub().PutPrivateData(orgCollection, flag.ID, flagJSONasBytes)
		if err != nil {
			return fmt.Errorf("failed to put flag into private data collecton: %v", err)
		}
		return nil
	}	

	// Check if asset already exists
	assetAsBytes, err := ctx.GetStub().GetState(dataInput.ID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if assetAsBytes != nil {
		fmt.Println("Asset already exists: " + dataInput.ID)
		return fmt.Errorf("this asset already exists: " + dataInput.ID)
	}

	// Get Recipe
	var recipe *Recipe
	recipeDetailsJSON, err := ctx.GetStub().GetPrivateData(orgCollection, dataInput.RecipeID)
	if err != nil {
		return fmt.Errorf("failed to read recipe details: %v", err)
	}
	if recipeDetailsJSON == nil {
		return fmt.Errorf("ReadPrivateRecipe for %v does not exist in collection %v", dataInput.RecipeID, orgCollection)
	}
	err = json.Unmarshal(recipeDetailsJSON, &recipe)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	
	//check if Product name matches the recipe 
	if recipe.Product == "FinalProduct"{
		return fmt.Errorf("Use the Final Product function to create the final product")
	}

	//check if Product name matches the recipe 
	if recipe.Product != dataInput.Name{
		return fmt.Errorf("The name of the new product does not match the recipe")
	}
	
	//check if the Assets exist and if the quanitity and ingredients match
	var count map[string]int
	count = make(map[string]int)
	var total_emissionsIDs []string 
	for _, s := range dataInput.Assets{
		//get asset info
		var asset *Asset
		assetDetailsJSON, err := ctx.GetStub().GetPrivateData(orgCollection, s)
		if err != nil {
			return fmt.Errorf("failed to read asset details in loop: %v", err)
		}
		if assetDetailsJSON == nil {
			return fmt.Errorf("Asset for %v does not exist in collection %v", s, orgCollection)
		}
		err = json.Unmarshal(assetDetailsJSON, &asset)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %v", err)
		}
		x := count[asset.Name]
		count[asset.Name] = x+1

		total_emissionsIDs = append(total_emissionsIDs, asset.EmissionsIDs...) // Check if that is properly working
	}	
	
	//check if ingredients match the recipe 
	if (len(count) != len(recipe.Ingredients)) || (len(count) != len(recipe.Quantity)){
		return fmt.Errorf("Number of ingredients does not match recipe")
	} 

	i := 0
	for k, v  := range count{
		for j, ingredient := range recipe.Ingredients{
			if (k == ingredient) && (v==recipe.Quantity[j]){
				i++
				break
			}
		}
	}
	//check if all items have been matched and the correct number is provided
	if i != len(count){
		return fmt.Errorf("The number of correct ingredients is not provided")
	}

	//All necessary checks have been carried out. Item can be created. Used assets are deleted
	//delete used assets
	for _, s := range dataInput.Assets{
		err = ctx.GetStub().DelPrivateData(orgCollection, s)
		if err != nil {
			return err
		}
	}

	//Create the public asset
	publicAsset := PublicAsset{
		ID:    			dataInput.ID,
		EmissionsIDs: 	append(total_emissionsIDs, dataInput.EmissionsIDs...),
		BasedOn: 		dataInput.Assets,
	}
	publicAssetJSONasBytes, err := json.Marshal(publicAsset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset_out into JSON: %v", err)
	}

	log.Printf("CreateAsset Put PublicAsset: ID %v", dataInput.ID)

	err = ctx.GetStub().PutState(dataInput.ID, publicAssetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset onto world state: %v", err)
	}
	


	// Mark Asset as outgoing
	asset_out := Asset{
		Name:  	dataInput.Name,
		ID:    	dataInput.ID,
		EmissionsIDs: 	append(total_emissionsIDs, dataInput.EmissionsIDs...),
		Dir: 	"out",
	}
	assetJSONasBytes, err := json.Marshal(asset_out)
	if err != nil {
		return fmt.Errorf("failed to marshal asset_out into JSON: %v", err)
	}

	log.Printf("CreateAsset Put: collection %v, ID %v", orgCollection, dataInput.ID)

	err = ctx.GetStub().PutPrivateData(orgCollection, dataInput.ID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}
	return nil
}


func (s *SmartContract) FinalProduct(ctx contractapi.TransactionContextInterface) error {

	err := deleteClaimedShippments(ctx)
	if err != nil{
		return fmt.Errorf("Failure with deleting claimed shipment: %v", err)
	}

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		// log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}

	type dataTransient struct {
		RecipeID 	string `json:"recipeID"`
		Name 		string `json:"assetName"`
		ID 			string `json:"assetID"`
		EmissionsIDs []string `json:"emissionsIDs"`
		Assets  	[]string `json:"assets"`
	}

	//get data and check it 
	var dataInput dataTransient
	err = json.Unmarshal(transientAssetJSON, &dataInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	if len(dataInput.RecipeID) == 0 {
		return fmt.Errorf("RecipeID field must be a non-empty string")
	}
	if len(dataInput.Name) == 0 {
		return fmt.Errorf("Name field must be a non-empty string")
	}
	if len(dataInput.ID) == 0 {
		return fmt.Errorf("assetID field must be a non-empty string")
	}
	if len(dataInput.EmissionsIDs) <= 0 {
		return fmt.Errorf("EmissionsIDs list must be greater than 0 %v",dataInput.EmissionsIDs)
	}
	if len(dataInput.Assets) == 0 {
		return fmt.Errorf("Assets slice must be a non-empty")
	}


	// Get ID of submitting client identity
	clientID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if clientID != "0"{clientID= "0"}

	// Verify that the client is submitting request to peer in their organization
	// This is to ensure that a client from another org doesn't attempt to read or
	// write private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateAsset cannot be performed: Error %v", err)
	}

	//get Owner Private Collection
	orgCollection, err := getCollectionName(ctx) // get owner collection from caller identity
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	//make sure that there is no attempt to set ID to "RIGHTS"
	if dataInput.ID =="RIGHTS"{
		//Find ID for the flag
		count:= 0
		for{
			flagAsBytes, err := ctx.GetStub().GetPrivateData(orgCollection, fmt.Sprintf("%s%d","F",count))
			if err != nil {
				return fmt.Errorf("failed to get flag: %v", err)
			} else if flagAsBytes == nil {
				break
			}
			count += 1
		}

		//get current Time and Date
		time_now := time.Now()

		flag := Flag{
			ID: 	fmt.Sprintf("%s%d","F",count),
			Date: 	time_now.String(),
			Mesg: 	"GRAVE: tried to set ID to RIGHTS",
		}

		flagJSONasBytes, err := json.Marshal(flag)
		if err != nil {
			return fmt.Errorf("failed to marshal flag into JSON: %v", err)
		}
		log.Printf("GRAVE: not allowed to set ID to RIGHTS")
		err = ctx.GetStub().PutPrivateData(orgCollection, flag.ID, flagJSONasBytes)
		if err != nil {
			return fmt.Errorf("failed to put flag into private data collecton: %v", err)
		}
		return nil
	}

	//check if the org has the right to create an asset
	rightsDetailsJSON, err := ctx.GetStub().GetPrivateData(orgCollection, "RIGHTS")
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if rightsDetailsJSON == nil {
		return fmt.Errorf("Need to receive rights to use this function")
	}

	var right *Rights
	err = json.Unmarshal(rightsDetailsJSON, &right)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	// if the Org is not a OEM create flag
	if right.Role != "OEM"{
		//Find ID for the flag
		count:= 0
		for{
			flagAsBytes, err := ctx.GetStub().GetPrivateData(orgCollection, fmt.Sprintf("%s%d","F",count))
			if err != nil {
				return fmt.Errorf("failed to get flag: %v", err)
			} else if flagAsBytes == nil {
				break
			}
			count += 1
		}

		//get current Time and Date
		time_now := time.Now()

		flag := Flag{
			ID: 	fmt.Sprintf("%s%d","F",count),
			Date: 	time_now.String(),
			Mesg: 	"Unauthorized attempt at invoking FinalProduct chaincode",
		}

		flagJSONasBytes, err := json.Marshal(flag)
		if err != nil {
			return fmt.Errorf("failed to marshal flag into JSON: %v", err)
		}
		log.Printf("Unauthorized attempt of access: function FinalProduct")
		err = ctx.GetStub().PutPrivateData(orgCollection, flag.ID, flagJSONasBytes)
		if err != nil {
			return fmt.Errorf("failed to put flag into private data collecton: %v", err)
		}
		return nil
	}
	// Check if asset already exists on the public chain
	assetAsBytes, err := ctx.GetStub().GetState(dataInput.ID)
	if err != nil {
		return fmt.Errorf("failed to get asset from world state: %v", err)
	} else if assetAsBytes != nil {
		fmt.Println("Asset already exists on world stage: " + dataInput.ID)
		return fmt.Errorf("this asset already exists on world stage: " + dataInput.ID)
	}

	// Get Recipe
	var recipe *Recipe
	recipeDetailsJSON, err := ctx.GetStub().GetPrivateData(orgCollection, dataInput.RecipeID)
	if err != nil {
		return fmt.Errorf("failed to read recipe details: %v", err)
	}
	if recipeDetailsJSON == nil {
		return fmt.Errorf("ReadPrivateRecipe for %v does not exist in collection %v", dataInput.RecipeID, orgCollection)
	}
	err = json.Unmarshal(recipeDetailsJSON, &recipe)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	
	//check if Product name matches the recipe 
	if recipe.Product != dataInput.Name{
		return fmt.Errorf("The name of the new product does not match the recipe")
	}
	
	//check if the Assets exist and if the quanitity and ingredients match
	var count map[string]int
	count = make(map[string]int)
	var total_emissionsIDs []string
	for _, s := range dataInput.Assets{
		//get asset info
		var asset *Asset
		assetDetailsJSON, err := ctx.GetStub().GetPrivateData(orgCollection, s)
		if err != nil {
			return fmt.Errorf("failed to read asset details in loop: %v", err)
		}
		if assetDetailsJSON == nil {
			return fmt.Errorf("Asset for %v does not exist in collection %v", s, orgCollection)
		}
		err = json.Unmarshal(assetDetailsJSON, &asset)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %v", err)
		}
		x := count[asset.Name]
		count[asset.Name] = x+1

		total_emissionsIDs = append(total_emissionsIDs, asset.EmissionsIDs...)
	}	
	
	//check if ingredients match the recipe 
	if (len(count) != len(recipe.Ingredients)) || (len(count) != len(recipe.Quantity)){
		return fmt.Errorf("Number of ingredients does not match recipe")
	} 

	i := 0
	for k, v  := range count{
		for j, ingredient := range recipe.Ingredients{
			if (k == ingredient) && (v==recipe.Quantity[j]){
				i++
				break
			}
		}
	}
	//check if all items have been matched and the correct number is provided
	if i != len(count){
		return fmt.Errorf("The number of correct ingredients is not provided")
	}

	//All necessary checks have been carried out. Item can be created. Used assets are deleted
	//delete used assets
	for _, s := range dataInput.Assets{
		err = ctx.GetStub().DelPrivateData(orgCollection, s)
		if err != nil {
			return err
		}
	}

	// Mark Asset as finished
	publicAsset := PublicAsset{
		ID:    			dataInput.ID,
		EmissionsIDs: 	append(total_emissionsIDs, dataInput.EmissionsIDs...),
		BasedOn: 		dataInput.Assets,
	}
	publicAssetJSONasBytes, err := json.Marshal(publicAsset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset_out into JSON: %v", err)
	}

	log.Printf("CreateAsset Put: World Stage ID %v",dataInput.ID)

	err = ctx.GetStub().PutState(dataInput.ID, publicAssetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset onto world stage: %v", err)
	}
	return nil
}


//function to package Assets as a Shipping and commit part of info to the shared DataCollection
func (s *SmartContract) CreateShipping(ctx contractapi.TransactionContextInterface) error {

	err := deleteClaimedShippments(ctx)
	if err != nil{
		return fmt.Errorf("Failure with deleting claimed shipment: %v", err)
	}

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		// log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}
	
	type shippingTransient struct {
		ID 			string `json:"shippingID"`
		Quantity 	int `json:"quantity"`
		List_ID 	[]string `json:"list_ID"`
		Name 		string `json:"assetName"`
		Date 		string `json:"date"`
		ShippedEmissionsIDs	[]string `json:"shipEmissionsIDs"`
	}

	//get data and check it 
	var shippingInput shippingTransient
	err = json.Unmarshal(transientAssetJSON, &shippingInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	if len(shippingInput.ID) == 0 {
		return fmt.Errorf("ShippingID field must be a non-empty string")
	}
	if shippingInput.Quantity <= 0 {
		return fmt.Errorf("Quantity field must be a non-empty, positive integer")
	}
	if len(shippingInput.List_ID) == 0 {
		return fmt.Errorf("List_ID slice must be non-empty")
	}
	if len(shippingInput.Name) == 0 {
		return fmt.Errorf("assetName must be a non-empty string")
	}
	if len(shippingInput.ShippedEmissionsIDs) <= 0 {
		return fmt.Errorf("ShipGHG must be larger than 0")
	}
	


	// Get ID of submitting client identity
	clientID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if clientID != "0"{clientID= "0"}

	// Verify that the client is submitting request to peer in their organization
	// This is to ensure that a client from another org doesn't attempt to read or
	// write private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateAsset cannot be performed: Error %v", err)
	}

	//get Owner Private Collection
	orgCollection, err := getCollectionName(ctx) // get owner collection from caller identity
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	//make sure that there is no attempt to set ID to "RIGHTS"
	if shippingInput.ID =="RIGHTS"{
		//Find ID for the flag
		count:= 0
		for{
			flagAsBytes, err := ctx.GetStub().GetPrivateData(orgCollection, fmt.Sprintf("%s%d","F",count))
			if err != nil {
				return fmt.Errorf("failed to get flag: %v", err)
			} else if flagAsBytes == nil {
				break
			}
			count += 1
		}

		//get current Time and Date
		time_now := time.Now()

		flag := Flag{
			ID: 	fmt.Sprintf("%s%d","F",count),
			Date: 	time_now.String(),
			Mesg: 	"GRAVE: tried to set ID to RIGHTS",
		}

		flagJSONasBytes, err := json.Marshal(flag)
		if err != nil {
			return fmt.Errorf("failed to marshal flag into JSON: %v", err)
		}
		log.Printf("GRAVE: not allowed to set ID to RIGHTS")
		err = ctx.GetStub().PutPrivateData(orgCollection, flag.ID, flagJSONasBytes)
		if err != nil {
			return fmt.Errorf("failed to put flag into private data collecton: %v", err)
		}
		return nil
	}

	//make sure that the date of creation is correct
	currentTime := time.Now()
	if shippingInput.Date != currentTime.Format("02-01-2006") {
		//Find ID for the flag
		count:= 0
		for{
			flagAsBytes, err := ctx.GetStub().GetPrivateData(orgCollection, fmt.Sprintf("%s%d","F",count))
			if err != nil {
				return fmt.Errorf("failed to get flag: %v", err)
			} else if flagAsBytes == nil {
				break
			}
			count += 1
		}

		//get current Time and Date
		time_now := time.Now()

		flag := Flag{
			ID: 	fmt.Sprintf("%s%d","F",count),
			Date: 	time_now.String(),
			Mesg: 	"GRAVE: tried to set ID to RIGHTS",
		}

		flagJSONasBytes, err := json.Marshal(flag)
		if err != nil {
			return fmt.Errorf("failed to marshal flag into JSON: %v", err)
		}
		log.Printf("GRAVE: not allowed to set ID to RIGHTS")
		err = ctx.GetStub().PutPrivateData(orgCollection, flag.ID, flagJSONasBytes)
		if err != nil {
			return fmt.Errorf("failed to put flag into private data collecton: %v", err)
		}
		return nil
	}

	// Check if shipping already exists
	assetAsBytes, err := ctx.GetStub().GetPrivateData(orgCollection, shippingInput.ID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if assetAsBytes != nil {
		fmt.Println("Asset already exists: " + shippingInput.ID)
		return fmt.Errorf("this asset already exists: " + shippingInput.ID)
	}
	
	//check if the Assets exist and if they are meant to be out-going
	//check if the Assets all have the same Name 
	var_name := "name"
	var total_EmissionsIDs [][]string
	for i, s := range shippingInput.List_ID{
		//get asset info
		var asset *Asset
		assetDetailsJSON, err := ctx.GetStub().GetPrivateData(orgCollection, s)
		if err != nil {
			return fmt.Errorf("failed to read asset details in loop: %v", err)
		}
		if assetDetailsJSON == nil {
			return fmt.Errorf("Asset for %v does not exist in collection %v", s, orgCollection)
		}
		err = json.Unmarshal(assetDetailsJSON, &asset)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %v", err)
		}
		//use info
		if i == 0{
			var_name = asset.Name
		}
		if asset.Name != var_name{
			return fmt.Errorf("All assets being shipped out must have the same name/type")
		}
		if asset.Dir != "out"{
			return fmt.Errorf("Asset %v is not meant to be shipped out", s)
		}
		total_EmissionsIDs = append(total_EmissionsIDs, asset.EmissionsIDs)
	}	

	//check if the List_ID matches the quantity
	if len(shippingInput.List_ID) != shippingInput.Quantity{
		return fmt.Errorf("Number of Asset IDs in List_ID must match Quantity")
	}

	//All necessary checks have been carried out. Item can be created. Used assets are deleted
	//delete used assets

	for _, s := range shippingInput.List_ID{
		err = ctx.GetStub().DelPrivateData(orgCollection, s)
		if err != nil {
			return err
		}
	}
	
	// Add the emissions of the shipping to the emissions of the assets
	for i, s := range total_EmissionsIDs{
		s = append(s, shippingInput.ShippedEmissionsIDs[i])
	}
	// Create the Private Shipping struct
	shippingPrivate := ShippingPrivate{
		ID:    		shippingInput.ID,
		Quantity: 	shippingInput.Quantity,
		List_ID: 	shippingInput.List_ID,
		Name: 		shippingInput.Name, 
		Date: 		shippingInput.Date,
		EmissionsIDs: total_EmissionsIDs,
	}
	shippingPrivateJSONasBytes, err := json.Marshal(shippingPrivate)
	if err != nil {
		return fmt.Errorf("failed to marshal shippingPrivate into JSON: %v", err)
	}

	log.Printf("CreateShipping Put: collection %v, ID %v", orgCollection, shippingInput.ID)
	//upload private data
	err = ctx.GetStub().PutPrivateData(orgCollection, shippingInput.ID, shippingPrivateJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}

	//get the ORG
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get verified MSPID: %v", err)
	}

	//Create the Public Shipping
	shippingPublic := ShippingPublic{
		ID:    		shippingInput.ID,
		SellerID: 	clientMSPID,
		Name: 		shippingInput.Name,
	}
	shippingPublicJSONasBytes, err := json.Marshal(shippingPublic)
	if err != nil {
		return fmt.Errorf("failed to marshal asset_out into JSON: %v", err)
	}

	log.Printf("CreateShipping Put: collection %v, ID %v", shippingCollection, shippingInput.ID)
	//upload public data
	err = ctx.GetStub().PutPrivateData(shippingCollection, shippingInput.ID, shippingPublicJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}
	
	return nil
}


//function to package Assets as a Shipping and commit part of info to the shared DataCollection
func (s *SmartContract) ClaimShipping(ctx contractapi.TransactionContextInterface) error {

	err := deleteClaimedShippments(ctx)
	if err != nil{
		return fmt.Errorf("Failure with deleting claimed shipment: %v", err)
	}

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		// log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}
	
	type shippingTransient struct {
		ID 			string `json:"shippingID"`
		Quantity 	int `json:"quantity"`
		List_ID 	[]string `json:"list_ID"`
		Name		string `json:"assetName"`
		Date 		string `json:"date"`
		EmissionsIDs [][]string `json:"emissionsIDs"`
	}

	//get data and check it 
	var shippingInput shippingTransient
	err = json.Unmarshal(transientAssetJSON, &shippingInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	if len(shippingInput.ID) == 0 {
		return fmt.Errorf("ShippingID field must be a non-empty string")
	}
	if shippingInput.Quantity <= 0 {
		return fmt.Errorf("Quantity field must be a non-empty, positive integer")
	}
	if len(shippingInput.List_ID) == 0 {
		return fmt.Errorf("List_ID slice must be non-empty")
	}
	if len(shippingInput.Name) == 0 {
		return fmt.Errorf("Name must be a non-empty string")
	}
	if len(shippingInput.Date) == 0 {
		return fmt.Errorf("Date must be a non-empty string")
	}
	if len(shippingInput.EmissionsIDs) <= 0 {
		return fmt.Errorf("EmissionsID List must be a non-empty list")
	}

	// Get ID of submitting client identity
	clientID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if clientID != "0"{clientID= "0"}

	// Verify that the client is submitting request to peer in their organization
	// This is to ensure that a client from another org doesn't attempt to read or
	// write private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateAsset cannot be performed: Error %v", err)
	}

	//get Owner Private Collection
	orgCollection, err := getCollectionName(ctx) // get owner collection from caller identity
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	//make sure that there is no attempt to set ID to "RIGHTS"
	if shippingInput.ID =="RIGHTS"{
		//Find ID for the flag
		count:= 0
		for{
			flagAsBytes, err := ctx.GetStub().GetPrivateData(orgCollection, fmt.Sprintf("%s%d","F",count))
			if err != nil {
				return fmt.Errorf("failed to get flag: %v", err)
			} else if flagAsBytes == nil {
				break
			}
			count += 1
		}

		//get current Time and Date
		time_now := time.Now()

		flag := Flag{
			ID: 	fmt.Sprintf("%s%d","F",count),
			Date: 	time_now.String(),
			Mesg: 	"GRAVE: tried to set ID to RIGHTS",
		}

		flagJSONasBytes, err := json.Marshal(flag)
		if err != nil {
			return fmt.Errorf("failed to marshal flag into JSON: %v", err)
		}
		log.Printf("GRAVE: not allowed to set ID to RIGHTS")
		err = ctx.GetStub().PutPrivateData(orgCollection, flag.ID, flagJSONasBytes)
		if err != nil {
			return fmt.Errorf("failed to put flag into private data collecton: %v", err)
		}
		return nil
	}

	// Check if shipping exists
	var shippingPublic *ShippingPublic
	assetAsBytes, err := ctx.GetStub().GetPrivateData(shippingCollection, shippingInput.ID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if assetAsBytes == nil {
		fmt.Println("Asset doesn't exist: " + shippingInput.ID)
		return fmt.Errorf("this asset doesn't exists: " + shippingInput.ID)
	}

	err = json.Unmarshal(assetAsBytes, &shippingPublic)
	if err != nil {
		return fmt.Errorf("failed to unmarshal shipping: %v", err)
	}
	//check if product is meant for the organization
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get verified MSPID: %v", err)
	}
	if clientMSPID == shippingPublic.SellerID{
		count:= 0
		for{
			flagAsBytes, err := ctx.GetStub().GetPrivateData(orgCollection, fmt.Sprintf("%s%d","F",count))
			if err != nil {
				return fmt.Errorf("failed to get flag: %v", err)
			} else if flagAsBytes == nil {
				break
			}
			count += 1
		}

		//get current Time and Date
		time_now := time.Now()

		flag := Flag{
			ID: 	fmt.Sprintf("%s%d","F",count),
			Date: 	time_now.String(),
			Mesg: 	"Attempt to claim its own shipment",
		}

		flagJSONasBytes, err := json.Marshal(flag)
		if err != nil {
			return fmt.Errorf("failed to marshal flag into JSON: %v", err)
		}
		log.Printf("Error: Attempt to claim own shipment")
		err = ctx.GetStub().PutPrivateData(orgCollection, flag.ID, flagJSONasBytes)
		if err != nil {
			return fmt.Errorf("failed to put flag into private data collecton: %v", err)
		}
		return nil
	}

	//Now compare the two hash values of the sippments 
	//get the seller Orgs Private Data Collection
	sellerCollection := shippingPublic.SellerID + "PrivateCollection"

	// Get hash of seller's shipment value
	sellerShippingHash, err := ctx.GetStub().GetPrivateDataHash(sellerCollection, shippingInput.ID)
	if err != nil {
		return fmt.Errorf("failed to get hash of shipment from seller collection %v: %v", sellerCollection, err)
	}
	if sellerShippingHash == nil {
		return fmt.Errorf("hash of shipment for %v does not exist in collection %v", shippingInput.ID, sellerCollection)
	}
	seller_hash := hex.EncodeToString(sellerShippingHash)

	//Get hash of buyer's shipment value
	shippingJsonAsBytes, err := json.Marshal(shippingInput)
	if err != nil{
		return fmt.Errorf("Failed to marshal shippingInput: %v", err)
	}
	sha := sha256.Sum256(shippingJsonAsBytes)
	buyer_hash := hex.EncodeToString(sha[:])

	// Verify that the two hashes match if not create flag
	if buyer_hash != seller_hash {
		//Find ID for the flag
		count:= 0
		for{
			flagAsBytes, err := ctx.GetStub().GetPrivateData(orgCollection, fmt.Sprintf("%s%d","F",count))
			if err != nil {
				return fmt.Errorf("failed to get flag: %v", err)
			} else if flagAsBytes == nil {
				break
			}
			count += 1
		}

		//get current Time and Date
		time_now := time.Now()

		flag := Flag{
			ID: 	fmt.Sprintf("%s%d","F",count),
			Date: 	time_now.String(),
			Mesg: 	"Failed attempt at claiming shipment",
		}

		flagJSONasBytes, err := json.Marshal(flag)
		if err != nil {
			return fmt.Errorf("failed to marshal flag into JSON: %v", err)
		}
		log.Printf("Unsuccessful attempt at claiming shipment")
		err = ctx.GetStub().PutPrivateData(orgCollection, flag.ID, flagJSONasBytes)
		if err != nil {
			return fmt.Errorf("failed to put flag into private data collecton: %v", err)
		}
		return nil
	}
	
	//////////////////////////////////////////////////////////////////////////////////////////
	//All necessary checks have been carried out. Item can be created. Used assets are deleted
	//////////////////////////////////////////////////////////////////////////////////////////

	//Create the Deletion List or append the Element that has to be deleted	
	delListJSON, err := ctx.GetStub().GetPrivateData(shippingCollection, "DEL")
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	//if the list doesn't exist
	} else if delListJSON == nil {
		//create deletion list element
		del_list := DeletionShippingList{
			ID:    			"DEL",
			Del_List:		[]string{shippingInput.ID},
		}

		delListJSONasBytes, err := json.Marshal(del_list)
		if err != nil {
			return fmt.Errorf("failed to marshal delList into JSON: %v", err)
		}

		log.Printf("NEW_LIST Put: collection %v, ID %v", shippingCollection, "DEL")
		err = ctx.GetStub().PutPrivateData(shippingCollection, "DEL", delListJSONasBytes)
		if err != nil {
			return fmt.Errorf("failed to put delList into public Shipping data collecton: %v", err)
		}

	//if it does exist, append new ID
	} else {

		var del_list *DeletionShippingList
		err = json.Unmarshal(delListJSON, &del_list)
		if err != nil {
			return fmt.Errorf("failed to unmarshal Del List: %v", err)
		}

		new_del_list := DeletionShippingList{
			ID:			"DEL",
			Del_List: 	append(del_list.Del_List,shippingInput.ID),
		}

		delListJSONasBytes, err := json.Marshal(new_del_list)
		if err != nil {
			return fmt.Errorf("failed to marshal delList into JSON: %v", err)
		}

		log.Printf("CreateRecipe Put: collection %v, ID %v", shippingCollection, "DEL")
		err = ctx.GetStub().PutPrivateData(shippingCollection, "DEL", delListJSONasBytes)
		if err != nil {
			return fmt.Errorf("failed to put delList into public Shipping data collecton: %v", err)
		}
	}

	//delete claimed shipping from shippingCollection
	log.Printf("Delete %v from %v", shippingInput.ID, shippingCollection)
	err = ctx.GetStub().DelPrivateData(shippingCollection, shippingInput.ID)
	if err != nil {
		return err
	}

	//create and unpack the new assets from the shipping in loop 
	for i, IDS := range shippingInput.List_ID{
		// Create the Asset
		asset := Asset{
			Name: 		shippingInput.Name,
			ID:    		IDS,
			EmissionsIDs: 		shippingInput.EmissionsIDs[i],
			Dir: 		"in",
		}
		
		assetPrivateJSONasBytes, err := json.Marshal(asset)
		if err != nil {
			return fmt.Errorf("failed to marshal asset into JSON: %v", err)
		}

		log.Printf("ClaimShipping Put: collection %v, ID %v", orgCollection, asset.ID)
		//upload to buyer private data collection
		err = ctx.GetStub().PutPrivateData(orgCollection, IDS, assetPrivateJSONasBytes)
		if err != nil {
			return fmt.Errorf("failed to put asset into private data collecton: %v", err)
		}
	}
	
	return nil
}


func deleteClaimedShippments(ctx contractapi.TransactionContextInterface) (error) {

	// Get the MSP ID of submitting client identity
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get verified MSPID: %v", err)
	}
	// Create the collection name
	orgCollection := clientMSPID + "PrivateCollection"

	//get the deletionList
	var del_list *DeletionShippingList
	delListJSON, err := ctx.GetStub().GetPrivateData(shippingCollection,"DEL")
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	//if there is no list return nil
	} else if delListJSON == nil {
		return nil
	}

	err = json.Unmarshal(delListJSON, &del_list)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(del_list.Del_List) == 0 {
		return nil
	}

	//delete all elements in the list
	for _, ele := range(del_list.Del_List){
		log.Printf("Delete %v from %v", ele, orgCollection)
		err = ctx.GetStub().DelPrivateData(orgCollection, ele)
		if err != nil {
			return err
 		}
	}

	//reset the del_list
	new_del_list := DeletionShippingList{
		ID: 		"DEL",
		Del_List: 	[]string{},
	}

	delJSONasBytes, err := json.Marshal(new_del_list)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}

	log.Printf("DeleteClaimedShipments Put: collection %v, ID %v", shippingCollection,"DEL")
	err = ctx.GetStub().PutPrivateData(shippingCollection, "DEL", delJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put delList into public shipping collection: %v", err)
	}

	return nil
}

// getCollectionName is an internal helper function to get collection of submitting client identity.
func getCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get the MSP ID of submitting client identity
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get verified MSPID: %v", err)
	}

	// Create the collection name
	orgCollection := clientMSPID + "PrivateCollection"

	return orgCollection, nil
}

// verifyClientOrgMatchesPeerOrg is an internal function used verify client org id and matches peer org id.
func verifyClientOrgMatchesPeerOrg(ctx contractapi.TransactionContextInterface) error {
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the client's MSPID: %v", err)
	}
	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the peer's MSPID: %v", err)
	}

	if clientMSPID != peerMSPID {
		return fmt.Errorf("client from org %v is not authorized to read or write private data from an org %v peer", clientMSPID, peerMSPID)
	}

	return nil
}

func submittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read clientID: %v", err)
	}
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	return string(decodeID), nil
}

//read asset from the chain
func (s *SmartContract) CustomerGetAsset(ctx contractapi.TransactionContextInterface, assetID string) (*PublicAsset, error) {

	log.Printf("Read Asset from world stage ID: %v", assetID)
	assetJSON, err := ctx.GetStub().GetState(assetID) //get the shipping from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset: %v", err)
	}
	// No asset found, return empty response
	if assetJSON == nil {
		log.Printf("%v does not exist on world stage", assetID)
		return nil, nil
	}

	var asset *PublicAsset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}



///////
	return asset, nil
}











///////


func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface, collection string) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(collection,"", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			continue
		}

		if asset.ID == "" {
			continue
		}

		assets = append(assets, &asset)
	}

	return assets, nil
}

func (s *SmartContract) GetAllPrivateShippings(ctx contractapi.TransactionContextInterface, collection string) ([]*ShippingPrivate, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(collection,"", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var shippings []*ShippingPrivate
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var shipping ShippingPrivate
		err = json.Unmarshal(queryResponse.Value, &shipping)
		if err != nil {
			continue
		}

		if (shipping.ID == "") || (shipping.Quantity == 0) {
			continue
		}

		shippings = append(shippings, &shipping)
	}

	return shippings, nil
}

func (s *SmartContract) GetAllPublicShippings(ctx contractapi.TransactionContextInterface, collection string) ([]*ShippingPublic, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(collection,"", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var shippings []*ShippingPublic
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var shipping ShippingPublic
		err = json.Unmarshal(queryResponse.Value, &shipping)
		if err != nil {
			continue
		}

		if shipping.SellerID == "" {
			continue
		}

		shippings = append(shippings, &shipping)
	}

	return shippings, nil
}

func (s *SmartContract) GetAllRecipes(ctx contractapi.TransactionContextInterface, collection string) ([]*Recipe, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(collection,"", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var recipes []*Recipe
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var recipe Recipe
		err = json.Unmarshal(queryResponse.Value, &recipe)
		if err != nil {
			continue
		}

		if len(recipe.Ingredients) == 0 {
			continue
		}

		recipes = append(recipes, &recipe)
	}

	return recipes, nil
}

func (s *SmartContract) GetAllFlags(ctx contractapi.TransactionContextInterface, collection string) ([]*Flag, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(collection,"", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var flags []*Flag
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var flag Flag
		err = json.Unmarshal(queryResponse.Value, &flag)
		if err != nil {
			continue
		}

		if flag.Mesg == "" {
			continue
		}

		flags = append(flags, &flag)
	}

	return flags, nil
}

//Read an Organization's rights
func (s *SmartContract) ReadRight(ctx contractapi.TransactionContextInterface, collection string) (*Rights, error) {
	
	log.Printf("ReadRight: collection %v", collection)
	rightsDetailsJSON, err := ctx.GetStub().GetPrivateData(collection, "RIGHTS") 
	// Get the shipping from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read rights details: %v", err)
	}
	if rightsDetailsJSON == nil {
		log.Printf("Rights do not exist in collection %v", collection)
		return nil, nil
	}

	var right *Rights
	err = json.Unmarshal(rightsDetailsJSON, &right)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return right, nil
}

//Read the Deletion list
func (s *SmartContract) ReadDelList(ctx contractapi.TransactionContextInterface) (*DeletionShippingList, error) {

	log.Printf("Read Del List: collection %v, ID %v", shippingCollection, "DEL")
	delListJSON, err := ctx.GetStub().GetPrivateData(shippingCollection, "DEL") //get the shipping from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read shipping: %v", err)
	}

	// No Shipping found, return empty response
	if delListJSON == nil {
		log.Printf("%v does not exist in collection %v", "DEL", shippingCollection)
		return nil, nil
	}

	var del_list *DeletionShippingList
	err = json.Unmarshal(delListJSON, &del_list)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return del_list, nil

}

//ReadPublicShipping reads the information from collection
func (s *SmartContract) ReadPublicShipping(ctx contractapi.TransactionContextInterface, shippingID string) (*ShippingPublic, error) {

	log.Printf("ReadPublicShipping: collection %v, ID %v", shippingCollection, shippingID)
	shippingJSON, err := ctx.GetStub().GetPrivateData(shippingCollection, shippingID) //get the shipping from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read shipping: %v", err)
	}

	// No Shipping found, return empty response
	if shippingJSON == nil {
		log.Printf("%v does not exist in collection %v", shippingID, shippingCollection)
		return nil, nil
	}

	var shipping *ShippingPublic
	err = json.Unmarshal(shippingJSON, &shipping)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return shipping, nil

}

//ReadPrivateShipping reads the information from collection
func (s *SmartContract) ReadPrivateShipping(ctx contractapi.TransactionContextInterface, collection string, shippingID string) (*ShippingPrivate, error) {
	
	log.Printf("ReadPrivateShippings: collection %v, ID %v", collection, shippingID)
	ShippingDetailsJSON, err := ctx.GetStub().GetPrivateData(collection, shippingID) 
	// Get the shipping from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read shipping details: %v", err)
	}
	if ShippingDetailsJSON == nil {
		log.Printf("ReadPrivateShippings for %v does not exist in collection %v", shippingID, collection)
		return nil, nil
	}

	var shipping *ShippingPrivate
	err = json.Unmarshal(ShippingDetailsJSON, &shipping)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return shipping, nil
}

// ReadPrivateAsset reads the asset private in organization specific collection
func (s *SmartContract) ReadPrivateAsset(ctx contractapi.TransactionContextInterface, collection string, assetID string) (*Asset, error) {
	
	log.Printf("ReadPrivateAsset: collection %v, ID %v", collection, assetID)
	assetDetailsJSON, err := ctx.GetStub().GetPrivateData(collection, assetID) // Get the asset from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset details: %v", err)
	}
	if assetDetailsJSON == nil {
		log.Printf("ReadPrivateAsset for %v does not exist in collection %v", assetID, collection)
		return nil, nil
	}

	var asset *Asset
	err = json.Unmarshal(assetDetailsJSON, &asset)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return asset, nil
}

// ReadRecipe reads the private recipe from a Data Collection
func (s *SmartContract) ReadRecipe(ctx contractapi.TransactionContextInterface, collection string, recipeID string) (*Recipe, error) {
	
	log.Printf("ReadPrivateRecipe: collection %v, ID %v", collection, recipeID)
	recipeDetailsJSON, err := ctx.GetStub().GetPrivateData(collection, recipeID) // Get the recipe from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read recipe details: %v", err)
	}
	if recipeDetailsJSON == nil {
		log.Printf("ReadPrivateRecipe for %v does not exist in collection %v", recipeID, collection)
		return nil, nil
	}

	var recipe *Recipe
	err = json.Unmarshal(recipeDetailsJSON, &recipe)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return recipe, nil
}






// main function starts up the chaincode in the container during instantiate
func main() {
	emissionsAuditChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating emissionsAudit chaincode: %v", err)
	}

	if err := emissionsAuditChaincode.Start(); err != nil {
		log.Panicf("Error starting emissionsAudit chaincode: %v", err)
	}
}
