package main

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

type AccessListEntry struct {
	Address     string   `json:"address"`
	StorageKeys []string `json:"storageKeys"`
}

type Transaction struct {
	BlockHash   string `json:"blockHash"`
	BlockNumber int    `json:"blockNumber"`
	From        string `json:"from"`
	Gas         int    `json:"gas"`
	GasPrice    string `json:"gasPrice"`

	MaxFeePerGas         string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`

	Hash             string            `json:"hash"`
	Input            string            `json:"input"`
	Nonce            int               `json:"nonce"`
	To               string            `json:"to"`
	TransactionIndex int               `json:"transactionIndex"`
	Value            string            `json:"value"`
	Type             string            `json:"type"`
	AccessList       []AccessListEntry `json:"accessList"`
	ChainId          string            `json:"chainId"`
	V                string            `json:"v"`
	R                string            `json:"r"`
	S                string            `json:"s"`
}

// Struct with only From and Nonce fields
type PartialTransaction struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
}

type QueryResult interface {
}

type FromQueryResult struct {
	Key       string       `json:"Key"`
	Record    *Transaction `json:"record"`
	Timestamp string       `json:"timestamp"`
}

type ToQueryResult struct {
	Key       string              `json:"Key"`
	Record    *PartialTransaction `json:"record"`
	Timestamp string              `json:"timestamp"`
}

// SimpleContract contract for handling writing and reading from the world state
type SmartContract struct {
}

func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (sc *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := stub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	switch function {
	case "InitLedger":
		return sc.InitLedger(stub)
	case "CreateBulk":
		return sc.CreateBulk(stub, args)
	case "CreateBulkParallel":
		return sc.CreateBulkParallel(stub, args)
	case "CreateBulkParallelTo":
		return sc.CreateBulkParallelTo(stub, args)
	case "Create":
		return sc.Create(stub, args)
	case "GetState":
		return sc.GetState(stub, args)
	case "GetHistoryForKey":
		return sc.GetHistoryForKey(stub, args)
	// Requires GetHistoryForKeyRange API
	case "GetHistoryForKeyRange":
		return sc.GetHistoryForKeyRange(stub, args)
	// Requires GetHistoryForVersionRange API
	case "GetHistoryForVersionRange":
		return sc.GetHistoryForVersionRange(stub, args)
	case "GetHistoryForBlockRange":
		return sc.GetHistoryForBlockRange(stub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
}

func (sc *SmartContract) InitLedger(stub shim.ChaincodeStubInterface) sc.Response {
	log.Println("'============= Initialized Ledger ==========='")
	return shim.Success(nil)

}

func (sc *SmartContract) Create(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var transaction Transaction
	json.Unmarshal([]byte(args[0]), &transaction)

	transactionBytes, err := json.Marshal(transaction)
	if err != nil {
		return shim.Error("Failed to marshal transaction JSON: " + err.Error())
	}

	transactionKey := transaction.From
	log.Printf("Appending transaction: %s\n", transactionKey)

	err = stub.PutState(transactionKey, transactionBytes)
	if err != nil {
		return shim.Error("failed to put transaction on ledger: " + err.Error())
	}

	return shim.Success(nil)

}

// Create a new key-value pair and send to state database
func (sc *SmartContract) CreateBulk(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	buffer := args[0]

	var transactions []Transaction
	json.Unmarshal([]byte(buffer), &transactions)

	for _, transaction := range transactions {

		transactionBytes, err := json.Marshal(transaction)
		if err != nil {
			return shim.Error("failed to marshal transaction JSON: " + err.Error())
		}

		transactionKey := transaction.From

		// Fabric key must be a string
		//fmt.Sprintf("%d", transaction.L_ORDERKEY)
		log.Printf("Appending transaction %s with gasPrice %d\n", transactionKey, transaction.GasPrice)
		err = stub.PutState(transactionKey, transactionBytes)
		if err != nil {
			return shim.Error("failed to put transaction on ledger: " + err.Error())
		}
	}

	return shim.Success(nil)

}

func (sc *SmartContract) CreateBulkParallel(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var transactions []Transaction
	json.Unmarshal([]byte(args[0]), &transactions)

	for _, transaction := range transactions {
		transactionBytes, err := json.Marshal(transaction)
		if err != nil {
			return shim.Error("Error marshaling transaction object: " + err.Error())
		}

		err = stub.PutState(transaction.From, transactionBytes)
		if err != nil {
			return shim.Error("Failed to create transaction: " + err.Error())
		}
	}
	return shim.Success(nil)
}

func (sc *SmartContract) CreateBulkParallelTo(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var transactions []Transaction
	json.Unmarshal([]byte(args[0]), &transactions)

	for _, transaction := range transactions {
		transactionBytes, err := json.Marshal(transaction)
		if err != nil {
			return shim.Error("Error marshaling transaction object: " + err.Error())
		}
		err = stub.PutState(transaction.From, transactionBytes)
		if err != nil {
			return shim.Error("Failed to create transaction: " + err.Error())
		}

		partial := PartialTransaction{
			From:  transaction.From,
			Value: transaction.Value,
		}
		partialBytes, err := json.Marshal(partial)
		if err != nil {
			return shim.Error("Error marshaling transaction object: " + err.Error())
		}
		err = stub.PutState("t-"+transaction.To, partialBytes)
		if err != nil {
			return shim.Error("Failed to create transaction: " + err.Error())
		}
	}
	return shim.Success(nil)
}

func (sc *SmartContract) GetState(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	log.Println("-----GetState Test-----")
	key := args[0]
	val, err := stub.GetState(key)
	if err != nil {
		shim.Error("Failed to get state: " + err.Error())
	}
	return shim.Success(val)
}

func ToOrFrom(historyData *queryresult.KeyModification) (QueryResult, error) {
	//Convert google.protobuf.Timestamp to string
	timestamp := time.Unix(historyData.Timestamp.Seconds, int64(historyData.Timestamp.Nanos)).String()

	var transaction Transaction
	json.Unmarshal(historyData.Value, &transaction)

	if strings.HasPrefix(transaction.To, "t-") {
		var partial PartialTransaction
		err := json.Unmarshal(historyData.Value, &partial)
		if err != nil {
			return nil, err
		}
		return ToQueryResult{Key: historyData.TxId, Record: &partial, Timestamp: timestamp}, nil
	}
	return FromQueryResult{Key: historyData.TxId, Record: &transaction, Timestamp: timestamp}, nil
}

// GetHistoryForKey calls built in GetHistoryForKey() API
func (sc *SmartContract) GetHistoryForKey(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	historyItr, err := stub.GetHistoryForKey(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	defer historyItr.Close()

	var history []QueryResult
	for historyItr.HasNext() {
		historyData, err := historyItr.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		result, err := ToOrFrom(historyData)
		if err != nil {
			return shim.Error(err.Error())
		}
		history = append(history, result)
	}

	historyAsBytes, _ := json.Marshal(history)
	return shim.Success(historyAsBytes)
}

// GetHistoryForKeyRange calls custom GetHistoryForKeyRange() API
func (sc *SmartContract) GetHistoryForKeyRange(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1 or more")
	}

	// calling the GetHistoryForKeyRange() API with keys as args
	historyItr, err := stub.GetHistoryForKeyRange(args) // historyIters in old version
	if err != nil {
		return shim.Error(err.Error())
	}

	var history []QueryResult
	for historyItr.HasNext() {
		historyData, err := historyItr.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		result, err := ToOrFrom(historyData)
		if err != nil {
			return shim.Error(err.Error())
		}
		history = append(history, result)
	}

	// var histories [][]FromQueryResult
	// for _, historyItr := range historyItrs {
	// 	var history []FromQueryResult
	// 	for historyItr.HasNext() {
	// 		historyData, err := historyItr.Next()
	// 		if err != nil {
	// 			return shim.Error(err.Error())
	// 		}

	// 		var transaction Transaction
	// 		json.Unmarshal(historyData.Value, &transaction)

	// 		history = append(history, FromQueryResult{Key: historyData.TxId, Record: &transaction})
	// 	}
	// 	histories = append(histories, history)
	// }

	// historiesAsBytes, _ := json.Marshal(histories)
	// return shim.Success(historiesAsBytes)

	historyAsBytes, _ := json.Marshal(history)
	return shim.Success(historyAsBytes)
}

func (sc *SmartContract) GetHistoryForVersionRange(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	start, _ := strconv.ParseUint(args[1], 10, 64)
	end, _ := strconv.ParseUint(args[2], 10, 64)

	versionIter, err := stub.GetHistoryForVersionRange(args[0], start, end)
	if err != nil {
		return shim.Error(err.Error())
	}

	var versions []QueryResult
	for versionIter.HasNext() {
		versionData, err := versionIter.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		result, err := ToOrFrom(versionData)
		if err != nil {
			return shim.Error(err.Error())
		}
		versions = append(versions, result)
	}

	versionAsBytes, _ := json.Marshal(versions)
	return shim.Success(versionAsBytes)
}

func (sc *SmartContract) GetHistoryForBlockRange(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	start, _ := strconv.ParseUint(args[0], 10, 64)
	end, _ := strconv.ParseUint(args[1], 10, 64)
	updates, _ := strconv.ParseUint(args[2], 10, 64)

	resultsIter, err := stub.GetHistoryForBlockRange(start, end, updates)
	if err != nil {
		return shim.Error(err.Error())
	}

	var results []QueryResult
	for resultsIter.HasNext() {
		resultData, err := resultsIter.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		result, err := ToOrFrom(resultData)
		if err != nil {
			return shim.Error(err.Error())
		}
		results = append(results, result)
	}

	resultsAsBytes, _ := json.Marshal(results)
	return shim.Success(resultsAsBytes)
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		log.Printf("Error starting chaincode: %v \n", err)
	}
}
