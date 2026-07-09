# TCP/HTTP Engine

A minimal HTTP/1.1 server built from scratch using only the Go standard library. This project touches the raw TCP wire to understand how HTTP works under the hood.

This project doesn't include `net/http` as it abstracts away many networking logics needed to bridge networking fundamentals.

## Imports

- `bufio`
- `io`
- `net`
- `strconv`
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

### `time`

Unreliable networks cause connections to hang forever. `time` lets you set absolute deadlines (`time.Now().Add(5 * time.Second)`) so dead clients cannot drain your server resources. This is critical to prevent Slowloris attacks.

## What This Engine Does NOT Include (By Design)

- No routing (you handle path matching manually).
- No middleware (logging, authentication, compression).
- No HTTP/2 (pure HTTP/1.1).
- No automatic body parsing (JSON, forms, file uploads are your responsibility).
- No connection pooling (each request gets its own goroutine).

This is a minimal, educational TCP/HTTP engine. It touches the wire directly so you understand the raw protocol before using high-level frameworks.
