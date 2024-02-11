protos:
	protoc ./router/proto/api.proto \
		--go_opt=module=github.com/octavore/nagax \
		--go_out=.

goimports:
	GO_DIRS=$$(find . -name "*.go" -exec dirname {} \; | sort -u); \
		$$GOBIN/goimports -w -local github.com/octavore/nagax $$GO_DIRS
