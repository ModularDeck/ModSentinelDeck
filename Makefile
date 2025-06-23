APP_NAME=sentinel
NAMESPACE=default
POSTGRES_SECRET=k8s/base/secrets.yml
POSTGRES_DEPLOYMENT=k8s/base/postgres-deployment.yml
POSTGRES_SERVICE=k8s/base/postgres-service.yml
APP_DEPLOYMENT=k8s/base/deployment.yml
APP_SERVICE=k8s/base/service.yml

IMAGE_TAG=v1.0.6

.PHONY: all up build-image postgres app port-forward run stop clean

# One-liner to start everything
all: up build-image postgres migrate app  port-forward validate

build: build-image app port-forward

# Start Minikube if not already running
up:
	@echo "🚀 Starting Minikube (if not already running)..."
	@minikube status | grep -q "Running" || (echo "Starting Minikube..." && minikube start)
	@echo "Minikube is running."

# Build Docker image inside Minikube
build-image:
	@echo "🐳 Building Docker image inside Minikube..."
	eval $(minikube -p minikube docker-env) 
	docker build -t $(APP_NAME):$(IMAGE_TAG) --no-cache .
	sleep 20
	minikube image load $(APP_NAME):$(IMAGE_TAG)
	minikube image list | grep $(APP_NAME)
	@echo "🐳 Docker image $(APP_NAME):$(IMAGE_TAG) built successfully."


# Apply Postgres secret & deployment
postgres:
	@echo "🔐 Applying Postgres secrets and deployment..."
	kubectl apply -f $(POSTGRES_SECRET)
	kubectl apply -f $(POSTGRES_DEPLOYMENT)
	kubectl apply -f $(POSTGRES_SERVICE)
	
	@echo "⏳ Waiting for Postgres pod to be ready..."
	kubectl wait --for=condition=ready pod -l app=$(APP_NAME)-postgres --timeout=30s
	@echo "🔐 Postgres secrets and deployment applied successfully."

# Deploy Go App (sentinel)
app:
	@echo "🚀 Deploying $(APP_NAME)..."
	kubectl apply -f $(APP_DEPLOYMENT)
	kubectl apply -f $(APP_SERVICE)
	@echo "⏳ Waiting for $(APP_NAME) pod to be ready..."
	kubectl wait --for=condition=ready pod -l app=$(APP_NAME) --timeout=30s

# Port-forward PG for local DB access (optional)
port-forward:
	@echo "🔁 Port-forwarding $(APP_NAME) to localhost:8080..."
	kubectl port-forward svc/$(APP_NAME)-service 8080:8080   
	@echo "🔁 Port-forwarding $(APP_NAME) to localhost:8080..."

# Run locally with K8s PG
run:
	@echo "🏃 Running Go app with env values..."
	go run cmd/server/main.go

# Stop port-forwarding
stop:
	@echo "🛑 Stopping port-forwarding..."
	@pkill -f "kubectl port-forward svc/$(APP_NAME)-service 8080:8080" || true

migrate:
	@echo "🔄 Running migrations..."
	kubectl apply -f k8s/base/init-db-configmap.yml
	kubectl apply -f k8s/base/sentinel-schema-script.yml
	kubectl apply -f k8s/base/db-init-job.yml
	@echo "🔄 Migrations completed."

# Cleanup everything
clean:
	@echo "🧹 Cleaning up resources..."
	kubectl delete -f $(APP_DEPLOYMENT) --ignore-not-found
	kubectl delete -f $(APP_SERVICE) --ignore-not-found
	kubectl delete job sentinel-db-init  --ignore-not-found
	kubectl delete -f $(POSTGRES_SERVICE) --ignore-not-found
	kubectl delete -f $(POSTGRES_DEPLOYMENT) --ignore-not-found
	kubectl delete -f $(POSTGRES_SECRET) --ignore-not-found
	kubectl delete -f k8s/base/init-db-configmap.yml --ignore-not-found
	kubectl delete -f k8s/base/sentinel-schema-script.yml --ignore-not-found
	kubectl delete -f k8s/base/db-init-job.yml --ignore-not-found
	kubectl delete pod -l app=$(APP_NAME)-postgres --ignore-not-found
	kubectl delete pod -l app=$(APP_NAME) --ignore-not-found
	@echo "🧹 All resources cleaned up."
	@echo "🧹 Cleanup completed."
# @echo "🧹 Stopping Minikube..."
# minikube stop
# @echo "🧹 Minikube stopped."

redeployment:
	@echo "🔄 Redeploying $(APP_NAME)..."
	docker build -t  $(APP_NAME):$(IMAGE_TAG) .
	minikube image list | grep $(APP_NAME)
	kubectl rollout restart deployment $(APP_NAME)-deployment
	@echo "⏳ Waiting for $(APP_NAME) pod to be ready..."
	kubectl wait --for=condition=ready pod -l app=$(APP_NAME) --timeout=30s
	@echo "🔄 Redeploying $(APP_NAME) completed."
	kubectl get pods
	kubectl port-forward svc/$(APP_NAME)-service 8080:8080   
	

validate:
	@echo "✅ Validating deployment..."
	kubectl get pods
# kubectl logs deployment/sentinel-deployment
# kubectl get services
# kubectl get deployments  
# minikube service list
# kubectl get secrets
# kubectl describe  pod -l app=sentinel	
# kubectl describe  pod -l app=sentinel-postgres
# kubectl exec -it postgres-deployment-d98596cf4-48vq7  -- psql -U sentineluser -d sentineldb
	@echo "✅ Validation completed."