# logging [![Go Reference][go-ref-image]][go-ref]

Using [uber-go/zap][zap] in all my project I found myself repeating a small amount of bootstrap and http middleware. Then when all my projects weren't Stackdriver compabitble I decided to create a small package for this:

- Highly opinionated
- Stackdriver compatible
- Heavy use of the context.Context


## Usage:

```golang
ctx := context.Background()

logger := logging.New(ctx, "myservice", "example")
logger.Info("hello, world!!")

ctx = logging.WithLogger(ctx, logger)
err := someOperation(ctx)
logger.Error("failed", zap.Error(err))
```
(see [this example](./cmd/example/main.go) for a more extensive example of using _logging_ in a http service)

#### Links

- [uber-go/zap][zap]

[go-ref-image]: https://pkg.go.dev/badge/github.com/koenbollen/logging.svg
[go-ref]: https://pkg.go.dev/github.com/koenbollen/logging
[zap]: https://github.com/uber-go/zap
