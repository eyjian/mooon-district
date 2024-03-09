target=mooon-district

all: ${target}

${target}: main.go district/district.go
	go build -o $@ $<

.PHONY: clean

clean:
	rm -f ${target}
