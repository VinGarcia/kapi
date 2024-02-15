package fiber

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/vingarcia/kapi"
	"golang.org/x/sync/errgroup"
)

func TestAdapter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := fiber.New()

	var handler func(c *fiber.Ctx) error
	app.Post("/:pathParam", func(c *fiber.Ctx) error {
		return handler(c)
	})

	g := errgroup.Group{}

	var port string
	err := getFreePorts(&port)
	if err != nil {
		t.Fatalf("unable to get free port for starting server: %s", err)
	}

	g.Go(func() error {
		return app.Listen(":" + port)
	})

	g.Go(func() error {
		<-ctx.Done()
		return app.Shutdown()
	})

	t.Run("fiber", func(t *testing.T) {
		kapi.AdaptTestSuite[*fiber.Ctx](t, func(fn any, test kapi.TestRequest, ctxValue any) error {
			url := fmt.Sprintf("http://localhost:%s/%s?queryParam=%s", port, test.PathParam, test.QueryParam)
			req, err := http.NewRequest("POST", url, strings.NewReader(test.Body))
			if err != nil {
				t.Fatalf("error creating request: %s", err)
			}

			req.Header.Set("headerParam", test.HeaderParam)

			handler = func(c *fiber.Ctx) error {
				c.Locals("userValue", ctxValue)

				return Handler(fn)(c)
			}

			_, err = http.DefaultClient.Do(req)
			return err
		})
	})

	cancel()
	g.Wait()
}

func getFreePorts(vars ...*string) error {
	for i := 0; i < len(vars); i++ {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			return err
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return err
		}
		defer l.Close()
		*vars[i] = strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	}

	return nil
}
