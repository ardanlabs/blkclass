package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
	"github.com/ethereum/go-ethereum/crypto"
)

var to = flag.String("t", "", "to")
var nonce = flag.Uint("n", 0, "nonce")
var value = flag.Uint("v", 0, "value")
var tip = flag.Uint("p", 0, "tip")

func main() {
	flag.Parse()

	err := sendTran()
	if err != nil {
		log.Fatalln(err)
	}
}

func sendTran() error {

	privateKey, err := crypto.LoadECDSA("zblock/accounts/kennedy.ecdsa")
	if err != nil {
		return err
	}

	toAccount, err := storage.ToAccount(*to)
	if err != nil {
		log.Fatal(err)
	}

	userTx, err := storage.NewUserTx(*nonce, toAccount, *value, *tip, nil)
	if err != nil {
		log.Fatal(err)
	}

	walletTx, err := userTx.Sign(privateKey)
	if err != nil {
		log.Fatal(err)
	}

	data, err := json.Marshal(walletTx)
	if err != nil {
		log.Fatal(err)
	}

	url := "http://localhost:8080"
	resp, err := http.Post(fmt.Sprintf("%s/v1/tx/submit", url), "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	return nil
}

func genkey() error {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	if err := crypto.SaveECDSA("./bill.ecdsa", privateKey); err != nil {
		log.Fatal(err)
	}

	return nil
}
