package main

import (
	log "github.com/Sirupsen/logrus"
	// "github.com/dainis/your_mother/bot"
	"github.com/dainis/your_mother/cmd"
	"github.com/evalphobia/logrus_sentry"
	"os"
)

func init() {
	sentryDSN := os.Getenv("SENTRY_DSN")

	if sentryDSN != "" {
		hook, err := logrus_sentry.NewSentryHook(sentryDSN, []log.Level{
			log.PanicLevel,
			log.WarnLevel,
			log.ErrorLevel,
		})

		if err == nil {
			log.AddHook(hook)
		}
	}
	log.SetLevel(log.DebugLevel)
}

func main() {
	cmd.RootCmd.Execute()
}
