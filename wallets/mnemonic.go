package wallet

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

func AddrFromSeed(seed string) string {
	seedBytes := []byte(seed)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(seedBytes)
	publicKeyECDSA, _ := crypto.GenerateKey()
	publicKey := publicKeyECDSA.Public().(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKey).Hex()

	return address
}
