# go-wasm-table

<iframe src="https://giphy.com/embed/xT0BKumCMrUb0dCypa" width="480" height="480" frameBorder="0" class="giphy-embed" allowFullScreen></iframe><p><a href="https://giphy.com/gifs/loop-infinite-xT0BKumCMrUb0dCypa">via GIPHY</a></p>

This is my attempt to learn Go through building something familiar: An infinite scroll table.

## Building

Building consists of just running the `build` step in the Makefile. This will compile the main application and client into the `build` directory.

```
make build
```

I've included the Go HTTP server example from the WebAssembly page to act as a web server. To build that,

```
go build -o server ./server.go
```

## Running

Once built, you will need a web server to serve the HTML and compiled wasm application. Either run the included web server, or some other server that serves `application/wasm` mime types for the main.wasm file.

### Included server

```
./server -dir ./build
listening on ":8080"...
```

## Demo gif

![infinite-sorta](https://user-images.githubusercontent.com/177652/44961068-ad644a00-aed8-11e8-9616-a3f5d4776aa6.gif)

## Todo

☐ Page rendering
☐ Feeding / Requesting data
☐ Mouse interactions