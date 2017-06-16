// Copyright (c) , zhoucb, Strong Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/zhouchangbo/ss-go

package sslog

import (
	"log"
	"os"
)

var (
	logfile, _ = os.OpenFile("ss.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	closeDebug = false
)

/*
func init() {
	LoggerInit()
}
*/
func LoggerInit() {
	flag := log.Flags()
	flag = flag | log.Lshortfile | log.LstdFlags
	log.SetFlags(flag)
	log.SetOutput(logfile)

	//关闭调试打印
	closeDebug = true
}

func LoggerDebug(format string, v ...interface{}) {
	if closeDebug {
		return
	}

	log.SetPrefix("[debug]: ")
	log.Printf(format, v...)
}

func LoggerErr(format string, v ...interface{}) {
	log.SetPrefix("[error]: ")
	log.Printf(format, v...)
}

func LoggerInfo(format string, v ...interface{}) {
	log.SetPrefix("[info]: ")
	log.Printf(format, v...)
}
