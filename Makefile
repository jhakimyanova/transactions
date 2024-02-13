.PHONY: build clean rebuild deploy

clean:
	rm -rf ./aws-sam

build:
	sam build

rebuild: clean build

deploy: rebuild
	sam deploy --guided
