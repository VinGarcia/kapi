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

For more technical information on how to use it, please read [the Docs][docs]

[docs]: https://pkg.go.dev/github.com/vingarcia/kapi

## Performance

This library uses reflection which brings performance concerns.

The use of reflection was made with caution using it only when necessary
and avoiding it on the critical sections of the code.

This granted a performance that isn't terrible:

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
