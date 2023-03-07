 
package wallet

type Wallet interface {
	GetMnemonic(path string) (string, error)
}