// Copyright (c) , zhoucb, Strong Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/zhouchangbo/ss-go

package module

import (
	"config"
	"encoding/json"
	sjson "github.com/bitly/go-simplejson"
	"strings"
)

const (
	/**视频更改请求*/
	CLIENT_VIDEOCHANGE = 1

	/** 服务端对于视频更改请求的回复*/
	SERVER_VIDEOCHANGE_REPLY = 101

	/** 服务端对于视频更改请求的通知 */
	SERVER_VIDEOCHANGE_INFORM = 201

	/**服务端对请求视频的回复*/
	SERVER_VIDEOOPEN_REPLY = 102
)

func VideoCommunicationHandler(sessionType int, js *sjson.Json, session *config.TcpSession) {
	switch sessionType {
	case CLIENT_VIDEOCHANGE:
		clientVideoChangeHandler(js, session)
	}
}

func clientVideoChangeHandler(js *sjson.Json, session *config.TcpSession) error {
	if session.Cid != 0 {
		pJson := js.Get("p")
		cid := session.Cid
		uid := session.Userid

		avideoUrl, _ := pJson.Get("avideoUrl").String()
		time := config.SystemSec()
		iTime := int(time)
		uDB, okuDB := config.GetUserByUid(cid, uid)
		if okuDB && uDB.Character == 1 {
			mid, _ := pJson.Get("mid").String()
			aid, _ := pJson.Get("aid").String()
			uMDB, okuMDB := config.GetUserByUid(cid, mid)
			uADB, okuADB := config.GetUserByUid(cid, aid)
			if okuMDB && okuADB {
				return nil
			}

			if strings.EqualFold(uADB.Userid, uMDB.Userid) {
				return nil
			}

			if strings.EqualFold(uADB.Userid, uDB.Userid) || strings.EqualFold(uMDB.Userid, uDB.Userid) {
				return nil
			}

			config.SessionListRWMutex.Lock()
			uMDB.Video = false
			uMDB.VideoUrl = ""
			uMDB.VideoSynId = ""
			config.SessionListRWMutex.Unlock()

			uADB.VideoUrl = avideoUrl
			uADB.VideoTime = iTime

			uMDBJs, _ := json.Marshal(uMDB)
			uMDBJson, _ := sjson.NewJson(uMDBJs)

			uADBJs, _ := json.Marshal(uADB)
			uADBJson, _ := sjson.NewJson(uADBJs)

			//视频变更通知
			//uMDB json
			config.BroadCastToClients(cid, config.VIDEO*1000+SERVER_VIDEOCHANGE_INFORM, uMDBJson, session)

			//uADB Json
			config.BroadCastToClients(cid, config.VIDEO*1000+SERVER_VIDEOCHANGE_INFORM, uADBJson, session)
		}
	}

	return nil
}
