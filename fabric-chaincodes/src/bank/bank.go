package main

import (
	"fmt"
	// JSON Encoding
	"bytes"
	"encoding/json"
	"strconv"

	//Fabric 2.0 Shim
	"github.com/hyperledger/fabric-chaincode-go/shim"

	peer "github.com/hyperledger/fabric-protos-go/peer"
	// KV Interface
)

type smartBank struct {
}

type User struct {
	ObjectType string `json:"Type"`
	UID        string `json:"uid"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Balance    uint64 `json:"balance"`
}

type Transactions struct {
	ObjectType     string `json:"Type"`
	Transaction_ID string `json:"transaction_ID"`
	To             string `json:"to"`
	From           string `json:"from"`
	Amount         uint64 `json:"amount"`
}

//define constant
//const	DepositLimit uint64	:= math.Max(uint64)
//const DepositLimit uint64 = 184467440737095516
const DepositLimit uint64 = 100000
const WithdrawLimit uint64 = 4000
const TransferLimit uint64 = 3000

func (t *smartBank) Init(stub shim.ChaincodeStubInterface) peer.Response {

	fmt.Println("Init executed for Bank contract!")
	t.SetupSampleData(stub)
	return shim.Success(nil)
}

func (t *smartBank) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	// Retrieve the Contract function and arguments
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("Chaincode Invoke Is executing " + function)

	if function == "addUser" {
		return t.addUser(stub, args)
	}
	if function == "queryUser" {
		return t.queryUser(stub, args)
	}
	if function == "Deposit" {
		return t.Deposit(stub, args)
	}
	if function == "Withdrawl" {
		return t.Withdrawl(stub, args)
	}
	if function == "Transfer" {
		return t.Transfer(stub, args)
	}
	if function == "balanceOf" {
		return t.balanceOf(stub, args)
	}
	if function == "queryTransactionsFrom" {
		return t.queryTransactionsFrom(stub, args)
	}
	if function == "queryTransactionsTo" {
		return t.queryTransactionsTo(stub, args)
	}
	if function == "updatePssword" {
		return t.updatePssword(stub, args)
	}

	fmt.Println("Bad Function Name" + function)
	return shim.Error("Invoke did not find this function " + function)
}

func (t *smartBank) addUser(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var err error

	if len(args) != 5 {
		return shim.Error("Incorrect Number of Aruments")
	}

	fmt.Println("Adding new User")
	if len(args[0]) <= 0 {
		return shim.Error("1st Argument Must be a Non-Empty String")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd Argument Must be a Non-Empty String")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd Argument Must be a Non-Empty String")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th Argument Must be a Non-Empty String")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th Argument Must be a Non-Empty String")
	}

	UID := args[0]
	name := args[1]
	email := args[2]
	Password := args[3]
	Balance, _ := strconv.ParseUint(args[2], 10, 64)

	// Check if User Already exists
	userAsBytes, err := stub.GetState(UID)
	if err != nil {
		return shim.Error("Transaction Failed with Error: " + err.Error())
	} else if userAsBytes != nil {
		return shim.Error("The Inserted User already Exists: " + UID)
	}

	// Create User Object and Marshal to JSON
	objectType := "user"
	user := &User{objectType, UID, name, email, Password, Balance}
	userJSONasBytes, err := json.Marshal(user)

	if err != nil {
		return shim.Error(err.Error())
	}

	//Save User to State
	err = stub.PutState(UID, userJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Return Success
	fmt.Println("Successfully Saved User")
	return shim.Success(nil)
}

func (t *smartBank) queryUser(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments")
	}

	UID := args[0]
	Password := args[1]

	queryString := fmt.Sprintf("{\"selector\":{\"Type\":\"user\",\"UID\":\"%s\",\"password\":\"%s\"}}", UID, Password)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (t *smartBank) Deposit(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments")
	}
	// get the state information
	bytes, _ := stub.GetState(args[0])
	if bytes == nil {
		return shim.Error("Provided UID not found!!!")
	}
	Amount, _ := strconv.ParseUint(args[2], 10, 64)
	// unmarshall the data
	// Read the JSON and Initialize the struct
	var user User
	_ = json.Unmarshal(bytes, &user)

	if user.Password != args[1] {
		return shim.Error("Current owner MUST match !!!")
	}
	if Amount > DepositLimit {
		return shim.Error("exceed Deposit Limit")
	}
	//balance overflow
	if user.Balance+Amount < user.Balance {
		return shim.Error("error..Bad behavior !!!")
	} else {
		user.Balance += Amount
	}

	jsonUser, _ := json.Marshal(user)

	stub.PutState(user.UID, jsonUser)

	return shim.Success([]byte("Balance Record Updated!!! " + string(jsonUser)))
}

func (t *smartBank) Withdrawl(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments")
	}
	// get the state information
	bytes, _ := stub.GetState(args[0])
	if bytes == nil {
		return shim.Error("Provided UID not found!!!")
	}
	Amount, _ := strconv.ParseUint(args[2], 10, 64)
	// unmarshall the data
	// Read the JSON and Initialize the struct
	var user User
	_ = json.Unmarshal(bytes, &user)

	if user.Password != args[1] {
		return shim.Error("Current owner MUST match !!!")
	}
	if Amount > WithdrawLimit || Amount < 0 {
		return shim.Error("bad request...exceed Withdraw Limit")
	}
	//"Subtraction: balance underflow"
	if Amount > user.Balance {
		return shim.Error("error..Bad behavior !!!")
	} else {
		user.Balance -= Amount
	}
	jsonUser, _ := json.Marshal(user)

	stub.PutState(user.UID, jsonUser)

	return shim.Success([]byte("Balance Record Updated!!! " + string(jsonUser)))
}

func (t *smartBank) Transfer(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var err error

	if len(args) != 5 {
		return shim.Error("Incorrect Number of Aruments")
	}

	fmt.Println("Adding new Transaction")
	// all argument must be set
	if len(args[0]) <= 0 {
		return shim.Error("1st Argument Must be a Non-Empty String")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd Argument Must be a Non-Empty String")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd Argument Must be a Non-Empty String")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th Argument Must be a Non-Empty String")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th Argument Must be a Non-Empty String")
	}

	transaction_ID := args[0]
	from := args[1]
	to := args[2]
	amount, _ := strconv.ParseUint(args[3], 10, 64)

	//Check if Transaction Already exists
	transactionsAsBytes, err := stub.GetState(transaction_ID)
	if err != nil {
		return shim.Error("Transaction Failed with Error: " + err.Error())
	} else if transactionsAsBytes != nil {
		return shim.Error("The Inserted transaction already Exists: " + transaction_ID)
	}

	if amount > TransferLimit || amount < 0 {
		return shim.Error("bad request...exceed Transfer Limit")
	}

	fromUserAsBytes, _ := stub.GetState(args[1])
	if fromUserAsBytes == nil {
		return shim.Error("Provided UID not found!!!")
	}
	// unmarshall  from user's data
	var fromUser User
	_ = json.Unmarshal(fromUserAsBytes, &fromUser)

	toUserAsBytes, _ := stub.GetState(args[2])
	if toUserAsBytes == nil {
		return shim.Error("Provided UID not found!!!")
	}
	// unmarshall to user's data
	var toUser User
	_ = json.Unmarshal(toUserAsBytes, &toUser)

	if fromUser.Password != args[4] {
		return shim.Error("Current owner MUST match !!!")
	}
	//"Subtraction:  fromBalance underflow"
	if amount > fromUser.Balance {
		return shim.Error("error..Bad behavior !!!")
	}
	//toBalance overflow
	if toUser.Balance+amount < toUser.Balance {
		return shim.Error("error..Bad behavior !!!")
	} else {
		fromUser.Balance -= amount
		toUser.Balance += amount
	}
	//write all changed assets to the ledger **** it contains two unhandeled error
	jsonFromUser, _ := json.Marshal(fromUser)
	stub.PutState(fromUser.UID, jsonFromUser)

	jsonToUser, err := json.Marshal(toUser)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(toUser.UID, jsonToUser)
	if err != nil {
		return shim.Error("Error  in writing updates for toUser account " + toUser.UID)
	}

	// Create transactions Object and Marshal to JSON
	objectType := "transactions"
	transactions := &Transactions{objectType, transaction_ID, from, to, amount}
	transactionsJSONasBytes, err := json.Marshal(transactions)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Save transactions to State
	err = stub.PutState(transaction_ID, transactionsJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Emit Transfer Event
	eventPayload := "{\"from\":\"" + from + "\", \"to\":\"" + to + "\",\"amount\":" + strconv.FormatInt(int64(amount), 10) + "}"
	stub.SetEvent("Transfer", []byte(eventPayload))
	fmt.Println("Transfer transaction completed successfully")
	return shim.Success([]byte("true"))
}

func (t *smartBank) balanceOf(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// Check if userID and password is in the arguments
	if len(args) < 2 {
		return shim.Error("Incorrect Number of Aruments")
	}
	UID := args[0]
	Password := args[1]
	// get the state information
	bytes, _ := stub.GetState(args[0])
	if bytes == nil {
		return shim.Error("Provided UID not found!!!")
	}
	/*bytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Provided UID not found!!!")
	}*/
	// unmarshall the data
	var user User
	_ = json.Unmarshal(bytes, &user)
	if user.Password != Password {
		return shim.Error("Current owner MUST match !!!")
	}
	balanceJSON(UID, string(bytes))
	// Return success
	return shim.Success(nil)
}

// balanceJSON creates a JSON for representing the balance
func balanceJSON(OwnerID, balance string) string {
	return "{\"owner\":\"" + OwnerID + "\", \"balance\":" + balance + "}"
}

func (t *smartBank) queryTransactionsFrom(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	from := args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"Type\":\"transactions\",\"from\":\"%s\"}}", from)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (t *smartBank) queryTransactionsTo(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	to := args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"Type\":\"transactions\",\"to\":\"%s\"}}", to)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (t *smartBank) updatePssword(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error

	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments")
	}

	UID := args[0]
	//Password := args[1]
	newPassword := args[2]

	userAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get user:" + err.Error())
	} else if userAsBytes == nil {
		return shim.Error("User does not exist")
	}
	userToupdate := User{}
	err = json.Unmarshal(userAsBytes, &userToupdate) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	if userToupdate.Password != args[1] {
		return shim.Error("Current owner MUST match !!!")
	} else {
		userToupdate.Password = newPassword
	}

	userJSONasBytes, _ := json.Marshal(userToupdate)
	err = stub.PutState(UID, userJSONasBytes) //rewrite the conversion
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("User password Successfully Updated (success)")
	return shim.Success(nil)
}

// Result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

// SetupSampleData
func (t *smartBank) SetupSampleData(stub shim.ChaincodeStubInterface) {

	AddData(stub, "1", "milad", "milad@gmail.com", "P1", 100)
	AddData(stub, "2", "sajjad", "sajjad@gmail.com", "P2", 200)
	AddData(stub, "3", "reza", "reza@gmail.com", "P3", 300)
	AddData(stub, "4", "ali", "ali@gmail.com", "P4", 400)

	fmt.Println("Initialized with the sample data!!")
}

func AddData(stub shim.ChaincodeStubInterface, uid string, name string, email string, password string, balance uint64) {
	objectType := "user"
	user := &User{ObjectType: objectType, UID: uid, Name: name, Email: email, Password: password, Balance: balance}

	jsonUser, _ := json.Marshal(user)
	stub.PutState(uid, jsonUser)
}

//Main Function starts up the Chaincode
func main() {
	err := shim.Start(new(smartBank))
	if err != nil {
		fmt.Printf("Smart Contract could not be run. Error Occured: %s", err)
	} else {
		fmt.Println("Smart Contract successfully Initiated")
	}
}
