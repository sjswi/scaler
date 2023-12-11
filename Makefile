.PHONY: build

export GO111MODULE ?= on
export GOPROXY ?= https://goproxy.cn
export GOSUMDB ?= sum.golang.org


LOCAL_ARCH := $(shell uname -m)
ifeq ($(LOCAL_ARCH),x86_64)
	TARGET_ARCH_LOCAL=amd64
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 5),armv8)
	TARGET_ARCH_LOCAL=arm64
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 4),armv)
	TARGET_ARCH_LOCAL=arm
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 5),arm64)
	TARGET_ARCH_LOCAL=arm64
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 7),aarch64)
	TARGET_ARCH_LOCAL=arm64
else
	TARGET_ARCH_LOCAL=amd64
endif
export GOARCH ?= $(TARGET_ARCH_LOCAL)

build:
	go build -o scaler main.go

docker-build:
	docker build -t 10.10.150.23:35000/ax-scaler:v0.0.1 .


docker-push:
	docker buipushld 10.10.150.23:35000/ax-scaler:v0.0.1
