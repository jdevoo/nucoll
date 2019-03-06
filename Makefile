BINARY = nucoll
NEWTAG := $(shell git describe --abbrev=0 --tags)
OLDTAG := $(shell git describe --abbrev=0 --tags `git rev-list --tags --skip=1 --max-count=1`)

NIX_BINARIES = linux/amd64/$(BINARY) darwin/amd64/$(BINARY)
WIN_BINARIES = windows/amd64/$(BINARY).exe
COMPRESSED_BINARIES = $(NIX_BINARIES:%=%.bz2) $(WIN_BINARIES:%.exe=%.zip)
COMPRESSED_TARGETS = $(COMPRESSED_BINARIES:%=target/%)

temp = $(subst /, ,$@)
OS = $(word 2, $(temp))
ARCH = $(word 3, $(temp))
GITHASH = $(shell git log -1 --pretty=format:"%h")
GOVER = $(word 3, $(shell go version))
LDFLAGS = -ldflags '-X main.version=$(NEWTAG) -X main.githash=$(GITHASH) -X main.golang=$(GOVER)'

RELEASE_TOOL = github-release
USER = jdevoo

all: $(BINARY)

target/linux/amd64/$(BINARY):
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) -o "$@"
target/darwin/amd64/$(BINARY):
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) -o "$@"
target/windows/amd64/$(BINARY).exe:
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) -o "$@"

%.bz2: %
	tar -cjf "$@" -C $(dir $<) $(BINARY)

%.zip: %.exe
	zip -j "$@" "$<"

$(BINARY):
	go build $(LDFLAGS) -o $(BINARY)

install:
	go install $(LDFLAGS)
	$(BINARY) -v

# git tag v0.1
release:
	$(MAKE) $(COMPRESSED_TARGETS)
	git push && git push --tags
	git log --pretty=format:"%s" $(OLDTAG)...$(NEWTAG) | $(RELEASE_TOOL) release -u $(USER) -r $(BINARY) -t $(NEWTAG) -n $(NEWTAG) -d - || true
	$(foreach FILE, $(COMPRESSED_BINARIES), $(RELEASE_TOOL) upload -u $(USER) -r $(BINARY) -t $(NEWTAG) -n $(subst /,-,$(FILE)) -f target/$(FILE);)

clean:
	rm -f $(BINARY)
	rm -rf target

test:
	go test -v ./...

.PHONY: install release clean test
