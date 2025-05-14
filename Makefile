APP_NAME=sentinel
NAMESPACE=default
POSTGRES_SECRET=k8s/base/secrets.yml
POSTGRES_DEPLOYMENT=k8s/base/postgres-deployment.yml
POSTGRES_SERVICE=k8s/base/postgres-service.yml
APP_DEPLOYMENT=k8s/base/deployment.yml
APP_SERVICE=k8s/base/service.yml

IMAGE_TAG=v1.0.3

.PHONY: all up build-image postgres app port-forward run stop clean

# One-liner to start everything
all: up build-image postgres app port-forward

# Start Minikube if not already running
up:
	@echo "🚀 Starting Minikube (if not already running)..."
	@minikube status >/dev/null || minikube start

# Build Docker image inside Minikube
build-image:
	@echo "🐳 Building Docker image inside Minikube..."
	@eval $$(minikube -p minikube docker-env)  &&  docker build -t $(APP_NAME):$(IMAGE_TAG) . --no-cache 
	@echo "🐳 Docker image $(APP_NAME):$(IMAGE_TAG) built successfully."
	@eval minikube image load $(APP_NAME):$(IMAGE_TAG)
	@echo "🐳 Docker image $(APP_NAME):$(IMAGE_TAG) loaded into Minikube."


# Apply Postgres secret & deployment
postgres:
	@echo "🔐 Applying Postgres secrets and deployment..."
	kubectl apply -f $(POSTGRES_SECRET)
	kubectl apply -f $(POSTGRES_DEPLOYMENT)
	kubectl apply -f $(POSTGRES_SERVICE)
	@echo "⏳ Waiting for Postgres pod to be ready..."
	kubectl wait --for=condition=ready pod -l app=$(APP_NAME)-postgres --timeout=30s

# Deploy Go App (sentinel)
app:
	@echo "🚀 Deploying $(APP_NAME)..."
	kubectl apply -f $(APP_DEPLOYMENT)
	kubectl apply -f $(APP_SERVICE)
	@echo "⏳ Waiting for $(APP_NAME) pod to be ready..."
	kubectl wait --for=condition=ready pod -l app=$(APP_NAME) --timeout=30s

# Port-forward PG for local DB access (optional)
port-forward:
	@echo "🔁 Port-forwarding Postgres to localhost:5432..."
	@pkill -f "kubectl port-forward svc/postgres-service 5432:5432" || true
	@nohup kubectl port-forward svc/postgres-service 5432:5432 > /dev/null 2>&1 &

# Run locally with K8s PG
run:
	@echo "🏃 Running Go app with env values..."
	go run cmd/server/main.go

# Stop port-forwarding
stop:
	@echo "🛑 Stopping port-forwarding..."
	@pkill -f "kubectl port-forward svc/postgres-service 5432:5432" || true

# Cleanup everything
clean:
	@echo "🧹 Cleaning up resources..."
	kubectl delete -f $(APP_DEPLOYMENT) --ignore-not-found
	kubectl delete -f $(APP_SERVICE) --ignore-not-found
	kubectl delete -f $(POSTGRES_SERVICE) --ignore-not-found
	kubectl delete -f $(POSTGRES_DEPLOYMENT) --ignore-not-found
	kubectl delete -f $(POSTGRES_SECRET) --ignore-not-found
# @echo "🧹 Stopping Minikube..."
# minikube stop
# @echo "🧹 Minikube stopped."