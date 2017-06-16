// Copyright (c) , zhoucb, Strong Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/zhouchangbo/ss-go

package module

import (
	"config"
	sjson "github.com/bitly/go-simplejson"
)

const (
	/**
	 * 文件上传完成
	 * **/
	UPLOAD_COMPLETE = 1
	/**
	 * 文件上传回复
	 * **/
	UPLOAD_REPLY = 2
	/**
	 * 文件上传通知
	 * **/
	UPLOAD_INFO = 3
)

func UploadCommunicationHandler(sessionType int, js *sjson.Json, session *config.TcpSession) {
	switch sessionType {
	case UPLOAD_COMPLETE:
		broadCastMsg(js, session)
	default:
	}
}

func broadCastMsg(js *sjson.Json, session *config.TcpSession) error {
	dataJson, _ := sjson.NewJson([]byte(`{}`))
	cid := session.Cid

	//上传成功通知
	config.BroadCastToClients(cid, config.UPLOAD*1000+UPLOAD_INFO, dataJson, session)

	return nil
}
