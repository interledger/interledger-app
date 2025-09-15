infrastructure:
	docker compose --profile infrastructure up

services:
	docker compose --profile services up

application:
	docker compose --profile application up --watch

backend:
	docker compose --profile backend up --watch

frontend:
	docker compose --profile frontend up --watch

all:
	docker compose --profile "*" up --watch

down:
	docker compose --profile "*" down

delete-volumes:
	docker compose --profile "*" down -v

.PHONY: infra services application down delete-volumes backend frontend all help

help:
	@echo "make <command>"
	@echo "e.g. make infrastructure"
	@echo ""
	@echo "Commands:"
	@echo "  infrastructure       Start infrastructure services - traefik, postgres, redis"
	@echo "  services             Start application services - kratos, temporal, rafiki services, requires infrastructure services"
	@echo "  application          Start wallet application - backend and frontend, requires services and infrastructure"
	@echo "  backend              Start only the backend service"
	@echo "  frontend             Start only the frontend service"
	@echo "  all                  Start all services"
	@echo "  down                 Stop all running services"
	@echo "  delete-volumes       Stop all services and remove associated volumes"
	@echo "  help                 Display this help message"