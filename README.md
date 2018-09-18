# HashAPI

Hash passwords...eventually.

```
Usage of ./hashapi:
  -cert string
    	path to ssl crt file
  -host string
    	hostname to serve
  -key string
    	path to ssl key file
  -port string
    	port to bind the api server to (default "8080")
```

## Test

```
make test
```

## Run

```
make run
```

## Docker

```
make docker-build
make docker-run
```

## SSL

```
./hashapi --cert /path/to/server.crt --key /path/to/server.key
```