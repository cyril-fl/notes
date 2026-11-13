FROM golang:1.27 AS build
WORKDIR /src
# Modules first: a code change does not download them again.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# The version shown to the admin: the context has no .git for go build to read the commit from.
ARG VERSION
RUN CGO_ENABLED=0 go build -ldflags "-X github.com/cyril-fl/notes/notes-go/config.version=$VERSION" -o /notes .
# scratch has no /tmp, where an import writes its body instead of holding it in memory (#119).
RUN install -d /image-tmp

FROM scratch
COPY --from=build /notes /notes
COPY --from=build --chown=65534:65534 /image-tmp /tmp
USER 65534
EXPOSE 8080
ENTRYPOINT ["/notes"]
