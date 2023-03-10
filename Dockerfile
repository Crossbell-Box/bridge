FROM golang:1.18 as builder

WORKDIR /opt

COPY . /opt/bridge
RUN cd bridge && make bridge

FROM debian:buster

RUN apt-get update  && apt-get install -y --no-install-recommends ca-certificates
RUN update-ca-certificates

WORKDIR "/opt"

ENV CONFIG_PATH ''
ENV RONIN_RPC ''
ENV RONIN_VALIDATOR_KEY ''
ENV RONIN_RELAYER_KEY ''
ENV RONIN_TASK_INTERVAL ''
ENV RONIN_TRANSACTION_CHECK_PERIOD ''
ENV RONIN_MAX_PROCESSING_TASKS ''

ENV RONIN_MAX_TASK_QUERY ''

ENV ETHEREUM_RPC ''
ENV ETHEREUM_VALIDATOR_KEY ''
ENV ETHEREUM_RELAYER_KEY ''
ENV ETHEREUM_GET_LOGS_BATCH_SIZE ''
ENV VERBOSITY 100
ENV NUMBER_OF_WORKERS 1024

ENV DB_HOST ''
ENV DB_PORT ''
ENV DB_NAME ''
ENV DB_USERNAME ''
ENV DB_PASSWORD ''
ENV DB_CONN_MAX_LIFE_TIME ''
ENV DB_MAX_IDLE_CONNS ''
ENV DB_MAX_OPEN_CONNS ''

COPY --from=builder /go/bin/bridge /usr/local/bin/bridge
COPY --from=builder /opt/bridge/config/ ./
COPY --from=builder /opt/bridge/docker/entrypoint.sh ./

RUN apt update && apt install -y jq

ENTRYPOINT ["./entrypoint.sh"]
