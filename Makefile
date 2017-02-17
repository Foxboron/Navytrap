 all:
	go build
	go build -buildmode=plugin ./plugins/*.go
