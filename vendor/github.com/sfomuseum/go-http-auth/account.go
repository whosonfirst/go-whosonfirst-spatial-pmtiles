package auth

// type Account is a struct that defines minimal information for an account.
type Account struct {
	// The unique ID associated with this account.
	Id int64 `json:"id"`
	// The name associated with this account.
	Name string `json:"name"`
}
