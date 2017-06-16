// Copyright (c) , zhoucb, Strong Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/zhouchangbo/ss-go

package socket

import (
	"config"
	sjson "github.com/bitly/go-simplejson"
	"module"
	"net"
	"runtime"
	"sslog"
	"strings"
)

func NewTcpSession(localPort int, conn net.Conn) *config.TcpSession {
	session := new(config.TcpSession)

	session.Conn = conn
	session.RecvCh = make(chan interface{}, 128)
	session.SendCh = make(chan interface{}, 128)

	session.IsVerify = false
	session.LocalPort = localPort

	go doRecv(session)
	go doSend(session)
	return session
}

//从tcp连接中读取数据，并处理后，放入到recvCh通道中
func doRecv(session *config.TcpSession) {
	recvbuf := make([]byte, 10240)
	//处理异常panic中断，保持协程继续进行
	defer func() {
		if err := recover(); err != nil {
			_, file, line, _ := runtime.Caller(0)
			sslog.LoggerDebug("[%s,%d]recover--doRecv failed with %s ", file, line+1, err)
		}
	}()

	for {
		//是否ping超时
		nowSec := config.SystemSec()
		if session.LastTime > 0 && nowSec-session.LastTime > int64(config.PingTimeout) {
			session.RecvCh <- "ping time out to close"
			break
		}

		//读取数据
		iLen, err := session.Conn.Read(recvbuf)
		if err != nil {
			session.RecvCh <- "recv failed to close"

			_, file, line, _ := runtime.Caller(0)
			sslog.LoggerErr("[%s,%d]doRecv--recv failed to close  err:%v ", file, line+1, err)
			break
		}

		_, file, line, _ := runtime.Caller(0)
		sslog.LoggerDebug("[%s,%d]doRecv 接收到的数据长度 iLen =%d", file, line+1, iLen)
		if iLen < 0 {
			_, file, line, _ := runtime.Caller(0)
			sslog.LoggerDebug("[%s:%d]doRecv--iLen < 0 continue", file, line+1)
			continue
		}

		readStr := string(recvbuf[0:iLen])
		if session.IsVerify == false {
			verify(readStr, session)
		} else {
			recvJson := WebSocketDecode(recvbuf[0:iLen], session)
			if recvJson == nil {
				sslog.LoggerErr("recvJson nil")
				continue
			} else {
				sslog.LoggerDebug("recvJson :%v", recvJson)
				session.RecvCh <- recvJson
			}
		}
	}

	sslog.LoggerErr("doRecv break")
}

func verify(readbuf string, session *config.TcpSession) {
	if session.LocalPort == 80 {
		if bKey := strings.Contains(readbuf, "LkCwIf8QOUGUOUajbgIIqA=="); bKey { //带验证请求
			//fmt.Println("readbuf :", readbuf)
			sslog.LoggerDebug("LkCwIf8QOUGUOUajbgIIqA== contains :%v ", bKey)
			err := ResponseWebsocketHandShake(session)
			if err == nil {
				session.IsVerify = true
			}
		}
	} else if session.LocalPort == 8080 {
		if bKey := strings.Contains(readbuf, "<policy-file-request/>"); bKey { //带验证请求
			sslog.LoggerDebug("<policy-file-request/>== contains :%v ", bKey)
			err := ResponseWebsocketHandShake(session)
			if err == nil {
				session.IsVerify = true
			}
		} else { //flash socket有不带验证的请求不需要验证通过的情况
			session.IsVerify = true
		}
	}
}

//从sendCh发送通道取数据，并简单处理后进行发送
func doSend(session *config.TcpSession) {
	defer func() {
		if err := recover(); err != nil {
			_, file, line, _ := runtime.Caller(0)
			sslog.LoggerDebug("[%s:%d]recover--doSend failed with %s ", file, line+1, err)
		}
	}()

	for {
		sendObj, ok := <-session.SendCh
		if ok == false {
			break
		}

		switch valueObj := sendObj.(type) { //判断发送数据类型
		case string:
			_, file, line, _ := runtime.Caller(0)
			sslog.LoggerDebug("[%s:%d]doSend valueObj[%s]", file, line+1, valueObj)
			if strings.Contains(valueObj, "close") {
				break
			} else if strings.Contains(valueObj, "http") {
				config.HttpPost(valueObj)
			} else {
				_, file, line, _ := runtime.Caller(0)
				sslog.LoggerErr("[%s:%d]string info unknow [%s]", file, line+1, valueObj)
			}

		case *sjson.Json:
			cid, err := valueObj.Get("cid").Int()
			if err != nil {
				_, file, line, _ := runtime.Caller(0)
				sslog.LoggerDebug("[%s:%d]doSend---err:%s", file, line+1, err.Error())
			}

			sendBuf := valueObj.Get("sendbuf")

			//sendBuf := valueObj.Get("sendbuf")    //要发送的json结构数据
			snBytes, err := sendBuf.MarshalJSON() //转成[]byte
			if err != nil {
				_, file, line, _ := runtime.Caller(0)
				sslog.LoggerDebug("[%s:%d]doSend---err:%s", file, line+1, err.Error())
			}

			encodeSnBytes := WebSocketEncode(snBytes, session)
			if cid == -1 {
				_, err := session.Conn.Write(encodeSnBytes) //单个应答
				if err != nil {
					//_, file, line, _ := runtime.Caller(0)
					//sslog.LoggerErr("[%s:%d]Write: %s [%s]", file, line+1, err.Error(), session.Conn.RemoteAddr().String())
					break
				} else {
					_, file, line, _ := runtime.Caller(0)
					sslog.LoggerDebug("[%s:%d]doSend---应答发送成功的数据长度:%d", file, line+1, len(snBytes))
				}

			} else {
				doBroadCast(cid, encodeSnBytes) //会议内通知
			}
		}

	}

	session.RecvCh <- "doSend to close"
	sslog.LoggerErr("doSend break")

}

//向与会的所有用户连接发送数据
func doBroadCast(cid int, snbuf []byte) {
	config.SessionListRWMutex.RLock()
	_, ok := config.SessionList[cid]
	if ok {
		cidMap, _ := config.SessionList[cid]
		for _, userInfo := range cidMap {
			sslog.LoggerDebug("username: %s, iosession[%v] \n", userInfo.Name, userInfo.IoSession)
			if userInfo.IoSession == nil {
				continue
			}
			_, err := userInfo.IoSession.Conn.Write(snbuf)
			if err != nil {
				//_, file, line, _ := runtime.Caller(0)
				userInfo.IoSession.Close()
				//sslog.LoggerErr("[%s:%d]Write: %s  [%s]", file, line+1, err.Error(), userInfo.IoSession.Conn.RemoteAddr().String())
			}
		}
	}
	config.SessionListRWMutex.RUnlock()
}

//从recvCh接收通道取数据，进行相应业务处理后，把待发送的数据放入到sendCh待发送通道中
func ProcessTcpSession(session *config.TcpSession) {
	for {
		recvObj, ok := <-session.RecvCh
		if ok == false {
			break
		}
		switch recvValue := recvObj.(type) {
		case string:
			if strings.Contains(recvValue, "close") {
				break
			} else {
				session.SendCh <- recvValue
			}
		case *sjson.Json:
			//模块处理
			processMessageHandle(session, recvValue)

		}
	}

	session.Close()
	sslog.LoggerErr("ProcessTcpSession break")

}

func processMessageHandle(session *config.TcpSession, recvJson *sjson.Json) error {
	/*
		defer func() {
			if err := recover(); err != nil {
				_, file, line, _ := runtime.Caller(0)
				sslog.LoggerDebug("[%s:%d]recover--processMessageHandle failed with %s ", file, line+1, err)
			}
		}()
	*/
	tsessionType, err := recvJson.Get("m").Int()
	if err != nil || tsessionType == 0 {
		_, file, line, _ := runtime.Caller(0)
		sslog.LoggerDebug("[%s:%d]not  Correct json or not json data: err[%s] || tsessionType[%v] recvJson= %v", file, line+1, err.Error(), tsessionType, recvJson)
		return err
	}

	sessionType := tsessionType / 1000
	moduleType := tsessionType % 1000
	jsStr, _ := recvJson.MarshalJSON()

	_, file, line, _ := runtime.Caller(0)
	sslog.LoggerDebug("[%s:%d]接收到的信息: %s", file, line+1, jsStr)
	switch sessionType {
	case config.LOGIN:
		module.LoginCommunicationHandler(moduleType, recvJson, session)
	case config.LIST:
		module.ListCommunicationHandler(moduleType, recvJson, session)
	case config.CHAT:
		module.ChatCommunicationHandler(moduleType, recvJson, session)
	case config.POWER:
		module.PowerCommunicationHandler(moduleType, recvJson, session)
	case config.UPLOAD:
		module.UploadCommunicationHandler(moduleType, recvJson, session)
	case config.VIDEO:
		module.VideoCommunicationHandler(moduleType, recvJson, session)
	case config.WHITEBOARD:
		module.WhiteBoardCommunicationHandler(moduleType, recvJson, session)
	case config.BROWER:
		module.BrowerCommunicationHandler(moduleType, recvJson, session)
	}

	return nil
}

func WebSocketEncode(inBytes []byte, session *config.TcpSession) []byte {
	var outBytes []byte
	outLen := 0

	//flash socket encode
	if session.LocalPort == 8080 {
		inLen := len(inBytes)
		outBytes = append(outBytes, byte(inLen>>24))
		outBytes = append(outBytes, byte(inLen>>16))
		outBytes = append(outBytes, byte(inLen>>8))
		outBytes = append(outBytes, byte(inLen))
		outBytes = append(outBytes, inBytes...)
		return outBytes
	}

	//对websocket加上头部字段，一条完整的文本消息
	outBytes = append(outBytes, byte(129)) // 详细参考http://blog.csdn.net/otypedef/article/details/51492188

	//对websocket 加上长度字段
	inLen := len(inBytes)
	if inLen <= 125 { //占1字节长度
		outLen = inLen
		outBytes = append(outBytes, byte(outLen))
	} else if inLen <= 65535 { //占3字节长度
		outLen = outLen + 126
		outBytes = append(outBytes, byte(outLen))
		outBytes = append(outBytes, byte(inLen>>8))
		outBytes = append(outBytes, byte(inLen))
	} else { //>65535 占8个字节长度
		outLen = outLen + 127
		outBytes = append(outBytes, byte(outLen))

		outBytes = append(outBytes, byte(0))
		outBytes = append(outBytes, byte(0))
		outBytes = append(outBytes, byte(0))
		outBytes = append(outBytes, byte(0))

		outBytes = append(outBytes, byte(inLen>>24))
		outBytes = append(outBytes, byte(inLen>>16))
		outBytes = append(outBytes, byte(inLen>>8))
		outBytes = append(outBytes, byte(inLen))
	}

	outBytes = append(outBytes, inBytes...)
	return outBytes
}

func WebSocketDecode(inBytes []byte, session *config.TcpSession) *sjson.Json {
	if session.LocalPort == 8080 {
		outJson, err := sjson.NewJson(inBytes[4:]) //4字节的长度信息
		if err != nil {
			//fmt.Println("recv json body not correct --------->>>headlength:", headlength, "inBytes[headlength:]", string(inBytes[headlength:]))
			//fmt.Println("inBytes:", string(inBytes), "inBytes[headlength]: ", inBytes[headlength])
			//fmt.Println("err:", err)
			return nil
		}

		return outJson
	}

	inLen := len(inBytes)
	if inLen < 2 {
		return nil
	}

	wsHeadflag := inBytes[0]   //web socket 1个字节头部标示
	wsFin := (wsHeadflag >> 7) //是否结尾包
	wsOp := wsHeadflag & 0xF   //消息类型

	if wsFin != 1 {
		sslog.LoggerErr("Decoder : Need Buffer")
		return nil
	}

	wsIsMsgData := false
	switch wsOp {
	case 0:
		sslog.LoggerErr("Decoder : Frame Data")
	case 1, 2:
		wsIsMsgData = true
	case 8, 9, 10:
		wsIsMsgData = false
	default:
		return nil
	}

	wsHeadpayload := inBytes[1] //websocket头部内容长度属性
	payloadlen := wsHeadpayload & 0x7F
	mask := wsHeadpayload >> 7
	headlength := 2 //至少2个字节的websocket头部长度

	//无数据处理
	if wsIsMsgData != true {
		return nil
	}

	//有数据需要处理, 定位到json数据开始的位置
	if payloadlen == 126 { //0x7E
		headlength = headlength + 2
	} else if payloadlen == 127 { //0x7F
		headlength = headlength + 8
	}

	if mask == 1 {
		headlength = headlength + 4
	}

	outJson, err := sjson.NewJson(inBytes[headlength:])
	if err != nil {
		//fmt.Println("recv json body not correct --------->>>headlength:", headlength, "inBytes[headlength:]", string(inBytes[headlength:]))
		//fmt.Println("inBytes:", string(inBytes), "inBytes[headlength]: ", inBytes[headlength])
		//fmt.Println("err:", err)
		return nil
	}

	return outJson
}

func ResponseWebsocketHandShake(session *config.TcpSession) error {
	var err error
	if session.LocalPort == 80 {
		serverKey := "WRVoTRt3LIcNYP32yfQM44zmAkY="

		httpStr := "HTTP/1.1 101 Web Socket Protocol Handshake\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + serverKey + "\r\n\r\n"
		_, err = session.Conn.Write([]byte(httpStr))
	} else {
		strFlashSocket := "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<!DOCTYPE cross-domain-policy SYSTEM \"http://www.adobe.com/xml/dtds/cross-domain-policy.dtd\">\n<cross-domain-policy>\n<site-control permitted-cross-domain-policies=\"all\"/>\n<allow-access-from domain=\"*\" to-ports=\"*\" />\n</cross-domain-policy>"
		_, err = session.Conn.Write([]byte(strFlashSocket))
	}

	return err
}
