NAME=rlcs
REGISTRY_URL=gcr.io/sousandrei
VERSION=$(shell git rev-parse --short=7 HEAD)

.PHONY: deploy

build:
	docker build . -t ${NAME}

push:
	docker tag ${NAME} ${REGISTRY_URL}/${NAME}:${VERSION}
	docker push ${REGISTRY_URL}/${NAME}:${VERSION}

deploy:
	kubectl kustomize deploy | sed 's/latest/${VERSION}/g' | kubectl apply -f -

destroy:
	kubectl delete -k deploy/
