
current_dir := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

devup:
	@echo "Installing local CA..."
	(cd ./dev/vagrant && ./installCA.sh)
	@echo "Bringing local environment online..."
	(cd ./dev/vagrant && vagrant up)
	@echo "Deploying to local environment..."
	(cd ./dev/vagrant && vagrant reload && vagrant ssh < deploy.sh)
	@echo "Done."

devdeploy:
	@echo "Deploying to local environment..."
	(cd ./dev/vagrant && vagrant ssh < deploy.sh)
	@echo "Done."

devdown:
	@echo "Deleting local environment..."
	(cd ./dev/vagrant && vagrant destroy)
	@echo "Done."

devnomad:
	@echo "Restarting Nomad..."
	(cd ./dev/vagrant && vagrant ssh -c "sudo systemctl restart nomad")
	@echo "Done."

devssh:
	(cd ./dev/vagrant && vagrant ssh)

localwallet:
	FYNBOS_ENV=local go run go/backend/cli/dev/main.go make wallet -k -l
