ARG GO_IMAGE=m.daocloud.io/docker.io/library/golang:1.25.2-alpine
ARG RUNTIME_IMAGE=m.daocloud.io/docker.io/library/alpine:3.20

FROM ${GO_IMAGE} AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM ${RUNTIME_IMAGE}

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/server /app/server

EXPOSE 9002

USER app

ENTRYPOINT ["/app/server"]
