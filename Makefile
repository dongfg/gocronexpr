all: build test
test:
	go test -v ./...
build:
	cd libs/ccronexpr && gcc -c ccronexpr.c -o ccronexpr.o && ar -crs libccronexpr.a ccronexpr.o \
		&& rm -f ccronexpr.o && mv libccronexpr.a ../
	go build -v
clean:
	cd libs/ccronexpr && rm -f ccronexpr.o && rm -f libccronexpr.a && rm -f ../libccronexpr.a