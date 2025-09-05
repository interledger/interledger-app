FROM alpine:3.18.4

ARG BUILDARCH=amd64
ARG BUILDOS=linux

RUN wget -q https://github.com/temporalio/cli/releases/download/v1.3.0/temporal_cli_1.3.0_${BUILDOS}_${BUILDARCH}.tar.gz -O temporal_cli.tar.gz \
 && tar -xf temporal_cli.tar.gz temporal -C /usr/local/bin/ \
 && rm -rf temporal_cli.tar.gz

EXPOSE 7233 8233

ENTRYPOINT ["temporal", "server", "start-dev", "--ip=0.0.0.0"]

CMD ["--namespace=default"]
