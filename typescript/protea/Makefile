
test:
	@echo "No tests to run"

format:
	@echo "No formatting"

lint:
	@echo "No Linting"

build:
	@echo "Building image without uploading to CDN"
	# Timstamp in UTC format YYYYMMDD-HHMMSS
	DOCKER_BUILDKIT=1 docker build -t "protea" -f ./Dockerfile .
	@echo "Image built"

# Use this to build the image locally and upload the public assets to the CDN
# Make sure you have the `gsa-key.conf` file in the current directory with
# the correct permissions to upload to the specified bucket.
build-and-upload:
	@echo "Building image and uploading to CDN if REMIX_PUBLIC_PATH is set"
	DOCKER_BUILDKIT=1 docker build -t "protea" -f ./Dockerfile --build-arg REMIX_PUBLIC_PATH=https://cdn.interledger.app/protea --build-arg CDN_BUCKET_NAME=common-interledger-cdn-assets --secret id=cdn,src=./rclone.conf  .
	@echo "Image built and assets uploaded"