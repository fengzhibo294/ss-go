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
	/**客户端的聊天请求*/
	CLENT_CHATSEND = 1

	/** 服务端的聊天请求的回复*/
	SERVER_CHATSEND_REPLY = 101

	/** 服务端的聊天请求的通知  */
	SERVER_CHATSEND_INFORM = 201
)

func ChatCommunicationHandler(sessionType int, js *sjson.Json, session *config.TcpSession) {
	switch sessionType {
	case CLENT_CHATSEND:
		clientChatSendHandler(js, session)
	}
}

func clientChatSendHandler(js *sjson.Json, session *config.TcpSession) error {
	if session.Cid != 0 {
		pJson := js.Get("p")
		cid := session.Cid
		uid := session.Userid
		if config.CheckHasPower(cid, uid, 2) == false {
			return nil
		}

		//聊天消息通知
		config.BroadCastToClients(cid, config.CHAT*1000+SERVER_CHATSEND_INFORM, pJson, session)
	}

	return nil
}
