# HTTPParser Adapter

This library was created to make it easier and less verbose the
parsing of request arguments when using the fasthttp framework.

A simple usage example is as follows:

```Go
  router.Post("/adapted/<id>", adapter.Adapt(func(ctx *routing.Context, args struct {
  	ID     uint64 `path:"id"`
  	Brand  string `header:"brand,optional"`
  	Qparam string `query:"qparam,required"`
  	myType MyType `uservalue:"my_type"`
  	Body   Foo    `content-type:"application/json"`
  }) error {
  	// Do stuff

  	return nil
  }
```

For a working example see the file `cmd/main.go`, to run this example (it is a simple server)
use `make run` and to test the api you can run the following command:

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

> Note: The `cmd/main.go` example contains two routes doing the same thing,
> one using the library and the other not using it, you can test the
> not adapted one replacing `adapted` by `not-adapted` on the example below.

For more technical information on how to use please read [the Docs][docs]

[docs]: https://pkg.go.dev/github.com/vingarcia/go-adapter

## Performance

This library uses reflection which brings performance concerns.

The use of reflection was made with care using only when necessary
and avoiding it on the critical sections of the code.

This granted a performance that isn't terrible:

```
go test -bench=. -benchtime=15s
goos: linux
goarch: amd64
pkg: github.com/vingarcia/go-adapter
BenchmarkAdapter/adapted_handler-4         	 5416262	      3402 ns/op
BenchmarkAdapter/not_adapted_handler-4     	11372478	      1666 ns/op
PASS
ok  	github.com/vingarcia/go-adapter	42.354s
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
