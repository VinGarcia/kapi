# HTTPParser Adapter

This library was created to make the parsing of
request arguments easier when using the fasthttp framework.

The behavior is simillar to how `json.Unmarshal` works, you give it a struct
with tags that will be used to inform the parser from where to extract
each attribute.

So when each request is received it is parsed and validated into more
abstract and useful data types that are ready to be used.

If there are problems parsing any of these values it will return
a routing.HTTPError with a BadRequest status code and a descriptive message.

A simple usage example is as follows:

```Go
  router.Post("/adapted/<id>", adapter.Adapt(func(ctx *routing.Context, args struct {
  	ID     uint64 `path:"id"`
  	Brand  string `header:"brand,optional"`
  	Qparam string `query:"qparam,required"`
  	MyType MyType `uservalue:"my_type"`
  	Body   Foo    `content-type:"application/json"`
  }) error {
  	fmt.Println("request received for brand: '%s'", args.Brand)

	// Do stuff

  	return nil
  }
```

For a working example see the file `adapters/fasthttp-routingV2/example/main.go`
(the `adapters/fiberV2/example/main.go` example is equivalent), to run this
example (it is a simple server) use `make run` and to test the api you can run
the following command:

```bash
$ curl -XPOST localhost:8765/adapted/42?qparam=barbar \
	-H 'Content-Type: application/json' \
	-H 'brand: Dito' \
	-d '{"id":32, "name":"John"}'
```

or simply:

```bash
make request
```

> Note: The example above contains two routes doing the same thing,
> one using the library and the other not using it, you can test the
> not adapted one replacing `adapted` by `not-adapted` on the example below.

For more technical information on how to use it, please read [the Docs][docs]

[docs]: https://pkg.go.dev/github.com/vingarcia/kapi

## Architecture

kapi follows a ports-and-adapters (hexagonal) design so a single request-parsing
engine can drive any HTTP framework. The roles and their allowed import directions:

- **Core engine — the root `kapi` package.** The framework-agnostic reflection logic
  that maps an HTTP request onto a handler's argument struct: `DecodeHandlerFunction`
  (inspects the struct's tags once at startup), `UnmarshalRequestAsStruct` (fills the
  struct from the request on each call), and the `BuildJSONResponse` helper. The parsing
  engine reaches a request only through the `RequestAdapter` port, so it does not depend
  on any concrete HTTP framework. (The one exception is `BuildJSONResponse`, which still
  imports `fasthttp-routing` directly and is not yet backend-agnostic.)
- **Port — `RequestAdapter` (in `contracts.go`).** The interface every backend must
  implement: `GetBody`, `GetPathParam`, `GetHeaderParam`, `GetQueryParam`,
  `GetContextValue`, `SetContextValue`, and `NewHTTPError`. This interface is the seam
  that inverts the dependency — the core depends on it, not on a framework.
- **Adapters — `adapters/*`.** One package per supported backend
  (`fasthttp-routingV2`, `fiberV2`). Each exposes an `Adapter` type that implements
  `RequestAdapter` over that framework's request context, plus an `Adapt(fn)` function
  that wraps a tagged handler into a native handler for that framework. Each also ships
  a runnable `example/main.go` that wires a server. Adapters import the core and their
  own framework; the core never imports an adapter.
- **Test helpers — `internal/testtools`.** Assertion, JSON, time and panic helpers
  shared by the test suites.

Dependencies point inward: adapters depend on the core, and for request parsing the core
depends only on the `RequestAdapter` port, not on any backend. Adding support for a new
framework means writing one adapter, with no change to the parsing engine.

```
        kapi (core engine)
        DecodeHandlerFunction / UnmarshalRequestAsStruct
                 │
                 │ uses
                 ▼
        RequestAdapter (port interface, contracts.go)
                 ▲
                 │ implements
        ┌────────┴─────────┐
adapters/fasthttp-      adapters/fiberV2
   routingV2

(each adapter also imports the core; the core imports no adapter)
```

## Performance

This library uses reflection which brings performance concerns.

The use of reflection was made with caution using it only when necessary
and avoiding it on the critical sections of the code.

This granted a performance that isn't terrible, as shown by this historical
benchmark result (the `BenchmarkAdapter` it came from is not currently part of
this repository's test suite):

```
go test -bench=. -benchtime=15s
goos: linux
goarch: amd64
pkg: github.com/vingarcia/kapi
BenchmarkAdapter/adapted_handler-4         	 5416262	      3402 ns/op
BenchmarkAdapter/not_adapted_handler-4     	11372478	      1666 ns/op
PASS
ok  	github.com/vingarcia/kapi	42.354s
```

The functions tested above are very common examples parsing one integer
from the path, one value from the request header and unmarshalling the body
as JSON.

The `adapted` version uses this library and the `not_adapted` version
uses normal calls to the routing.Context received as argument.

The results above show that using the library is almost exactly two times slower
than the version without the library, for most use cases this is ok, since
either the performance gain is not necessary on this route or when the actual
task made by this route includes an external request or a database access.

However, for routes where the performance is critical we do not recomend the use
of this library.

The good news is that you can use this library only on the routes where performance
is not critical, getting the best of both worlds.
