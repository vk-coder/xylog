package xylog_test

import (
	"testing"

	"github.com/xybor-x/xycond"
	"github.com/xybor-x/xylog"
	"github.com/xybor-x/xylog/test"
)

func TestGetLogger(t *testing.T) {
	var names = []string{"", "foo", "foo.bar"}
	for i := range names {
		var logger1 = xylog.GetLogger(names[i])
		var logger2 = xylog.GetLogger(names[i])
		xycond.ExpectEqual(logger1, logger2).Test(t)
	}
}

func TestLoggerLogMethods(t *testing.T) {
	var fixedMsg = test.GetRandomMessage()
	test.WithLogger(t, func(logger *xylog.Logger, w *test.MockWriter) {
		var tests = []struct {
			methodf func(string, ...any)
			method  func(...any)
		}{
			{logger.Debugf, logger.Debug},
			{logger.Infof, logger.Info},
			{logger.Warnf, logger.Warn},
			{logger.Warningf, logger.Warning},
			{logger.Errorf, logger.Error},
			{logger.Fatalf, logger.Fatal},
			{logger.Criticalf, logger.Critical},
		}

		logger.SetLevel(xylog.DEBUG)
		for i := range tests {
			w.Reset()
			var msg = test.GetRandomMessage()
			tests[i].method(msg)
			xycond.ExpectIn(msg, w.Captured).Test(t)

			w.Reset()
			var msgf = test.GetRandomMessage()
			tests[i].methodf(msgf)
			xycond.ExpectIn(msgf, w.Captured).Test(t)
		}
		w.Reset()
		logger.Log(xylog.DEBUG, fixedMsg)
		xycond.ExpectIn(fixedMsg, w.Captured).Test(t)
		w.Reset()
		logger.Logf(xylog.DEBUG, fixedMsg)
		xycond.ExpectIn(fixedMsg, w.Captured).Test(t)

		logger.SetLevel(xylog.NOTLOG)
		for i := range tests {
			w.Reset()
			var msg = test.GetRandomMessage()
			tests[i].method(msg)
			xycond.ExpectNotIn(msg, w.Captured).Test(t)

			w.Reset()
			var msgf = test.GetRandomMessage()
			tests[i].methodf(msgf)
			xycond.ExpectNotIn(msgf, w.Captured).Test(t)
		}
		w.Reset()
		logger.Log(xylog.DEBUG, fixedMsg)
		xycond.ExpectNotIn(fixedMsg, w.Captured).Test(t)
		w.Reset()
		logger.Logf(xylog.DEBUG, fixedMsg)
		xycond.ExpectNotIn(fixedMsg, w.Captured).Test(t)
	})
}

func TestLoggerCallHandlerHierarchy(t *testing.T) {
	test.WithLogger(t, func(logger *xylog.Logger, w *test.MockWriter) {
		var child = xylog.GetLogger(t.Name() + ".main")
		logger.SetLevel(xylog.INFO)

		var msg = test.GetRandomMessage()
		child.Log(xylog.WARN, msg)
		xycond.ExpectIn(msg, w.Captured).Test(t)

		msg = test.GetRandomMessage()
		child.Log(xylog.DEBUG, msg)
		xycond.ExpectNotIn(msg, w.Captured).Test(t)
	})
}

func TestLoggerStack(t *testing.T) {
	test.WithLogger(t, func(logger *xylog.Logger, w *test.MockWriter) {
		logger.SetLevel(xylog.DEBUG)
		logger.Stack(xylog.DEBUG)
		xycond.ExpectIn("xylog.(*Logger).Stack", w.Captured).Test(t)
	})
}

func TestLoggerFilterLog(t *testing.T) {
	test.WithLogger(t, func(logger *xylog.Logger, w *test.MockWriter) {
		for _, h := range logger.Handlers() {
			h.AddFilter(&test.LoggerNameFilter{Name: t.Name() + ".main"})
		}
		var main = xylog.GetLogger(t.Name() + ".main")
		var other = xylog.GetLogger(t.Name() + ".other")

		var msg = test.GetRandomMessage()
		main.Error(msg)
		xycond.ExpectIn(msg, w.Captured).Test(t)

		w.Reset()
		other.Error(msg)
		xycond.ExpectNotIn(msg, w.Captured).Test(t)
	})
}

func TestLoggerLogInvalidJSON(t *testing.T) {
	test.WithLogger(t, func(logger *xylog.Logger, w *test.MockWriter) {
		logger.Handlers()[0].SetFormatter(xylog.NewJSONFormatter())
		logger.Event("test").Field("func", func() {}).Error()
		xycond.ExpectIn("An error occurred while formatting the message",
			w.Captured).Test(t)
	})
}

func TestLoggerAddExtraMacro(t *testing.T) {
	test.WithLogger(t, func(logger *xylog.Logger, w *test.MockWriter) {
		var formatter = xylog.NewTextFormatter().AddMacro("foo", "custom")
		logger.AddExtraMacro("custom", "this-is-a-custom-field")
		logger.Handlers()[0].SetFormatter(formatter)
		logger.Event("test").Error()
		xycond.ExpectIn(`foo="this-is-a-custom-field" event="test"`,
			w.Captured).Test(t)
	})
}

func TestLoggerAddField(t *testing.T) {
	test.WithLogger(t, func(logger *xylog.Logger, w *test.MockWriter) {
		logger.AddField("custom", "this-is-a-custom-field")
		logger.Event("test").Error()
		xycond.ExpectIn(`event="test" custom="this-is-a-custom-field"`,
			w.Captured).Test(t)
	})
}

func TestLoggerHandlers(t *testing.T) {
	var handler = xylog.GetHandler("")
	var logger = xylog.GetLogger(t.Name())
	logger.AddHandler(handler)

	xycond.ExpectEqual(len(logger.Handlers()), 1).Test(t)
	xycond.ExpectEqual(logger.Handlers()[0], handler).Test(t)

	logger.RemoveHandler(handler)
	xycond.ExpectEqual(len(logger.Handlers()), 0).Test(t)
}

func TestLoggerFilters(t *testing.T) {
	var filter = &test.LoggerNameFilter{Name: "foo"}
	var logger = xylog.GetLogger(t.Name())
	logger.AddFilter(filter)

	xycond.ExpectEqual(len(logger.Filters()), 1).Test(t)
	xycond.ExpectEqual(logger.Filters()[0], filter).Test(t)

	logger.RemoveFilter(filter)
	xycond.ExpectEqual(len(logger.Filters()), 0).Test(t)
}
