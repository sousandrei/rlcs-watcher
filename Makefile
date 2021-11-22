NAME=rlcs
REGISTRY_URL=gcr.io/sousandrei
VERSION=$(shell git rev-parse --short=7 HEAD)


build:
	docker build . -t ${NAME}

push:
	docker tag ${NAME} ${REGISTRY_URL}/${NAME}:${VERSION}
	docker push ${REGISTRY_URL}/${NAME}:${VERSION}

deploy: build push
	k apply -k deploy

destroy:
	k delete -k deploy
