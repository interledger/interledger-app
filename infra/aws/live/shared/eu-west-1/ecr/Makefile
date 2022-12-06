
login:
	@echo "Login to container registry"
	skopeo login -u AWS -p "$(shell aws-vault exec $(account) -- aws ecr get-login-password)" 823058932981.dkr.ecr.eu-west-1.amazonaws.com
	@echo "Logged in"

sync:
	@echo "Syncing containers"
	skopeo sync --src yaml --scoped --dest docker sync.yml 823058932981.dkr.ecr.eu-west-1.amazonaws.com --all
	@echo "Containers synced"

synctemporal:
	@echo "Syncing temporal containers"
	{ \
  		skopeo copy docker://temporalio/auto-setup:1.18.5 docker://823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/auto-setup:1.18.5; \
  		skopeo copy docker://temporalio/server:1.18.5 docker://823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/server:1.18.5; \
  		skopeo copy docker://temporalio/admin-tools:1.18.5 docker://823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/admin-tools:1.18.5; \
   		skopeo copy docker://temporalio/ui:2.9.0 docker://823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/ui:2.9.0; \
    }
	@echo "Syncing temporal synced"