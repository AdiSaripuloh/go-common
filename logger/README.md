# Package `logger`

Library code that's ok to use by external applications to log their application to console.

# Usage
```go
package main

import (
	"context"
	
	"github.com/AdiSaripuloh/go-common/logger"
)

func main() {
	logger.Init(false)
	defer logger.Sync()
	
	logger.Info(context.Background(), "logger without stacktrace")
	logger.Info(context.Background(), "logger with data", logger.Field{Key: "Key", Value: "Value"})
}
```