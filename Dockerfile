# build stage
FROM golang:1.24 AS build-env
RUN apt update && apt install -y build-essential git bzr mercurial gcc
ADD . /src
RUN cd /src && GO111MODULE=on go build -o cgw cmd/cmd.go

# final stage
FROM ubuntu
WORKDIR /app
COPY --from=build-env /src/cgw /app/cgw
COPY --from=build-env /src/src/template/form.html /app/src/template/form.html
CMD ['./cgw', ' --config', './config.yaml']
