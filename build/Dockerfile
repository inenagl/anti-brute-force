FROM golang:1.21 as build

ENV CODE_DIR "/go/src/"
ENV SERVICE_BINARY "/opt/abf/abf-app"
ENV CLI_BINARY "/opt/abf/abf-cli"
ENV MIGRATIONS_BINARY "/opt/abf/abf-migrate"
ENV GOOSE_REPO "github.com/pressly/goose/v3/cmd/goose@latest"

WORKDIR ${CODE_DIR}

RUN go install ${GOOSE_REPO}

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

ARG LDFLAGS
RUN CGO_ENABLED=0 GOOS=linux go build \
        -ldflags "$LDFLAGS" \
        -o ${SERVICE_BINARY} cmd/abf/* &&\
    CGO_ENABLED=0 GOOS=linux go build \
        -ldflags "$LDFLAGS" \
        -o ${CLI_BINARY} cmd/cli/* &&\
    CGO_ENABLED=0 GOOS=linux go build \
        -o ${MIGRATIONS_BINARY} cmd/migrate/*

FROM alpine:3.9

LABEL ORGANIZATION=""
LABEL SERVICE="anti-brute-force"
LABEL MAINTAINERS="inenagl@hotmail.com"

RUN apk add --no-cache bash

ENV WORK_DIR "/opt/abf"

ENV SERVICE_BINARY "${WORK_DIR}/abf-app"
ENV CLI_BINARY "${WORK_DIR}/abf-cli"
ENV MIGRATIONS_BINARY "${WORK_DIR}/abf-migrate"

ENV CONFIG_FILE "${WORK_DIR}/config.yaml"
ENV MIGRATIONS_DIR "/opt/abf/migrations"

WORKDIR ${WORK_DIR}

COPY --from=build ${SERVICE_BINARY} ${SERVICE_BINARY}
COPY --from=build ${CLI_BINARY} ${CLI_BINARY}
COPY --from=build ${MIGRATIONS_BINARY} ${MIGRATIONS_BINARY}

COPY ./build/sh/wait-for-it.sh /sh/wait-for-it.sh
RUN chmod +x /sh/wait-for-it.sh

COPY --from=build /go/src/migrations ${MIGRATIONS_DIR}
COPY ./configs/config.yaml ${CONFIG_FILE}

EXPOSE 8082

CMD ${MIGRATIONS_BINARY} up migrations && \
${SERVICE_BINARY} -config ${CONFIG_FILE}
