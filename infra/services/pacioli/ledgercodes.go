package pacioli

func LocalClusterLedgerCodes() map[string]uint16 {
	const base uint16 = 0
	ret := map[string]uint16{
		// don't use 0 to not confuse with default uint16 = 0.
		"backend-usd": base + 840,
	}

	return ret
}

func DevClusterLedgerCodes() map[string]uint16 {
	const base uint16 = 1000
	ret := map[string]uint16{
		// don't use 0 to not confuse with default uint16 = 0.
		"backend-usd": base + 840,
	}

	return ret
}
