# URL Shortener

A distributed url shortener written in Go and gRPC for high performance.

# Project Structure

backend -> backend files (Go)
frontend -> frontend files (ReactJS)
docs -> diagrams (ERD, C4, sequence)
deploy -> gitops files (k8s, helm, argocd)
infra -> infrastructure files (terraform, ansible)

## How To Run

NOTE: instructions ain't fully done yet.

#### Locally

apps runs locally
databases in docker

make sure you have Docker installed (check with docker --version)
clone the project (git clone https://github.com/AbdElAziz333/url-shortener)
cd into the project `cd url-shortener`
first, run the databases with `make docker_run_dbs_local` via docker compose
wait some time to run
second, run the backend with `make run_apps_local` via gradle
wait some time to run

#### Docker

apps runs in docker
databases in docker

make sure you have Docker installed (check with docker --version)


- First, run make docker-build in cmd (builds docker images)
- Second, run make run-prod in cmd and volah

#### Kubernetes

apps runs in k8s
databases in k8s
