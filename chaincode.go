package main

import (
	"bytes"
	_ "bytes"
	"encoding/json"
	_ "encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	_ "strconv"
	_ "time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"

	_ "github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

type SmartContractPrinter struct {
}

type Person struct {
	Name   string `json:"name"`
	Surname  string `json:"surname"`
	Ident string `json:"ident"`
	IdentType  string `json:"identType"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Name   string `json:"name"`
	Endorsers []string `json:"endorsers"`
	ViewPermissions []ViewPermission `json:"viewPermissions"`
}

type ViewPermission struct {
	RequesterId string `json:"requesterId"`
	Endorsers []string `json:"endorsers"`
}

var logger = flogging.MustGetLogger("fabcar_cc");

func (s *SmartContractPrinter) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	client, err := cid.New(stub)
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}

	id, err := client.GetID()
	mspId, err := client.GetMSPID()

	var clientId string = mspId + id

	function, args := stub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))
	logger.Infof("ID KLIJENTA JE:  %d", clientId)

	switch function {
	case "insertPerson":
		return s.insertPerson(stub, args)
	case "readPerson":
		return s.readPerson(stub, args)
	case "readAllPersons":
		return s.readAllPersons(stub, args)
	case "readCert":
		return s.readCert(stub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}


	return shim.Success(nil)
}
func (s *SmartContractPrinter) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContractPrinter) insertPerson(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	var person = Person{Name: args[0], Surname: args[1], Ident: args[2], IdentType: args[3]}
	personAsBytes, _ := json.Marshal(person)
	APIstub.PutState(person.Ident, personAsBytes)
	return shim.Success(personAsBytes)
}

func (s *SmartContractPrinter) readAllPersons(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	resultsIterator, err := APIstub.GetStateByRange("","")
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
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

	fmt.Printf("- queryAllCars:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContractPrinter) readPerson(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	personAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(personAsBytes)
}

func (s *SmartContractPrinter) readCert(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	var buffer bytes.Buffer

	client, err := cid.New(APIstub)

	cert, err := client.GetX509Certificate()

	creator, err := APIstub.GetCreator()
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
	id, err := client.GetID()
	buffer.WriteString("Creator " + string(creator))
	buffer.WriteString(", Signature " + string(cert.Signature))
	buffer.WriteString(", MSPId " + string(cert.Signature))
	buffer.WriteString(", ID " + string(id))
	return shim.Success(buffer.Bytes())

}

func main() {
	err := shim.Start(new(SmartContractPrinter))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
