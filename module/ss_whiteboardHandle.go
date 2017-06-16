// Copyright (c) , zhoucb, Strong Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/zhouchangbo/ss-go

package module

import (
	"config"
	sjson "github.com/bitly/go-simplejson"
	"sslog"
	"sync"
)

const (

	/**
	 * 客户端打开白板
	 * **/
	CLIENT_OPEN_BOARD = 1

	/**
	 * 回复客户端打开白板
	 * **/
	SERVER_OPEN_BOARD_INFO = 2

	/**
	 * 回复打开/关闭白板的是否成功
	 * **/
	SERVER_OPEN_BOARD_REPLY = 3

	/**
	 * 客户端切换白板
	 * **/
	CLIENT_SWITCH_BOARD = 11

	/**
	 * 回复客户端切换白板
	 * **/
	SERVER_SWITCH_BOARD_INFO = 12

	/**
	 * 画图
	 * **/
	CLIENT_DRAW_BOARD = 21

	/**
	 * 画图事件广播
	 * **/
	SERVER_DRAW_BOARD_INFO = 22

	/**
	 * 查询画图数据
	 * **/
	CLIENT_GET_BOARD_DATA = 31

	/**
	 * 返回画图数据
	 * **/
	SETVER_POST_BOARD_DATA = 32

	/**
	 * 客户端翻页
	 * **/
	CLIENT_DOCUMENT_TOOL_CHANGE = 41

	/**
	 * 广播翻页
	 * **/
	SERVER_DOCUMENT_TOOL_CHANGE = 42

	/**
	 * resize事件
	 * **/
	CLIENT_RESIZE_WINDOWS = 51

	/**
	 * RESIZE事件广播
	 * **/
	SERVER_RESIZE_WINDOWNS = 52

	/**
	 * 客户端播放动画事件
	 * **/
	CLIENT_PLAY_ANIMATION = 61

	/**
	 * 服务端广播动画消息
	 * **/
	SERVER_PLAY_ANIMATION = 62
)

type MessageDealer struct {
	cid       int                            //会议id
	curBoards *sjson.Json                    //当前显示的白板
	whites    map[int]*sjson.Json            //白板数据
	boards    map[string]map[int]*sjson.Json //已建立的白板
	docs      map[int]*sjson.Json            //文档数据
}

func NewMessageDealer(cid int) *MessageDealer {
	deal := new(MessageDealer)
	deal.cid = cid
	deal.whites = make(map[int]*sjson.Json)
	deal.boards = make(map[string]map[int]*sjson.Json)
	deal.docs = make(map[int]*sjson.Json)
	return deal
}

var (
	MsgDealerRWMutex sync.RWMutex
	MsgDealer        map[int]*MessageDealer
)

//自动初始化
func init() {
	MsgDealer = make(map[int]*MessageDealer)
}

func WhiteBoardCommunicationHandler(sessionType int, js *sjson.Json, session *config.TcpSession) {
	iType := sessionType % 1000
	cid := session.Cid

	sslog.LoggerDebug("WhiteBoardCommunicationHandler==   iType[%d] cid[%d]", iType, cid)

	MsgDealerRWMutex.Lock()
	dealer, ok := MsgDealer[cid]
	if ok == false {
		sslog.LoggerDebug("WhiteBoardCommunicationHandler==  NewMessageDealer")

		tmpdealer := NewMessageDealer(cid) //new MessageDealer
		MsgDealer[cid] = tmpdealer
	}
	sslog.LoggerDebug("WhiteBoardCommunicationHandler==  ok[%v]", ok)
	MsgDealerRWMutex.Unlock()

	if iType != CLIENT_GET_BOARD_DATA {
		config.ConferenceListRWMutex.Lock()
		cFDB := config.ConferenceList[cid]
		cFDB.SelectIndex = 1
		config.ConferenceListRWMutex.Unlock()
	}

	dealer.CommunicationHandler(iType, js, session)
}

func OnCloseConference(cid int) {
	MsgDealerRWMutex.Lock()
	delete(MsgDealer, cid)
	MsgDealerRWMutex.Unlock()

}

func (this *MessageDealer) CommunicationHandler(sessionType int, js *sjson.Json, session *config.TcpSession) {
	iType := sessionType % 1000
	this.DealMsg(iType, js, session)
}

// 白板消息处理
func (this *MessageDealer) DealMsg(sessionType int, js *sjson.Json, session *config.TcpSession) {
	switch sessionType {
	case CLIENT_OPEN_BOARD:
		this.OnClientOpenBoard(js, session)
	case CLIENT_SWITCH_BOARD:
		this.OnClientSwitchBoard(js, session)
	case CLIENT_DRAW_BOARD:
		this.OnClientDrawBoard(js, session)
	case CLIENT_GET_BOARD_DATA:
		this.OnClientGetDrawData(session)
	case CLIENT_DOCUMENT_TOOL_CHANGE:
		this.OnClientChange(js, session)
	case CLIENT_RESIZE_WINDOWS:
		this.OnClientResize(js, session)
	case CLIENT_PLAY_ANIMATION:
		this.OnPlayAnimationm(js, session)
	}
}

//当收到客户端播放动画消息
func (this *MessageDealer) OnPlayAnimationm(js *sjson.Json, session *config.TcpSession) {
	pJson := js.Get("p")
	cid := session.Cid

	//播放动画通知
	config.BroadCastToClients(cid, config.WHITEBOARD*1000+SERVER_PLAY_ANIMATION, pJson.Get("data"), session)
}

//客户端最大化/最小化消息
func (this *MessageDealer) OnClientResize(js *sjson.Json, session *config.TcpSession) {
	pJson := js.Get("p")
	cid := session.Cid

	//客户端最大化/最小化通知
	config.BroadCastToClients(cid, config.WHITEBOARD*1000+SERVER_RESIZE_WINDOWNS, pJson, session)
}

//客户端翻页
func (this *MessageDealer) OnClientChange(js *sjson.Json, session *config.TcpSession) {
	realdataJson := js.GetPath("p", "data")
	cid := session.Cid
	bordIndex, _ := realdataJson.Get("index").Int()
	page, _ := realdataJson.Get("page").Int()
	size, _ := realdataJson.Get("size").Float64()

	MsgDealerRWMutex.Lock()
	tmpIndex, ok := this.curBoards.Get("index").Int()
	if ok == nil && tmpIndex == bordIndex {
		this.curBoards.Set("curPage", page)
		this.curBoards.Set("size", size)
	}

	docMap := this.boards["doc"] // map[int] sjson.json
	for _, value := range docMap {
		objJson := value

		jsIndex, _ := objJson.Get("index").Int()
		if jsIndex == bordIndex {
			objJson.Set("curPage", page)
			objJson.Set("size", size)
		}
	}
	MsgDealerRWMutex.Unlock()

	//客户端翻页通知
	config.BroadCastToClients(cid, config.WHITEBOARD*1000+SERVER_DOCUMENT_TOOL_CHANGE, js.Get("p"), session)
}

//客户端请求画板数据
func (this *MessageDealer) OnClientGetDrawData(session *config.TcpSession) {
	MsgDealerRWMutex.RLock()
	defer MsgDealerRWMutex.RUnlock()

	dataJson, _ := sjson.NewJson([]byte(`{}`))
	/*
		dataJson.Set("whiteBord", this.boards["white"])
		dataJson.Set("docBord", this.boards["doc"])
		dataJson.Set("curBord", this.curBoards)
		dataJson.Set("whiteDatas", this.whites)
		dataJson.Set("docDatas", this.docs)
	*/
	nullarrayJs, _ := sjson.NewJson([]byte(`[]`))
	dataJson.Set("whiteBord", nullarrayJs)
	dataJson.Set("docBord", nullarrayJs)
	dataJson.Set("whiteDatas", nullarrayJs)
	dataJson.Set("docDatas", nullarrayJs)

	//发送画板请求数据应答
	config.SendToClient(config.WHITEBOARD*1000+SETVER_POST_BOARD_DATA, dataJson, session)
}

//画图
func (this *MessageDealer) OnClientDrawBoard(js *sjson.Json, session *config.TcpSession) {
	pJson := js.Get("p")
	cid := session.Cid
	paramJson := pJson.Get("data")

	sslog.LoggerDebug("OnClientDrawBoard : cid[%d] \n", cid)
	//画图通知
	config.BroadCastToClients(cid, config.WHITEBOARD*1000+SERVER_DRAW_BOARD_INFO, pJson, session)

	this.SaveDrawInfo(paramJson)
}

//存储绘图数据
func (this *MessageDealer) SaveDrawInfo(param *sjson.Json) {
	MsgDealerRWMutex.Lock()
	defer MsgDealerRWMutex.Unlock()

	index, _ := param.Get("index").Int()
	isDoc, _ := param.Get("isDoc").Bool()

	if isDoc {
		this.docs[index] = param
	} else {
		this.whites[index] = param
	}

}

//收到切换白板
func (this *MessageDealer) OnClientSwitchBoard(js *sjson.Json, session *config.TcpSession) {
	pJson := js.Get("p")
	cid := session.Cid

	//切换白板通知
	config.BroadCastToClients(cid, config.WHITEBOARD*1000+SERVER_SWITCH_BOARD_INFO, pJson, session)

	this.ServerSwitchBord(pJson)
}

//服务端切换board
func (this *MessageDealer) ServerSwitchBord(js *sjson.Json) {
	realdataJson := js.Get("data")
	this.curBoards = realdataJson
}

// 收到打开白板/文档消息
func (this *MessageDealer) OnClientOpenBoard(js *sjson.Json, session *config.TcpSession) {
	dataJson := js.Get("p")
	cid := session.Cid
	realdataJson := dataJson.Get("data")

	isClose, _ := realdataJson.Get("open").Bool()
	isOk := false
	if isClose {
		isOk = this.NewBord(realdataJson)
	} else {
		isOk = this.DeleteBord(realdataJson)
	}

	if isOk {
		//白板打开通知
		config.BroadCastToClients(cid, config.WHITEBOARD*1000+SERVER_OPEN_BOARD_INFO, dataJson, session)
	} else {
		//打开白板失败应答
		objJson, _ := sjson.NewJson([]byte(`{}`))
		objJson.Set("fail", true)
		isDoc, _ := realdataJson.Get("isDoc").Bool()
		objJson.Set("isDoc", isDoc)
		objJson.Set("isDelete", isClose)

		config.SendToClient(config.WHITEBOARD*1000+SERVER_OPEN_BOARD_REPLY, objJson, session)
	}
}

//删除一个白板
func (this *MessageDealer) DeleteBord(data *sjson.Json) bool {
	MsgDealerRWMutex.Lock()
	defer MsgDealerRWMutex.Unlock()

	isDeleteComplete := false
	isDoc, _ := data.Get("isDoc").Bool()
	dataIndex, _ := data.Get("index").Int()

	arrMap := this.boards["doc"]
	if isDoc == false {
		arrMap = this.boards["white"]
	}

	//删除已建立的白板集合中对应索引的白板
	if _, ok := arrMap[dataIndex]; ok {
		isDeleteComplete = true
	}
	delete(arrMap, dataIndex)

	//有可能重置当前白板
	maxIndex := -1
	for key, _ := range arrMap {
		if maxIndex < key {
			maxIndex = key
		}
	}
	this.curBoards = arrMap[maxIndex]

	//删除相应白板中的数据
	if isDoc {
		delete(this.docs, dataIndex)
	} else {
		delete(this.whites, dataIndex)
	}

	return isDeleteComplete
}

//新建一个白板
func (this *MessageDealer) NewBord(data *sjson.Json) bool {
	MsgDealerRWMutex.Lock()
	defer MsgDealerRWMutex.Unlock()

	isNewComplete := false
	dataIndex, _ := data.Get("index").Int()
	isDoc, _ := data.Get("isDoc").Bool()

	this.curBoards = data
	if isDoc {
		docSize := len(this.boards["doc"])
		if docSize < 5 {
			tmpDocs := make(map[int]*sjson.Json)
			tmpDocs[dataIndex] = data
			this.boards["doc"] = tmpDocs
			this.docs[dataIndex] = nil
			isNewComplete = true
		}
	} else {
		whiteSize := len(this.boards["white"])
		if whiteSize < 5 {
			tmpWhites := make(map[int]*sjson.Json)
			tmpWhites[dataIndex] = data
			this.boards["white"] = tmpWhites
			this.whites[dataIndex] = nil
			isNewComplete = true
		}
	}

	return isNewComplete
}
