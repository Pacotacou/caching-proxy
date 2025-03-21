# Caching Proxy

A lightweight caching HTTP proxy server written in Go. This tool forwards HTTP requests to a specified origin server and caches the responses. Subsequent identical requests are served from the cache, improving response times and reducing load on the origin server.

[Project URL](https://roadmap.sh/projects/caching-server)

## Features

- Forward HTTP requests to a specified origin server
- Cache responses in memory for faster subsequent requests
- Clear indication of cache hits and misses via response headers
- Simple command-line interface
- Cache clearing functionality

## Installation

### Prerequisites

- Go 1.16 or later

### Building from source

1. Clone the repository
   ```bash
   git clone https://github.com/yourusername/caching-proxy.git
   cd caching-proxy
   ```

2. Download dependencies
   ```bash
   go mod tidy
   ```

3. Build the application
   ```bash
   go build -o caching-proxy
   ```

## Usage

### Starting the proxy server

```bash
./caching-proxy --port <port_number> --origin <origin_url>
```

Parameters:
- `--port`: The port on which the proxy server will listen
- `--origin`: The URL of the origin server to which requests will be forwarded

Example:
```bash
./caching-proxy --port 3000 --origin http://dummyjson.com
```

This starts the proxy server on port 3000, forwarding requests to http://dummyjson.com.

### Clearing the cache

```bash
./caching-proxy --clear-cache --port <port_number>
```

This command sends a request to a running proxy server on the specified port to clear its cache. The port number must match the port of the running proxy server.

Example:
```bash
# Clear the cache of the proxy server running on port 3000
./caching-proxy --clear-cache --port 3000
```

## How It Works

1. When a request is received, the proxy first checks if a matching response is stored in its cache.
2. If found (cache hit), it returns the cached response with the header `X-Cache: HIT`.
3. If not found (cache miss), it forwards the request to the origin server, caches the response, and returns it with the header `X-Cache: MISS`.

## Example

Start the proxy server:
```bash
./caching-proxy --port 3000 --origin http://dummyjson.com
```

Make a request to the proxy:
```bash
curl -v http://localhost:3000/products
```

The first request will show `X-Cache: MISS` in the response headers. Make the same request again:
```bash
curl -v http://localhost:3000/products
```

The second request should show `X-Cache: HIT`, indicating that the response was served from the cache.



Clear the cache either through the API:
```bash
curl -X DELETE http://localhost:3000/admin/cache
```

Or using the command-line tool:
```bash
./caching-proxy --clear-cache --port 3000
```

## Project Structure

```
caching-proxy/
├── main.go        # Entry point for the application
├── cmd/           # Command-line interface handling
│   └── root.go    # Root command definition
├── proxy/         # Proxy server implementation
│   └── server.go  # HTTP proxy server
├── cache/         # Caching implementation
│   └── cache.go   # In-memory cache
└── go.mod         # Go module file
```

## Limitations and Future Improvements

- Currently, the cache has no expiration mechanism
- No cache size limits are implemented
- Cache is stored in memory and not persisted between restarts
- Limited support for different HTTP methods and complex request scenarios
- No HTTPS support
- Basic authentication could be added to the admin endpoints for security
- Metrics and monitoring could be improved (request counts, cache hit ratios, etc.)
