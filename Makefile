.PHONY: run
run:
	docker-compose up

.PHONY: build
build:
	docker build -f ./maple/Dockerfile . -t maple
