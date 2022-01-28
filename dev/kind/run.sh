#!/usr/bin/env sh

# Create cluster
ctlptl apply -f "$(dirname "$0")/config.yaml"

# Build backend base docker and upload to cluster
docker build "$(dirname "$0")/../../go/" -f "$(dirname "$0")/../../go/backend/Dockerfile" -t localhost:5005/backend:latest
docker push localhost:5005/backend:latest

# Build protea base docker and upload to cluster
docker build "$(dirname "$0")/../../typescript/protea" -f "$(dirname "$0")/../../typescript/protea/Dockerfile" --target dev -t localhost:5005/protea:latest
docker push localhost:5005/protea:latest