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
	/**客户端的浏览器同步*/
	CLIENT_SYNCBROWSE = 1

	/**客户端的浏览器同步回复*/
	SERVER_SYNCBROWSE_REPLY = 101

	/**客户端的浏览器同步通知*/
	SERVER_SYNCBROWSE_INFORM = 201
)

func BrowerCommunicationHandler(sessionType int, js *sjson.Json, session *config.TcpSession) {
	switch sessionType {
	case CLIENT_SYNCBROWSE:
		clientSyncBrowseHandler(js, session)
	}
}

func clientSyncBrowseHandler(js *sjson.Json, session *config.TcpSession) {
	if session.Cid != 0 {
		pJson := js.Get("p")
		cid := session.Cid
		uid := session.Userid

		cFDB := config.GetConference(cid)

		config.ConferenceListRWMutex.Lock()
		cFDB.SelectIndex = 2
		cFDB.BrowseData = pJson
		config.ConferenceListRWMutex.Unlock()

		uDB, _ := config.GetUserByUid(cid, uid)

		if uDB.Character == 1 {
			//浏览同步通知
			config.BroadCastToClients(cid, config.BROWER*1000+SERVER_SYNCBROWSE_INFORM, pJson, session)
		}
	}

}
