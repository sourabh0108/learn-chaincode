package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// AccessRecord used to denote a accessRecord of a patient
type AccessRecord struct {
	DoctorID  string `json:"doctorId"`
	TestID    string `json:"testId"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

//PatientAsset where patient denotes which doctor can access with testId
type PatientAsset struct {
	PatientID     string         `json:"patientId"`
	AccessRecords []AccessRecord `json:"accessRecords"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	err := stub.PutState("patientDoctorApp", []byte(args[0]))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Invoke is our entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)

	} else if function == "write" {
		return t.write(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error

	return nil, errors.New("Received unknown function invocation: " + function)
}

//write method is used to add/update/delete assets to/from the world state.
func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0]   // patient id
	value = args[1] // list of accessRecords

	username, _ := GetCertAttribute(stub, "username")
	if username == "user_type1_0" || username == "user_type1_1" {
		err = stub.PutState(key, []byte(value)) //write the variable into the chaincode state
	} else {
		return nil, errors.New(username + " does not have access to create a patient asset")
	}

	if err != nil {
		return nil, err
	}
	return nil, nil
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function) //error

	return nil, errors.New("Received unknown function query: " + function)
}

//read method is used to read from the wrold state and return the patient accessRecords asset. Arguments to be passed patient and doctor id
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting patientID")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	if key != "patientDoctorApp" {

		doctorUsername, _ := GetCertAttribute(stub, "username")
		var accessRecords []AccessRecord
		var newAccessRecordList []AccessRecord
		var index int
		index = 0
		err = json.Unmarshal(valAsbytes, &accessRecords)
		if err != nil {
			jsonResp = "{\"Error\":\"Failed to convert byteArray to accessRecords list.\"}"
			return nil, errors.New(jsonResp)
		}

		//This loop checks if the logged in doctor has access to any of patient's shared EMRs
		for _, accessRecord := range accessRecords {
			if doctorUsername == accessRecord.DoctorID {
				newAccessRecordList[index] = accessRecord
			}
		}

		if len(newAccessRecordList) < 1 {
			jsonResp = "{\"Error\":\"You dont have access to any EMR of Patient " + key + "\"}"
			return nil, errors.New(jsonResp)
		}

		newAccessRecordsAsbytes, err := json.Marshal(&newAccessRecordList)
		if err != nil {
			jsonResp = "{\"Error\":\"Failed to convert newAccessRecords to byteArray.\"}"
			return nil, errors.New(jsonResp)
		}
		valAsbytes = newAccessRecordsAsbytes
	}

	return valAsbytes, nil
}

//GetCertAttribute fetches the value of attribute from the certificate
func GetCertAttribute(stub shim.ChaincodeStubInterface, attributeName string) (string, error) {
	fmt.Println("Entering GetCertAttribute")
	attr, err := stub.ReadCertAttribute(attributeName)
	if err != nil {
		return "", errors.New("Couldn't get attribute " + attributeName + ". Error: " + err.Error())
	}
	attrString := string(attr)
	return attrString, nil
}
