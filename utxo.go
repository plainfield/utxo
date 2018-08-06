/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/*
 * The sample smart contract for documentation topic:
 * Writing Your First Blockchain Application
 */

package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
        //"crypto/sha1"
        //"encoding/hex"
        //"time"
	"math"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

//Define the utxo structure
type Utxo struct {
  Txid    string `json:"txid"`
  Index   string `json:"index"`
  Amount  string `json:"amount"`
  Address string `json:"address"`
  Spent   string `json:"spent"`
}

// Define the tranction structure
type Transaction struct {
  Utxo_ins  []Utxo `json:"utxo_ins"`            
  Utxo_outs []Utxo `json:"utxo_outs"`
}

/*
 * The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryTransaction" {
		return s.queryTransaction(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "utxoTransfer" {
		return s.utxoTransfer(APIstub, args)
	} else if function == "queryUtxoByAddr" {
		return s.queryUtxoByAddr(APIstub, args)
	} else if function == "queryUtxo" {
		return s.queryUtxo(APIstub, args)
	}else if function == "queryAllUtxo" {
		return s.queryAllUtxo(APIstub)
	}else if function == "queryAllTransaction" {
		return s.queryAllTransaction(APIstub)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {

	txid := APIstub.GetTxID()

	//math.MaxUint32 means it's a coinbase
	utxo_ins := []Utxo{
		makeUtxo(txid,strconv.Itoa(math.MaxUint32),"50","System coinbase","true"),
	}

	utxo_outs := []Utxo{
		makeUtxo(txid,"1","50","UserA","false"),
	}

	transaction := makeTransaction(utxo_ins,utxo_outs)

	//store utxo
	storeUtxo(APIstub,txid,utxo_outs[0])
	//store transaction
	storeTransaction(APIstub,txid,transaction)

	return shim.Success(nil)
}

func storeUtxo(APIstub shim.ChaincodeStubInterface, txid string, utxo Utxo) error {

	utxoKey := txid+":"+utxo.Index
	utxoAsBYtes, _ := json.Marshal(utxo)

	//Utxo key is transaction_ID:index
	APIstub.PutState(utxoKey,utxoAsBYtes)

	return nil
}

func storeTransaction(APIstub shim.ChaincodeStubInterface, txid string, transaction Transaction) error {

	transactionAsBYtes, _ := json.Marshal(transaction)

	APIstub.PutState(txid,transactionAsBYtes)

	return nil
}

func makeUtxo(txid string, index string, amount string, address string, spent string) (Utxo) {

	//Index, Address, Amount, Spent
	utxo := Utxo{txid, index,amount,address,spent}

	return utxo
}

func makeTransaction(utxo_ins []Utxo, utxo_outs []Utxo) (Transaction) {

	transaction := Transaction{utxo_ins,utxo_outs}

	return transaction
}

func (s *SmartContract) queryAllTransaction(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := ""
	endKey := ""

	resultsIterator, err := APIstub.GetStateByRange(startKey,endKey)

	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		//skip utxo
		if strings.Contains(queryResponse.Key,":") {
			continue
		}
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllTransaction:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) queryAllUtxo(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := ""
	endKey := ""

	resultsIterator, err := APIstub.GetStateByRange(startKey,endKey)

	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		
		if strings.Contains(queryResponse.Key,":") {
			// Add a comma before array members, suppress it for the first array member
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"Key\":")
			buffer.WriteString("\"")
			buffer.WriteString(queryResponse.Key)
			buffer.WriteString("\"")

			buffer.WriteString(", \"Record\":")
			// Record is a JSON object, so we write as-is
			buffer.WriteString(string(queryResponse.Value))
			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
		}
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllUtxo:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) queryTransaction(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	transactionAsBytes, err := APIstub.GetState(args[0])

        if err != nil {
	  return shim.Error(err.Error())
        }

	return shim.Success(transactionAsBytes)
}

func (s *SmartContract) queryUtxo(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	utxoAsBytes, err := APIstub.GetState(args[0])

        if err != nil {
	  return shim.Error(err.Error())
        }

	return shim.Success(utxoAsBytes)
}

func (s *SmartContract) queryUtxoByAddr(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	address := args[0]

	//Get all Utxo 
	startKey := ""
	endKey := ""
	resultsIterator, err := APIstub.GetStateByRange(startKey,endKey)

	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
 
		if err != nil {
			return shim.Error(err.Error())
		} 
		
		if strings.Contains(queryResponse.Key,":") {
			var utxo Utxo
			json.Unmarshal(queryResponse.Value,&utxo)

			//check utxo owner ship, only return the one with address equal to with args[0]
			if checkUtxoOwnerShip(utxo, address) == false {
				continue
			}

			// Add a comma before array members, suppress it for the first array member
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"Key\":")
			buffer.WriteString("\"")
			buffer.WriteString(queryResponse.Key)
			//transaction value, Record is a JSON object, so we write as-is
			buffer.WriteString("\"")
			buffer.WriteString(", \"Record\":")
			buffer.WriteString(string(queryResponse.Value))

			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
		}
	}
	buffer.WriteString("]")

	fmt.Printf("- queryUtxoByAddr:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

//at this point, there is no encoding and decoding, so just check the address and spent
func checkUtxoOwnerShip(utxo Utxo, address string) bool {
	return utxo.Address == address && utxo.Spent == "false"
}

func (s *SmartContract) utxoTransfer(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	//from = args[0]
	//to = args[1]
	//amount = args[2]

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	addrFrom := args[0]
	addrTo := args[1]
	amount, _ := strconv.ParseFloat(args[2], 64)

	//get all utxo
	startKey := ""
	endKey := ""
	resultsIterator, err := APIstub.GetStateByRange(startKey,endKey)

	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultsIterator.Close()

	//the utxo selection rule is simple, add the utxo belong to addrFrom to utxo_int one by one until the accumulate value equal or greater than amount
	var utxo_ins = []Utxo{}
	var currValue float64 = 0.0
	txid := APIstub.GetTxID()

	for resultsIterator.HasNext() && currValue < amount{

		//loop the utxo one by one
		queryResponse, err := resultsIterator.Next()
 
		if err != nil {
			return shim.Error(err.Error())
		} 

		//prepare utxo to spend
		var utxo Utxo
		json.Unmarshal(queryResponse.Value,&utxo)

		//skip the utxo belong to others
		if checkUtxoOwnerShip(utxo, addrFrom) == false {
			continue
		}

		utxo_ins = append(utxo_ins,utxo)
	
		//delete utxo which have been spent
		APIstub.DelState(queryResponse.Key)

		//accumulate currValue
		newValue, _ := strconv.ParseFloat(utxo.Amount, 64)
		currValue += newValue
	}

	//no enough utxo to spend
	if currValue < amount {
		return shim.Error("No enough amount to spend")
	}

	//create new utxo_outs
	var utxo_outs = []Utxo{}
	var utxo1 = makeUtxo(txid,"1",args[2],addrTo,"false") 
	var utxo2 Utxo
	utxo_outs = append(utxo_outs,utxo1)
	//give the extract amount back to addrFrom
	if currValue > amount {
		utxo2 = makeUtxo(txid,"2",strconv.FormatFloat(float64(currValue-amount), 'f', 2, 64),addrFrom,"false")
		utxo_outs = append(utxo_outs,utxo2)
	} 

	//store new utxo
	storeUtxo(APIstub,txid,utxo1)
	storeUtxo(APIstub,txid,utxo2)

	//store new transaction
	transaction := makeTransaction(utxo_ins,utxo_outs)
	storeTransaction(APIstub,txid,transaction)
	
	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}