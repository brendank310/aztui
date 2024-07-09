DESTDIR?=bin

$(DESTDIR):
	mkdir -p $(DESTDIR)

bin/banjo: bin
	go build -o $(DESTDIR)/banjo cmd/main.go

all: bin/banjo
