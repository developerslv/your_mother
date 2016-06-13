VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods -nilfunc -printf -rangeloops -shift -structtags -unsafeptr
GIT_HASH=$(shell git rev-parse HEAD)
GIT_REPO=$(shell git config --get remote.origin.url)

updatedeps:
	go get -u github.com/kardianos/govendor
	govendor fetch +vendor


# test runs the unit tests and vets the code
test:
	go test -timeout=30s -parallel=4 ./...
	@$(MAKE) vet

# testrace runs the race checker
testrace:
	go test -race

cover:
	@go tool cover 2>/dev/null; if [ $$? -eq 3 ]; then \
		go get -u golang.org/x/tools/cmd/cover; \
	fi
	go list -f '{{if gt (len .TestGoFiles) 0}}"go test -covermode count -coverprofile {{.Name}}.coverprofile -coverpkg ./... {{.ImportPath}}"{{end}}' ./... | xargs -I {} bash -c {}
	gocovmerge `ls *.coverprofile` > coverage.out
	go tool cover -html=coverage.out
	rm coverage.out
	rm *.coverprofile

# vet runs the Go source code static analysis tool `vet` to find
# any common errors.
vet:
	@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@echo "go tool vet $(VETARGS) ."
	@go tool vet $(VETARGS) . ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for reviewal."; \
	fi

updateproto:
	wget https://raw.githubusercontent.com/dweinstein/google-play-proto/master/googleplay.proto -O protobuf/googleplay.proto
	protoc -I=protobuf --go_out=protobuf protobuf/googleplay.proto

build:
	go build -ldflags "-X github.com/dainis/your_mother/bot.GitHash=$(GIT_HASH)" -ldflags "-X github.com/dainis/your_mother/bot.GitRepo=$(GIT_REPO)" -o bin/your_mom

publish:
	@${MAKE} build
	docker build -t dainis/your_mother_base -f dockerfiles/base ./
	docker build -t dainis/your_mother_rpc -f dockerfiles/rpc ./
	docker build -t dainis/your_mother_irc -f dockerfiles/irc ./
	docker push dainis/your_mother_rpc
	docker push dainis/your_mother_irc

run_rpc:
	go run main.go rpc -v --irc_logs=irc_logs

run_irc:
	go run main.go irc -v --nick="your_mom_test" --channel="#your_mom_test"

.PHONY: updatedeps vet testrace test cover run build publish
