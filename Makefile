all: district_tool

district_tool: main.go
	go build -o $@ $<

.PHONY: clean

clean:
	rm -f district_tool
