# docker-detective

## Demo that the FS->JSON Golang code works

```bash
docker build -t dac:test -f test.Dockerfile .
go run main.go --image dac:test | jq
```

## TODO

1. Get the JSON representation of the individual image layers
2. Add in the HTTP server to serve the JSON as a web app
3. Build the front end in D3

How do we want this product to work, exactly?
