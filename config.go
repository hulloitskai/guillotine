package guillotine

import "github.com/sirupsen/logrus"

// WithLogger configures a Guillotine to write logs with log.
func WithLogger(log logrus.FieldLogger) Option {
	return func(cfg *Config) { cfg.Logger = log }
}

type (
	// A Config configures a Guillotine.
	Config struct {
		Logger logrus.FieldLogger
	}

	// An Option modifies a Config.
	Option func(*Config)
)
