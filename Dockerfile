# Built following https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324

# STEP 1 build executable binary
FROM golang:alpine as builder
# Install SSL ca certificates
RUN apk update && apk add --no-cache \
  git=2.38.3-r1 \
  ca-certificates=20220614-r4
# Create appuser
RUN adduser -D -g '' appuser
COPY . $GOPATH/src/alertmanager-discord/
WORKDIR $GOPATH/src/alertmanager-discord/
ARG APPLICATION_VERSION
ENV APPLICATION_VERSION="${APPLICATION_VERSION:-"0.0.0"}"
#get dependancies
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s -X 'github.com/specklesystems/alertmanager-discord/pkg/version.Version=${APPLICATION_VERSION}'" -o /go/bin/alertmanager-discord main.go


# STEP 2 build a small image
# start from scratch
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
# Copy our static executable
COPY --from=builder /go/bin/alertmanager-discord /bin/alertmanager-discord

ENV LISTEN_ADDRESS=0.0.0.0:9094
EXPOSE 9094
USER appuser
ENTRYPOINT ["/bin/alertmanager-discord"]
