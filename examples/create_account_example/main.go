package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/Arculus-Holdings-L-L-C/hedera-sdk-go"
)

func main() {

	args := os.Args[1:]

	// How to run...
	// args: network sender-paper-key sender-account receiver-paper-key amount(HBAR float, minimum is 1)
	// e.e. go run create-account.go mainnet "12 word phrase SENDER" 0.0.12345 "12 word phrase RECEIVER" 1

	if len(args) < 4 {
		fmt.Printf("Not enough arguments\n")
		return
	}

	var err error

	network := args[0]
	fromPaperKey := args[1]
	fromAccountId := args[2]
	toPaperKey := args[3]
	amount, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		fmt.Printf("Error parsing amount: %s\n", err.Error())
		return
	}

	var client *hedera.Client

	// Retrieving network type from environment variable HEDERA_NETWORK
	client, err = hedera.ClientForName(network)
	if err != nil {
		fmt.Printf("error creating client %s\n", err.Error())
		return
	}

	// Retrieving operator ID from environment variable OPERATOR_ID
	operatorAccountID, err := hedera.AccountIDFromString(fromAccountId)
	if err != nil {
		fmt.Printf("error creating client %s\n", err.Error())
		return
	}

	// Retrieving operator key from environment variable OPERATOR_KEY
	mn, err := hedera.MnemonicFromString(fromPaperKey)
	pk, err := mn.ToPrivateKey("")
	if err != nil {
		fmt.Printf("error creating client %s\n", err.Error())
		return
	}

	// Defaults the operator account ID and key such that all generated transactions will be paid for
	// by this account and be signed by this key
	client.SetOperator(operatorAccountID, pk)

	// Now the recipient
	toMn, err := hedera.MnemonicFromString(toPaperKey)
	toPk, err := toMn.ToPrivateKey("")
	if err != nil {
		println(err.Error(), ": error converting string to PrivateKey")
		return
	}

	// Assuming that the target shard and realm are known.
	// For now they are virtually always 0 and 0.
	publicKey := toPk.PublicKey()
	aliasAccountID := publicKey.ToAccountID(0, 0)

	fmt.Printf("Transfering %v HBAR from %s to %s\n", amount, operatorAccountID.String(), aliasAccountID.String())
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Press Y to continue N to cancel: ")
	text, _ := reader.ReadString('\n')
	if text != "Y\n" {
		fmt.Printf("\nTransfer aborted...\n")
		return
	}

	println("Transferring...")
	resp, err := hedera.NewTransferTransaction().
		AddHbarTransfer(client.GetOperatorAccountID(), hedera.NewHbar(amount).Negated()).
		AddHbarTransfer(*aliasAccountID, hedera.NewHbar(amount)).
		Execute(client)
	if err != nil {
		println(err.Error(), ": error executing transfer transaction")
		return
	}

	receipt, err := resp.GetReceipt(client)
	println(receipt.Status.String())
	if receipt.AccountID != nil {
		println(receipt.AccountID.String())
	}
	if err != nil {
		println(err.Error(), ": error getting transfer transaction receipt")
		return
	}

	balance, err := hedera.NewAccountBalanceQuery().
		SetAccountID(*aliasAccountID).
		Execute(client)
	if err != nil {
		println(err.Error(), ": error retrieving balance")
		return
	}

	println("Balance of the new account:", balance.Hbars.String())

	info, err := hedera.NewAccountInfoQuery().
		SetAccountID(*aliasAccountID).
		Execute(client)
	if err != nil {
		println(err.Error(), ": error retrieving account info")
		return
	}

	println("New account info:")
	println("The normal account ID:", info.AccountID.String())
	println("The alias key:", info.AliasKey.String())
	println("Example complete")
	err = client.Close()
	if err != nil {
		println(err.Error(), ": error closing client")
		return
	}
}
