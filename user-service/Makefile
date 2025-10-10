.PHONY: proto
proto:
	mkdir -p pb
	protoc --go_out=./pb --go_opt=paths=source_relative \
		--go-grpc_out=./pb --go-grpc_opt=paths=source_relative \
		--proto_path=../protos/user \
	user.proto

	mkdir -p auth-pb
	protoc --go_out=./auth-pb --go_opt=paths=source_relative \
		--go-grpc_out=./auth-pb --go-grpc_opt=paths=source_relative \
		--proto_path=../protos/auth \
	auth.proto

.PHONY: run
run:
	go run main.go

.PHONY: clean
clean:
	rm -rf pb/