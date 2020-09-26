GOBUILD = go build

ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
# TC_PATH := $(ROOT_DIR)/toolchain
# TC_PATH := /home/sk/buildroot/buildroot-rpi1-oled/output/host
TC_PATH := /home/sk/buildroot/buildroot-2018.08.1/output/host/

TC_PREFIX := arm-buildroot-linux-uclibcgnueabihf
# TC_PREFIX := arm-buildroot-linux-gnueabihf
CC := ${TC_PREFIX}-gcc
LD := ${TC_PREFIX}-ld
AS := ${TC_PREFIX}-as
CXX := ${TC_PREFIX}-g++

export PKG_CONFIG_PATH := $(TC_PATH)/$(TC_PREFIX)/sysroot/usr/lib/pkgconfig:$(PKG_CONFIG_PATH)

SDL2_CFLAGS = `sdl2-config --cflags`
SDL2_LDFLAGS = `sdl2-config --libs`

export PATH := $(TC_PATH)/bin:$(TC_PATH)/$(TC_PREFIX)/sysroot/usr/bin:$(PATH)

DEPLOY_HOST := root@192.168.178.96
GOOS := linux
GOARCH := arm
GOARM := 6

.PHONY: all clean
.DEFAULT: all

all: gosdl-arm gosdl-local

gosdl-arm: *.go
	env \
	CGO_ENABLED="1" \
	CXX=$(CXX) \
	CC=$(CC) \
	CGO_CFLAGS="$(SDL2_CFLAGS)" \
	CGO_LDFLAGS="$(SDL2_LDFLAGS)" \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	GOARM=$(GOARM) \
	$(GOBUILD) -tags egl -x -o gosdl-$(GOOS)-$(GOARCH)$(GOARM)


gosdl-local: *.go
	env \
	CGO_ENABLED=1 \
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

deploy: gosdl-arm
	# rsync -h -v -r -P -t gosdl-linux-arm6 root@192.168.178.96:/root/
	ssh $(DEPLOY_HOST) /etc/init.d/S03app stop
	scp gosdl-linux-arm6 marken.ttf db.png $(DEPLOY_HOST):/root/
	ssh $(DEPLOY_HOST) /etc/init.d/S03app start

remote: gosdl-arm
	ssh $(DEPLOY_HOST) killall gosdl-linux-arm6 || true
	scp gosdl-linux-arm6 marken.ttf ProggyTiny.ttf db.png $(DEPLOY_HOST):/root/
	ssh $(DEPLOY_HOST) /root/gosdl-linux-arm6