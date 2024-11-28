package logger

import (
	"fmt"
	"log"
	"os"
)

var infoLogger = log.New(os.Stdout, "Info: ", log.Ldate|log.Ltime)
var warningLogger = log.New(os.Stderr, "Warning: ", log.Ldate|log.Ltime)
var errorLogger = log.New(os.Stderr, "Error: ", log.Ldate|log.Ltime|log.Lshortfile)

func Println(v ...any) {
	infoLogger.Output(2, fmt.Sprintln(v...))
}

func Printf(format string, v ...any) {
	infoLogger.Output(2, fmt.Sprintf(format, v...))
}

func Print(v ...any) {
	infoLogger.Output(2, fmt.Sprint(v...))
}

func Warningln(v ...any) {
	warningLogger.Output(2, fmt.Sprintln(v...))
}

func Warningf(format string, v ...any) {
	warningLogger.Output(2, fmt.Sprintf(format, v...))
}

func Warning(v ...any) {
	warningLogger.Output(2, fmt.Sprint(v...))
}

func Errorln(v ...any) {
	errorLogger.Output(2, fmt.Sprintln(v...))
}

func Errorf(format string, v ...any) {
	errorLogger.Output(2, fmt.Sprintf(format, v...))
}

func Error(v ...any) {
	errorLogger.Output(2, fmt.Sprint(v...))
}

func Fatalln(v ...any) {
	errorLogger.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	errorLogger.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func Fatal(v ...any) {
	errorLogger.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

func ShouldNeverHappen(v ...any) {
	errorLogger.Output(2, "SHOULDNEVERHAPPEND "+fmt.Sprint(v...))
}
