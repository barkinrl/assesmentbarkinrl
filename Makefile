# This work is licensed under Apache License, Version 2.0 or later.
# Please read and understand latest version of Licence.

all: build

build:
	docker build \
		--build-arg REV=$(shell git describe --long --tags --match='v*' --dirty 2>/dev/null || git rev-list -n1 HEAD) \
		--build-arg BUILD_TIME=$(shell date +'%Y-%m-%d_%T') \
		-t go_react_mui:latest .

build-local:
	./build.sh
