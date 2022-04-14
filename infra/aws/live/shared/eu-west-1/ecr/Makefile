
login:
	@echo "Login to container registry"
	skopeo login -u AWS -p "$(shell aws-vault exec $(account) -- aws ecr get-login-password)" 823058932981.dkr.ecr.eu-west-1.amazonaws.com
	@echo "Logged in"

sync:
	@echo "Syncing containers"
	skopeo sync --src yaml --dest docker sync.yml 823058932981.dkr.ecr.eu-west-1.amazonaws.com --all
	@echo "Containers synced"