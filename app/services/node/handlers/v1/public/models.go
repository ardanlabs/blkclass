package public

import "github.com/ardanlabs/blockchain/foundation/blockchain/storage"

type info struct {
	Account storage.Account `json:"account"`
	Name    string          `json:"name"`
	Balance uint            `json:"balance"`
	Nonce   uint            `json:"nonce"`
}

type actInfo struct {
	LastestBlock string `json:"lastest_block"`
	Uncommitted  int    `json:"uncommitted"`
	Accounts     []info `json:"accounts"`
}
