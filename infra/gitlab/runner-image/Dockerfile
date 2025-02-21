FROM debian:bullseye

RUN apt-get -y update && apt-get -y install apt-transport-https ca-certificates curl gnupg-agent software-properties-common
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
RUN add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable"
RUN apt-get update -y && apt-get -y install docker-ce docker-ce-cli containerd.io awscli unzip wget amazon-ecr-credential-helper make build-essential jq

# Install goland
ENV PATH /usr/local/go/bin:$PATH
COPY --from=golang:1.23.0 /usr/local/go/ /usr/local/go/
ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.1
