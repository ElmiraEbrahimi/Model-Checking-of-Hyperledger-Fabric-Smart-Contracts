package main

import (
	"fmt"
	// JSON Encoding
	"encoding/json"
	"strconv"

	//Fabric 2.0 Shim
	"github.com/hyperledger/fabric-chaincode-go/shim"

	peer "github.com/hyperledger/fabric-protos-go/peer"
	// KV Interface
)

type casino struct {
}
type csnOwner struct {
	ObjectType string `json:"Type"`
	UID        string `json:"uid"`
	Password   string `json:"password"`
	Balance    uint64 `json:"balance"`
	CoinResult bool   `json:"coinResult"`
	GameState  uint   `json:"gameState"`
}

type player struct {
	ObjectType   string `json:"Type"`
	UID          string `json:"uid"`
	Password     string `json:"password"`
	Balance      uint64 `json:"balance"`
	GuessedValue bool   `json:"guessedValue"`
	BetValue     uint64 `json:"betValue"`
}

//define constant
const stateGameStoped uint = 0
const stateGameStarted uint = 1
const stateGameBetPlaced uint = 2
const WithdrawLimit uint64 = 2000

var nonce uint = 10

func (t *casino) Init(stub shim.ChaincodeStubInterface) peer.Response {

	fmt.Println("Init executed for casino contract!")
	return shim.Success(nil)
}

func (t *casino) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	function, args := stub.GetFunctionAndParameters()
	fmt.Println("Chaincode Invoke Is executing " + function)

	if function == "casino" {
		return t.casino(stub, args)
	}
	if function == "registerPlayer" {
		return t.registerPlayer(stub, args)
	}

	if function == "withdraw" {
		return t.withdraw(stub, args)
	}
	if function == "deposit" {
		return t.deposit(stub, args)
	}
	if function == "tossACoin" {
		return t.tossACoin(stub, args)
	}

	if function == "placeBet" {
		return t.placeBet(stub, args)
	}
	if function == "endGame" {
		return t.endGame(stub, args)
	}

	fmt.Println("Bad Function Name" + function)
	return shim.Error("Invoke did not find this function " + function)
}

func (t *casino) casino(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect Number of Aruments")
	}

	var UID string = "1"
	Password := args[1]

	Balance, _ := strconv.ParseUint(args[2], 10, 64)
	CoinResult := false
	GameState := stateGameStoped
	// Check if owner Already exists return error
	ownerAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Transaction Failed with Error: " + err.Error())
	} else if ownerAsBytes != nil {
		return shim.Error("The Inserted owner already Exists")
	}

	// Create owner Object and Marshal to JSON once
	objectType := "owner"
	owner := &csnOwner{objectType, UID, Password, Balance, CoinResult, GameState}
	ownerJSONasBytes, err := json.Marshal(owner)

	if err != nil {
		return shim.Error(err.Error())
	}

	//Save owner to State
	err = stub.PutState(UID, ownerJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Return Success
	fmt.Println("Successfully Saved owner")
	return shim.Success(nil)
}

func (t *casino) registerPlayer(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var err error
	if len(args) != 2 {
		return shim.Error("Incorrect Number of Aruments")
	}

	UID := args[0]
	Password := args[1]
	var Balance uint64 = 10
	var BetValue uint64 = 0
	var GuessedValue bool = false
	// Check if player Already registers
	playerAsBytes, err := stub.GetState(UID)
	if err != nil {
		return shim.Error("Transaction Failed with Error: " + err.Error())
	} else if playerAsBytes != nil {
		return shim.Error("The Inserted player already Exists")
	}

	// Create player Object and Marshal to JSON once
	objectType := "player"
	player := &player{objectType, UID, Password, Balance, GuessedValue, BetValue}
	playerJSONasBytes, err := json.Marshal(player)

	if err != nil {
		return shim.Error(err.Error())
	}

	//Save player to State
	err = stub.PutState(UID, playerJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Return Success
	fmt.Println("Successfully Saved player")
	return shim.Success(nil)
}

func (t *casino) withdraw(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments")
	}
	// get the state information
	bytes, _ := stub.GetState(args[0])
	if bytes == nil {
		return shim.Error("Provided UID not found!!!")
	}
	// get ammount as string and convert it to uint
	Amount, _ := strconv.ParseUint(args[2], 10, 64)
	// unmarshall the data
	// Read the JSON and Initialize the struct
	var owner csnOwner
	_ = json.Unmarshal(bytes, &owner)

	if owner.Password != args[1] {
		return shim.Error("Current owner MUST match !!!")
	}
	if Amount > WithdrawLimit || Amount < 0 {
		return shim.Error("bad request...exceed Withdraw Limit")
	}
	//"Subtraction: balance underflow"
	if Amount > owner.Balance {
		return shim.Error("error..Bad behavior !!!")
	}
	if owner.GameState != stateGameStoped {
		return shim.Error("error..can't withdraw now...try later !!!")
	} else {
		owner.Balance -= Amount
	}
	jsonOwner, _ := json.Marshal(owner)

	stub.PutState(owner.UID, jsonOwner)

	return shim.Success([]byte("Balance Record Updated!!! " + string(jsonOwner)))
}

func (t *casino) deposit(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments")
	}
	// get the state information
	bytes, _ := stub.GetState(args[0])
	if bytes == nil {
		return shim.Error("Provided UID not found!!!")
	}
	// get ammount as string and convert it to uint
	Amount, _ := strconv.ParseUint(args[2], 10, 64)
	// unmarshall the data
	// Read the JSON and Initialize the struct
	var owner csnOwner
	_ = json.Unmarshal(bytes, &owner)

	if owner.Password != args[1] {
		return shim.Error("Current owner MUST match !!!")
	}

	//balance overflow
	if owner.Balance+Amount < owner.Balance {
		return shim.Error("error..Bad behavior !!!")
	} else {
		owner.Balance += Amount
	}
	jsonOwner, _ := json.Marshal(owner)

	stub.PutState(owner.UID, jsonOwner)

	return shim.Success([]byte("Balance Record Updated!!! " + string(jsonOwner)))
}

func (t *casino) tossACoin(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err2 error
	var secretNum int64
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments")
	}
	// get the state information
	bytes, _ := stub.GetState(args[0])
	if bytes == nil {
		return shim.Error("Provided UID not found!!!")
	}
	// unmarshall the data
	// Read the JSON and Initialize the struct
	var owner csnOwner
	_ = json.Unmarshal(bytes, &owner)

	if owner.Password != args[1] {
		return shim.Error("Current owner MUST match !!!")
	}
	///////////////////////////////////////////////////////////////////////
	if owner.GameState != stateGameStoped {
		return shim.Error("error..can't toss a coin now...try later !!!")
	} else {
		owner.GameState = stateGameStarted
		fmt.Printf("Game started")

	}

	ts, err2 := stub.GetTxTimestamp()

	if err2 != nil {
		fmt.Printf("Error getting transaction timestamp: %s", err2)
	}
	secretNum = (ts.seconds + nonce) / 10
	//owner.CoinResult = (secretNum % 2 == 1)? true: false
	if secretNum%2 == 1 {
		owner.CoinResult = true
	} else {
		owner.CoinResult = false
	}

	nonce += 1
	jsonOwner, _ := json.Marshal(owner)
	stub.PutState(owner.UID, jsonOwner)
	return shim.Success([]byte("owner toss a coin!!! " + string(jsonOwner)))
}

func (t *casino) placeBet(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 5 {
		return shim.Error("Incorrect number of arguments")
	}
	cUID := args[0]
	pUID := args[1]
	pPassword := args[2]
	pGuessedValue, _ := strconv.ParseBool(args[3])
	pBetValue, _ := strconv.ParseUint(args[4], 10, 64)

	// get the state information
	ownerAsBytes, _ := stub.GetState(cUID)
	if ownerAsBytes == nil {
		return shim.Error("Provided UID not found!!!")
	}

	var owner csnOwner
	_ = json.Unmarshal(ownerAsBytes, &owner)

	if owner.GameState != stateGameStarted {
		return shim.Error("The casino owner has not tossed a coin yet.try later !!!")
	}

	playerAsBytes, _ := stub.GetState(pUID)
	if playerAsBytes == nil {
		return shim.Error("Provided UID not found!!!")
	}
	var player player
	_ = json.Unmarshal(playerAsBytes, &player)

	if player.Password != pPassword {
		return shim.Error("Current player MUST match !!!")
	}

	if player.Balance < pBetValue {
		return shim.Error("Your account does not have enough money !!!")
	} else {

		player.BetValue = pBetValue
		// must check overflow and underflow
		player.Balance -= pBetValue
		owner.Balance += pBetValue
	}

	player.GuessedValue = pGuessedValue
	owner.GameState = stateGameBetPlaced

	jsonOwner, _ := json.Marshal(owner)
	stub.PutState(owner.UID, jsonOwner)

	jsonPlayer, _ := json.Marshal(player)
	stub.PutState(player.UID, jsonPlayer)
	// Emit bet Event
	eventPayload := "{\"player bet\":\"" + pUID + "\"}"
	stub.SetEvent("betPlaced", eventPayload)
	return shim.Success("you placed a bet..!!! ")
}

func (t *casino) endGame(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	pUID := args[3]
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments")
	}
	// get the state information
	ownerAsBytes, _ := stub.GetState(args[0])
	if ownerAsBytes == nil {
		return shim.Error("Provided UID not found!!!")
	}

	var owner csnOwner
	_ = json.Unmarshal(ownerAsBytes, &owner)
	if owner.Password != args[1] {
		return shim.Error("Current owner MUST match !!!")
	}

	playerAsBytes, _ := stub.GetState(args[3])
	if playerAsBytes == nil {
		return shim.Error("Provided UID not found!!!")
	}
	var player player
	_ = json.Unmarshal(playerAsBytes, &player)

	if owner.GameState == stateGameBetPlaced {
		if owner.CoinResult == player.GuessedValue {
			// must check overflow and underflow
			owner.Balance -= (18 * player.BetValue / 10)
			player.Balance += (18 * player.BetValue / 10)

			eventPayload := "{\"The current player wins  \":\"" + pUID + "\"}"
			stub.SetEvent("winnerAnnounced", eventPayload)
		}
		owner.GameState = stateGameStoped
	}

	jsonOwner, _ := json.Marshal(owner)
	stub.PutState(owner.UID, jsonOwner)

	jsonPlayer, _ := json.Marshal(player)
	stub.PutState(player.UID, jsonPlayer)
	return shim.Success("you placed a bet..!!! ")

}

//Main Function starts up the Chaincode
func main() {
	err := shim.Start(new(casino))
	if err != nil {
		fmt.Printf("Smart Contract could not be run. Error Occured: %s", err)
	} else {
		fmt.Println("Smart Contract successfully Initiated")
	}
}
