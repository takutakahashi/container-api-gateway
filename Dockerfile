# build stage
FROM golang:alpine AS build-env
RUN apk --no-cache add build-base git bzr mercurial gcc
ADD . /src
RUN cd /src && GO111MODULE=on go build -o cgw cmd/cmd.go

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/cgw /app/cgw
COPY --from=build-env /src/src/template/form.html /app/src/template/form.html
CMD ['./cgw', ' --config', './config.yaml']
