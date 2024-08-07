# Stage 1: Build the Ant app
FROM golang:1.21.6-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN apk add --no-cache make 

WORKDIR /go/src/github.com/cloud-barista/cm-ant

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH make build


# Stage 2: Run the Ant app
FROM ubuntu:22.04 as prod

RUN apt update && \
    apt install -y sudo curl

# ANT ROOT PATH
ENV ANT_ROOT_PATH=/app

WORKDIR $ANT_ROOT_PATH

COPY --from=builder /go/src/github.com/cloud-barista/cm-ant/ant /app/ant
COPY --from=builder /go/src/github.com/cloud-barista/cm-ant/config.yaml /app/config.yaml
COPY --from=builder /go/src/github.com/cloud-barista/cm-ant/test_plan /app/test_plan
COPY --from=builder /go/src/github.com/cloud-barista/cm-ant/script /app/script
COPY --from=builder /go/src/github.com/cloud-barista/cm-ant/meta /app/meta
COPY --from=builder /go/src/github.com/cloud-barista/cm-ant/web /app/web

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s \
   CMD curl -f "http://cm-ant:8880/ant/api/v1/readyz" || exit 1   


EXPOSE 8880

ENTRYPOINT ["./ant"]