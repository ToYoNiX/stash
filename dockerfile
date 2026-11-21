FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . /src
RUN apk add --no-cache gcc musl-dev
# Why CGO_ENABLED? https://stackoverflow.com/questions/61515186/when-using-cgo-enabled-is-must-and-what-happens
# Sqlite needs it. and it was disabled by default at some point so we need to enable it manually
RUN CGO_ENABLED=1 go build -o /bin/stash

FROM alpine
EXPOSE 8090
COPY ./migrations /migrations
COPY --from=build /bin/stash /bin/stash
CMD ["/bin/stash", "-db", "/data/db.sqlite3", "-uploads", "/data/uploads", "-addr", ":8090", "-migrations", "/migrations"]
