PREFIX ?= example
PROJECT := url-shortener

AWS_BUCKET_NAME := $(PREFIX)-$(PROJECT)-artifacts
AWS_STACK_NAME := $(PREFIX)-$(PROJECT)-stack
AWS_REGION ?= eu-west-1
AWS_PROFILE ?= default 

FILE_TEMPLATE = ./template.yml
FILE_PACKAGE = ./dist/stack.yml

configure:
	@ aws --profile $(AWS_PROFILE) s3api create-bucket \
		--bucket $(AWS_BUCKET_NAME) \
		--region $(AWS_REGION) \
		--create-bucket-configuration LocationConstraint=$(AWS_REGION)

package:
	@ mkdir -p dist
	@ aws --profile $(AWS_PROFILE) cloudformation package \
		--template-file $(FILE_TEMPLATE) \
		--s3-bucket $(AWS_BUCKET_NAME) \
		--output-template-file $(FILE_PACKAGE) \
		--region $(AWS_REGION)

deploy:
	@ aws --profile $(AWS_PROFILE) cloudformation deploy \
		--template-file $(FILE_PACKAGE) \
		--region $(AWS_REGION) \
		--capabilities CAPABILITY_IAM \
		--stack-name $(AWS_STACK_NAME) \
		--force-upload \
		--parameter-overrides \
			PREFIX=$(PREFIX) \
			PROJECT=$(PROJECT)

destroy:
	@ aws --profile $(AWS_PROFILE) cloudformation delete-stack \
		--stack-name $(AWS_STACK_NAME) \

describe:
	@ aws --profile $(AWS_PROFILE) cloudformation describe-stacks \
		--region $(AWS_REGION) \
		--stack-name $(AWS_STACK_NAME) \
		--query "Stacks[0].Outputs"

.PHONY: clean configure package deploy describe outputs