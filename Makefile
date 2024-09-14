build:
	@go build -o bin/blockchain


run: build
	@./bin/blockchain



proto:
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/types.proto


test-crypto:
	@go test -v ./crypto/

bench-crypto:
	@go test  -benchmem -bench=.  ./crypto/ 



test:
	@go test -v ./...