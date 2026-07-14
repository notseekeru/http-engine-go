#!/usr/bin/env bash
# Brutal endpoint test suite for http-engine-go
# Usage: ./test_endpoints.sh [host:port]
PASS=0
FAIL=0

ok()   { echo "  ✓ $1"; ((PASS++)); }
fail() { echo "  ✗ $1"; ((FAIL++)); }
req()  { timeout 2 nc "$@" 2>&1 || true; }
check() { local label="$1" pattern="$2" input="$3"
  if echo "$input" | grep -q "$pattern"; then ok "$label"; else fail "$label"; fi
}

HOST="${1:-localhost:8080}"

# --- start server ---
cleanup() { kill "$SERVER_PID" 2>/dev/null || true; }
trap cleanup EXIT

/tmp/http-engine-test &
SERVER_PID=$!
sleep 0.3

echo "=== GET / (index.html) ==="
R=$(printf "GET / HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
echo "$R"
check "/ returns 200"            "200 OK"           "$R"
check "/ HTML body"              "<!DOCTYPE html>"  "$R"
check "/ Content-Type text/html" "Content-Type: text/html" "$R"

echo ""
echo "=== GET /ping ==="
R=$(printf "GET /ping HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
echo "$R"
check "/ping returns 200" "200 OK" "$R"
check "/ping body pong"   "pong"   "$R"

echo ""
echo "=== GET /nonexistent (404) ==="
R=$(printf "GET /nonexistent HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
echo "$R"
check "/nonexistent 404" "404 Not Found" "$R"

echo ""
echo "=== GET with relativeURI params ==="
R=$(printf "GET /ping?foo=bar&baz=qux HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
echo "$R"
check "relativeURI params 200" "200 OK" "$R"

echo ""
echo "=== GET malformed relativeURI (trailing &) ==="
R=$(printf "GET /ping?foo=bar& HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
echo "$R"
if ! echo "$R" | grep -qi "panic"; then ok "trailing & no crash"; else fail "trailing & no crash"; fi

echo ""
echo "=== GET malformed relativeURI (no value) ==="
R=$(printf "GET /ping?foo HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
echo "$R"
if ! echo "$R" | grep -qi "panic"; then ok "no-value relativeURI no crash"; else fail "no-value relativeURI no crash"; fi

echo ""
echo "=== POST /ping ==="
R=$(printf "POST /ping HTTP/1.1\r\nHost: $HOST\r\nContent-Length: 11\r\nContent-Type: text/plain\r\nConnection: close\r\n\r\nhello world" | req localhost 8080)
echo "$R"
check "POST /ping 200" "200 OK" "$R"

echo ""
echo "=== PUT /ping (405) ==="
R=$(printf "PUT /ping HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
echo "$R"
check "PUT 405" "405" "$R"

echo ""
echo "=== HTTP/1.0 (400) ==="
R=$(printf "GET /ping HTTP/1.0\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
echo "$R"
check "HTTP/1.0 400" "400" "$R"

echo ""
echo "=== invalid request line (400) ==="
R=$(printf "GET /ping HTTP/1.1 extra\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
echo "$R"
check "extra param 400" "400\|Too many" "$R"

echo ""
echo "=== Keep-Alive: two requests ==="
R=$(printf "GET /ping HTTP/1.1\r\nHost: $HOST\r\n\r\nGET /ping HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | timeout 3 nc localhost 8080 2>&1 || true)
echo "$R"
COUNT=$(echo "$R" | grep -c "200 OK" || true)
[ "$COUNT" -eq 2 ] && ok "keep-alive 2x200" || fail "keep-alive 2x200 (got $COUNT)"

echo ""
echo "=== Content-Length:0 ==="
R=$(printf "POST /ping HTTP/1.1\r\nHost: $HOST\r\nContent-Length: 0\r\nConnection: close\r\n\r\n" | req localhost 8080)
echo "$R"
check "Content-Length 0 200" "200 OK" "$R"

echo ""
echo "=========================================="
echo "  PASS: $PASS  FAIL: $FAIL"
echo "=========================================="

[ "$FAIL" -eq 0 ]
