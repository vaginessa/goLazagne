include .env
export
ssh_opt=-o StrictHostKeyChecking=no -p $(DEPLOY_PORT)
define upload
	rsync -vz \
		--checksum \
		--rsh="ssh $(ssh_opt)" \
		$(1) \
		${DEPLOY_HOST}:$(DEPLOY_PATH)/$(patsubst %,%,$(2))
endef
define download
	rsync -avz \
		--checksum \
		--rsh="ssh $(ssh_opt)" \
		${DEPLOY_HOST}:$(DEPLOY_PATH)/$(patsubst %,%,$(1)) \
		$(2)
endef

define command
	ssh $(ssh_opt)  $(DEPLOY_HOST) "cd $(DEPLOY_PATH) && "$(1)
endef
.PHONY: build upload run clean test
GO_SRCS := $(wildcard *.go) $(shell find . -type f -name '*.go')


all: upload
target_exe=goLazagne.exe
go=GOOS=windows GOARCH=amd64 CGO_ENABLED="1" CC="x86_64-w64-mingw32-gcc" go

build: $(target_exe)
$(target_exe): $(GO_SRCS) Makefile
	$(go) build  -ldflags='-s -w' -o $(target_exe) -v
	upx $(target_exe)
	ls -lh $(target_exe)

debug:
	$(go) build -gcflags "all=-N -l" -o $(target_exe) -v
	ls -lh $(target_exe)
	$(call upload, $(target_exe), ./)
	$(call command,"bash ./kill_dlv.sh")
	$(call command, \
		"dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./$(target_exe) ")


upload: build
	$(call upload, $(target_exe), ./)
run: upload
	$(call command,"bash ./kill_dlv.sh")
	$(call command, " ./$(target_exe) ")

clean:
	find . -type f -name '*.exe' -delete
	#rm -f $(target_exe)


upload_sh:
	$(call upload, ./kill_dlv.sh, ./)
