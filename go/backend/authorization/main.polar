allow(user: User, action: String, org: Organisation) if
	org.OwnerID = user.ID and
	action in ["read"];

actor User {}
