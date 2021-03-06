# pingo

This is a naive implementation of the `ping` command using Go's `x/net/icmp` package. This is a learning/experimentation exercise.

![Stranger Pings](stranger-pings.jpg)

_Image credits:_ https://twitter.com/alicegoldfuss/status/923947130074439680

## Building

A `make build` should build a binary called `pingo`.

## Running

A `make run` should build and run the program. Currently, these are the parameters you can change when running `pingo`:

```sh
Usage: ./pingo host
  -c uint
        number of packets to be sent and received; if not specified, ./pingo will send requests until interrupted
  -s uint
        number of data bytes to be sent in each request (default 56)
  -t uint
        timeout in seconds for each request (default 1)
```

**Note:** You need `sudo` privileges in order to send ping requests, so `make run` uses `sudo` for running the binary that is built.


## Acknowledgements

I've used [go-fastping](https://github.com/tatsushid/go-fastping) a lot as a reference  while trying to understand how to implement ping, and a few fuctions were directly copied from it (see comments in the code) and adapted for use with my own implementation.
