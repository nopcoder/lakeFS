package config

import "errors"

var ErrConfigNotLoaded = errors.New("configuration not found in context")
var ErrCLIContextNotLoaded = errors.New("CLI context not found in context")
