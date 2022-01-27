allow(user: User, action: String, accHolder: AccountHolder) if
	accHolder.ProfileID = user.ID and
	action in ["read"];

actor User {}
