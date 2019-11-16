# container-api-gateway

This is the frontend API with common Docker container

### WARNING  
This app has **NO** concern about security issues.  
**DO NOT USE** in production and **EXPOSE** the Internet.  

## Requirement

- Docker

## Usage

This is the tutorial for using a `hello-world` container image.

```
 % docker run -it hello-world
Hello from Docker!
This message shows that your installation appears to be working correctly.

To generate this message, Docker took the following steps:
 1. The Docker client contacted the Docker daemon.
 2. The Docker daemon pulled the "hello-world" image from the Docker Hub.
    (amd64)
 3. The Docker daemon created a new container from that image which runs the
    executable that produces the output you are currently reading.
 4. The Docker daemon streamed that output to the Docker client, which sent it
    to your terminal.

To try something more ambitious, you can run an Ubuntu container with:
 $ docker run -it ubuntu bash

Share images, automate workflows, and more with a free Docker ID:
 https://hub.docker.com/

For more examples and ideas, visit:
 https://docs.docker.com/get-started/
```

### 1. Prepare config.yaml

You need to prepare config.yaml.

```
endpoints:
  - path: /test
    method: GET
    container:
      image: hello-world
```

### 2. Execute binary

User must be able to execute docker API (default: root).

```
# cgw --config path/to/config.yaml
```

### 3. Execute API

If you want to call API, use `curl` ex:

```
 % curl http://localhost:8080/test
{"stdout":"\nHello from Docker!\nThis message shows that your installation appears to be working correctly.\n\nTo generate this message, Docker took the following steps:\n 1. The Docker client contacted the Docker daemon.\n 2. The Docker daemon pulled the \"hello-world\" image from the Docker Hub.\n    (amd64)\n 3. The Docker daemon created a new container from that image which runs the\n    executable that produces the output you are currently reading.\n 4. The Docker daemon streamed that output to the Docker client, which sent it\n    to your terminal.\n\nTo try something more ambitious, you can run an Ubuntu container with:\n $ docker run -it ubuntu bash\n\nShare images, automate workflows, and more with a free Docker ID:\n https://hub.docker.com/\n\nFor more examples and ideas, visit:\n https://docs.docker.com/get-started/\n\n","stderr":""}
```

## Features

### 1. Parameters

You can define some parameters.

```
endpoints
  - path: /hello
    method: GET
    async: false
    params:
      - name: fullname
        optional: false
    container:
      image: ubuntu
      command:
        - "bash"
        - "-c"
        - "echo hello, {{ fullname }}" # set params[].name
```

Output:

```
 % curl "http://localhost:8080/hello"
 required param fullname was not found.%                               
 % curl "http://localhost:8080/hello?fullname=bob"
{"stdout":"hello, bob\n","stderr":""}
```


If you want to get all features, please see `example/config.yaml`.
