include scripts/serverless.mk

PATH_FUNCTIONS := ./src/
LIST_FUNCTIONS := $(subst $(PATH_FUNCTIONS),,$(wildcard $(PATH_FUNCTIONS)*))

clean:
	@ rm -rf ./dist

test: export GO111MODULE=on
test:
	@ go test ./...

build-%: export GO111MODULE=on
build-%:
	@ GOOS=linux GOARCH=amd64 go build -o ./dist/handler/$* ./src/$*

build: clean
build:
	@ make $(foreach FUNCTION,$(LIST_FUNCTIONS),build-$(FUNCTION))
