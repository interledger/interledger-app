#!/usr/bin/env sh

# Create cluster
ctlptl apply -f "$(dirname "$0")/config.yaml"

# Build backend base docker and upload to cluster
docker build "$(dirname "$0")/../../go/" -f "$(dirname "$0")/../../go/backend/Dockerfile" -t localhost:5005/backend:latest
docker push localhost:5005/backend:latest

# Build protea base docker and upload to cluster
docker build "$(dirname "$0")/../../typescript/protea" -f "$(dirname "$0")/../../typescript/protea/Dockerfile" --target dev -t localhost:5005/protea:latest
docker push localhost:5005/protea:latest

# Build pacioli base docker and upload to cluster
docker build "$(dirname "$0")/../../go/" -f "$(dirname "$0")/../../go/pacioli/Dockerfile" -t localhost:5005/pacioli:latest
docker push localhost:5005/pacioli:latest

# Pull and push temporal
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalite
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalite localhost:5005/temporalite
docker push localhost:5005/temporalite

# Pull and push tigerbeetle
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/tigerbeetle:patch-1
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/tigerbeetle:patch-1 localhost:5005/tigerbeetle:patch-1
docker push localhost:5005/tigerbeetle:patch-1

# Pull and push rafiki
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/rafiki-backend
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/rafiki-backend localhost:5005/rafiki-backend
docker push localhost:5005/rafiki-backend