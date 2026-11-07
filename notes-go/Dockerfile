FROM golang:1.27 AS build
WORKDIR /src
# Modules first: a code change does not download them again.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /notes .

FROM scratch
COPY --from=build /notes /notes
USER 65534
EXPOSE 8080
ENTRYPOINT ["/notes"]
