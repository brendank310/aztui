# Variables
DESTDIR ?= $(PWD)/bin
BINARY_NAME ?= aztui
VERSION ?= 0.0.1
RELEASE ?= 1
ARCH ?= $(shell uname -m)
RPMBUILD_DIR ?= $(PWD)/rpmbuild
TARBALL ?= $(BINARY_NAME)-$(VERSION).tar.gz
SRC_DIR ?= $(PWD)/src

# Directories
BINDIR ?= /usr/bin
SPECDIR ?= $(RPMBUILD_DIR)/SPECS
SOURCEDIR ?= $(RPMBUILD_DIR)/SOURCES
BUILDDIR ?= $(RPMBUILD_DIR)/BUILD

$(DESTDIR):
	mkdir -p $(DESTDIR)

$(DESTDIR)/$(BINARY_NAME): $(DESTDIR) $(shell find . -name "*.go")
	cd $(SRC_DIR) && \
	go build -o $(DESTDIR)/$(BINARY_NAME) cmd/main.go && \
	cd ..

clean:
	rm -rf $(DESTDIR)

run:
	go run cmd/main.go

all: $(DESTDIR)/$(BINARY_NAME)

# RPM target (binary RPM)
rpm: prepare_rpm_structure
	# Build the binary from source
	rpmbuild --define "_topdir $(RPMBUILD_DIR)" -bb $(SPECDIR)/$(BINARY_NAME).spec

# Source RPM target
srpm: prepare_rpm_structure tarball
	# Copy tarball to SOURCEDIR
	cp $(TARBALL) $(SOURCEDIR)/
	# Build the source RPM
	rpmbuild --define "_topdir $(RPMBUILD_DIR)" -bs $(SPECDIR)/$(BINARY_NAME).spec

# Create tarball for source RPM
tarball:
	# Create source tarball including current directory (SRC_DIR)
	tar czf $(TARBALL) --transform "s,^,$(BINARY_NAME)-$(VERSION)/," -C $(SRC_DIR) .

# Prepare RPM directories and .spec file
prepare_rpm_structure: tarball
	@mkdir -p $(SPECDIR) $(SOURCEDIR) $(BUILDDIR)

	# Create SPEC file using echo to avoid any target pattern issues
	@echo "Name:           $(BINARY_NAME)" > $(SPECDIR)/$(BINARY_NAME).spec
	@echo "Version:        $(VERSION)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "Release:        $(RELEASE)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "Summary:        My Go binary packaged for CBL Mariner" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "License:        MIT" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "URL:            http://example.com" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "Source0:        $(TARBALL)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "BuildArch:      $(ARCH)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%description" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "A Go application packaged for CBL Mariner." >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%prep" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%setup -q" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%build" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "go build -o \$(BINARY_NAME) ." >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%install" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "install -D -m 0755 \$(BINARY_NAME) \$(BUILDROOT)$(BINDIR)/$(BINARY_NAME)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%files" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "$(BINDIR)/$(BINARY_NAME)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%changelog" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "* $(shell date +"%a %b %d %Y") Me <me@example.com> - $(VERSION)-$(RELEASE)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "- Initial package" >> $(SPECDIR)/$(BINARY_NAME).spec
