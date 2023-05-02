package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	//Fabric 2.0 Shim
	"github.com/hyperledger/fabric-chaincode-go/shim"

	peer "github.com/hyperledger/fabric-protos-go/peer"
)

type LocalEnergy struct {
}
type Prosumer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Coins  uint64 `json:"coins"`
	Tokens uint64 `json:"Tokens"`
	Status bool   `json:"Status"`
}

type Transaction struct {
	Timestamp      string `json:"timestamp"`
	Consumer       string `json:"consumer"`
	Producer       string `json:"producer"`
	RequiredTokens uint64 `json:"RequiredTokens"`
	Price          uint64 `json:"price"`
}

func (t *LocalEnergy) Init(stub shim.ChaincodeStubInterface) peer.Response {

	fmt.Println("Init executed for LocalEnergy contract!")
	return shim.Success(nil)
}

func (t *LocalEnergy) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	function, args := stub.GetFunctionAndParameters()
	fmt.Println("Chaincode Invoke Is executing " + function)

	// register a participant by ID
	if function == "register" {
		return t.register(stub, args)
	}

	if function == "Energymarket" {
		return t.Energymarket(stub, args)
	}
	if function == "Endmarket" {
		return t.Endmarket(stub, args)
	}

	fmt.Println("Bad Function Name" + function)
	return shim.Error("Invoke did not find this function " + function)

}

func (t *LocalEnergy) register(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect Number of Arguments")
	}

	ID := args[0]
	Name := args[1]
	Coins, _ := strconv.ParseUint(args[2], 10, 64)
	Tokens, _ := strconv.ParseUint(args[3], 10, 64)
	Status := false
	fmt.Println("registration started")
	if len(ID) <= 0 {
		return shim.Error("id must be a non-empty integer")
	}
	// Check if player Already registers
	prosumerAsBytes, err := stub.GetState(ID)
	if err != nil {
		return shim.Error("Transaction Failed with Error: " + err.Error())
	} else if prosumerAsBytes != nil {
		return shim.Error("The Inserted Prosumer already Exists")
	}

	// Create Prosumer Object and Marshal to JSON once
	Prosumer := &Prosumer{ID, Name, Coins, Tokens, Status}
	prosumerJSONasBytes, err := json.Marshal(Prosumer)

	if err != nil {
		return shim.Error(err.Error())
	}

	//Save Prosumer to State
	err = stub.PutState(ID, prosumerJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Return Success
	fmt.Println("Successfully Saved Prosumer")
	return shim.Success(nil)
}

func (t *LocalEnergy) Energymarket(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect Number of Arguments")
	}
	Timestamp, _ := stub.GetTxTimestamp()

	// ======Check if Transaction Already exists

	transactionsAsBytes, err := stub.GetState(Timestamp)
	if err != nil {
		return shim.Error("Transaction Failed with Error: " + err.Error())
	} else if transactionsAsBytes != nil {
		return shim.Error("The Inserted transaction already Exists: " + Timestamp)
	}
	Consumer := args[0]
	Producer := args[1]
	RequiredTokens, _ := strconv.ParseUint(args[2], 10, 64)
	Price := RequiredTokens * 100
	// check if customer has enough Coins for purchase

	ConsumerAsBytes, _ := stub.GetState(Consumer)
	if ConsumerAsBytes == nil {
		return shim.Error("Provided Consumer not found!!!")
	}

	var Prosumer Prosumer
	_ = json.Unmarshal(ConsumerAsBytes, &Prosumer)

	if Prosumer.Coins <= Price {
		return shim.Error("Your account balance is not enough to buy energy units !!!")
	}
	Prosumer.Status = true
	jsonP, _ := json.Marshal(Prosumer)

	stub.PutState(Prosumer.ID, jsonP)

	// Create Transaction Object and Marshal to JSON once
	Transaction := &Transaction{Timestamp, Consumer, Producer, RequiredTokens, Price}
	TransactionJSONasBytes, err := json.Marshal(Transaction)

	if err != nil {
		return shim.Error(err.Error())
	}

	//Save Transaction to State
	err = stub.PutState(Timestamp, TransactionJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return Success
	fmt.Println("Successfully Saved  Transaction with Timestamp:" + Timestamp)
	return shim.Success(nil)

}

func (t *LocalEnergy) Endmarket(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect Number of Arguments")
	}
	Timestamp1 := args[0]
	Timestamp2 := args[1]

	Transaction1AsBytes, err := stub.GetState(Timestamp1)
	if err != nil {
		return shim.Error("Transaction Failed with Error: " + err.Error())
	} else if Transaction1AsBytes == nil {
		return shim.Error("Provided Transaction not found!!!")
	}
	var TxcConsumer1 Transaction
	_ = json.Unmarshal(Transaction1AsBytes, &TxcConsumer1)

	Transaction2AsBytes, err := stub.GetState(Timestamp2)
	if err != nil {
		return shim.Error("Transaction Failed with Error: " + err.Error())
	} else if Transaction2AsBytes == nil {
		return shim.Error("Provided Transaction not found!!!")
	}

	var TxcConsumer2 Transaction
	_ = json.Unmarshal(Transaction2AsBytes, &TxcConsumer2)

	ProducerAsBytes, err := stub.GetState(TxcConsumer1.Producer)
	if err != nil {
		return shim.Error("Transaction Failed with Error: " + err.Error())
	} else if ProducerAsBytes == nil {
		return shim.Error("Provided Producer not found!!!")
	}

	var Producer Prosumer
	_ = json.Unmarshal(ProducerAsBytes, &Producer)

	BOBAsBytes, err := stub.GetState(TxcConsumer1.Consumer)
	if err != nil {
		return shim.Error("Transaction Failed with Error: " + err.Error())
	} else if ProducerAsBytes == nil {
		return shim.Error("Provided Consumer not found!!!")
	}

	var BOB Prosumer
	_ = json.Unmarshal(BOBAsBytes, &BOB)

	TOMAsBytes, err := stub.GetState(TxcConsumer2.Consumer)
	if err != nil {
		return shim.Error("Transaction Failed with Error: " + err.Error())
	} else if ProducerAsBytes == nil {
		return shim.Error("Provided Consumer not found!!!")
	}

	var TOM Prosumer
	_ = json.Unmarshal(TOMAsBytes, &TOM)

	if Producer.Tokens > TxcConsumer1.RequiredTokens+TxcConsumer2.RequiredTokens {
		// must check overflow and underflow
		Producer.Coins += TxcConsumer1.Price
		Producer.Tokens -= TxcConsumer1.RequiredTokens
		BOB.Tokens += TxcConsumer1.RequiredTokens
		BOB.Status = false
		Producer.Coins += TxcConsumer2.Price
		Producer.Tokens -= TxcConsumer2.RequiredTokens
		TOM.Tokens += TxcConsumer2.RequiredTokens
		TOM.Status = false
		eventPayload := "{\"The tokens were transferred from Pruducer \":\"" + Producer.ID + "\"}"
		stub.SetEvent("MarketAnnounced", eventPayload)
	}
	if Producer.Tokens <= TxcConsumer1.RequiredTokens+TxcConsumer2.RequiredTokens {
		fairsell1 := (Producer.Tokens * TxcConsumer1.RequiredTokens) / (TxcConsumer1.RequiredTokens + TxcConsumer2.RequiredTokens)
		fairsell2 := (Producer.Tokens * TxcConsumer2.RequiredTokens) / (TxcConsumer1.RequiredTokens + TxcConsumer2.RequiredTokens)
		Producer.Tokens -= fairsell1
		BOB.Tokens += fairsell1
		BOB.Status = false
		Producer.Coins += (fairsell1 + fairsell2) * 100
		Producer.Tokens -= fairsell2
		TOM.Tokens += fairsell2
		TOM.Status = false
	}

	jsonProducer, _ := json.Marshal(Producer)
	stub.PutState(Producer.ID, jsonProducer)

	jsonBOB, _ := json.Marshal(BOB)
	stub.PutState(BOB.ID, jsonBOB)

	jsonTOM, _ := json.Marshal(TOM)
	stub.PutState(TOM.ID, jsonTOM)

	return shim.Success("market closed successfully..!!! ")
}

//Main Function starts up the Chaincode
func main() {
	err := shim.Start(new(LocalEnergy))
	if err != nil {
		fmt.Printf("Smart Contract could not be run. Error Occured: %s", err)
	} else {
		fmt.Println("Smart Contract successfully Initiated")
	}
}
