package main

import (
	"fmt"
	"os"

	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

func main() {
	appLogger, err := logger.NewZapLogger()
	if err != nil {
		fmt.Println("Error creating logger: ", err)
		os.Exit(1)
	}
	defer appLogger.Sync()

	err = config.InitConfig(config.GophermartFlagsSet)
	if err != nil {
		fmt.Println("Error initializing config: ", err)
		os.Exit(1)
	}
}
