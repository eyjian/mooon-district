ifeq ($(OS),Windows_NT)
	target=mooon-district.exe
else
	target=mooon-district
endif

all: ${target}

${target}: main.go district/district.go
ifeq ($(OS),Windows_NT)
	set GOOS=windows
	set GOARCH=amd64
endif
	go mod tidy && go build -ldflags "-X 'main.buildTime=`date +%Y%m%d%H%M%S`'" -o $@ $<

.PHONY: clean

clean:
	rm -f ${target}

install: ${target}
ifeq ($(OS),Windows_NT)
	copy ${target} %GOPATH%\bin\
else
	cp ${target} $$GOPATH/bin/
endif