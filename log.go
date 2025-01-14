package main

import (
	"net/http"

	logrus "github.com/sirupsen/logrus"
)

func initLog() {
	logFormat := "pretty"
	strEnvConfig(&logFormat, "IMGPROXY_LOG_FORMAT")

	switch logFormat {
	case "structured":
		logrus.SetFormatter(&logStructuredFormatter{})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(newLogPrettyFormatter())
	}

	logLevel := "info"
	strEnvConfig(&logLevel, "IMGPROXY_LOG_LEVEL")

	levelLogLevel, err := logrus.ParseLevel(logLevel)
	if err != nil {
		levelLogLevel = logrus.DebugLevel
	}

	logrus.SetLevel(levelLogLevel)

	if isSyslogEnabled() {
		slHook, err := newSyslogHook()
		if err != nil {
			logFatal("Unable to connect to local syslog daemon")
		}

		logrus.AddHook(slHook)
	}
}

func logRequest(reqID string, r *http.Request) {
	path := r.URL.RequestURI()

	logrus.WithFields(logrus.Fields{
		"request_id": reqID,
		"method":     r.Method,
	}).Infof("Started %s", path)
}

func logResponse(reqID string, r *http.Request, status int, err *imgproxyError, imageURL *string, po *processingOptions) {
	var level logrus.Level

	switch {
	case status >= 500:
		level = logrus.ErrorLevel
	case status >= 400:
		level = logrus.WarnLevel
	default:
		level = logrus.InfoLevel
	}

	fields := logrus.Fields{
		"request_id": reqID,
		"method":     r.Method,
		"status":     status,
	}

	if err != nil {
		fields["error"] = err

		if stack := err.FormatStack(); len(stack) > 0 {
			fields["stack"] = stack
		}
	}

	if imageURL != nil {
		fields["image_url"] = *imageURL
	}

	if po != nil {
		fields["processing_options"] = po
	}

	logrus.WithFields(fields).Logf(
		level,
		"Completed in %s %s", getTimerSince(r.Context()), r.URL.RequestURI(),
	)
}

func logNotice(f string, args ...interface{}) {
	logrus.Infof(f, args...)
}

func logWarning(f string, args ...interface{}) {
	logrus.Warnf(f, args...)
}

func logFatal(f string, args ...interface{}) {
	logrus.Fatalf(f, args...)
}

func logDebug(f string, args ...interface{}) {
	logrus.Debugf(f, args...)
}
