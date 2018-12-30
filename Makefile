SHELL:=/bin/bash

default:
	make clean && make build && make test

prepare:
	if ! [[ $$(which go) ]]; then \
		echo Go not found, please make sure that go is installed and in the \$$PATH; \
		echo For example, on ubuntu this package would be obtained with: apt install golang; \
	else \
		echo "Go found, should be OK to run: make build && make install"; \
	fi

clean:
	rm -rf jirahours

build:
	go build

test:
	go test

format:
	go test

install:
	cp -prv jirahours /usr/local/bin/ || \
	cp -prv jirahours /usr/bin/ || \
	cp -prv jirahours ~/.bin/ || \
	cp -prv jirahours ~/bin/ || \
	cp -prv jirahours ~/go/bin/

uninstall:
	mv -t /tmp /usr/local/bin/jirahours || \
	mv -t /tmp /usr/bin/jirahours || \
	mv -t /tmp ~/.bin/jirahours || \
	mv -t /tmp ~/bin/jirahours || \
	mv -t /tmp ~/go/bin/jirahours
