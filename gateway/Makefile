.PHONY: proto
proto:
	mkdir -p protos/ts
	protoc --plugin=protoc-gen-ts_proto=./node_modules/.bin/protoc-gen-ts_proto \
		--ts_proto_out=./protos/ts \
		--ts_proto_opt=nestJs=true,forceLong=string,oneof=unions,useOptionals=messages \
		--proto_path=../protos \
		user/user.proto

	protoc --plugin=protoc-gen-ts_proto=./node_modules/.bin/protoc-gen-ts_proto \
		--ts_proto_out=./protos/ts \
		--ts_proto_opt=nestJs=true,forceLong=string,oneof=unions,useOptionals=messages \
		--proto_path=../protos \
		auth/auth.proto

	protoc --plugin=protoc-gen-ts_proto=./node_modules/.bin/protoc-gen-ts_proto \
		--ts_proto_out=./protos/ts \
		--ts_proto_opt=nestJs=true,forceLong=string,oneof=unions,useOptionals=messages \
		--proto_path=../protos \
		lastseen/lastseen.proto

	protoc --plugin=protoc-gen-ts_proto=./node_modules/.bin/protoc-gen-ts_proto \
		--ts_proto_out=./protos/ts \
		--ts_proto_opt=nestJs=true,forceLong=string,oneof=unions,useOptionals=messages \
		--proto_path=../protos \
		gateway/gateway.proto

.PHONY: run
run:
	go run main.go

.PHONY: clean
clean:
	rm -rf pb/