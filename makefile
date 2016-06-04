dev: clean build-cli
	mkdir -p ./repositories
	./gittp -port 8080 -autocreate -masteronly -debug -path ./repositories 

build-cli:
	go build -race ./cli/gittp

clean:
	rm -rf ./repositories
	rm -rf ./gittp
	
test:
	go test -cover

git_debug:
	export HTTP_PROXY=http://localhost:8080
	export GIT_CURL_VERBOSE=1

check:
	golint $$(pwd)/*.go
