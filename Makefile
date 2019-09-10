BIN := euraxess-pull
DEPLOYTMPLDIR := ~/var/euraxess
DEPLOYBINDIR := ~/bin

.PHONY: clean linux darwin

clean:
	rm -rf build/

deploy: linux
	scp build/linux/$(BIN) ims.ut.ee:$(DEPLOYBINDIR) && scp offers.tmpl ims.ut.ee:$(DEPLOYTMPLDIR)

all: linux darwin

linux:
	mkdir -p build/linux
	GOOS=linux GOARCH=amd64 go build -o build/linux/$(BIN)

darwin:
	mkdir -p build/linux
	GOOS=darwin GOARCH=amd64 go build -o build/darwin/$(BIN)
