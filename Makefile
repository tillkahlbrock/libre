.PHONY: build clean deploy

build:
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/parse parse/main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

test:
	go test ./...

deploy: guard-USERNAME guard-PASSWORD guard-BASE_URL clean build test
	sls deploy --verbose

guard-%:
	@ if [ "${${*}}" = "" ]; then \
		echo "Environment variable $* not set"; \
		exit 1; \
	fi