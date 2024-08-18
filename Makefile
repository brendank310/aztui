DESTDIR?=bin

$(DESTDIR):
	mkdir -p $(DESTDIR)

bin/banjo: $(DESTDIR) cmd/main.go
	go build -o $(DESTDIR)/banjo cmd/main.go

all: bin/banjo
