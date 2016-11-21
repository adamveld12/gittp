all: clean check dev git_debug test

dev: gittp
	mkdir -p ./repositories
	./gittp -addr :8080 -autocreate -masteronly -debug -path ./repositories

setup:
	go get -u github.com/golang/lint/golint
	go get -t -d -v ./...

clean:
	rm -rf ./repositories
	rm -rf ./gittp

test:
	go test -v -cover

test_cover: c.out
	go tool -v cover -html=c.out

git_debug:
	export HTTP_PROXY=http://localhost:8080
	export GIT_CURL_VERBOSE=1

check: test
	golint $$(pwd)/*.go

.PHONY: check git_debug test_cover clean test dev all

c.out:
	go test -v -coverprofile=c.out

gittp:
	go build -a -v -o ./gittp ./cli/gittp
