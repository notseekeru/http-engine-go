# TCP/HTTP Engine

A minimal HTTP/1.1 server built from scratch using only the Go standard library. This project touches the raw TCP wire to understand how HTTP works under the hood.

This project doesn't include `net/http` as it abstracts away many networking logics needed to bridge networking fundamentals.

I wrote this code entirely myself without relying on AI for core logic or conventions. I used AI solely for syntax, boilerplate generation, and as a critique tool to better understand underlying systems and hand-picked adviced implementation on engineering the engine.

## Features

### Raw TCP Listener

Binds to `:8080` with `net.Listen` and accepts connections in a loop. No framework wrappers — the socket is yours.

**Why:** You see exactly what `net/http` hides: the listener lifecycle, the `Accept` block, and the deferred `Close` that releases the port.

### Per-Connection Goroutines

Each accepted connection spawns its own goroutine via `go handleConnection(conn)`. The main loop goes back to listening immediately.

**Why:** Blocking on sequential handling would let one slow client stall everyone. Goroutines give you M:N concurrency with zero thread-pool configuration.

### Connection Deadline (5s)

`conn.SetDeadline(time.Now().Add(5 * time.Second))` kills unresponsive clients after 5 seconds.

**Why:** Without this, a client that opens a TCP socket and never sends data holds the goroutine forever — a Slowloris primitive. A hard deadline is the simplest defense.

### Request Keep-Alive Loop

After writing the response, the goroutine loops back to read the next request on the same connection. Exits only on client disconnect (`EOF`) or when the client sends `Connection: close`.

**Why:** Real pages load many resources (CSS, JS, images). Reusing one TCP socket for all of them avoids the TCP handshake + slow-start penalty per file. The loop is the minimum plumbing needed — no connection-level state, no pipelining reordering, just serial request handling on a warm socket.

### Buffered I/O with `bufio`

Wraps the raw `net.Conn` in a `bufio.Reader` before reading the request line and headers.

**Why:** Raw TCP gives you whatever the kernel delivers — fragmented headers, partial lines, leftover bytes between reads. `bufio.Reader` handles buffering and lets you read line-by-line with `ReadString('\n')`, matching HTTP's line-delimited header format.

### Request Line Parsing

Splits the first line on spaces into `[Method, Path, Version]`. Rejects anything that doesn't match the triplet format with `400 Bad Request`.

**Why:** HTTP/1.1 mandates `METHOD /path HTTP/1.1`. A malformed request line is unrecoverable — reject early before any header parsing.

### HTTP Method Validation

Only `GET` and `POST` are accepted. Anything else gets `405 Method Not Allowed`.

**Why:** This engine doesn't implement `PUT`, `DELETE`, etc. Accepting them without implementation would lie to the client. A proper allowlist prevents silent misbehavior.

### HTTP Version Check

The server only speaks `HTTP/1.1`. Anything else gets `400 Bad Request`.

**Why:** HTTP/1.0 lacks mandatory `Host` headers and connection semantics differ. Supporting multiple versions adds complexity that doesn't fit this project's scope.

### Header Key-Value Parsing

Iterates header lines until a blank line (`\r\n`) — the end-of-headers marker — then stores each in a `map[string]string`.

**Why:** Headers are structured request metadata (Content-Type, User-Agent, etc.) needed for request routing and body handling. The blank-line delimiter is part of the HTTP spec; matching it exactly is correct parsing, not magic.

### Content-Length Body Reading

If `Content-Length` is present, parses it to `int64` and reads exactly that many bytes with `io.LimitReader` + `io.ReadAll`.

**Why:** Without `Content-Length`, you don't know where the headers end and the body ends. Reading blindly into the next request (HTTP pipelining) would corrupt data. `LimitReader` bounds the read so a malicious or broken client can't exhaust memory.

### Query String Parsing

Extracts `?key=val` from the request path into `map[string]string` with the endpoint path stored under the `"endpoint"` key. Guarded against paths without `?`.

**Why:** Real endpoints read query parameters. Without parsing, `/search?q=go` is just `/search` and you lose data.

### Route Dispatch

A `switch` on the raw request path (`requestParts[1]`):

- `/` → serves `index.html` as `text/html`
- `/ping` → returns `pong` as `text/plain`
- anything else → `404 Not Found`

**Why:** Manual dispatch makes routing explicit — no regex, no trie, no framework. You control exactly which paths exist and what they return.

**Note:** Query parameters are parsed but the route switch still matches the full raw path — a request to `/ping?foo=bar` won't match the `/ping` case. This is a known gap for later fixing.

### Response Builder (`MyHTTPMessage`)

Assembles a full HTTP/1.1 response: status line, `Date`, `Server`, `Content-Length`, `Content-Type`, `Connection` headers, blank line, body. Accepts variadic `contentType` to switch between raw text and HTML file serving.

**Why:** HTTP responses must follow the wire format precisely: status line, headers (each `Key: Value\r\n`), blank line (`\r\n`), body. One function enforces that format in every code path, eliminating duplicate header-writing bugs.

### File Serving

`/` reads `index.html` from disk via `os.ReadFile` and sends it with `Content-Type: text/html`.

**Why:** Static file serving is the most common HTTP use case. Reading the file once per request is naive (no caching), but it's the simplest implementation that works — and the ceiling is marked for replacement with `os.ReadFile` + sync or in-memory cache.

### Request Logging to Stdout

Prints the request line and each header key-value pair as they're parsed.

**Why:** When debugging a raw TCP server, you can't use browser devtools — the wire is opaque. Printing each line lets you confirm the parser is consuming exactly what the client sent.

## Imports

- `bufio`
- `fmt`
- `io`
- `net`
- `os`
- `strconv`
- `strings`
- `time`

## The Only Net Docs You Need

From the entire `net` package documentation, these are the only functions/interfaces you need:

- `Listen` – Creates the listener. Reserves a port and starts tracking incoming connections.
- `Accept` – Blocks until a client connects. Returns a dedicated `net.Conn` for that client.
- `conn.Read` – Pulls raw bytes from the network buffer into your application.
- `conn.Write` – Pushes raw bytes from your application down the network pipe.
- `conn.SetDeadline` – Enforces timeouts to prevent hanging connections.

## Why These Imports?

### `bufio`

Reading raw TCP streams byte-by-byte is slow and resource-heavy. `bufio` reduces system calls by reading data in 4KB chunks into memory, then serving you line-by-line from that buffer. Without it, you'd manually handle packet fragmentation, leftover bytes, and syscall spam.

### `io`

HTTP bodies can be arbitrarily large. Reading them in one shot (`io.ReadAll`) works for small payloads but blows memory for large ones. `io.LimitReader` wraps the buffered reader and stops after `Content-Length` bytes, preventing over-read. `io.ReadAll` then drains that bounded stream into a byte slice — safe because the limit is already enforced.

### `net`

Your direct bridge to the operating system's network stack. Manages low-level socket creation, IP binding, and port management required to speak TCP.

### `strconv`

HTTP is a text-based protocol, but computers need numbers. `strconv` converts string representations of numbers (like `"22"`) into integers (`22`) so you can validate and manipulate them. It also converts integers back to strings for generating `Content-Length` headers dynamically.

### `fmt`

Formats and prints request lines, headers, and body info to stdout during parsing.

**Why:** Raw TCP has no request inspector — `fmt.Printf` is your debugger. Without it, parsing bugs are invisible.

### `os`

Reads files from disk (`os.ReadFile`) to serve static content.

**Why:** A TCP/HTTP server that only returns hardcoded strings is useless. `os.ReadFile` bridges the filesystem into the HTTP response pipeline with zero dependencies.

### `strings`

Trims, splits, and normalizes raw HTTP text — removing `\r\n` delimiters, parsing header lines into key-value pairs, and case-folding the `Connection` header for keep-alive logic.

**Why:** HTTP is text. Every line needs trimming, splitting, or comparison. The `strings` package handles all of it without regex overhead.

### `time`

Unreliable networks cause connections to hang forever. `time` lets you set absolute deadlines (`time.Now().Add(5 * time.Second)`) so dead clients cannot drain your server resources. This is critical to prevent Slowloris attacks.

## What This Engine Does NOT Include (By Design)

- No routing (you handle path matching manually).
- No middleware (logging, authentication, compression).
- No HTTP/2 (pure HTTP/1.1).
- No automatic body parsing (JSON, forms, file uploads are your responsibility).
- No connection pooling (each request gets its own goroutine).

This is a minimal, educational TCP/HTTP engine. It touches the wire directly so you understand the raw protocol before using high-level frameworks.

## Optional Features (Not Implemented)

These build on the existing engine in increasing complexity. Ordered from least to most effort.

### Content-Type Negotiation

Map file extensions (`.html`, `.css`, `.js`, `.png`) to MIME types instead of hardcoding `text/html` or `text/plain`.

**Why:** Browsers rely on `Content-Type` to render resources correctly. Serving CSS as `text/plain` disables styling.

### Chunked Transfer Encoding

Use `Transfer-Encoding: chunked` to stream bodies without knowing `Content-Length` upfront.

**Why:** Dynamic responses (proxied data, real-time events, large DB results) can't compute their length before sending. Chunked encoding lets you flush data as it arrives.

### POST Form Parsing (`application/x-www-form-urlencoded`)

Parse URL-encoded bodies (e.g., `name=alice&age=30`) into a key-value map.

**Why:** HTML forms submit as URL-encoded POST by default. Without this, the engine can't receive user input from a browser.

### Gzip Response Compression

Check for `Accept-Encoding: gzip`, compress the body with `compress/gzip`, and set `Content-Encoding: gzip`.

**Why:** Text compresses 5-10x. Uncompressed HTML/CSS/JS wastes bandwidth and increases page load time on slow networks.

### ETag / Conditional Requests

Hash the response body, send it as `ETag: "<hash>"`, and return `304 Not Modified` if the client sends `If-None-Match: "<hash>"`.

**Why:** Re-sending unchanged resources wastes bandwidth and battery. Conditional requests let clients cache aggressively and only refetch when content changes.

### Minimal Router (Prefix Trie)

Replace the `switch` statement with a radix tree that matches paths by prefix and supports path parameters (`/users/:id`).

**Why:** A flat switch doesn't scale past 5-10 routes. A trie gives O(k) matching (k = path length) and supports dynamic segments that the switch can't express.

### Graceful Shutdown

Catch `SIGINT`/`SIGTERM`, stop accepting new connections, drain active handlers, then exit.

**Why:** Killing the process mid-request drops responses on the floor. A production server must finish in-flight work before shutting down — especially during deploys.

### HTTP/2 Cleartext (h2c)

Implement HTTP/2 framing (frames, streams, HPACK header compression) over TCP without TLS.

**Why:** HTTP/2 eliminates head-of-line blocking and enables multiplexed requests over one connection. h2c sidesteps TLS complexity so you can learn the binary framing protocol in isolation.

**Note:** h2c is a significant project — ~2000+ lines of framing, HPACK, flow control. Only attempt if the goal is understanding the HTTP/2 wire format itself, not just getting a faster server.
