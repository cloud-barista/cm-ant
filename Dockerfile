# Stage 1: Build the Ant app
FROM golang:1.21.6-alpine AS builder

RUN apk add --no-cache make gcc sqlite-libs sqlite-dev build-base

WORKDIR /go/src/github.com/cloud-barista/cm-ant

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 make build


# Stage 2: Run the Ant app
FROM alpine:latest as prod

# ANT ROOT PATH
ENV ANT_ROOT_PATH=/app

WORKDIR $ANT_ROOT_PATH

COPY --from=builder /go/src/github.com/cloud-barista/cm-ant/ant /app/ant
COPY --from=builder /go/src/github.com/cloud-barista/cm-ant/config.yaml /app/config.yaml
COPY --from=builder /go/src/github.com/cloud-barista/cm-ant/test_plan /app/test_plan
COPY --from=builder /go/src/github.com/cloud-barista/cm-ant/script /app/script
COPY --from=builder /go/src/github.com/cloud-barista/cm-ant/meta /app/meta
COPY --from=builder /go/src/github.com/cloud-barista/cm-ant/web /app/web

EXPOSE 8880

ENTRYPOINT ["./ant"]