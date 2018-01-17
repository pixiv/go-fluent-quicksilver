package logger

type Logger interface {
	Fatal(...interface{}) Logger
	Fatalf(string, ...interface{}) Logger
	Println(...interface{}) Logger
	Print(...interface{}) Logger
	Printf(string, ...interface{}) Logger
}

type Logger struct{}

func (l *Logger) Fatal(v ...interface{}) {

}

func (l *Logger) Println(v ...interface{}) {

}

func (l *Logger) Print(...interface{}) {

}
