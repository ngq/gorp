# grpc

gRPC server adapters live here when the service needs exported gRPC registration.

This directory is reserved for Proto-first gRPC register / adapter code. Keep business orchestration in `internal/service/`, and use this layer only for transport registration and adapter glue.
