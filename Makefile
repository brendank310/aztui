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
	cd $(SRC_DIR) && \
	AZTUI_CONFIG_PATH=$(PWD)/conf/default.yaml go run cmd/main.go && \
	cd ..

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
	@mkdir -p $(SOURCEDIR)
	# Create source tarball including current directory (SRC_DIR)
	tar czf $(SOURCEDIR)/$(TARBALL) --transform "s,^,$(BINARY_NAME)-$(VERSION)/," -C $(SRC_DIR) .

# Prepare RPM directories and .spec file
prepare_rpm_structure: tarball
	@mkdir -p $(SPECDIR) $(SOURCEDIR) $(BUILDDIR)

	# Create SPEC file using echo to avoid any target pattern issues
	@echo "Name:           $(BINARY_NAME)" > $(SPECDIR)/$(BINARY_NAME).spec
	@echo "Version:        $(VERSION)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "Release:        $(RELEASE)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "Summary:        A text UI for managing Azure resources" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "License:        Apache2" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "URL:            http://github.com/brendank310/aztui" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "Source0:        $(TARBALL)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "BuildArch:      $(ARCH)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%description" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "A text UI for managing Azure resources." >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%prep" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%setup -q" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%build" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "go build -o $(BINARY_NAME) $(SRC_DIR)/cmd/main.go" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%install" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "install -D -m 0755 $(BINARY_NAME) %{buildroot}$(BINDIR)/$(BINARY_NAME)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%files" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "$(BINDIR)/$(BINARY_NAME)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "%changelog" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "* $(shell date +"%a %b %d %Y") Brendan Kerrigan <bkerrig1@binghamton.edu> - $(VERSION)-$(RELEASE)" >> $(SPECDIR)/$(BINARY_NAME).spec
	@echo "- Initial package" >> $(SPECDIR)/$(BINARY_NAME).spec

format:
	gofmt -s -w ./src/
