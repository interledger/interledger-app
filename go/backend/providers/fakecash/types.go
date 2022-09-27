package fakecash

type (
	Account struct {
		ID               string
		AvailableBalance uint64
	}

	CreateArgs struct {
		ID string
	}
)
