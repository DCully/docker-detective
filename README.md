# docker-detective

## The DB building step is slow

`ros:melodic-robot`
   with an on-disk DB
      takes ~3 minutes
      takes ~1 minute without doing the unfinished file size updates at all
      takes ~41 seconds when SaveFile is stubbed out
   with an in-memory DB
      takes ~43 seconds
      takes ~41 seconds without doing the unfinished file size updates at all
      takes ~39 seconds when SaveFile is stubbed out
   with no reader interactions whatsoever, 22 seconds

it takes 20 seconds to docker save ros:melodic-robot, very similar to our reader interaction time
compilation takes ~ 2 seconds
parallelizing things makes the whole in-memory DB building take ~17 seconds


## Demo that the FS->JSON Golang code works

```bash
docker build -t dac:test -f test.Dockerfile .
go run main.go --image dac:test | jq
```

JSON is like:
```
{
    "image": {... a root file system metadata tree...},
    "layers": {
        "config": {...},
        "manifest": [...],
        "layerIdsToRootFSs": {
            "a layer sha": {... a root file system metadata tree...},
            "another sha": {... a root file system metadata tree...},
            ...
        }
    }
}
```

## TODO

1. Build the front end for data visualization
2. Build a website that uses one hard-coded JSON blob to show a demo
3. Drive traffic to the website and measure click-through rate, collect emails
4. Build the rest of the product
   1. Extend the CLI so that it accepts a token and can upload to our API
   2. Write our API backend to accept these authenticated uploads
   3. Design and write our product backend to render our product app (how to design this?)

How do we want this product to work, exactly?
