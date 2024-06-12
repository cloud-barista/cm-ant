swag:
	swag init -g cmd/cm-ant/main.go --output api/

run:
	go run cmd/cm-ant/main.go

build:
	go build -o ant ./cmd/cm-ant/...

.PHONY: swag run build