package auth

// type Account is an interface that defines minimal information for an account.
type Account interface {
	// The unique ID associated with this account.
	Id() int64
	// The name associated with this account.
	Name() string
}

// NewAccount returns a new instance of `BasicAccount` (which implements the `Account` interface) for 'id' and 'name'.
func NewAccount(id int64, name string) Account {

	return &BasicAccount{
		AccountId:   id,
		AccountName: name,
	}
}

// BasicAccount is the simplest (most basic) implementation of the `Account` interface for wrapping a unique account ID and an account name.
type BasicAccount struct {
	Account `json:",omitempty"`
	// The unique ID associated with this account.
	AccountId int64 `json:"id"`
	// The name associated with this account.
	AccountName string `json:"name"`
}

// Returns the unique ID associated with 'a'.
func (a *BasicAccount) Id() int64 {
	return a.AccountId
}

// Returns the name associated with 'a'.
func (a *BasicAccount) Name() string {
	return a.AccountName
}
