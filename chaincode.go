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
	cert, err := client.GetX509Certificate()

	var clientId string = cert.Subject.CommonName

	function, args := stub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))
	logger.Infof("ID KLIJENTA JE:  %d", clientId)
	logger.Infof("COMMON NAME:  %d", cert.Subject.CommonName)

	switch function {
	case "insertPerson":
		return s.insertPerson(stub, args)
	case "readPerson":
		return s.readPerson(stub, args)
	case "readAllPersons":
		return s.readAllPersons(stub, args)
	case "personAddField":
		return s.personAddFields(stub, args)
	case "approveAccess":
		return s.approveAccess(stub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}


	return shim.Success(nil)
}

func (s *SmartContractPrinter) insertPerson(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	var person = Person{Name: args[0], Surname: args[1], Ident: args[2], IdentType: args[3]}
	if(len(args) > 4) {
		var fields []Field
		err := json.Unmarshal([]byte(args[4]), &fields)
		if err != nil {
			fmt.Printf("NIJE USPEO UNMARSHALING: %s", err)
		} else{
			person.Fields = fields
		}
	}
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

func (s *SmartContractPrinter) personAddFields(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	var personIdent string = args[0]
	logger.Critical("ADDING FIELD FOR PERSON IDENT " + personIdent)
	logger.Critical("FIELD ARG " + args[1])

	personAsBytes, _ := APIstub.GetState(args[0])

	if(personAsBytes == nil) {
		shim.Error("Person ident " + personIdent + " not found !")
	}
	var person Person
	json.Unmarshal(personAsBytes, &person)

	var fields []Field
	err := json.Unmarshal([]byte(args[1]), &fields)

	if(err != nil) {
		fmt.Printf("NIJE USPEO UNMARSHALING: %s", err)
	}

	if(len(person.Fields) == 0) {
		logger.Critical("OSOBA NEMA POLJA, UPISUJEMO !")
		person.Fields = fields
	} else {
		for i, newField := range fields {
			logger.Critical("Iter " + string(i))
			var updatedExistingField bool = false
			for j, personField := range person.Fields {
				if (newField.Name == personField.Name) {
					logger.Critical("VEC POSTOJI POLJE NAZIV: " + newField.Name + ", RADIMO UPDATE !")
					person.Fields[j] = newField
					updatedExistingField = true
				}
			}
			if(!updatedExistingField) {
				logger.Critical("RADIMO UNOS NOVOG POLJA NAZIV: " + newField.Name)
				person.Fields = append(person.Fields, newField)
			}
		}
	}
	logger.Critical("CUVAMO NOVU KONFIGURACIJU POLJA !")
	insertPersonAsBytes, _ := json.Marshal(person)
	APIstub.PutState(person.Ident, insertPersonAsBytes)
	return shim.Success(insertPersonAsBytes)


	return shim.Success([]byte(fmt.Sprintf("%v", fields)))
}

func (s *SmartContractPrinter) approveAccess(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	client, err := cid.New(APIstub)
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}

	cert, err := client.GetX509Certificate()

	var clientId string = cert.Subject.CommonName
	var personIdent string = args[0]
	var requesterId = args[1]
	var fieldNames []string

	_ = json.Unmarshal([]byte(args[2]), &fieldNames)

	logger.Infof("Person ident is:  %d", personIdent)
	logger.Infof("Requesting client id:  %d", requesterId)
	logger.Infof("Field names:  %d", fieldNames)

	personBytes, err := APIstub.GetState(personIdent)

	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
	if personBytes == nil {
		return shim.Error("Person " + personIdent + " not found!")
	}
	var person Person
	json.Unmarshal(personBytes, &person)
	var isChanged bool = false

	for i, f := range person.Fields {
		fmt.Println(i, f.Name)
		for fni, fn := range  fieldNames {
			fmt.Println(fni, fn)
			if fn == f.Name {
				isEndorser := contains(f.Endorsers, clientId)
				if !isEndorser {
					logger.Critical("CLIENT NOT ENDORSER ON FIELD " + fn)
					continue
				}
				logger.Critical("CLIENT IS ENDORSER ON FIELD " + fn)
				for vpi, vp := range f.ViewPermissions {
					fmt.Println(vpi, vp)
					if vp.RequesterId == requesterId {
						logger.Critical("REQUESTER ON FIELD FOUND " + vp.RequesterId)
						var isEndorsementAdded bool = contains(vp.Endorsers, clientId)
						if !isEndorsementAdded {
							logger.Critical("ENDORSEMENT NOT FOUND, ADDING NEW")
							vp.Endorsers = append(vp.Endorsers, clientId)
							f.ViewPermissions[vpi] = vp
							isChanged = true
						}
					}
				}
			}
		}
		person.Fields[i] = f
	}
	if isChanged {
		logger.Critical("CHANGE DONE, SAVING")
		insertPersonAsBytes, _ := json.Marshal(person)
		APIstub.PutState(person.Ident, insertPersonAsBytes)
	}
	return shim.Success(nil)
}

func contains(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}

func main() {
	err := shim.Start(new(SmartContractPrinter))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
func (s *SmartContractPrinter) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}