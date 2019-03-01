FROM alpine:3.9
WORKDIR /function/
COPY bin/oci-lego-sslupdate .
RUN apk update && \
	apk add --no-cache ca-certificates
ENTRYPOINT ["/function/oci-lego-sslupdate"]
