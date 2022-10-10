# Build
FROM golang:1.18-alpine AS builder

ARG CGO_ENABLED=0
ARG GO111MODULE=on
ARG GOARCH=amd64
ARG GOOS=linux


WORKDIR /repo

COPY go.mod go.sum ./
RUN git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"
RUN go mod download

COPY . .
RUN export VERSION=$(git rev-parse --short HEAD)
RUN go build -o /repo/app cmd/app/main.go
RUN go build -o /repo/execution cmd/execution/main.go

# Deploy
FROM alpine:3.13 as deployer

WORKDIR /repo

COPY --from=builder /repo/app /repo/app
COPY --from=builder /repo/execution /repo/execution

COPY scripts /repo/scripts
COPY configs /repo/configs
COPY migrations /repo/migrations
COPY api /repo/api

RUN chmod +x /repo/scripts/run.sh

CMD ["/repo/scripts/run.sh", "app"]
