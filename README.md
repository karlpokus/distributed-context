# distributed-context
Playing with propagated http deadlines and cancelation signals in go.

We have 3 services; A, B and C. A calls B and B calls C and C returns some response to A via B. A will initialize the context and propagate it throughout the http calls.

My goal is to make sure that a canceled context - in any service - release all resources (http connections, pending work) in the call path.

# usage
```bash
$ go run ctx.go
```

# license
MIT
