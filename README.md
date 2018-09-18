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

## Endpoints

    POST /hash with payload=`password=some_password`
      - returns a hash id

    GET /hash/id
      - will return the hashed password if 5 seconds since creation have elapsed

    GET /stats
      - returns a JSON response with total hashes posted and the average time to handle those posts

    GET /shutdown
      - graceful shutdown of service

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