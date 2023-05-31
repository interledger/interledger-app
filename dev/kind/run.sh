#!/usr/bin/env sh
set -ex

# Create cluster
ctlptl apply -f "$(dirname "$0")/config.yaml"

# Build backend base docker and upload to cluster
docker build "$(dirname "$0")/../../go/" -f "$(dirname "$0")/../../go/backend/Dockerfile" -t localhost:5005/backend:latest
docker push localhost:5005/backend:latest

# Build protea base docker and upload to cluster
docker build "$(dirname "$0")/../../typescript/protea" -f "$(dirname "$0")/../../typescript/protea/Dockerfile" --target dev -t localhost:5005/protea:latest
docker push localhost:5005/protea:latest

docker build "$(dirname "$0")/../../typescript/botanist" -f "$(dirname "$0")/../../typescript/botanist/Dockerfile" --target dev -t localhost:5005/botanist:latest
docker push localhost:5005/botanist:latest

# Pull and push temporal
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalite
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalite localhost:5005/temporalite
docker push localhost:5005/temporalite

# Pull and push temporal
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/auto-setup:1.20.3.0
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/auto-setup:1.20.3.0 localhost:5005/temporalio/auto-setup:1.20.3.0
docker push localhost:5005/temporalio/auto-setup:1.20.3.0
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/ui:2.15.0
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/ui:2.15.0 localhost:5005/temporalio/ui:2.15.0
docker push localhost:5005/temporalio/ui:2.15.0
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/admin-tools:1.20.3.0
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/admin-tools:1.20.3.0 localhost:5005/temporalio/admin-tools:1.20.3.0
docker push localhost:5005/temporalio/admin-tools:1.20.3.0
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/server:1.20.3.0
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/server:1.20.3.0 localhost:5005/temporalio/server:1.20.3.0
docker push localhost:5005/temporalio/server:1.20.3.0

# Pull and push cert watcher
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/certwatcher:3.17.0
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/certwatcher:3.17.0 localhost:5005/certwatcher
docker push localhost:5005/certwatcher
