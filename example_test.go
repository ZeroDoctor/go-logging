package logging

import (
	"os"
	"testing"
)

func TestExample(t *testing.T) {
	// This call is for testing purposes and will set the time to unix epoch.
	InitForTesting(DEBUG)

	var log = MustGetLogger("example")

	// For demo purposes, create two backend for os.Stdout.
	//
	// os.Stderr should most likely be used in the real world but then the
	// "Output:" check in this example would not work.
	backend1 := NewLogBackend(os.Stdout, "", 0)
	backend2 := NewLogBackend(os.Stdout, "", 0)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	var format = MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	backend2Formatter := NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend2
	backend2Leveled := AddModuleLevel(backend2Formatter)
	backend2Leveled.SetLevel(ERROR, "")

	// Set the backends to be used and the default level.
	SetBackend(backend1, backend2Leveled)

	log.Debugf("debug %s", Password("secret"))
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("err")
	log.Critical("crit")

	// Output:
	// debug arg
	// error
	// 00:00:00.000 Example E error
}
