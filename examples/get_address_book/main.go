package main

import (
	"os"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func main() {
	var client *hedera.Client
	var err error

	// Retrieving network type from environment variable HEDERA_NETWORK
	client, err = hedera.ClientForName(os.Getenv("HEDERA_NETWORK"))
	if err != nil {
		println(err.Error(), ": error creating client")
		return
	}

	// Retrieving operator ID from environment variable OPERATOR_ID
	operatorAccountID, err := hedera.AccountIDFromString(os.Getenv("OPERATOR_ID"))
	if err != nil {
		println(err.Error(), ": error converting string to AccountID")
		return
	}

	// Retrieving operator key from environment variable OPERATOR_KEY
	operatorKey, err := hedera.PrivateKeyFromString(os.Getenv("OPERATOR_KEY"))
	if err != nil {
		println(err.Error(), ": error converting string to PrivateKey")
		return
	}

	// Setting the client operator ID and key
	client.SetOperator(operatorAccountID, operatorKey)

	fileQuery := hedera.NewFileContentsQuery().
		// Set the file ID for address book which is 0.0.102
		SetFileID(hedera.FileIDForAddressBook())

	println("the network that address book is for:", client.GetNetworkName().String())

	cost, err := fileQuery.GetCost(client)
	if err != nil {
		println(err.Error(), ": error getting file contents query cost")
		return
	}

	println("file contents cost:", cost.String())

	// Have to always set the cost a little bigger, otherwise it is possible it won't go through
	fileQuery.SetMaxQueryPayment(hedera.NewHbar(1))

	// Execute the file content query
	contents, err := fileQuery.Execute(client)
	if err != nil {
		println(err.Error(), ": error executing file contents query")
		return
	}

	// Create the file the address book will be read into
	// For string
	fileString, err := os.Create("address-book-string.proto.bin")
	if err != nil {
		println(err.Error(), ": error creating address-book-string.proto.bin")
		return
	}
	// For []byte
	fileByte, err := os.Create("address-book-byte.proto.bin")
	if err != nil {
		println(err.Error(), ": error creating address-book-byte.proto.bin")
		return
	}

	// Make sure they are empty
	err = fileString.Truncate(0)
	if err != nil {
		println(err.Error(), ": error emptying file")
		return
	}
	err = fileByte.Truncate(0)
	if err != nil {
		println(err.Error(), ": error emptying file")
		return
	}

	// Write the contents (string([]byte)) into the string file
	leng, err := fileString.WriteString(string(contents))
	if err != nil {
		println(err.Error(), ": error writing contents to file")
		return
	}
	// Write the contents ([]byte) into the byte file
	_, err = fileByte.Write(contents)
	if err != nil {
		println(err.Error(), ": error writing contents to file")
		return
	}

	temp := make([]byte, leng)

	_, err = fileString.Read(temp)

	// Close the files
	err = fileString.Close()
	if err != nil {
		println(err.Error(), ": error closing the file")
		return
	}
	err = fileByte.Close()
	if err != nil {
		println(err.Error(), ": error closing the file")
		return
	}
}
