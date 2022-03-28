package storage

// UserTx is the transactional data submitted by a user.
type UserTx struct {
	Nonce uint   `json:"nonce"` // Unique id for the transaction supplied by the user.
	From  string `json:"from"`  // Account sendig the money.
	To    string `json:"to"`    // Account receiving the benefit of the transaction.
	Value uint   `json:"value"` // Monetary value received from this transaction.
	Tip   uint   `json:"tip"`   // Tip offered by the sender as an incentive to mine this transaction.
	Data  []byte `json:"data"`  // Extra data related to the transaction.
}
