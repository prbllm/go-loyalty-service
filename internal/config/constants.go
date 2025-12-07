package config

import "time"

const (
	GophermartFlagsSet = "gophermart"
	AccrualFlagsSet    = "accrual"
)

const (
	DefaultRunAddress           = ":8080"
	DefaultDatabaseURI          = ""
	DefaultAccrualSystemAddress = ""
)

const (
	RunAddressFlag           = "a"
	DatabaseURIFlag          = "d"
	AccrualSystemAddressFlag = "r"
)

const (
	RunAddressEnv           = "RUN_ADDRESS"
	DatabaseURIEnv          = "DATABASE_URI"
	AccrualSystemAddressEnv = "ACCRUAL_SYSTEM_ADDRESS"
	LogLevelEnv             = "LOG_LEVEL"
)

const (
	RunAddressDescription           = "server address"
	DatabaseURIDescription          = "database URI"
	AccrualSystemAddressDescription = "accrual system address"
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

const (
	ShutdownTimeout = 5 * time.Second
)
