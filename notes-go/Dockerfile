# Build with `make api`: plum comes from .docker-vendor, go.mod replaces it by a path that only exists on the host.
FROM golang:1.27 AS build
WORKDIR /src
COPY . .
RUN mv .docker-vendor vendor && CGO_ENABLED=0 go build -mod=vendor -o /notes .

FROM scratch
COPY --from=build /notes /notes
USER 65534
EXPOSE 8080
ENTRYPOINT ["/notes"]
