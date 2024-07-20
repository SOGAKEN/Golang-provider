# 変数
PROJECT_ID := your-project-id
REGION := asia-northeast1
SERVICE_NAME := your-service-name
IMAGE_NAME := your-image-name
ARTIFACT_REPO := your-artifact-repo

# Goのビルドとテスト
.PHONY: build
build:
	go build -o bin/server cmd/server/main.go

.PHONY: run
run:
	go run cmd/server/main.go

.PHONY: test
test:
	go test ./...

.PHONY: migrate
migrate:
	go run cmd/migrate/main.go

# Dockerイメージのビルドとプッシュ
.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE_NAME) .

.PHONY: docker-push
docker-push:
	docker push $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(ARTIFACT_REPO)/$(IMAGE_NAME)

# ビルドしてArtifact Registryへプッシュ
.PHONY: build-push
build-push: docker-build
	docker tag $(IMAGE_NAME) $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(ARTIFACT_REPO)/$(IMAGE_NAME)
	docker push $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(ARTIFACT_REPO)/$(IMAGE_NAME)

# Cloud Runへのデプロイ
.PHONY: deploy
deploy:
	gcloud run deploy $(SERVICE_NAME) \
		--image $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(ARTIFACT_REPO)/$(IMAGE_NAME) \
		--platform managed \
		--region $(REGION) \
		--allow-unauthenticated

# 全てのステップを実行（ビルド、プッシュ、デプロイ）
.PHONY: all
all: build test build-push deploy

# クリーンアップ
.PHONY: clean
clean:
	rm -rf bin

.DEFAULT_GOAL := build
