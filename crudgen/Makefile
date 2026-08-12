all: l m w

l:
	GOOS=linux GOARCH=amd64 go build -o crudgen_linux

m:
	GOOS=darwin GOARCH=amd64 go build -o crudgen_mac

w:
	GOOS=windows GOARCH=amd64 go build -o crudgen_win.exe
