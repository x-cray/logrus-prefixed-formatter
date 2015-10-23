package main

import (
	"github.com/Sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var log = logrus.New()

func init() {
	log.Formatter = new(prefixed.TextFormatter)
	log.Level = logrus.DebugLevel
}

func main() {
	defer func() {
		err := recover()
		if err != nil {
			log.WithFields(logrus.Fields{
				"prefix": "main",
				"omg":    true,
				"number": 100,
			}).Fatal("The ice breaks!")
		}
	}()

	log.WithFields(logrus.Fields{
		"prefix": "main",
		"animal": "walrus",
		"number": 8,
	}).Debug("Started observing beach")

	log.WithFields(logrus.Fields{
		"prefix": "main",
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")

	log.WithFields(logrus.Fields{
		"prefix": "main",
		"omg":    true,
		"number": 122,
	}).Warn("The group's number increased tremendously!")

	log.WithFields(logrus.Fields{
		"prefix":      "sensor",
		"temperature": -4,
	}).Debug("Temperature changes")

	log.WithFields(logrus.Fields{
		"prefix": "sensor",
		"animal": "orca",
		"size":   9009,
	}).Panic("It's over 9000!")
}
