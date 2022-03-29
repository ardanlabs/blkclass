package storage

import "fmt"

// UserTx is the transactional data submitted by a user.
type UserTx struct {
	Nonce uint   `json:"nonce"` // Unique id for the transaction supplied by the user.
	From  string `json:"from"`  // Account sendig the money.
	To    string `json:"to"`    // Account receiving the benefit of the transaction.
	Value uint   `json:"value"` // Monetary value received from this transaction.
	Tip   uint   `json:"tip"`   // Tip offered by the sender as an incentive to mine this transaction.
	Data  []byte `json:"data"`  // Extra data related to the transaction.
}

// NewUserTx constructs a new user transaction.
func NewUserTx(nonce uint, from string, to string, value uint, tip uint, data []byte) (UserTx, error) {
	userTx := UserTx{
		Nonce: nonce,
		To:    to,
		Value: value,
		Tip:   tip,
		Data:  data,
	}

	return userTx, nil
}

// String implements the fmt.Stringer interface for logging.
func (tx UserTx) String() string {
	return fmt.Sprintf("%s:%d", tx.From, tx.Nonce)
}
