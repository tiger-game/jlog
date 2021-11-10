package jlog_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/tiger-game/tiger/jlog"
)

type Person struct {
	A string
	B int64
	C int
}

func TestJlog(t *testing.T) {
	jlog.GLogInit(jlog.LogDir("./log"), jlog.LogLevel(jlog.ERROR))
	defer jlog.CloseGLog()
	a := Person{"Hello", 64, 89}
	jlog.Infof("asda%v", 123)
	jlog.Info("asdads", a, "asdasd", "adsd", "")
	jlog.Info(errors.New("Test Print Error"))
}

var args = []interface{}{
	13, 28, 334,
}
var format = "asdasd%vasdasda%v,dasdaad%v"

var argvs = []interface{}{
	"asdasd", 13, "asdasda", 28, ",dasdaad", 334,
}

func Benchmark_DebugBufferAppend(b *testing.B) {
	var buffer = &jlog.Buffer{}
	for i := 0; i < b.N; i++ {
		buffer.Reset()
		for _, arg := range argvs {
			jlog.DebugBufferAppend(buffer, arg)
		}

	}
	b.ReportAllocs()
	// fmt.Println(buffer.String())
}

func Benchmark_Fprintf(b *testing.B) {
	var buf = &jlog.Buffer{}
	for i := 0; i < b.N; i++ {
		buf.Reset()
		fmt.Fprintf(buf, format, args...)

	}
	b.ReportAllocs()
	// fmt.Println(buf.String())
}

func Benchmark_Fprint(b *testing.B) {
	var buf = &jlog.Buffer{}
	for i := 0; i < b.N; i++ {
		buf.Reset()
		fmt.Fprint(buf, argvs...)
	}
	b.ReportAllocs()
	// fmt.Println(buf.String())
}
