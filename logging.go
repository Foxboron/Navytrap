package navytrap

import (
	log "github.com/Sirupsen/logrus"
)

var logger = log.New()

func init() {
	logger.Level = log.DebugLevel
	logger.Info("test")
}
