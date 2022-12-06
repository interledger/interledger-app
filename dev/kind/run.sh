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

# Pull and push temporal
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalite
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalite localhost:5005/temporalite
docker push localhost:5005/temporalite

# Pull and push temporal
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/auto-setup:1.18.5
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/auto-setup:1.18.5 localhost:5005/temporalio/auto-setup
docker push localhost:5005/temporalio/auto-setup
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/ui:2.9.0
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/ui:2.9.0 localhost:5005/temporalio/ui
docker push localhost:5005/temporalio/ui
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/admin-tools:1.18.5
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/admin-tools:1.18.5 localhost:5005/temporalio/admin-tools
docker push localhost:5005/temporalio/admin-tools
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/server:1.18.5
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalio/server:1.18.5 localhost:5005/temporalio/server
docker push localhost:5005/temporalio/server