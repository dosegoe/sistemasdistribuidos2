BUILDPATH=$(CURDIR)
GO=$(shell which go)
GOBUILD=$(GO) build
GOINSTALL=$(GO) run
GOCLEAN=$(GO) clean

EXENAME=datanode/datanode.go

export GOBIN=$(CURDIR)/bin

build:

	$(GOINSTALL) $(EXENAME)

all: build 
