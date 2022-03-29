package main

import (
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	err := genkey()
	if err != nil {
		log.Fatalln(err)
	}
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
