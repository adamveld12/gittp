dev: build-cli
	mkdir -p ./repositories
	./gittp -port 8080 -debug -path ./repositories -masteronly

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
