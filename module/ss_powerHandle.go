// Copyright (c) , zhoucb, Strong Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/zhouchangbo/ss-go

package module

import (
	"config"
	"encoding/json"
	sjson "github.com/bitly/go-simplejson"
)

const (
	/** 客户端的权限修改请求 */
	CLIENT_CHANGEPOWER = 1

	/**服务端的的权限修改回复 */
	SERVER_CHANGEPOWER_REPLY = 101

	/**服务端的的权限修改通知 */
	SERVER_CHANGEPOWER_INFORM = 201
)

func PowerCommunicationHandler(sessionType int, js *sjson.Json, session *config.TcpSession) {
	switch sessionType {
	case CLIENT_CHANGEPOWER:
		ClientChangePowerHandler(js, session)
	}
}

func ClientChangePowerHandler(js *sjson.Json, session *config.TcpSession) error {
	if session.Cid != 0 {
		pJson := js.Get("p")
		cid := session.Cid
		uid := session.Userid

		uDB, okuDB := config.GetUserByUid(cid, uid)
		if okuDB && uDB.Character == 1 {
			cF := config.GetConference(cid)
			bAudio, _ := pJson.Get("audio").Bool()
			bVideo, _ := pJson.Get("video").Bool()
			bHand, _ := pJson.Get("hand").Bool()

			if cF.Audio != false && bAudio {
				config.ClosePower(cid, 1, true)
				if cF.Audio == false && len(cF.MapAudioList) > 0 {
					config.ConferenceListRWMutex.Lock()
					cF.MapAudioList = nil
					config.ConferenceListRWMutex.Unlock()
				}
			}

			if cF.Video == false && bVideo {
				config.ClosePower(cid, 2, true)
				if cF.Audio == false && len(cF.MapVideoList) > 0 {
					config.ConferenceListRWMutex.Lock()
					cF.MapVideoList = nil
					config.ConferenceListRWMutex.Unlock()
				}
			}

			if cF.Audio && bAudio == false {
				config.ClosePower(cid, 1, false)
			}

			if cF.Video && bVideo == false {
				config.ClosePower(cid, 2, false)
			}

			if cF.Hand && bHand == false {
				config.ClosePower(cid, 3, false)
			}

			config.ConferenceListRWMutex.Lock()
			cF.Audio = bAudio
			cF.Chat, _ = pJson.Get("chat").Bool()
			cF.Hand = bHand
			cF.Video = bVideo
			config.ConferenceListRWMutex.Unlock()

			cfJs, _ := json.Marshal(cF)
			cfJson, _ := sjson.NewJson(cfJs)

			//权限改变通知
			config.BroadCastToClients(cid, config.POWER*1000+SERVER_CHANGEPOWER_INFORM, cfJson, session)
		}

	}

	return nil
}
