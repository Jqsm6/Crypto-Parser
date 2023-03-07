package metamask

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/crypto/pbkdf2"
)

type Metamask struct{}

type Data struct {
	KeyringController struct {
		Vault string `json:"vault"`
	} `json:"KeyringController"`
}

type Vault struct {
	Data string `json:"data"`
	IV   string `json:"iv"`
	Salt string `json:"salt"`
}

type Wallet struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (Metamask) GetMnemonic(path string, passwords []string) (string, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return "", err
	}

	dataBytes, err := db.Get([]byte("data"), nil)
	if err != nil {
		return "", err
	}

	var data Data

	if err = json.Unmarshal(dataBytes, &data); err != nil {
		return "", err
	}

	var vault Vault

	if err = json.Unmarshal([]byte(data.KeyringController.Vault), &vault); err != nil {
		return "", err
	}

	salt, err := base64.StdEncoding.DecodeString(vault.Salt)
	if err != nil {
		return "", err
	}

	iv, err := base64.StdEncoding.DecodeString(vault.IV)
	if err != nil {
		return "", err
	}

	encrypted, err := base64.StdEncoding.DecodeString(vault.Data)
	if err != nil {
		return "", err
	}

	for _, password := range passwords {
		key := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)
		block, err := aes.NewCipher(key)
		if err != nil {
			continue
		}
		gcm, err := cipher.NewGCMWithNonceSize(block, len(iv))
		if err != nil {
			continue
		}

		decrypted, err := gcm.Open(nil, iv, encrypted, nil)
		if err != nil {
			continue
		}

		var wallets []Wallet
		if err = json.Unmarshal(decrypted, &wallets); err != nil {
			return "", err
		}
		for _, wallet := range wallets {
			if wallet.Type == "HD Key Tree" {
				var data struct {
					Mnemonic []byte `json:"mnemonic"`
				}
				if err = json.Unmarshal(wallet.Data, &data); err != nil {
					return "", err
				}

				return string(data.Mnemonic), nil
			}
		}
	}

	return "", fmt.Errorf("no valid password for %s", path)
}
