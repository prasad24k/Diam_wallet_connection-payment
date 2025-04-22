package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/diamcircle/go/clients/auroraclient"
	"github.com/diamcircle/go/txnbuild"
)

const (
	port              = ":3000"
	networkPassphrase = "Diamante Testnet 2024"
	baseFee           = txnbuild.MinBaseFee
)

type CreateTransactionRequest struct {
	SourcePublicKey string `json:"sourcePublicKey"`
	Destination     string `json:"destination"`
	Amount          string `json:"amount"`
}

type CreateTransactionResponse struct {
	TransactionXDR string `json:"transactionXDR"`
}

func createTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	// Validate input
	if req.SourcePublicKey == "" || req.Destination == "" || req.Amount == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create Aurora client
	client := auroraclient.DefaultTestNetClient

	// Get source account
	sourceAccount, err := client.AccountDetail(auroraclient.AccountRequest{
		AccountID: req.SourcePublicKey,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching source account: %v", err), http.StatusBadRequest)
		return
	}

	// Create payment operation
	paymentOp := txnbuild.Payment{
		Destination: req.Destination,
		Amount:      req.Amount,
		Asset:       txnbuild.NativeAsset{},
	}

	// Create transaction
	txParams := txnbuild.TransactionParams{
		SourceAccount:        &sourceAccount,
		IncrementSequenceNum: true,
		Operations:           []txnbuild.Operation{&paymentOp},
		BaseFee:             baseFee,
		Timebounds:          txnbuild.NewTimeout(300),
	}

	tx, err := txnbuild.NewTransaction(txParams)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to XDR without signing
	txeB64, err := tx.Base64()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	response := CreateTransactionResponse{
		TransactionXDR: txeB64,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	http.HandleFunc("/create-transaction", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		if r.Method == http.MethodOptions {
			return
		}
		createTransactionHandler(w, r)
	})

	log.Printf("API listening at http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}