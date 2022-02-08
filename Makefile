
current_dir := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

kindup:
	@echo "Bringing Cluster Online"
	./dev/kind/run.sh
	@echo "Cluster online"

kinddown:
	@echo "Deleting Cluster"
	./dev/kind/nuke.sh
	@echo "Cluster deleted"

kindpulumiup:
	@echo "Running pulumi against Kind cluster"
	mkdir -p .fynbos
	PULUMI_K8S_SUPPRESS_HELM_HOOK_WARNINGS=true PULUMI_CONFIG_PASSPHRASE= PULUMI_BACKEND_URL="file://$(current_dir)/.fynbos/" \
	pulumi up -y -s local -C ./infra/clusters/local/

kindpulumidown:
	@echo "Deleting pulumi"
	rm -r ./.fynbos/.pulumi

kindpulumi:
	@echo "Running pulumi $(CMD) against Kind cluster"
	mkdir -p .fynbos
	PULUMI_K8S_SUPPRESS_HELM_HOOK_WARNINGS=true PULUMI_CONFIG_PASSPHRASE= PULUMI_BACKEND_URL="file://$(current_dir)/.fynbos/" \
	pulumi $(CMD) -y -s local -C ./infra/clusters/local/

buildgo:
	DOCKER_BUILDKIT=1 docker build $(current_dir)/go -f $(current_dir)/go/$(target)/Dockerfile -t localhost:5005/$(target):latest
	docker push localhost:5005/$(target):latest

tiltup:
	@echo "Running tilt"
	tilt up -f ./dev/tilt/Tiltfile