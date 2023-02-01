# SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>
#
# SPDX-License-Identifier: CC0-1.0

.PHONY: test
test:
	go test ./...

.PHONY: deps
deps: deps-npm deps-pip deps-go

.PHONY: deps-pip
deps-pip:
	pip install --user reuse

.PHONY: deps-npm
deps-npm: node_modules

node_modules: package.json
	npm install

.PHONY: deps-go
deps-go:
	go mod download
	go install golang.org/x/tools/cmd/goimports@latest

.PHONY: lint
lint: lint-md lint-license lint-go

.PHONY: lint-fix
lint-fix: lint-md-fix lint-go-fix

.PHONY: lint-md
lint-md: node_modules
	npx remark .

.PHONY: lint-md-fix
lint-md-fix: node_modules
	npx remark . -o

.PHONY: lint-license
lint-license:
	reuse lint

.PHONY: lint-go
lint-go:
	@echo goimports -d '**/*.go'
	@goimports -d $(shell git ls-files "*.go")

.PHONY: lint-go-fix
lint-go-fix:
	@echo goimports -d -w '**/*.go'
	@goimports -d -w $(shell git ls-files "*.go")
