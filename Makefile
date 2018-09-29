GOBUILD = go build

ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
# TC_PATH := $(ROOT_DIR)/toolchain
TC_PATH := /home/sk/buildroot/buildroot-rpi1/output/host

TC_PREFIX := arm-buildroot-linux-uclibcgnueabihf
CC := ${TC_PREFIX}-gcc
LD := ${TC_PREFIX}-ld
AS := ${TC_PREFIX}-as
CXX := ${TC_PREFIX}-g++

SDL2_CFLAGS = `sdl2-config --cflags`
SDL2_LDFLAGS = `sdl2-config --libs`

export PATH := $(TC_PATH)/bin:$(TC_PATH)/$(TC_PREFIX)/sysroot/usr/bin:$(PATH)

GOOS := linux
GOARCH := arm
GOARM := 6

.PHONY: all clean
.DEFAULT: all

all: gosdl-arm gosdl-local

gosdl-arm:
	env \
	CGO_ENABLED="1" \
	CXX=$(CXX) \
	CC=$(CC) \
	CGO_CFLAGS="$(SDL2_CFLAGS)" \
	CGO_LDFLAGS="$(SDL2_LDFLAGS)" \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	GOARM=$(GOARM) \
	$(GOBUILD) -x -o gosdl-$(GOOS)-$(GOARCH)$(GOARM)


gosdl-local:
	$(GOBUILD) -x -o gosdl-local

clean:
	-rm -f gosdl-*

test:
	# echo $(PATH)
	# echo -----------
	# which sdl2-config
	# echo -----------
	# sdl2-config --cflags
	# echo -----------
	echo $(SDL2_CFLAGS)
	echo $(SDL2_LDFLAGS)