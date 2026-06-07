# The Protocol (RESP)

Redis uses the **REdis Serialization Protocol (RESP)**. It is a human-readable, text-based protocol that is surprisingly simple.

One of the strongest technical features of RedisLite is that we implement this wire protocol directly over raw TCP. **We do not invent our own protocol.** This allows any existing, standard Redis client (like `redis-cli`, or Node.js `redis` packages) to connect to RedisLite without modification.

## How RESP Works

In RESP, the first byte of a message determines its data type:

- `+` Simple Strings
- `-` Errors
- `:` Integers
- `$` Bulk Strings (binary safe strings)
- `*` Arrays

Every part of a RESP message is terminated by `\r\n` (CRLF).

## Client Requests

When a client sends a command to the server, it always formats it as an **Array of Bulk Strings**.

For example, if you type `SET mykey myvalue` into `redis-cli`, the CLI translates that into:

```text
*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$7\r\nmyvalue\r\n
```

### Breakdown:
- `*3\r\n` - An array of 3 elements follows
- `$3\r\nSET\r\n` - A bulk string of length 3: "SET"
- `$5\r\nmykey\r\n` - A bulk string of length 5: "mykey"
- `$7\r\nmyvalue\r\n` - A bulk string of length 7: "myvalue"

Our custom parser reads this byte-by-byte, extracts the strings, and hands `["SET", "mykey", "myvalue"]` to the command handler.

## Server Responses

The server responds based on the outcome of the command:

**Success (Simple String):**
```text
+OK\r\n
```

**Errors:**
```text
-ERR unknown command 'FOO'\r\n
```

**Integers (e.g., DEL returning count):**
```text
:1\r\n
```

**Missing Data (Null Bulk String):**
```text
$-1\r\n
```

**Actual Data (Bulk String):**
```text
$7\r\nmyvalue\r\n
```

By strictly adhering to these rules, RedisLite perfectly mimics a real Redis instance on the network layer.
