// Copyright (c) , zhoucb, Strong Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/zhouchangbo/ss-go

package module

import (
	"config"
	"encoding/json"
	sjson "github.com/bitly/go-simplejson"
	"runtime"
	"sslog"
	"strconv"
	"strings"
)

const (
	/**举手*/
	CLIENT_LISTACTION = 1

	/** 服务端对于客户端举手的回复*/
	SERVER_LISTACTION_REPLY = 101

	/** 服务端对于其他客户端举手的通知 */
	SERVER_LISTACTION_INFORM = 201

	/** 客户端菜单的7种请求 */
	CLIENT_MENUACTION = 2

	/** 服务端对于客户端菜单的7种请求的回复 */
	SERVER_MENUACTION_REPLY = 102

	/** 服务端对于其他客户端菜单的7种请求的通知 */
	SERVER_MENUACTION_INFORM = 202

	/**客户端的设备*/
	CLIENT_EQUIPMENTINFO = 3

	/**客户端的设备回复*/
	SERVER_EQUIPMENTINFO_REPLY = 103

	/**客户端的点名*/
	CLIENT_CALL = 4

	/**客户端的点名的回复*/
	SERVER_CALL_REPLY = 104

	/**客户端的点名的通知*/
	SERVER_CALL_INFORM = 204

	/**客户端的点名确认*/
	CLIENT_SURECALL = 5
)

func ListCommunicationHandler(sessionType int, js *sjson.Json, session *config.TcpSession) {
	switch sessionType {
	case CLIENT_LISTACTION:
		clientHandHandler(js, session)
	case CLIENT_MENUACTION:
		clientMenuHandler(js, session)
	case CLIENT_EQUIPMENTINFO:
		clientEquipHandler(js, session)
	case CLIENT_CALL:
		clientCall(js, session)
	case CLIENT_SURECALL:
		clientSureCall(js, session)
	}
}

func clientHandHandler(js *sjson.Json, session *config.TcpSession) error {
	if session.Cid != 0 {
		cid := session.Cid
		userid := session.Userid

		pJs := js.Get("p")

		config.SessionListRWMutex.Lock()
		defer config.SessionListRWMutex.Unlock()
		uDB, _ := config.GetUserByUid(cid, userid)

		sDB := config.GetConferenceHost(cid, 2)
		iType, _ := pJs.Get("type").Int()
		b, _ := pJs.Get("b").Bool()

		time := config.SystemSec()
		iTime := int(time)

		if iType == 1 {
			audioUrl, _ := pJs.Get("audioUrl").String()
			if strings.EqualFold(audioUrl, "") == false && config.CheckHasPower(cid, userid, 4) {
				return nil
			}

			uDB.Audio, _ = pJs.Get("audio").Bool()
			uDB.AudioUrl, _ = pJs.Get("audioUrl").String()
			if _, err := pJs.Get("audioSynId").String(); err == nil {
				uDB.AudioSynId, _ = pJs.Get("audioSynId").String()
			}
		} else if iType == 2 {
			videoUrl, _ := pJs.Get("videoUrl").String()
			if strings.EqualFold(videoUrl, "") == false && config.CheckHasPower(cid, userid, 3) {
				return nil
			}

			bv, _ := pJs.Get("video").Bool()
			if uDB.Character == 3 && bv {
				if len(config.GetOpenVideoList(uDB.Cid)) < 4 || strings.EqualFold(uDB.VideoUrl, "") == false || sDB == nil && len(config.GetOpenVideoList(uDB.Cid)) < 5 {
					uDB.Video, _ = pJs.Get("video").Bool()
					uDB.VideoUrl, _ = pJs.Get("vidoeUrl").String()
					if _, err := pJs.Get("videoSynId").String(); err == nil {
						uDB.VideoSynId, _ = pJs.Get("videoSynId").String()
					}

					if uDB.VideoTime == 0 {
						uDB.VideoTime = iTime
					}
				} else {
					uDB.Video, _ = pJs.Get("video").Bool()
					uDB.VideoUrl = ""
					uDB.VideoTime = 0

					hostDB := config.GetConferenceHost(uDB.Cid, 1)
					if hostDB != nil && hostDB.IoSession != nil {

						//服务端对视频请求回复
						uDBbyte, _ := json.Marshal(uDB)
						uDBJson, _ := sjson.NewJson(uDBbyte)
						config.SendToClient(config.VIDEO*1000+SERVER_VIDEOOPEN_REPLY, uDBJson, session)
					}
				}
			} else if uDB.Character == 2 {
				if len(config.GetOpenVideoList(uDB.Cid)) < 5 {
					uDB.Video, _ = pJs.Get("video").Bool()
					uDB.VideoUrl, _ = pJs.Get("videoUrl").String()
					if _, err := pJs.Get("videoSynId").String(); err == nil {
						uDB.VideoSynId, _ = pJs.Get("videoSynId").String()
					}
					uDB.VideoTime = iTime
				} else {
					tmpDBSlice := config.GetOpenVideoList(uDB.Cid)
					qDB := tmpDBSlice[0]
					qDB.VideoUrl = ""
					qDB.VideoTime = 0
					qDB.AudioSynId = ""

					qDBJs, _ := json.Marshal(qDB)
					qSjon, _ := sjson.NewJson(qDBJs)
					qSjon.Set("type", iType)

					//对举手的通知
					config.SendToClient(config.LIST*1000+SERVER_LISTACTION_INFORM, qSjon, session)

					uDB.Video, _ = pJs.Get("video").Bool()
					uDB.VideoUrl, _ = pJs.Get("videoUrl").String()
					if _, err := pJs.Get("videoSynId").String(); err == nil {
						uDB.VideoSynId, _ = pJs.Get("videoSynId").String()
					}
					uDB.VideoTime = iTime
				}
			} else {
				uDB.Video, _ = pJs.Get("video").Bool()
				uDB.VideoUrl, _ = pJs.Get("videoUrl").String()
				if _, err := pJs.Get("videoSynId").String(); err == nil {
					uDB.VideoSynId, _ = pJs.Get("videoSynId").String()
				}
				uDB.VideoTime = iTime
			}
		} else if iType == 3 {
			if config.CheckHasPower(cid, userid, 1) == false {
				return nil
			}
			uDB.Hand, _ = pJs.Get("hand").Bool()
		} else if config.DesktopType == iType {
			if config.CheckHasPower(cid, userid, 1) == false {
				return nil
			}

			if _, err := pJs.Get("desktop").Bool(); err == nil {
				uDB.Desktop, _ = pJs.Get("desktop").Bool()
			}

			if _, err := pJs.Get("desktopUrl").String(); err == nil {
				uDB.DesktopUrl, _ = pJs.Get("desktopUrl").String()
			}

			if _, err := pJs.Get("desktopSynId").String(); err == nil {
				uDB.DesktopSynId, _ = pJs.Get("desktopSynId").String()
			}
		}

		sendJs, _ := json.Marshal(uDB)
		sendJson, _ := sjson.NewJson(sendJs)
		sendJson.Set("type", iType)
		//对举手的客户端应答
		config.SendToClient(config.LIST*1000+SERVER_LISTACTION_REPLY, sendJson, session)

		if b {
			//增加记录信息
			strVideo := "开启"
			strAudio := "开启"

			if uDB.Video == false {
				strVideo = " 关闭"
			}

			if uDB.Audio == false {
				strAudio = "关闭"
			}

			strWriteBuf := "会议号" + strconv.Itoa(uDB.Cid) + "\t" + "视频使能: " + strVideo + "\t" +
				"视频地址: " + uDB.VideoUrl + "\t" +
				"音频使能: " + strAudio + "\t" +
				"音频地址: " + uDB.AudioUrl + "\t" +
				"用户ID: " + uDB.Userid + "\n"

			_, file, line, _ := runtime.Caller(0)
			sslog.LoggerDebug("[%s:%d]clientHandHandler--- strWriteBuf: %s", file, line+1, strWriteBuf)

			//对举手的通知
			config.BroadCastToClients(uDB.Cid, config.LIST*1000+SERVER_LISTACTION_INFORM, sendJson, session)
		} else {
			//有点疑问
			config.SendToClient(config.LIST*1000+SERVER_LISTACTION_INFORM, sendJson, session)
		}
	}

	return nil
}

func clientMenuHandler(js *sjson.Json, session *config.TcpSession) error {
	if session.Cid != 0 {
		pJs := js.Get("p")
		cid := session.Cid
		uid := session.Userid
		uDB, _ := config.GetUserByUid(cid, uid)
		cF := config.GetConference(cid)
		if uDB.Character == 1 {
			iType, _ := pJs.Get("type").Int()
			userid, _ := pJs.Get("userid").String()
			aDB, okaDB := config.SessionList[cid][userid]
			sDB := config.GetConferenceHost(cid, 2)

			if strings.EqualFold(uDB.Userid, userid) == false {
				if okaDB {
					return nil
				}

				if iType == 1 {
					aDB.Character = 1

					config.SessionListRWMutex.Lock()
					uDB.Character = 3
					config.SessionListRWMutex.Unlock()

					if aDB.Hand {
						aDB.Hand = false
					}

					if strings.EqualFold(uDB.VideoUrl, "") == false {
						config.SessionListRWMutex.Lock()
						uDB.VideoUrl = ""
						config.SessionListRWMutex.Unlock()
					}

					if strings.EqualFold(uDB.AudioUrl, "") == false && cF.Audio == false {
						config.SessionListRWMutex.Lock()
						uDB.AudioUrl = ""
						config.SessionListRWMutex.Unlock()
					}

					arr := strings.Split(aDB.Userid, "-")
					auid := arr[0]

					arr = strings.Split(uDB.Userid, "-")
					uuid := arr[0]

					strUrl := "http://" + config.HttpAddress + "/index.php/user/update_partner_role/" + auid + "/1"
					session.SendCh <- strUrl

					strUrl = "http://" + config.HttpAddress + "/index.php/user/update_partner_role/" + uuid + "/3"
					session.SendCh <- strUrl
				} else if iType == 2 {
					if sDB != nil {
						sDB.Character = 3
						sDB.Desktop = false
						sDB.DesktopUrl = ""
					}

					aDB.Character = 2
					if aDB.Hand {
						aDB.Hand = false
					}
				} else if iType == 3 {
					aDB.Character = 3
					aDB.Desktop = false
					aDB.DesktopUrl = ""
				} else if iType == 4 {
					aDB.AudioUrl = ""
				} else if iType == 5 {
					aDB.VideoUrl = ""
					aDB.VideoTime = 0
				} else if iType == 6 {
					config.QuiteUser(uDB.Cid, userid, session, config.LIST*1000+SERVER_MENUACTION_REPLY, config.LIST*1000+SERVER_MENUACTION_INFORM)
				} else if iType == 7 {
					aDB.Hand = false
				} else if iType == 8 {
					config.ConferenceListRWMutex.Lock()
					cF.MapVideoList[aDB.Userid] = true
					config.ConferenceListRWMutex.Unlock()
				} else if iType == 9 {
					config.ConferenceListRWMutex.Lock()
					cF.MapAudioList[aDB.Userid] = true
					config.ConferenceListRWMutex.Unlock()
					if aDB.Hand {
						aDB.Hand = false
					}
				} else if iType == 10 {
					config.ConferenceListRWMutex.Lock()
					if _, ok := cF.MapVideoList[aDB.Userid]; ok {
						delete(cF.MapVideoList, aDB.Userid)
					}
					config.ConferenceListRWMutex.Unlock()

					if strings.EqualFold(aDB.VideoUrl, "") == false {
						aDB.VideoUrl = ""
					}
				} else if iType == 11 {
					config.ConferenceListRWMutex.Lock()
					if _, ok := cF.MapAudioList[aDB.Userid]; ok {
						delete(cF.MapAudioList, aDB.Userid)
					}
					config.ConferenceListRWMutex.Unlock()

					if strings.EqualFold(aDB.AudioUrl, "") == false {
						aDB.AudioUrl = ""
					}

					aJS, _ := json.Marshal(uDB)
					bJS, _ := json.Marshal(aDB)

					aJson, _ := sjson.NewJson(aJS)
					bJson, _ := sjson.NewJson(bJS)

					aJson.Set("type", iType)
					bJson.Set("type", iType)
					if iType == 1 {

						//通知与会的消息aJson
						config.BroadCastToClients(uDB.Cid, config.LIST*1000+SERVER_MENUACTION_INFORM, aJson, session)

						//通知与会的消息bJson
						config.BroadCastToClients(uDB.Cid, config.LIST*1000+SERVER_MENUACTION_INFORM, bJson, session)
					} else if iType == 8 || iType == 9 || iType == 10 || iType == 11 {
						cfJs, _ := json.Marshal(cF)
						cfJson, _ := sjson.NewJson(cfJs)
						cfJson.Set("type", iType)

						//菜单通知消息
						config.BroadCastToClients(uDB.Cid, config.LIST*1000+SERVER_MENUACTION_INFORM, cfJson, session)

						if iType == 11 {
							bJson.Set("type", 4)
							config.BroadCastToClients(uDB.Cid, config.LIST*1000+SERVER_MENUACTION_INFORM, bJson, session)
						} else if iType == 10 {
							bJson.Set("type", 5)
							config.BroadCastToClients(uDB.Cid, config.LIST*1000+SERVER_MENUACTION_INFORM, bJson, session)
						}
					} else if iType == 12 || iType == 13 {
						if aDB.IoSession != nil {
							config.SendToClient(config.LIST*1000+SERVER_MENUACTION_INFORM, bJson, session)
						}
					} else if iType != 6 && iType != 8 && iType != 9 && iType != 10 && iType != 11 {
						//通知消息
						config.BroadCastToClients(uDB.Cid, config.LIST*1000+SERVER_MENUACTION_INFORM, bJson, session)

						if iType == 2 && sDB != nil {
							sJs, _ := json.Marshal(sDB)
							ssJson, _ := sjson.NewJson(sJs)
							ssJson.Set("type", iType)

							config.BroadCastToClients(uDB.Cid, config.LIST*1000+SERVER_MENUACTION_INFORM, ssJson, session)
						}
					}
				}
			}
		}
	}

	return nil
}

func clientEquipHandler(js *sjson.Json, session *config.TcpSession) {
	if session != nil && session.Cid != 0 {
		cid := session.Cid
		uDB := config.GetConferenceHost(cid, 1)
		if uDB != nil {
			if uDB.IoSession != nil {
				pJson := js.Get("p")

				//设备端应答
				config.SendToClient(config.LIST*1000+SERVER_MENUACTION_INFORM, pJson, session)
			}
		}
	}
}

func clientCall(js *sjson.Json, session *config.TcpSession) error {
	if session.Cid != 0 {
		cid := session.Cid
		uid := session.Userid

		uDB, okuDB := config.GetUserByUid(cid, uid)

		cF := config.GetConference(cid)

		time := config.SystemSec()
		iTime := int(time)

		config.ConferenceListRWMutex.Lock()
		if okuDB && uDB.Character == 1 {
			if cF.CallTime == 0 {
				cF.CallTime = iTime
			}

			if iTime-cF.CallTime < 30 && iTime-cF.CallTime > 0 {

			} else {
				pJson := js.Get("p")
				pJson.Set("callTime", 30)

				//通知点名消息
				config.BroadCastToClients(cid, config.LIST*1000+SERVER_CALL_INFORM, pJson, session)
			}
		}
		config.ConferenceListRWMutex.Unlock()
	}

	return nil
}

func clientSureCall(js *sjson.Json, session *config.TcpSession) error {
	if session.Cid != 0 {
		cid := session.Cid
		uid := session.Userid

		config.ConferenceListRWMutex.Lock()
		uDB, okuDB := config.SessionList[cid][uid]
		cF, okcF := config.ConferenceList[cid]
		if okcF && okuDB {
			time := config.SystemSec()
			iTime := int(time)
			if iTime-cF.CallTime <= 30 && iTime-cF.CallTime > 0 {
				cF.MapCallTimeList[uDB.Userid] = true
			}
		}
		config.ConferenceListRWMutex.Unlock()
	}

	return nil
}
