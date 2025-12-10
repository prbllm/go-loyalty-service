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
	DefaultJWTSecret            = "test-secret-key"
)

const (
	PathUserRegister  = "/api/user/register"
	PathUserLogin     = "/api/user/login"
	PathUserOrders    = "/api/user/orders"
	PathUserBalance   = "/api/user/balance"
	PathUserWithdraw  = "/api/user/balance/withdraw"
	PathWithdrawals   = "/api/user/withdrawals"
	AccrualOrdersPath = "/api/orders"
)

const (
	HeaderAuthorization = "Authorization"
	BearerPrefix        = "Bearer "
	HeaderContentType   = "Content-Type"
	ContentTypeJSON     = "application/json"
	HeaderRetryAfter    = "Retry-After"
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
	JWTSecretEnv            = "JWT_SECRET"
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
	ShutdownTimeout  = 5 * time.Second
	CompressionLevel = 5
)
