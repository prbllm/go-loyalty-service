package main

import (
	"os"

	"github.com/prbllm/go-loyalty-service/internal/accrual/config"
)

func main() {
	_, err := config.New(os.Args[1:])
	if err != nil {

	}
}
