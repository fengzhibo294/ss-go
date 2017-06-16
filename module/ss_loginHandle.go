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
	"time"
)

const (
	/**登陆成功*/
	CLIENT_LOGIN = 1

	/** 服务端对于客户端登陆的回复 */
	SERVER_LOGIN_REPLY = 101

	/** 服务端对于其他客户端登陆的通知 */
	SERVER_LOGIN_INFORM = 201

	/**登出通知*/
	SERVER_LOGOUT_REPLY = 102

	/**登出*/
	SERVER_LOGOUT_INFORM = 202

	/**关闭系统*/
	CLIENT_CLOSE = 3

	/**关闭系统*/
	SERVER_CLOSE_INFORM = 203

	//客户端的ping包
	CLIENT_PING = 4

	//服务器对客户端的ping回包
	SERVER_PING_RES = 104
)

func LoginCommunicationHandler(sessionType int, js *sjson.Json, session *config.TcpSession) {
	switch sessionType {
	case CLIENT_LOGIN:
		clientLoginHandler(js, session)
	case CLIENT_CLOSE:
		clientCloseHandler(js, session)
	case CLIENT_PING:
		clientPingHandler(js, session)
	}

}

func clientLoginHandler(js *sjson.Json, session *config.TcpSession) {
	uDB := addSession(js, session)
	if uDB != nil {
		_, file, line, _ := runtime.Caller(0)
		sslog.LoggerDebug("[%s:%d]clientLoginHandler--登陆的人: %s, id=%s", file, line+1, uDB.Name, uDB.Userid)
		jList, _ := config.GetConferenceUserList(uDB.Cid)
		if jList != nil {
			cFDB := config.GetConference(uDB.Cid)

			//b, _ := json.Marshal(cFDB)
			//cfJs, _ := sjson.NewJson(b)
			cfJs := cFDB.GetJson()
			jList.Set("conference", cfJs)
			jList.SetPath([]string{"sync", "selectIndex"}, cFDB.SelectIndex)
			if cFDB.BrowseData != nil {
				jList.SetPath([]string{"sync", "browseData"}, cFDB.BrowseData)
			}

			if uDB.Character == 1 {
				_, file, line, _ := runtime.Caller(0)
				sslog.LoggerDebug("[%s:%d]clientLoginHandler--会议号: %d 会议开始时间:%s", file, line, uDB.Cid, time.Now())
			}

			//写入发送通道
			//-->应答该客户端登入
			config.SendToClient(config.LOGIN*1000+SERVER_LOGIN_REPLY, jList, session)

			//通知与会消息
			config.BroadCastToClients(uDB.Cid, config.LOGIN*1000+SERVER_LOGIN_INFORM, uDB.GetJson(), session)
		}
	}
}

func clientCloseHandler(js *sjson.Json, session *config.TcpSession) {
	if session.Cid != 0 {
		cid := session.Cid
		userid := session.Userid
		uDB, okuDB := config.GetUserByUid(cid, userid)
		if okuDB {
			if uDB.Character == 1 {
				config.CloseConference(uDB.Cid, session, config.LOGIN*1000+SERVER_CLOSE_INFORM)
			}
		}
	}
}

func clientPingHandler(js *sjson.Json, session *config.TcpSession) {
	nJs, _ := sjson.NewJson([]byte(`{}`))
	nResult := 0

	nJs.Set("result", nResult)
	//ping回包
	config.SendToClient(config.LOGIN*1000+SERVER_PING_RES, nJs, session)

	//更新ping包的lastTime
	session.LastTime = config.SystemSec()
}

func addSession(js *sjson.Json, session *config.TcpSession) *config.UserDB {
	pJson := js.Get("p")

	cid, _ := pJson.Get("cid").Int()
	uid, _ := pJson.Get("userid").String()
	name, _ := pJson.Get("name").String()
	character, _ := pJson.Get("character").Int()
	uDB := &config.UserDB{cid, uid, name, character, 0, false, false, false, "", "", 0, 0, 0, "", "", "", false, "", nil}

	uDB.IsOnline = 1
	uDB.IoSession = session
	time := config.SystemMs()
	if uDB.Cid == -1 {
		config.LoginNewConference(uDB)
	} else {
		if _, ok := config.SessionList[uDB.Cid]; ok {
			if tDB, ok1 := config.GetUserByUid(uDB.Cid, uDB.Userid); ok1 {
				arr := strings.Split(uDB.Userid, "-")
				if len(arr) == 1 {
					uDB.Userid = uDB.Userid + strconv.Itoa(int(time))
					uDB.IsOnline = 1
					config.SessionListRWMutex.Lock()
					config.SessionList[uDB.Cid][uDB.Userid] = uDB
					config.SessionListRWMutex.Unlock()
				} else {
					if tDB.IoSession == nil {
						//登出消息
						tDBJs, _ := json.Marshal(tDB) //自带json库
						tDBJson, _ := sjson.NewJson(tDBJs)
						config.SendToClient(config.LOGIN*1000+SERVER_LOGOUT_REPLY, tDBJson, session)
					}

					config.SessionListRWMutex.Lock()
					tDB.IoSession = session
					tDB.IsOnline = 1
					tDB.Name = uDB.Name
					tDB.Audio = false
					tDB.AudioSynId = ""
					tDB.AudioUrl = ""
					tDB.Desktop = false
					tDB.DesktopSynId = ""
					tDB.DesktopUrl = ""
					tDB.Hand = false
					tDB.Video = false
					tDB.VideoSynId = ""
					tDB.VideoUrl = ""
					tDB.VideoTime = 0
					config.SessionListRWMutex.Unlock()

					uDB = tDB
					cF, okcF := config.ConferenceList[uDB.Cid]
					if okcF && strings.EqualFold(cF.Pid, uDB.Userid) {
						cF.Pid = ""
						nDB := config.GetConferenceHost(tDB.Cid, 1)
						nDB.Character = 3
						nDB.Desktop = false

						nDBJs, _ := json.Marshal(nDB) //自带json库

						uDB.Character = 1

						aJS, _ := sjson.NewJson(nDBJs)
						uDBJs, _ := json.Marshal(uDB) //自带json库
						bJS, _ := sjson.NewJson(uDBJs)

						//放入发送通道中
						aJS.Set("type", 14)
						bJS.Set("type", 14)

						//通知菜单请求反馈消息
						config.BroadCastToClients(nDB.Cid, config.LIST*1000+SERVER_MENUACTION_INFORM, aJS, session)

						//通知菜单请求反馈消息
						config.BroadCastToClients(uDB.Cid, config.LIST*1000+SERVER_MENUACTION_INFORM, bJS, session)
					} //if cF != nil && cF.pid != nil && strings.EqualFold(cF.pid, uDB.userid)
				} //if len(arr) == 1  else
			} else { //if _, ok1 := config.SessionList[uDB.cid][uDB.userid]; ok1  else
				//不存在的用户添加到 config.SessionList资源信息中
				config.InsertUserDB(uDB)

			}
		} else { //if _, ok := config.SessionList[uDB.cid]; ok
			err := config.AddNewConferenceList(uDB)
			if err != nil {
				_, file, line, _ := runtime.Caller(0)
				sslog.LoggerErr("[%s:%d ] AddNewConferenceList :%s", file, line+1, err.Error())
			}
		}

		ok := config.CheckConference(uDB.Cid)
		if ok == false {
			err := config.AddNewConference(uDB)
			if err != nil {
				_, file, line, _ := runtime.Caller(0)
				sslog.LoggerErr("[%s:%d]config.AddNewConferenceList :%s", file, line+1, err.Error())
			}
		}
	} //if uDB.cid == -1 else

	if uDB.EnterTime == 0 {
		uDB.EnterTime = int(time)
	}

	session.Cid = uDB.Cid
	session.Userid = uDB.Userid

	return uDB
}
