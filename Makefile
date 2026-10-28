ifneq (,$(wildcard .env))
	include .env
	export
endif

.PHONY: generate_proto \
		run_gateway run_shortener run_redirect \
		run_gateway_unit_tests run_shortener_unit_tests run_redirect_unit_tests run_all_unit_tests \
		run_gateway_integration_tests run_shortener_integration_tests run_redirect_integration_tests run_all_integration_tests \
		run_gateway_performance_tests run_shortener_performance_tests run_redirect_performance_tests run_all_performance_tests \
		migrate_gateway migrate_shortener run_all_migrations \
		docker_build_gateway docker_build_shortener docker_build_redirect build_each_separately \
		docker_push_gateway docker_push_shortener docker_push_redirect docker_push_all

# Go Build

build_gateway:
	cd backend/gateway && go build cmd/main.go
build_shortener:
	cd backend/shortener && go build cmd/main.go
build_redirect:
	cd backend/redirect && go build cmd/main.go

# Go Run

run_gateway:
	cd backend/gateway && go run cmd/main.go
run_shortener:
	cd backend/shortener && go run cmd/main.go
run_redirect:
	cd backend/redirect && go run cmd/main.go

# Go Test

# Run Unit Tests

run_gateway_unit_tests:
	cd backend/gateway && go test ./...
run_shortener_unit_tests:
	cd backend/shortener && go test ./...
run_redirect_unit_tests:
	cd backend/redirect && go test ./...
run_all_unit_tests: run_gateway_unit_tests run_shortener_unit_tests run_redirect_unit_tests

# Run Integration Tests

run_gateway_integration_tests:
	cd backend/gateway && go test -tags=integration ./...
run_shortener_integration_tests:
	cd backend/shortener && go test -tags=integration ./...
run_redirect_integration_tests:
	cd backend/redirect && go test -tags=integration ./...
run_all_integration_tests: run_gateway_integration_tests run_shortener_integration_tests run_redirect_integration_tests

# Run Performance Tests

run_gateway_performance_tests:
	k6 run backend/perf/gateway.js
run_shortener_performance_tests:
	k6 run backend/perf/shortener.js
run_redirect_performance_tests:
	k6 run backend/perf/redirect.js
run_all_performance_tests: run_gateway_performance_tests run_shortener_performance_tests run_redirect_performance_tests

# npm

run_npm:
	cd frontend && npm run dev

# Protobuf

generate_proto:
	cd backend/shortener && mkdir -p internal/pb && protoc --proto_path=../proto --go_out=internal/pb --go_opt=paths=source_relative --go-grpc_out=internal/pb --go-grpc_opt=paths=source_relative ../proto/link.proto
	cd backend/gateway && mkdir -p internal/pb/shortener && protoc --proto_path=../proto --go_out=internal/pb/shortener --go_opt=paths=source_relative --go_opt=Mlink.proto=aziz.dev/gateway/internal/pb/shortener --go-grpc_out=internal/pb/shortener --go-grpc_opt=paths=source_relative --go-grpc_opt=Mlink.proto=aziz.dev/gateway/internal/pb/shortener ../proto/link.proto

# SQL Migrations

migrate_gateway:
	migrate -database postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/gateway?sslmode=disable -path backend/gateway/migrations up
migrate_shortener:
	migrate -database postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/shortener?sslmode=disable -path backend/migrations/shortener up
run_all_migrations: migrate_gateway migrate_shortener

# Docker

# Docker Build

docker_build_gateway:
	docker compose build gateway
docker_build_shortener:
	docker compose build shortener
docker_build_redirect:
	docker compose build redirect
docker_build_frontend:
	docker compose build frontend

docker_build_backend: docker_build_gateway docker_build_shortener docker_build_redirect
docker_build_all: docker_build_gateway docker_build_shortener docker_build_redirect docker_build_frontend

# Docker Push

docker_push_gateway: docker_build_gateway
	docker push abdelaziz333/url-gateway:0.0.6
docker_push_shortener: docker_build_shortener
	docker push abdelaziz333/url-shortener:0.0.6
docker_push_redirect: docker_build_redirect
	docker push abdelaziz333/url-redirect:0.0.6
docker_push_all: docker_push_gateway docker_push_shortener docker_push_redirect

# Docker Run

docker_run_apps:
	docker compose -f docker-compose.yaml --profile app up

docker_run_dbs:
	docker compose -f docker-compose.yaml --profile database up

docker_run_all:
	docker compose -f docker-compose.yaml --profile all up

docker_down_apps:
	docker compose -f docker-compose.yaml --profile app down -v

docker_down_dbs:
	docker compose -f docker-compose.yaml --profile database down -v

docker_down_all:
	docker compose -f docker-compose.yaml --profile all down -v

# apply-helm-infra:
# 	helm repo add bitnami https://charts.bitnami.com/bitnami 2>/dev/null || true
# 	helm repo update
# 	helmfile -f helm/infra/helmfile.yaml sync

# apply-helm-sealed-secrets:
# 	helm install my-sealed-secrets bitnami-labs/sealed-secrets --version 2.18.6 -f ./helm/infra/sealed-secrets/values.yaml -n url-shortener

apply-helm-postgres:
	helm install my-postgres bitnami/postgresql --version 18.6.7 -f ./helm/infra/postgres/values.yaml -n url-shortener

apply-helm-redis:
	helm install my-redis bitnami/redis --version 25.5.3 -f ./helm/infra/redis/values.yaml -n url-shortener

apply-k8s-minikube:
	kubectl apply -k k8s/overlays/minikube

apply-k8s-base:
	kubectl apply -k k8s/base

apply-argo:
	kubectl apply -f root-app