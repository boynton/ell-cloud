PKG=github.com/boynton/ell-cloud
BIN=$(GOPATH)/bin/cell

all: bin/cell
#	go install $(PKG)/cell

prebuild:
	go get -u github.com/boynton/ell
	go get -u github.com/boynton/repl
#	go get -u github.com/pborman/uuid

bin/cell: cloud.go
	go install $(PKG)/cell

test:
#	go run cloudtest.go

run:
	./bin/cell

clean:
	go clean $(PKG)
	rm -rf *~ $(BIN)

check:
	@(cd $(GOPATH)/src/$(PKG); go vet $(PKG); go fmt $(PKG))
