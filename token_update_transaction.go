package hedera

import (
	"time"

	"github.com/hashgraph/hedera-sdk-go/proto"
)

// Updates an already created Token.
// If no value is given for a field, that field is left unchanged. For an immutable tokens
// (that is, a token created without an adminKey), only the expiry may be updated. Setting any
// other field in that case will cause the transaction status to resolve to TOKEN_IS_IMMUTABlE.
type TokenUpdateTransaction struct {
	Transaction
	pb *proto.TokenUpdateTransactionBody
}

func NewTokenUpdateTransaction() *TokenUpdateTransaction {
	pb := &proto.TokenUpdateTransactionBody{}

	transaction := TokenUpdateTransaction{
		pb:          pb,
		Transaction: newTransaction(),
	}

	return &transaction
}

// The Token to be updated
func (transaction *TokenUpdateTransaction) SetTokenID(tokenID TokenID) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.pb.Token = tokenID.toProtobuf()
	return transaction
}

// The new Symbol of the Token. Must be UTF-8 capitalized alphabetical string identifying the token.
func (transaction *TokenUpdateTransaction) SetSymbol(symbol string) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.pb.Symbol = symbol
	return transaction
}

// The new Name of the Token. Must be a string of ASCII characters.
func (transaction *TokenUpdateTransaction) SetName(name string) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.pb.Name = name
	return transaction
}

// The new Treasury account of the Token. If the provided treasury account is not existing or
// deleted, the response will be INVALID_TREASURY_ACCOUNT_FOR_TOKEN. If successful, the Token
// balance held in the previous Treasury Account is transferred to the new one.
func (transaction *TokenUpdateTransaction) SetTreasury(treasury AccountID) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.pb.Treasury = treasury.toProtobuf()
	return transaction
}

// The new Admin key of the Token. If Token is immutable, transaction will resolve to
// TOKEN_IS_IMMUTABlE.
func (transaction *TokenUpdateTransaction) SetAdminKey(publicKey PublicKey) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.pb.AdminKey = publicKey.toProtoKey()
	return transaction
}

// The new KYC key of the Token. If Token does not have currently a KYC key, transaction will
// resolve to TOKEN_HAS_NO_KYC_KEY.
func (transaction *TokenUpdateTransaction) SetKycKey(publicKey PublicKey) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.pb.KycKey = publicKey.toProtoKey()
	return transaction
}

// The new Freeze key of the Token. If the Token does not have currently a Freeze key, transaction
// will resolve to TOKEN_HAS_NO_FREEZE_KEY.
func (transaction *TokenUpdateTransaction) SetFreezeKey(publicKey PublicKey) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.pb.FreezeKey = publicKey.toProtoKey()
	return transaction
}

// The new Wipe key of the Token. If the Token does not have currently a Wipe key, transaction
// will resolve to TOKEN_HAS_NO_WIPE_KEY.
func (transaction *TokenUpdateTransaction) SetWipeKey(publicKey PublicKey) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.pb.WipeKey = publicKey.toProtoKey()
	return transaction
}

// The new Supply key of the Token. If the Token does not have currently a Supply key, transaction
// will resolve to TOKEN_HAS_NO_SUPPLY_KEY.
func (transaction *TokenUpdateTransaction) SetSupplyKey(publicKey PublicKey) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.pb.SupplyKey = publicKey.toProtoKey()
	return transaction
}

// The new account which will be automatically charged to renew the token's expiration, at
// autoRenewPeriod interval.
func (transaction *TokenUpdateTransaction) SetAutoRenewAccount(autoRenewAccount AccountID) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.pb.AutoRenewAccount = autoRenewAccount.toProtobuf()
	return transaction
}

// The new interval at which the auto-renew account will be charged to extend the token's expiry.
func (transaction *TokenUpdateTransaction) SetAutoRenewPeriod(autoRenewPeriod uint64) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.pb.AutoRenewPeriod = autoRenewPeriod
	return transaction
}

// The new expiry time of the token. Expiry can be updated even if admin key is not set. If the
// provided expiry is earlier than the current token expiry, transaction wil resolve to
// INVALID_EXPIRATION_TIME
func (transaction *TokenUpdateTransaction) SetExpirationTime(expirationTime uint64) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.pb.Expiry = expirationTime
	return transaction
}

//
// The following methods must be copy-pasted/overriden at the bottom of **every** _transaction.go file
// We override the embedded fluent setter methods to return the outer type
//

func tokenUpdateTransaction_getMethod(request request, channel *channel) method {
	return method{
		transaction: channel.getToken().UpdateToken,
	}
}

func (transaction *TokenUpdateTransaction) IsFrozen() bool {
	return transaction.isFrozen()
}

// Sign uses the provided privateKey to sign the transaction.
func (transaction *TokenUpdateTransaction) Sign(
	privateKey PrivateKey,
) *TokenUpdateTransaction {
	return transaction.SignWith(privateKey.PublicKey(), privateKey.Sign)
}

func (transaction *TokenUpdateTransaction) SignWithOperator(
	client *Client,
) (*TokenUpdateTransaction, error) {
	// If the transaction is not signed by the operator, we need
	// to sign the transaction with the operator

	if client.operator == nil {
		return nil, errClientOperatorSigning
	}

	if !transaction.IsFrozen() {
		transaction.FreezeWith(client)
	}

	return transaction.SignWith(client.operator.publicKey, client.operator.signer), nil
}

// SignWith executes the TransactionSigner and adds the resulting signature data to the Transaction's signature map
// with the publicKey as the map key.
func (transaction *TokenUpdateTransaction) SignWith(
	publicKey PublicKey,
	signer TransactionSigner,
) *TokenUpdateTransaction {
	if !transaction.IsFrozen() {
		transaction.Freeze()
	}

	if transaction.keyAlreadySigned(publicKey) {
		return transaction
	}

	for index := 0; index < len(transaction.transactions); index++ {
		signature := signer(transaction.transactions[index].GetBodyBytes())

		transaction.signatures[index].SigPair = append(
			transaction.signatures[index].SigPair,
			publicKey.toSignaturePairProtobuf(signature),
		)
	}

	return transaction
}

// Execute executes the Transaction with the provided client
func (transaction *TokenUpdateTransaction) Execute(
	client *Client,
) (TransactionResponse, error) {
	if !transaction.IsFrozen() {
		transaction.FreezeWith(client)
	}

	transactionID := transaction.id

	if !client.GetOperatorID().isZero() && client.GetOperatorID().equals(transactionID.AccountID) {
		transaction.SignWith(
			client.GetOperatorKey(),
			client.operator.signer,
		)
	}

	resp, err := execute(
		client,
		request{
			transaction: &transaction.Transaction,
		},
		transaction_shouldRetry,
		transaction_makeRequest,
		transaction_advanceRequest,
		transaction_getNodeId,
		tokenUpdateTransaction_getMethod,
		transaction_mapResponseStatus,
		transaction_mapResponse,
	)

	if err != nil {
		return TransactionResponse{}, err
	}

	return TransactionResponse{
		TransactionID: transaction.id,
		NodeID:        resp.transaction.NodeID,
	}, nil
}

func (transaction *TokenUpdateTransaction) onFreeze(
	pbBody *proto.TransactionBody,
) bool {
	pbBody.Data = &proto.TransactionBody_TokenUpdate{
		TokenUpdate: transaction.pb,
	}

	return true
}

func (transaction *TokenUpdateTransaction) Freeze() (*TokenUpdateTransaction, error) {
	return transaction.FreezeWith(nil)
}

func (transaction *TokenUpdateTransaction) FreezeWith(client *Client) (*TokenUpdateTransaction, error) {
	transaction.initFee(client)
	if err := transaction.initTransactionID(client); err != nil {
		return transaction, err
	}

	if !transaction.onFreeze(transaction.pbBody) {
		return transaction, nil
	}

	return transaction, transaction_freezeWith(&transaction.Transaction, client)
}

func (transaction *TokenUpdateTransaction) GetMaxTransactionFee() Hbar {
	return transaction.Transaction.GetMaxTransactionFee()
}

// SetMaxTransactionFee sets the max transaction fee for this TokenUpdateTransaction.
func (transaction *TokenUpdateTransaction) SetMaxTransactionFee(fee Hbar) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.Transaction.SetMaxTransactionFee(fee)
	return transaction
}

func (transaction *TokenUpdateTransaction) GetTransactionMemo() string {
	return transaction.Transaction.GetTransactionMemo()
}

// SetTransactionMemo sets the memo for this TokenUpdateTransaction.
func (transaction *TokenUpdateTransaction) SetTransactionMemo(memo string) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.Transaction.SetTransactionMemo(memo)
	return transaction
}

func (transaction *TokenUpdateTransaction) GetTransactionValidDuration() time.Duration {
	return transaction.Transaction.GetTransactionValidDuration()
}

// SetTransactionValidDuration sets the valid duration for this TokenUpdateTransaction.
func (transaction *TokenUpdateTransaction) SetTransactionValidDuration(duration time.Duration) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.Transaction.SetTransactionValidDuration(duration)
	return transaction
}

func (transaction *TokenUpdateTransaction) GetTransactionID() TransactionID {
	return transaction.Transaction.GetTransactionID()
}

// SetTransactionID sets the TransactionID for this TokenUpdateTransaction.
func (transaction *TokenUpdateTransaction) SetTransactionID(transactionID TransactionID) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.id = transactionID
	transaction.Transaction.SetTransactionID(transactionID)
	return transaction
}

func (transaction *TokenUpdateTransaction) GetNodeAccountIDs() []AccountID {
	return transaction.Transaction.GetNodeAccountIDs()
}

// SetNodeTokenID sets the node TokenID for this TokenUpdateTransaction.
func (transaction *TokenUpdateTransaction) SetNodeAccountIDs(nodeID []AccountID) *TokenUpdateTransaction {
	transaction.requireNotFrozen()
	transaction.Transaction.SetNodeAccountIDs(nodeID)
	return transaction
}
