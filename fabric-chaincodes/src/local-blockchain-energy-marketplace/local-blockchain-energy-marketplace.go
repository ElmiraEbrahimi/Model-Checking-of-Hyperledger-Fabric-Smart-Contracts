package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	// April 2020, Updated to Fabric 2.0 Shim

	// KV Interface
	//"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type LocalEnergyTrade struct {
}

var customersKey = "_customers"
var offersKey = "_offers"
var transactionsKey = "_transactions"
var pendingTransactionKey = "_pendingtransaction" //  tracking the pending transaction

type Transaction struct {
	TXID   int64          `json:"txid"`
	Offers map[string]int `json:"offers"`
	Buyer  string         `json:"buyer"`
	Cost   int            `json:"cost"`
	Energy int            `json:"energy"`
	Status string         `json:"status"`
}

// Query response structs
type QueryResponseInt struct {
	Success bool `json:"success"`
	Data    int  `json:"data"`
}

type QueryResponseMap struct {
	Success bool           `json:"success"`
	Data    map[string]int `json:"data"`
}

type QueryResponseString struct {
	Success bool   `json:"success"`
	Data    string `json:"data"`
}

type QueryResponseTransactions struct {
	Success bool          `json:"success"`
	Data    []Transaction `json:"data"`
}

type QueryResponseBytes struct {
	Success bool   `json:"success"`
	Data    []byte `json:"data"`
}

// Init - reset the chaincode
func (t *LocalEnergyTrade) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var input int
	var err error
	var message string

	if len(args) != 1 {
		message = "Incorrect number of arguments. Expecting 1: Initial value"
		return []byte(message), errors.New(message)
	}

	// Get initial value
	input, err = strconv.Atoi(args[0])
	if err != nil {
		message = "Incorrect number of arguments. Expecting 1: Initial value"
		return []byte(message), errors.New(message)
	}

	err = stub.PutState("ece", []byte(strconv.Itoa(input)))
	if err != nil {
		return nil, err
	}

	// Clear list of customers
	// "owner" represents the owner of market
	emptyCustomers := make(map[string]int)
	emptyCustomers["owner"] = 0
	err = marshalAndPut(stub, customersKey, emptyCustomers)
	if err != nil {
		return nil, err
	}

	// Clear the list of offers
	emptyOffers := make(map[string]int)
	err = marshalAndPut(stub, offersKey, emptyOffers)
	if err != nil {
		return nil, err
	}
	// Clear the list of transactions
	var emptyTransactions []Transaction
	err = marshalAndPut(stub, transactionsKey, emptyTransactions)
	if err != nil {
		return nil, err
	}

	// Clear the pendingTransaction array
	var emptyPendingTransaction []Transaction
	err = marshalAndPut(stub, pendingTransactionKey, emptyPendingTransaction)
	if err != nil {
		return nil, err
	}
	// Successful init return
	message = "Chaincode initialized."
	return []byte(message), nil
}

// Invoke function
func (t *LocalEnergyTrade) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	fmt.Println("Chaincode Invoke Is executing " + function)

	switch function {
	case "addOfferQuantity":
		return addOfferQuantity(stub, args)
	case "subtractOfferQuantity":
		return subtractOfferQuantity(stub, args)
	case "addCustomer":
		return addCustomer(stub, args)
	case "addCustomerFunds":
		return addCustomerFunds(stub, args)
	case "acceptOffer":
		return acceptOffer(stub, args)
	case "completeTransaction":
		return completeTransaction(stub)
	case "init":
		return t.Init(stub, "init", args)
	default:

		fmt.Println("Invoke() did not find function: " + function)
		return []byte("Invoke() did not find function: " + function), errors.New("Received unknown function invocation: " + function)
	}

	return nil, nil

}

func (t *LocalEnergyTrade) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	//function, args := stub.GetFunctionAndParameters()
	fmt.Println("Chaincode Query Is executing " + function)

	if function == "getPendingTransaction" {
		return getPendingTransaction(stub)
	}
	if function == "getOffers" {
		return getOffers(stub)
	}

	if function == "getCustomer" {
		return getCustomer(stub, args)
	}
	if function == "getTotalEnergyForSale" {
		return getTotalEnergyForSale(stub)
	}

	fmt.Println("Bad Function Name" + function)
	return createQueryResponseString(false, "Query() did not find function"+function)
}

func getPendingTransaction(stub shim.ChaincodeStubInterface) ([]byte, error) {

	var pt []Transaction
	fmt.Println(" getting the pending transaction")

	transactionAsBytes, err := stub.GetState(pendingTransactionKey)
	if err != nil {
		return createQueryResponseString(false, "Failed to get pending transaction")
	}
	json.Unmarshal(transactionAsBytes, &pt)

	// Make sure there isn't more than 1 pending transaction
	// Otherwise, return the pt array
	if len(pt) > 1 {
		return createQueryResponseString(false, "More than 1 pending transaction, something is wrong!")
	} else {
		return createQueryResponseTransactions(true, pt)
	}

}

// Get all of the available offers
func getOffers(stub shim.ChaincodeStubInterface) ([]byte, error) {

	var offers map[string]int
	fmt.Println(" getting the available offers")

	// Get the available offers from the chaincode state
	offersAsBytes, err := stub.GetState(offersKey)
	if err != nil {
		return createQueryResponseString(false, "Failed to get available offers")
	}
	json.Unmarshal(offersAsBytes, &offers)

	return createQueryResponseMap(true, offers)

}
func getCustomer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var customers map[string]int

	if len(args) != 1 {
		return createQueryResponseString(false, "Incorrect number of arguments. Expecting 1: Customer ID")

	}
	if len(args[0]) == 0 {
		return createQueryResponseString(false, "First argument (customer name) cannot be an empty string")
	}

	customerID := strings.ToLower(args[0])

	customersAsBytes, err := stub.GetState(customersKey)
	if err != nil {
		return createQueryResponseString(false, "Failed to get customers")
	}
	json.Unmarshal(customersAsBytes, &customers)

	//  customer is in the list and is valid
	if val, ok := customers[customerID]; ok {
		return createQueryResponseInt(true, val)
	} else {
		return createQueryResponseString(false, "Failed to find customer with ID "+customerID)
	}

}

// total number of energy
func getTotalEnergyForSale(stub shim.ChaincodeStubInterface) ([]byte, error) {

	var offers map[string]int

	offersAsBytes, err := stub.GetState(offersKey)
	if err != nil {
		return createQueryResponseString(false, "Failed to get available offers")
	}
	json.Unmarshal(offersAsBytes, &offers)

	// Sum the values for sale over all of the keys
	total := 0
	for j := range offers {
		total += offers[j]
		fmt.Println("Key: " + j + ", Value: " + strconv.Itoa(offers[j]) + ". Total is now " + strconv.Itoa(total))
	}
	return createQueryResponseInt(true, total)
}

//////////////////////////////////////// INVOKE  ////////////////////////////////////////
func addOfferQuantity(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var retStr string
	var offers map[string]int
	//int totalCount =0

	if len(args) != 2 {
		retStr = "Incorrect number of arguments. Expecting 2"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	if len(args[0]) == 0 {
		retStr = "First argument (offer ID) cannot be an empty string"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	if len(args[1]) == 0 {
		retStr = "Second argument (quantity to add) cannot be an empty string"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	//offer ID is valid
	offerIDInt, err := strconv.Atoi(args[0])
	if err != nil {
		retStr = "Offer ID must be an integer string"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	if offerIDInt <= 0 {
		retStr = "Offer ID must not be less than zero"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	offerID := args[0]

	// quantity is not less than  0
	quantity, err := strconv.Atoi(args[1])
	if err != nil {
		retStr = "Second argument (quantity to add) must be an integer string"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	if quantity <= 0 {
		retStr = "Second argument (quantity to add) must not be less than zero"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	// available offers
	offersAsBytes, err := stub.GetState(offersKey)
	if err != nil {
		retStr = "Could not get offersKey from chaincode state"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	json.Unmarshal(offersAsBytes, &offers)

	//["4","100"]
	if _, ok := offers[offerID]; ok {
		offers[offerID] += quantity
	} else {
		offers[offerID] = quantity
	}

	marshalAndPut(stub, offersKey, offers)
	retStr = "Successfully added " + args[1] + " to offer " + args[0]
	fmt.Println(retStr)
	return []byte(retStr), nil

}
func subtractOfferQuantity(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var retStr string
	var offers map[string]int

	if len(args) != 2 {
		retStr = "Incorrect number of arguments. Expecting 2: offer ID, quantity to subtract from the offer"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	///  ["5","200"]
	// Check to make sure offer ID is a valid integer
	offerIDInt, err := strconv.Atoi(args[0])
	if err != nil {
		retStr = "Offer ID must be an integer string"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	if offerIDInt <= 0 {
		retStr = "Offer ID must not be less than zero"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	offerID := args[0]

	//  quantity to subtract is valid?
	quantity, err := strconv.Atoi(args[1])
	if err != nil {
		retStr = "Second argument (quantity to subtract) must be an integer string"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	if quantity <= 0 {
		retStr = "Second argument (quantity to subtract) must not be less than zero"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	//available offers
	offersAsBytes, err := stub.GetState(offersKey)
	if err != nil {
		retStr = "Could not get offersKey from chaincode state"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	json.Unmarshal(offersAsBytes, &offers)

	if val, ok := offers[offerID]; ok {
		if quantity < val {
			offers[offerID] -= quantity
		} else {
			delete(offers, offerID)
		}
	} else {
		retStr = "Offer ID " + offerID + " does not exist"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	marshalAndPut(stub, offersKey, offers)

	retStr = "Successfully subtracted " + args[1] + " from offer " + offerID
	fmt.Println(retStr)
	return []byte(retStr), errors.New(retStr)

}

func addCustomer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var retStr string
	var err error
	var customers map[string]int

	if len(args) != 1 {
		retStr = "Incorrect number of arguments. Expecting 1: new customer ID"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	newCustomer := strings.ToLower(args[0])

	// Get the list of customers
	customerListBytes, err := stub.GetState(customersKey)
	if err != nil {
		retStr = "Could not get customersKey "
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	json.Unmarshal(customerListBytes, &customers)

	//  new customer already exists
	if _, ok := customers[newCustomer]; ok {
		retStr = "Cannot add customer '" + newCustomer + "': customer already exists"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	// init value of balance
	customers[newCustomer] = 0
	marshalAndPut(stub, customersKey, customers)
	fmt.Println("Successfully added new customer")
	retStr = "Successfully added new customer"
	return []byte(retStr), nil

}

// Add amount to customer's balance
func addCustomerFunds(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var retStr string
	var err error
	var customers map[string]int

	if len(args) != 2 {
		retStr = "Incorrect number of arguments. Expecting 2: customer ID, amount to add"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	if len(args[0]) == 0 {
		retStr = "First argument (customer ID) cannot be an empty string"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	if len(args[1]) == 0 {
		retStr = "Second argument (amount to add) cannot be an empty string"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	customerName := strings.ToLower(args[0])
	funds, err := strconv.Atoi(args[1])
	if err != nil {
		retStr = "funds must be a numeric string"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	// Funds is valid?
	if funds <= 0 {
		retStr = "funds must not be negative"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	customerListBytes, err := stub.GetState(customersKey)
	if err != nil {
		retStr = "Could not get customersKey from chaincode state"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	json.Unmarshal(customerListBytes, &customers)
	/////// add
	if _, ok := customers[customerName]; ok {

		customers[customerName] += funds

		marshalAndPut(stub, customersKey, customers)

		fmt.Println("Successfully added " + strconv.Itoa(funds) + " to " + customerName + "'s balance")
		retStr = "Successfully added " + strconv.Itoa(funds) + " to " + customerName + "'s balance"
		return []byte(retStr), nil
	} else {
		// Customer wasn't found, return error message
		retStr = "Could not find customer " + customerName + " to add funds"
		fmt.Println(retStr)
		return []byte(retStr), nil
	}

}

func acceptOffer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var retStr string
	var pendingTransaction []Transaction
	var newTransaction Transaction
	var offers map[string]int
	var customers map[string]int

	// Check parameters ali ->100
	if len(args) != 2 {
		retStr = "Incorrect number of arguments. Expecting  customer ID, units of energy "
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	fmt.Println("Checking if there is a pending transaction")
	pendingTransactionsBytes, err := stub.GetState(pendingTransactionKey)
	if err != nil {
		retStr = "Could not get pendingTransactionsKey from chaincode state"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	json.Unmarshal(pendingTransactionsBytes, &pendingTransaction)
	//There is already a pending transaction
	if len(pendingTransaction) > 0 {
		retStr = "Cannot accept an offer while a transaction is in progress."
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	buyer := strings.ToLower(args[0])
	requestedQuantity, _ := strconv.Atoi(args[1])

	if requestedQuantity <= 0 {
		retStr = "amount of energy cannot be less than 0"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	// requestedQuantity will be altered
	newTransaction.Energy = requestedQuantity

	//  available offers
	offerListBytes, err := stub.GetState(offersKey)
	if err != nil {
		retStr = "Could not get offersKey from chaincode state"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	json.Unmarshal(offerListBytes, &offers)

	totalAvailable := 0
	for i, val := range offers {
		totalAvailable += val
		fmt.Println("Key: " + i + ", Value: " + strconv.Itoa(val) + ". Total available is now " + strconv.Itoa(totalAvailable))
	}
	if totalAvailable < requestedQuantity {
		retStr = "Requested " + args[1] + " with only " + strconv.Itoa(totalAvailable) + " available"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	//  valid customers
	customerListBytes, err := stub.GetState(customersKey)
	if err != nil {
		retStr = "Could not get customersKey from chaincode state"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	json.Unmarshal(customerListBytes, &customers)

	//  is buyer a valid customer
	if _, ok := customers[buyer]; !ok {
		retStr = args[0] + " is not a valid buyer"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	newTransaction.Offers = make(map[string]int)
	// Make an array of price per unit offers in ascending order
	ascendingOfferKeys := getMapStringKeysAsAscendingInts(offers)
	totalCost := 0
	//pricePerUnit is the key
	for _, pricePerUnit := range ascendingOfferKeys {
		pricePerUnitStr := strconv.Itoa(pricePerUnit)
		unitsAvailable := offers[pricePerUnitStr]
		if unitsAvailable > requestedQuantity {

			offers[pricePerUnitStr] -= requestedQuantity
			totalCost += requestedQuantity * pricePerUnit
			newTransaction.Offers[pricePerUnitStr] = requestedQuantity
			// Done
			break
		} else {

			requestedQuantity -= unitsAvailable
			totalCost += unitsAvailable * pricePerUnit
			newTransaction.Offers[pricePerUnitStr] = unitsAvailable
			// Delete the map key
			delete(offers, pricePerUnitStr)
			if requestedQuantity == 0 {
				break
			}
			// Continue to the next upper unit price
		}
	}

	// customer has enough funds?
	if customers[buyer] < totalCost {
		retStr = "Buyer does not have enough funds: total cost = " + strconv.Itoa(totalCost) + ", available funds = " + strconv.Itoa(customers[buyer])
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	// Subtract funds from customer and add funds to "owner"
	customers[buyer] -= totalCost
	customers["owner"] += totalCost

	// Field update
	newTransaction.Status = "Pending"
	newTransaction.Buyer = buyer
	newTransaction.Cost = totalCost
	// newTransaction.Energy
	// newTransaction.Offers

	newTransaction.TXID = 0

	// Update pending transactions
	fmt.Println("Adding new transaction to pending transaction")
	pendingTransaction = append(pendingTransaction, newTransaction)
	err = marshalAndPut(stub, pendingTransactionKey, pendingTransaction)
	if err != nil {
		retStr = "Could not write pendingTransactionKey to chaincode "
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	// Update customer accounts
	fmt.Println("Writing updated customer list to chaincode state")
	err = marshalAndPut(stub, customersKey, customers)
	if err != nil {
		retStr = "Could not write customersKey to chaincode state"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	// Update available offers
	fmt.Println("Writing updated available offers to chaincode state")
	err = marshalAndPut(stub, offersKey, offers)
	if err != nil {
		retStr = "Could not write offersKey to chaincode state"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	retStr = "Successfully accepted the offer"
	fmt.Println(retStr)
	return []byte(retStr), nil

}
func completeTransaction(stub shim.ChaincodeStubInterface) ([]byte, error) {

	var retStr string
	var pendingTransaction []Transaction
	var newTransaction Transaction
	var pastTransactions []Transaction

	// pending transaction ?
	pendingTransactionsBytes, err := stub.GetState(pendingTransactionKey)
	if err != nil {
		retStr = "Could not get pendingTransactionsKey from chaincode"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	json.Unmarshal(pendingTransactionsBytes, &pendingTransaction)
	if len(pendingTransaction) == 0 {
		retStr = "No pending transaction "
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	//transaction must added to the transactions list
	newTransaction = pendingTransaction[0]
	newTransaction.Status = "Completed"
	// the current UTC timestamp
	newTransaction.TXID = time.Now().Unix()

	//  list of past transactions
	transactionListBytes, err := stub.GetState(transactionsKey)
	if err != nil {
		retStr = "Could not get transactionsKey from chaincode state"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	json.Unmarshal(transactionListBytes, &pastTransactions)

	// Append the new transaction to the list of completed transactions
	pastTransactions = append(pastTransactions, newTransaction)
	err = marshalAndPut(stub, transactionsKey, pastTransactions)
	if err != nil {
		retStr = "Could not write pendingTransactionsKey to chaincode state"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}

	// Clear the pending transaction
	var emptyTransactions []Transaction
	err = marshalAndPut(stub, pendingTransactionKey, emptyTransactions)

	if err != nil {
		retStr = "Could not write pendingTransactionsKey to chaincode"
		fmt.Println(retStr)
		return []byte(retStr), errors.New(retStr)
	}
	retStr = "Successfully completed the pending transaction"
	fmt.Println(retStr)
	return []byte(retStr), nil

}

/////////////////////////////////////////////////////////////
func marshalAndPut(stub shim.ChaincodeStubInterface, key string, v interface{}) error {

	var err error
	jsonAsBytes, _ := json.Marshal(v)
	err = stub.PutState(key, jsonAsBytes)
	if err != nil {
		return err
	}
	return nil

}
func createQueryResponseString(success bool, data string) ([]byte, error) {
	var response QueryResponseString
	response.Success = success
	response.Data = data
	r, _ := json.Marshal(response)
	if success {
		return r, nil
	} else {
		return r, errors.New(data)
	}
}
func createQueryResponseMap(success bool, data map[string]int) ([]byte, error) {
	var response QueryResponseMap
	response.Success = success
	response.Data = data
	r, _ := json.Marshal(response)
	return r, nil
}

func createQueryResponseInt(success bool, data int) ([]byte, error) {
	var response QueryResponseInt
	response.Success = success
	response.Data = data
	r, _ := json.Marshal(response)
	return r, nil
}

func createQueryResponseTransactions(success bool, data []Transaction) ([]byte, error) {
	var response QueryResponseTransactions
	response.Success = success
	response.Data = data
	r, _ := json.Marshal(response)
	return r, nil
}

func createQueryResponseBytes(success bool, data []byte) ([]byte, error) {
	var response QueryResponseBytes
	response.Success = success
	response.Data = data
	r, _ := json.Marshal(response)
	return r, nil
}
func getMapStringKeysAsAscendingInts(m map[string]int) []int {
	// Create keys int array
	keys := make([]int, len(m))
	i := 0

	for j := range m {
		keys[i], _ = strconv.Atoi(j)
		i++
	}
	// Sort ascending
	sort.Ints(keys)
	//fmt.Println("Sorted integers:", keys)
	return keys
}
func main() {
	err := shim.Start(new(LocalEnergyTrade))
	if err != nil {
		fmt.Printf("Error starting Local Energy Trade: %s", err)
	}
}
