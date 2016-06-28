all:
	PROTO_FILES=$$(find . -name "*.proto") ; \
		PROTO_DIR=$$(find . -name "*.proto" -exec dirname {} \; | sort -u | sed -e 's/^/-I/'); \
		protoc $$PROTO_DIR $$PROTO_FILES \
		--go_out=import_prefix_proto=github.com/octavore/nagax/proto/,plugins=setter+grpc:./proto

