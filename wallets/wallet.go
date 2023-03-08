package wallets

type Wallet interface {
	GetMnemonic(path string) (string, error)
}