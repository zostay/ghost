package http

//go:generate protoc -I ./ ./secrets.proto --go_out=.. --go-grpc_out=..
