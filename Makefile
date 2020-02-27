BIN := euraxess-pull
DEPLOYTMPLDIR := ~/var/euraxess
DEPLOYBINDIR := ~/bin

.PHONY: clean deploy

all: linux darwin

clean:
	rm -rf build/

linux:
	mkdir -p build/linux
	GOOS=linux GOARCH=amd64 go build -o build/linux/$(BIN)

darwin:
	mkdir -p build/linux
	GOOS=darwin GOARCH=amd64 go build -o build/darwin/$(BIN)

deploy: linux
	scp build/linux/$(BIN) ims.ut.ee:$(DEPLOYBINDIR) && scp offers.tmpl ims.ut.ee:$(DEPLOYTMPLDIR)
