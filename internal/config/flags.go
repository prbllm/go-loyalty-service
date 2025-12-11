package config

import (
	"flag"
)

func ParseFlags(flagsetName string, args []string, flagErrorHandling flag.ErrorHandling) *Config {
	config := defaultConfig()
	fs := flag.NewFlagSet(flagsetName, flagErrorHandling)

	fs.StringVar(&config.RunAddress, RunAddressFlag, config.RunAddress, RunAddressDescription)
	fs.StringVar(&config.DatabaseURI, DatabaseURIFlag, config.DatabaseURI, DatabaseURIDescription)

	if flagsetName == GophermartFlagsSet {
		fs.StringVar(&config.AccrualSystemAddress, AccrualSystemAddressFlag, config.AccrualSystemAddress, AccrualSystemAddressDescription)
	}
	fs.Parse(args)
	return config
}
