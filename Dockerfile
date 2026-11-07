# The image of the server: the API and the web app on one origin. The stages run on the build machine and
# cross-compile, so an arm64 image costs no emulation.
FROM --platform=$BUILDPLATFORM oven/bun:1.4.4 AS web
WORKDIR /src
COPY notes-app/package.json notes-app/bun.lock ./
RUN bun install --frozen-lockfile
COPY notes-app .
# Empty API base: the app calls /api on its own origin.
RUN NUXT_PUBLIC_API_BASE= bun --bun run generate

FROM --platform=$BUILDPLATFORM golang:1.27 AS api
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY notes-go/go.mod notes-go/go.sum ./
RUN go mod download
COPY notes-go .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /notes .

FROM scratch
COPY --from=api /notes /notes
COPY --from=web /src/.output/public /web
USER 65534
EXPOSE 8080
ENTRYPOINT ["/notes"]
CMD ["serve", "--addr", ":8080", "--web", "/web"]
