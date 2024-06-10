swag:
	swag init -g cmd/cm-ant/main.go --output api/

run:
	go run cmd/cm-ant/main.go

.PHONY: swag run