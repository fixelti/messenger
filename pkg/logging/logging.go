package logging

import "github.com/sirupsen/logrus"

//TODO: разобраться во всех блоках кода

var e *logrus.Entry

type Logger struct {
	*logrus.Entry
}

func GetLogger() *Logger {
	return &Logger{e}
}

func init() {
	l := logrus.New()
	e = logrus.NewEntry(l)
}
