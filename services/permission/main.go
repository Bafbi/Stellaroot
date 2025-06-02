package main

import (
	"log/slog"
	"os"

	"github.com/casbin/casbin/v2"
	casbinredisadapter "github.com/casbin/redis-adapter/v2"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	logger.Info("Hello permissions service")

	adapter := casbinredisadapter.NewAdpaterWithOption(
		casbinredisadapter.WithNetwork("tcp"),
		casbinredisadapter.WithAddress("localhost:6379"),
		casbinredisadapter.WithPassword("your_password"),
		casbinredisadapter.WithKey("your_key"),
	)
	enforcer, err := casbin.NewEnforcer("model.conf", adapter)
	if err != nil {
		logger.Error("Failed to create Casbin enforcer", "error", err)
		os.Exit(1)
	}
	err = enforcer.LoadPolicy()
	if err != nil {
		logger.Error("Failed to load policies from adapter", "error", err)
		os.Exit(1)
	}
	logger.Info("Casbin policies loaded.")
}
