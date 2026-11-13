FROM alpine:3.24.3
RUN apk --no-cache upgrade && apk --no-cache add ca-certificates

ARG TARGETOS
ARG TARGETARCH

ADD "bin/${TARGETOS}_${TARGETARCH}/provider" /usr/local/bin/provider

# The NetBox Terraform provider is compiled into the provider binary and served
# in-process (upjet no-fork mode): no Terraform CLI or provider plugin needed.
ENV USER_ID=65532
USER ${USER_ID}
EXPOSE 8080

ENTRYPOINT ["provider"]
