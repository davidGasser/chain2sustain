package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Import hyperledger fabric SmartContract
type SmartContract struct {
	contractapi.Contract
}

// Emissions Record describes which emissions are being tracked
// Alphabetic order to achieve determinism accross languages
type EmissionsRecord struct {
	ID    string `json:"ID"`
	KgCO2 int    `json:"KgCO2"` // Total emissions in Kg of CO2
}

// EmissionsRecordPrivateDetails describes details that are private to owner/creator of the emissions record
type EmissionsRecordPrivateDetails struct {
	ID    string `json:"ID"`
	Owner string `json:"Owner"` // Identifier based on MSPID and ID of the client's identity
}


// CreateEmissionsRecord adds a new emissions record to the ledger
// Maybe this should be a private function, but for now it is public
func (s *SmartContract) CreateEmissionsRecord(ctx contractapi.TransactionContextInterface, id string, kgCO2 int) error {
	// Check if the emissions record already exists
	exists, err := s.EmissionsRecordExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the Emissions Record with ID %s already exists", id)
	}

	// Create a new emissions record
	emissionsRecord := EmissionsRecord{
		ID:    id,
		KgCO2: kgCO2,
	}
	recordJSON, err := json.Marshal(emissionsRecord) // Convert the emissions record to JSON
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, recordJSON) // Write the emissions record to the ledger
}


// CreateEmissionsRecordPrivateDetails adds new private emissions record details to the private data collection
// Transient Data: ownerID string
// Maybe this should be a private function, but for now it is public
func (s *SmartContract) CreateEmissionsRecordPrivateDetails(ctx contractapi.TransactionContextInterface, id string) error {
	// Get new emissions data from transient map
	transientOwnerID, err := getTransientData(ctx, "ownerID") // OPTIONAL TODO: Check if the ownerID is a valid ID
	if err != nil {
		return fmt.Errorf("failed to get transient: %v", err)
	}
	// Check if emissons record already exists on private data collection
	// Get collection name for this organization.
	orgCollection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}
	recordAsBytes, err := ctx.GetStub().GetPrivateData(orgCollection, id)
	if err != nil {
		return fmt.Errorf("failed to get private emissionsRecordDetails from collection: %v", err)
	} else if recordAsBytes != nil {
		fmt.Println("EmissionsRecordID already exists: " + id)
		return fmt.Errorf("EmissionsRecord asset already exists: " + id)
	}

	// Verify that the client is submitting request to peer in their organization
	// This is to ensure that a client from another org doesn't attempt to read or
	// write private data from this peer.
	// I think this can also be defined in the collections_config.json file, so it might be redundant
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateAsset cannot be performed: Error %v", err)
	}

	// Save emissionRecordDetails to collection visible to owning organization
	emissionsRecordPrivateDetails := EmissionsRecordPrivateDetails{
		ID:    id,
		Owner: string(transientOwnerID),
	}

	emissionsRecordPrivateDetailsAsBytes, err := json.Marshal(emissionsRecordPrivateDetails) // marshal private record details to JSON
	if err != nil {
		return fmt.Errorf("failed to marshal into JSON: %v", err)
	}

	// Put private details of emissions Record into owners org specific private data collection
	err = ctx.GetStub().PutPrivateData(orgCollection, id, emissionsRecordPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put emissionsRecord private details: %v", err)
	}
	return nil
}

// GetEmissionsRecord returns the record stored in the ledger with the given id.
func (s *SmartContract) GetEmissionsRecord(ctx contractapi.TransactionContextInterface, id string) (*EmissionsRecord, error) {
	recordJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from ledger: %v", err)
	}
	if recordJSON == nil {
		return nil, fmt.Errorf("the emissions record with ID %s does not exist", id)
	}

	var record EmissionsRecord
	err = json.Unmarshal(recordJSON, &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// GetEmissionsRecordsList returns all emissions records matching the specified IDs stored in the ledger
func (s *SmartContract) GetEmissionsRecordsList(ctx contractapi.TransactionContextInterface, ids []string) ([]*EmissionsRecord, error) {
	var records []*EmissionsRecord
	for _, id := range ids {
		record, err := s.GetEmissionsRecord(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("the emissionsRecordID %v does not exist", id)
		}
		records = append(records, record)
	}
	return records, nil
}

// GetAllEmissionsRecords returns all emissions records stored in the ledger
// This function should be used only for testing purposes
func (s *SmartContract) GetAllEmissionsRecords(ctx contractapi.TransactionContextInterface) ([]*EmissionsRecord, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []*EmissionsRecord
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var record EmissionsRecord
		err = json.Unmarshal(queryResponse.Value, &record)
		if err != nil {
			return nil, err
		}
		records = append(records, &record)
	}

	return records, nil
}

// GetEmissionsRecordsOfOwner returns all emissions records stored in the private data collection for the invoking owner
// Transient Data: ownerID string
func (s *SmartContract) GetEmissionsRecordsOfOwner(ctx contractapi.TransactionContextInterface) ([]*EmissionsRecordPrivateDetails, error) {
	// Get collection name for this organization.
	orgCollection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}
	// Get ownerID from transient data
	ownerID, err := getTransientData(ctx, "ownerID")
	if err != nil {
		return nil, fmt.Errorf("failed to get ownerID from transient data: %v", err)
	}

	queryString := fmt.Sprintf(`{"selector":{"Owner":"%s"}}`, ownerID)
	resultsIterator, err := ctx.GetStub().GetPrivateDataQueryResult(orgCollection, queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to execute the private data query: %w", err)
	}
	defer resultsIterator.Close()

	var recordsDetails []*EmissionsRecordPrivateDetails

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over private data query results: %w", err)
		}

		var record EmissionsRecordPrivateDetails
		err = json.Unmarshal(queryResponse.Value, &record)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal struct JSON: %w", err)
		}

		recordsDetails = append(recordsDetails, &record)
	}

	return recordsDetails, nil
}

// GetEmissionsRecordPrivateDetails reads the private details of an emissions Record in organization specific collection
func (s *SmartContract) GetEmissionsRecordPrivateDetails(ctx contractapi.TransactionContextInterface, recordID string) (*EmissionsRecordPrivateDetails, error) {
	// Get collection name for this organization.
	orgCollection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	recordDetailsJSON, err := ctx.GetStub().GetPrivateData(orgCollection, recordID) // Get the asset from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset details: %v", err)
	}
	if recordDetailsJSON == nil {
		log.Printf("AssetPrivateDetails for %v does not exist in collection %v", recordID, orgCollection)
		return nil, nil
	}

	var recordDetails *EmissionsRecordPrivateDetails
	err = json.Unmarshal(recordDetailsJSON, &recordDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return recordDetails, nil
}

// GetAllEmissionsRecordsPrivateDetails returns all emissions records private details stored in the private data collection of the invoking organization
// This function should be used only for testing purposes
func (s *SmartContract) GetAllEmissionsRecordsPrivateDetails(ctx contractapi.TransactionContextInterface) ([]*EmissionsRecordPrivateDetails, error) {
	// Get collection name for this organization.
	orgCollection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}
	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(orgCollection, "", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []*EmissionsRecordPrivateDetails
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var record EmissionsRecordPrivateDetails
		err = json.Unmarshal(queryResponse.Value, &record)
		if err != nil {
			return nil, err
		}
		records = append(records, &record)
	}

	return records, nil
}

// EmissionsRecordExists returns true when record with given ID exists in ledger
func (s *SmartContract) EmissionsRecordExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	recordJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from ledger: %v", err)
	}

	return recordJSON != nil, nil
}

// AuditEmissions takes emissions data from an organization, checks its validity and adds it to the ledger
func (s *SmartContract) AuditEmissions(ctx contractapi.TransactionContextInterface, id string, prevEmissionsIDs []string, kgCO2 int, info string) error {
	// Check if the emissions record already exists on public ledger
	exists, err := s.EmissionsRecordExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the Emissions Record with ID %s already exists", id)
	}

	// Check if the emissions matche the expected value

	// If no records exist, no outlier detection is possible
	if len(prevEmissionsIDs) > 0 {
		// Get the emissions records from the ledger
		records, err := s.GetEmissionsRecordsList(ctx, prevEmissionsIDs)
		if err != nil {
			return fmt.Errorf("unable to retrieve emissionsRecords from ledger")
		}

		if len(records) == 0 {
			return fmt.Errorf("records on public ledger do not match records in private collection")
		}

		// Check if submitted emissions match expected range with a simple outlier detection
		// Extract Emissions from records
		previousEmissions := make([]int, len(records))
		for i, record := range records {
			previousEmissions[i] = record.KgCO2
		}
		// Check if the submitted emissions are an outlier
		if isOutlier(kgCO2, previousEmissions) {
			return fmt.Errorf("the submitted emissions do not pass the automated Audit, please contact an administrator for manual reauditing")
		}
	}

	// Emissions are valid, add them to the ledger
	// Create a new emissions record
	err = s.CreateEmissionsRecord(ctx, id, kgCO2)
	if err != nil {
		return fmt.Errorf("failed to create emissions record: %v", err)
	}
	return s.CreateEmissionsRecordPrivateDetails(ctx, id)
}

// HELPER FUNCTION isOutlier checks if a value is an outlier in a given set of values based on the Median Absolute Deviation (MAD) method
func isOutlier(value int, values []int) bool {
	// No outlier detection is possible if there are no values
	if len(values) == 0 {
		return false
	}
	// For 1-4 values, use the simple outlier detection method where values can differentiate by 50%
	if len(values) < 5 {
		// Calculate the median
		median := calculateMedian(values)
		lowerFence := median * 0.5
		upperFence := median * 1.5
		return float64(value) < lowerFence || float64(value) > upperFence
	}

	// For more than 4 values, calculate quantiles 
	// Sort the values in ascending order
	sortedValues := make([]int, len(values))
	copy(sortedValues, values)
	sort.Ints(sortedValues)

	// Calculate the first quartile (25th percentile) and the third quartile (75th percentile)
	q1 := calculatePercentile(sortedValues, 25)
	q3 := calculatePercentile(sortedValues, 75)

	// Calculate the interquartile range (IQR)
	iqr := q3 - q1

	// Calculate the lower and upper fences
	lowerFence := q1 - (1.5 * float64(iqr))
	upperFence := q3 + (1.5 * float64(iqr))

	// Check if the value is an outlier
	return float64(value) < lowerFence || float64(value) > upperFence
}

// HELPER FUNCTION calculatePercentile calculates the percentile of a given set of values
func calculatePercentile(sortedValues []int, percentile int) float64 {
	length := len(sortedValues)
	index := (percentile * (length + 1) / 100) - 1

	if index < 0 {
		return float64(sortedValues[0])
	} else if index >= length-1 {
		return float64(sortedValues[length-1])
	}

	lowerValue := float64(sortedValues[index])
	upperValue := float64(sortedValues[index+1])

	return lowerValue + (upperValue-lowerValue)*(float64(percentile%100)/100.0)
}

// HELPER FUNCTION calculateMedian calculates the median of a given set of values
func calculateMedian(data []int) float64 {
	dataCopy := make([]int, len(data))
	copy(dataCopy, data)

	sort.Ints(dataCopy)

	var median float64
	l := len(dataCopy)
	if l == 0 {
		return 0
	} else if l%2 == 0 {
		median = float64(dataCopy[l/2-1]+dataCopy[l/2]) / 2
	} else {
		median = float64(dataCopy[l/2])
	}

	return median
}

// HELPER FUNCTIONS verifyClientOrgMatchesPeerOrg is an internal function used verify client org id and matches peer org id.
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

// HELPER FUNCTION getCollectionName is an internal helper function to get collection of submitting client identity.
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

// HELPER FUNCTION getTransientData to extract transient data from the transaction proposal
func getTransientData(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	// Get data from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}

	transientData, ok := transientMap[key]
	if !ok {
		return "", fmt.Errorf("key \"%v\" not found in the transient map input", key)
	}

	if len(transientData) == 0 {
		return "", fmt.Errorf("key \"%v\" field in the transient map must be a non-empty string", key)
	}
	return string(transientData), nil
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
