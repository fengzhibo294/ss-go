// Copyright (c) , zhoucb, Strong Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/zhouchangbo/ss-go

package config

import (
	"db2mysql"
	"encoding/json"
	sjson "github.com/bitly/go-simplejson"
	"net"
	"net/http"
	"os"
	"runtime"
	"sslog"
	"strconv"
	"strings"
	"sync"
	"time"
)

//tcp会话链接属性结构
type TcpSession struct {
	Conn        net.Conn
	RecvCh      chan interface{}
	SendCh      chan interface{}
	IsVerify    bool
	LocalPort   int
	pingTimeout int
	LastTime    int64
	Cid         int
	Userid      string
}

var (
	SqlUrl      = "videoconference:videoconference@tcp(192.168.20.68:3306)/videoconference?charset=utf8"
	HttpAddress = "huiyi.cnstrongwx.cn"
	AudioType   = 1
	VideoType   = 2
	HandType    = 3
	DesktopType = 4
)

var (
	PingTimeout = 30 //sec
)

const (
	/**登陆登出模块*/
	LOGIN = 1
	/**列表模块*/
	LIST = 2
	/**聊天模块*/
	CHAT = 3
	/**权限模块*/
	POWER = 4
	/**上传模块*/
	UPLOAD = 5
	/**白板模块*/
	WHITEBOARD = 6
	/**视频模块*/
	VIDEO = 7
	/**浏览器模块*/
	BROWER = 8
)

type ConferenceDB struct {
	Cid             int             `json:"cid"`         //会议ID
	Hostid          int             `json:"hostid"`      //主持人ID
	Hand            bool            `json:"hand"`        //举手权限
	Chat            bool            `json:"chat"`        //聊天权限
	Video           bool            `json:"video"`       //视频权限
	Audio           bool            `json:"audio"`       //音频权限
	BeginTime       int64           `json:"begintime"`   //开始时间
	Pid             string          `json:"pid"`         //前一个主持人ID
	Status          int             `json:"status"`      //状态
	MapAudioList    map[string]bool `json:"audioList"`   //音频用户列表
	MapVideoList    map[string]bool `json:"videoList"`   //视频用户列表
	SelectIndex     int             `json:"selectindex"` //当前选中的页签 1白板 2协同浏览
	BrowseData      *sjson.Json     `json:"browsedata"`  //浏览的数据
	CallTime        int             `json:"calltime"`    //点名时间
	MapCallTimeList map[string]bool `json:"callList"`    //点名列表
}

func KeystringMapToJson(strMap map[string]bool) *sjson.Json {
	var tmpSringSlice []string

	for tmpStr, _ := range strMap {
		tmpSringSlice = append(tmpSringSlice, tmpStr)
	}

	tmpstrBytes, err := json.Marshal(tmpSringSlice)
	if tmpstrBytes == nil || err != nil {
		return nil
	}

	retJs, _ := sjson.NewJson(tmpstrBytes)
	return retJs
}

type UserDB struct {
	Cid          int         `json:"cid"`       //会议ID
	Userid       string      `json:"userid"`    //角色ID
	Name         string      `json:"name"`      //角色名字
	Character    int         `json:"character"` //角色类型
	IsOnline     int         `json:"isOnline"`  //是否在线
	Video        bool        `json:"video"`     //视频是否开启
	Audio        bool        `json:"audio"`     //音频是否开启
	Hand         bool        `json:"hand"`      //是否举手
	VideoUrl     string      `json:"videoUrl"`
	AudioUrl     string      `json:"audioUrl"`
	VideoTime    int         `json:"videoTime"`
	EnterTime    int         `json:"enterTime"`
	QuiteTime    int         `json:"quiteTime"`
	AudioSynId   string      `json:"audioSynId"`
	VideoSynId   string      `json:"videoSynId"`
	DesktopUrl   string      `json:"desktopUrl"`   //桌面共享vpp
	Desktop      bool        `json:"desktop"`      //是否桌面共享
	DesktopSynId string      `json:"desktopSynId"` //桌面共享Synid
	IoSession    *TcpSession `json:"iosession"`
}

func (this *ConferenceDB) GetJson() *sjson.Json {
	retJs, _ := sjson.NewJson([]byte(`{}`))
	nullarrayJs, _ := sjson.NewJson([]byte(`[]`))

	//MapAudioList转换成数组格式
	audioJs := KeystringMapToJson(this.MapAudioList)

	//MapVideoList转换成数组格式
	videoJs := KeystringMapToJson(this.MapVideoList)

	//MapCallTimeList转成数组格式
	callListJs := KeystringMapToJson(this.MapCallTimeList)

	retJs.Set("cid", this.Cid)
	retJs.Set("video", this.Video)
	retJs.Set("audio", this.Audio)
	retJs.Set("hand", this.Hand)
	retJs.Set("chat", this.Chat)

	ntime := SystemSec()
	keeptime := ntime - this.BeginTime
	retJs.Set("keeptime", keeptime)
	if audioJs != nil {
		retJs.Set("audioList", audioJs)
	} else {
		retJs.Set("audioList", nullarrayJs)
	}

	if videoJs != nil {
		retJs.Set("videoList", videoJs)

	} else {
		retJs.Set("videoList", nullarrayJs)
	}

	retJs.Set("callTime", 30-ntime+int64(this.CallTime))
	if callListJs != nil {
		retJs.Set("callList", callListJs)
	} else {
		retJs.Set("callList", nullarrayJs)
	}

	return retJs
}

func (this *UserDB) GetJson() *sjson.Json {
	retJs, _ := sjson.NewJson([]byte(`{}`))

	retJs.Set("cid", this.Cid)
	retJs.Set("userid", this.Userid)
	retJs.Set("name", this.Name)
	retJs.Set("character", this.Character)
	retJs.Set("isOnline", this.IsOnline)
	retJs.Set("video", this.Video)
	retJs.Set("audio", this.Audio)
	retJs.Set("hand", this.Hand)
	retJs.Set("videoUrl", this.VideoUrl)
	retJs.Set("audioUrl", this.AudioUrl)
	retJs.Set("videoTime", this.VideoTime)
	retJs.Set("audioSynId", this.AudioSynId)
	retJs.Set("videoSynId", this.VideoSynId)
	retJs.Set("desktop", this.Desktop)
	retJs.Set("desktopSynId", this.DesktopSynId)
	retJs.Set("desktopUrl", this.DesktopUrl)

	return retJs
}

var (
	ConferenceListRWMutex sync.RWMutex               //对会议列表加锁处理，go map是非线程安全的
	SessionListRWMutex    sync.RWMutex               //对分块处理表加锁处理
	ConferenceList        map[int]*ConferenceDB      //会议列表
	UserList              map[string]*UserDB         //用户列表
	SessionList           map[int]map[string]*UserDB //分块化处理session
)

func init() {
	ConferenceList = make(map[int]*ConferenceDB)
	UserList = make(map[string]*UserDB)
	SessionList = make(map[int]map[string]*UserDB)
}

//关闭Tcp会话属性链接
func (this *TcpSession) Close() bool {
	_, file, line, _ := runtime.Caller(0)
	sslog.LoggerErr("[%s:%d]TcpSession Close---tcp to close [%s]", file, line+1, this.Conn.RemoteAddr().String())
	this.Conn.Close()

	if _, ok := <-this.SendCh; ok {
		close(this.SendCh)
	}
	if _, ok := <-this.RecvCh; ok {
		close(this.RecvCh)
	}

	return true
}

func LoadGlobaldata() error {
	connect, err := db2mysql.GetConnection(SqlUrl)
	if err != nil {
		return err
	}
	defer db2mysql.DbClose(connect)

	//查询会议列表信息
	sql := "select a.meetingId,a.hostId,a.beginTime,a.status,b.speak,b.chat,b.audio,b.video from meeting a, competence b where a.meetingId = b.meetingId"
	rows, err := db2mysql.SelectSql(connect, sql)
	if err != nil {
		return err
	}
	defer db2mysql.RowsClose(rows)

	//对全局ConferenceList数据初始化
	var meetingId int
	var hostId int
	var strbeginTime string
	var status int
	var speak uint8
	var chat uint8
	var audio uint8
	var video uint8

	for rows.Next() {
		err := rows.Scan(&meetingId, &hostId, &strbeginTime, &status, &speak, &chat, &audio, &video)
		if err != nil {
			_, file, line, _ := runtime.Caller(0)
			sslog.LoggerErr("[%s:%d][LoadGlobaldata:rows.Scan -ConferenceDB] err=%s", file, line+1, err.Error())
			return err
		}

		beginTime := SecStringtimeToInt64(strbeginTime)

		cfDB := &ConferenceDB{meetingId, hostId, speak == 1, chat == 1, video == 1, audio == 1, beginTime, "", status, nil, nil, 0, nil, 0, nil}
		cfDB.MapVideoList = make(map[string]bool)
		cfDB.MapAudioList = make(map[string]bool)
		ConferenceList[cfDB.Cid] = cfDB //此处不需要加锁
	}

	_, file, line, _ := runtime.Caller(0)
	sslog.LoggerInfo("[%s:%d]初始化会议: 会员列表数 %d", file, line+1, len(ConferenceList))

	//查询user信息
	sql = "select `id`,`meetingId`,`partnerId`,`partnerName`,`partnerRole` from `partner` where 1"
	rows, err = db2mysql.SelectSql(connect, sql)
	if err != nil {
		return err
	}

	//对全局SessionList数据初始化
	var usermeetingId int
	var id int
	var partnerId *string
	var partnerName string
	var partnerRole int

	var userid string
	userCount := 0
	for rows.Next() {
		err := rows.Scan(&id, &usermeetingId, &partnerId, &partnerName, &partnerRole)
		if err != nil {
			_, file, line, _ := runtime.Caller(0)
			sslog.LoggerErr("[%s:%d][LoadGlobaldata:rows.Scan -UserDB] err=%s", file, line+1, err.Error())
			return err
		}

		userCount += 1
		if partnerId == nil { //partnerid为nil，不能进行操作，go语言强类型特性
			userid = strconv.Itoa(id) + "-"
		} else {
			userid = strconv.Itoa(id) + "-" + *partnerId
		}

		uDB := &UserDB{meetingId, userid, partnerName, partnerRole, 0, false, false, false, "", "", 0, 0, 0, "", "", "", false, "", nil}

		UserList[userid] = uDB
		SessionList[meetingId] = UserList //此处不需要加锁
	}

	_, file, line, _ = runtime.Caller(0)
	sslog.LoggerInfo("[%s:%d]初始化会议人数 %d", file, line+1, userCount)

	//超时关闭会议
	go func() {
		for {
			time.Sleep(3600 * 1e9) //一小时检测一次
			closeOutTime()
		}
	}()

	//检测点名
	go func() {
		for {
			time.Sleep(2 * 1e9) //2sec 检测一次
			checkCall()
		}
	}()
	return nil
}

func SystemMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func SystemSec() int64 {
	return time.Now().Unix()
}

func SecStringtimeToInt64(strTime string) int64 {
	tTime, err := time.Parse("2006-01-02 15:04:05", strTime)
	if err != nil {
		return int64(0)
	}

	int64Time := tTime.Unix()
	return int64Time
}

func HttpPost(strUrl string) {
	//生成client 参数为默认
	client := &http.Client{}

	//提交请求
	reqest, err := http.NewRequest("POST", strUrl, strings.NewReader("username=kevin&password=*********"))
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		sslog.LoggerErr("[%s:%d]http.NewRequest: url[%s] err[%s]", file, line+1, strUrl, err.Error())
		os.Exit(1)
	}

	//设置编码8859_1
	reqest.Header.Set("Accept-Charset", "UTF-8;q=1, ISO-8859-1;q=0")

	//处理返回结果
	response, _ := client.Do(reqest)

	//返回的状态码
	status := response.StatusCode
	if status != 200 {
		_, file, line, _ := runtime.Caller(0)
		sslog.LoggerErr("[%s:%d]http.NewRequest: url[%s] response.StatusCode[%s]", file, line+1, strUrl, status)
	}

}

/**如果内容里面没有该列表 那就去数据库里面获取该列表*/
func AddNewConferenceList(uDB *UserDB) error {
	connect, err := db2mysql.GetConnection(SqlUrl)
	if err != nil {
		return err
	}
	defer db2mysql.DbClose(connect)

	sql := "select `id`,`meetingId`,`partnerId`,`partnerName`,`partnerRole` from `partner` where `meetingId`= " + strconv.Itoa(uDB.Cid)
	rows, err := db2mysql.SelectSql(connect, sql)
	if err != nil {
		return err
	}
	defer db2mysql.RowsClose(rows)

	var id int
	var meetingId int
	var partnerId string
	var partnerName string
	var partnerRole int
	for rows.Next() {
		err := rows.Scan(&id, &meetingId, &partnerId, &partnerName, &partnerRole)
		if err != nil {
			_, file, line, _ := runtime.Caller(0)
			sslog.LoggerErr("[%s:%d]rows.Scan %s", file, line+1, err.Error())
			return err
		}
		userid := strconv.Itoa(id) + "-" + partnerId
		tmpDB := &UserDB{meetingId, userid, partnerName, partnerRole, 0, false, false, false, "", "", 0, 0, 0, "", "", "", false, "", nil}
		if strings.EqualFold(tmpDB.Userid, uDB.Userid) {
			tmpDB.Name = uDB.Name
			tmpDB.IoSession = uDB.IoSession
			tmpDB.IsOnline = 1
			uDB.Character = tmpDB.Character
		}

		SessionListRWMutex.Lock()
		if value, ok := SessionList[tmpDB.Cid]; ok {
			value[tmpDB.Userid] = tmpDB
		} else {
			tmpUserList := make(map[string]*UserDB)
			tmpUserList[tmpDB.Userid] = tmpDB
			SessionList[tmpDB.Cid] = tmpUserList
		}
		SessionListRWMutex.Unlock()
	}

	return nil
}

/**增加新的会议*/
func AddNewConference(uDB *UserDB) error {
	connect, err := db2mysql.GetConnection(SqlUrl)
	if err != nil {
		return err
	}
	defer db2mysql.DbClose(connect)

	sql := "select a.meetingId,a.hostId,a.beginTime,a.status,b.speak,b.chat,b.audio,b.video from meeting a, competence b where a.meetingId = b.meetingId and a.meetingId = " + strconv.Itoa(uDB.Cid)
	rows, err := db2mysql.SelectSql(connect, sql)
	if err != nil {
		return err
	}
	defer db2mysql.RowsClose(rows)

	var meetingId int
	var hostId int
	var strbeginTime string
	var status int
	var speak int
	var chat int
	var audio int
	var video int

	for rows.Next() {
		err := rows.Scan(&meetingId, &hostId, &strbeginTime, &status, &speak, &chat, &audio, &video)
		if err != nil {
			return err
		}

		beginTime := SecStringtimeToInt64(strbeginTime)
		cfDB := &ConferenceDB{meetingId, hostId, speak == 1, chat == 1, video == 1, audio == 1, beginTime, "", status, nil, nil, 0, nil, 0, nil}
		if cfDB.Cid != 0 {
			ConferenceListRWMutex.Lock() //写锁
			ConferenceList[cfDB.Cid] = cfDB
			ConferenceListRWMutex.Unlock()
		}
	}

	return nil
}

type NewUserDB struct {
	Cid       int    `json:"cid"`       //会议ID
	Userid    string `json:"userid"`    //角色ID
	Name      string `json:"name"`      //角色名字
	Character int    `json:"character"` //角色类型
	IsOnline  int    `json:"isOnline"`  //是否在线
	Video     bool   `json:"video"`     //视频是否开启
	Audio     bool   `json:"audio"`     //音频是否开启
	Hand      bool   `json:"hand"`      //是否举手
	VideoUrl  string `json:"videoUrl"`
	AudioUrl  string `json:"audioUrl"`
	VideoTime int    `json:"videoTime"`
	//EnterTime    int    `json:"enterTime"`
	//QuiteTime    int    `json:"quiteTime"`
	AudioSynId   string `json:"audioSynId"`
	VideoSynId   string `json:"videoSynId"`
	DesktopUrl   string `json:"desktopUrl"`   //桌面共享vpp
	Desktop      bool   `json:"desktop"`      //是否桌面共享
	DesktopSynId string `json:"desktopSynId"` //桌面共享Synid
}

//解决map字段不可用作变量名
type NewUserDBslice struct {
	Users []NewUserDB `json:"map"` //"Users" --> "map"
}

/**获得该会议里面所有人的列表*/
func GetConferenceUserList(cid int) (*sjson.Json, error) {
	var uSlice NewUserDBslice

	SessionListRWMutex.RLock()
	defer SessionListRWMutex.RUnlock()
	if _, ok := SessionList[cid]; ok {
		for _, value := range SessionList[cid] {
			uDB := *value
			uSlice.Users = append(uSlice.Users, NewUserDB{uDB.Cid, uDB.Userid, uDB.Name, uDB.Character, uDB.IsOnline, uDB.Video, uDB.Audio,
				uDB.Hand, uDB.VideoUrl, uDB.AudioUrl, uDB.VideoTime, uDB.AudioSynId, uDB.VideoSynId, uDB.DesktopUrl, uDB.Desktop, uDB.DesktopSynId})
			//打印用户信息
			_, file, line, _ := runtime.Caller(0)
			sslog.LoggerDebug("[%s:%d]GetConferenceUserList--uDB value:%v ", file, line+1, value)
		}

		b, err := json.Marshal(uSlice)
		if err != nil {
			_, file, line, _ := runtime.Caller(0)
			sslog.LoggerErr("[%s:%d]GetConferenceUserList  json err:%s", file, line+1, err.Error())
			sslog.LoggerErr("b=%s,  %v", string(b), b)
			return nil, err
		}

		//打印用户列表信息信息
		_, file, line, _ := runtime.Caller(0)
		sslog.LoggerDebug("[%s:%d]GetConferenceUserList -------json string:%s", file, line+1, string(b))
		newJs, _ := sjson.NewJson(b) //{"map":[UserDB1, UserDB2, ...]}
		return newJs, nil
	}

	return nil, nil
}

func CloseConference(cid int, session *TcpSession, msgType int) error {
	SessionListRWMutex.Lock()
	if _, ok := SessionList[cid]; ok {
		//会议状态更新
		str := "http://" + HttpAddress + "/index.php/meeting/update_meeting_status/" + strconv.Itoa(cid) + "/3"
		session.SendCh <- str //写入到发送通道中

		ConferenceListRWMutex.Lock() //写锁
		cF := ConferenceList[cid]
		if _, ok := ConferenceList[cid]; ok {
			cF.Status = 3
		}
		ConferenceListRWMutex.Unlock()

		//会议结束 用户的离开时间
		tmpUserList := SessionList[cid]
		time := SystemMs() / 1000
		iTime := int(time)
		for _, value := range tmpUserList {
			uDB := value
			if uDB.IoSession != nil {
				arr := strings.Split(uDB.Userid, "-")
				usqlid := arr[0]
				strUrl := "http://" + HttpAddress + "/index.php/user/update_partner_leavetime/" + usqlid + "/" + strconv.Itoa(iTime)
				session.SendCh <- strUrl
			}

		}

		//通知与会客户端会议结束
		jsNull, _ := sjson.NewJson([]byte(`{}`))
		newJs, _ := sjson.NewJson([]byte(`{}`))
		newJs.Set("cid", cid) //通知
		newJs.SetPath([]string{"sendbuf", "m"}, msgType)
		newJs.SetPath([]string{"sendbuf", "p"}, jsNull)
		session.SendCh <- newJs //放入到发送通道中待发送

		//移除该会议
		delete(SessionList, cid)

		//关闭白板
	}
	SessionListRWMutex.Unlock()

	return nil
}

/**
 * 获得会议中的主持人或主讲人
 * @param cid 会议ID
 * @param type 类型 1主持人 2主讲人
 *
 * */
func GetConferenceHost(cid int, iType int) *UserDB {
	SessionListRWMutex.RLock()
	defer SessionListRWMutex.RUnlock()

	if _, ok := SessionList[cid]; ok {
		tmpUserList := SessionList[cid]
		for _, value := range tmpUserList {
			uDB := value
			if uDB.Character == iType {
				return uDB
			}
		}
	}

	return nil
}

/**查看是否有权限
* @param cid 会议ID
* @param uid 角色ID
* @param type 1举手 2聊天 3视频 4音频
* */
func CheckHasPower(cid int, uid string, iType int) bool {
	ConferenceListRWMutex.RLock() //读锁
	defer ConferenceListRWMutex.RUnlock()

	cf := ConferenceList[cid]
	uDB := SessionList[cid][uid]
	if _, ok := ConferenceList[cid]; ok == false {
		return false
	}

	if _, ok := SessionList[cid][uid]; ok == false {
		return false
	}

	if iType == 1 && cf.Hand == false && uDB.Character == 3 {
		return false
	}

	if iType == 2 && cf.Chat == false && uDB.Character == 3 {
		return false
	}

	if iType == 3 && cf.Video == false && uDB.Character == 3 && cf.MapAudioList[uid] == false {
		return false
	}

	if iType == 4 && cf.Audio == false && uDB.Character == 3 && cf.MapAudioList[uid] == false {
		return false
	}

	return true
}

/**获得会议中开启视频的人数*/
func GetOpenVideoList(cid int) []*UserDB {
	var usrSlice []*UserDB

	SessionListRWMutex.RLock()
	if _, ok := SessionList[cid]; ok {
		tmpUserList := SessionList[cid]
		for _, value := range tmpUserList {
			uDB := value
			if uDB.Character == 3 && uDB.Video && strings.EqualFold(uDB.VideoUrl, "") == false {
				usrSlice = append(usrSlice, uDB)
			}
		}
	}
	SessionListRWMutex.RUnlock()

	return usrSlice
}

//踢出用户
func QuiteUser(cid int, uid string, session *TcpSession, responseType int, informType int) error {
	SessionListRWMutex.RLock()
	if _, ok := SessionList[cid]; ok {
		if _, ok = SessionList[cid][uid]; ok {
			uDB := SessionList[cid][uid]

			sendObj, _ := sjson.NewJson([]byte(`{}`))
			sendObj.Set("userid", uid)
			sendObj.Set("type", 6)

			sendJs, _ := sjson.NewJson([]byte(`{}`))
			sendJs.SetPath([]string{"sendbuf", "p"}, sendObj)

			if uDB.IoSession != nil {
				//应答
				sendJs.Set("cid", -1) //cid -1单发回复
				sendJs.SetPath([]string{"sendbuf", "m"}, responseType)
				session.SendCh <- sendJs //放入到发送通道中待发送
			}

			sendJs.Set("cid", cid) //通知其他与会客户端
			sendJs.SetPath([]string{"sendbuf", "m"}, informType)
			session.SendCh <- sendJs //放入到发送通道中待发送
		}
	}
	SessionListRWMutex.RUnlock()

	return nil
}

/**
 * 修改用户权限
 * @param cid 会议ID
 * @param type 类型 1音频2视频3举手
 * */
func ClosePower(cid int, iType int, b bool) error {
	ConferenceListRWMutex.RLock()
	if _, ok := ConferenceList[cid]; ok == false {
		ConferenceListRWMutex.RUnlock()
		return nil
	}
	ConferenceListRWMutex.RUnlock()

	SessionListRWMutex.Lock()
	if _, ok := SessionList[cid]; ok {

		tmpUserList := SessionList[cid]
		for _, value := range tmpUserList {

			uDB := value
			if uDB.Character == 3 {
				if iType == 1 && uDB.Audio {
					uDB.AudioUrl = ""
					if b {
						uDB.Audio = false
					} else if iType == 2 && uDB.Video {
						uDB.VideoUrl = ""
						if b {
							uDB.Video = false
						}
					} else if iType == 3 && uDB.Hand {
						uDB.Hand = false
					}
				}
			}
		}
	}
	SessionListRWMutex.Unlock()

	return nil
}

func closeOutTime() error {

	ConferenceListRWMutex.Lock()
	//遍历会议列表
	for _, value := range ConferenceList {
		cF := value

		iTime := SystemSec()
		if iTime-cF.BeginTime > 86400*7 && cF.Status == 2 {
			strUrl := "http://" + HttpAddress + "/index.php/meeting/update_meeting_status/" + strconv.Itoa(cF.Cid) + "/3"
			HttpPost(strUrl)
			cF.Status = 3
		}
	}
	ConferenceListRWMutex.Unlock()

	//尝试连接库，并测试连通性
	connect, err := db2mysql.GetConnection(SqlUrl)
	if err != nil {
		return err
	}
	defer db2mysql.DbClose(connect)

	sqlStr := "select *from `meeting` where `meetingId` = 1000"
	rows, err := db2mysql.SelectSql(connect, sqlStr)
	if err != nil {
		return err
	}
	defer db2mysql.RowsClose(rows)

	_, file, line, _ := runtime.Caller(0)
	sslog.LoggerDebug("[%s:%d] closeOutTime -- 测试数据库连接正常", file, line+1)
	return nil
}

func checkCall() {
	const SERVER_CALL_REPLY = 104
	iTime := int(SystemSec())

	ConferenceListRWMutex.Lock()
	for _, value := range ConferenceList {
		cF := value
		if cF.CallTime > 0 && iTime-cF.CallTime >= 30 {
			cFJson, _ := json.Marshal(cF)
			sendJs, _ := sjson.NewJson(cFJson)

			uDB := GetConferenceHost(cF.Cid, 1)
			if uDB.IoSession != nil {
				//写入发送通道
				newJs, _ := sjson.NewJson([]byte(`{}`))
				newJs.Set("cid", -1) //cid -1单发回复
				newJs.SetPath([]string{"sendbuf", "m"}, LIST*1000+SERVER_CALL_REPLY)
				newJs.SetPath([]string{"sendbuf", "p"}, sendJs)
				uDB.IoSession.SendCh <- newJs //放入到发送通道中待发送
			}

			cF.CallTime = 0
			cF.MapCallTimeList = make(map[string]bool)
		}
	}
	ConferenceListRWMutex.Unlock()
}

/**
 * 应答客户端
 * @param iType 消息类型
 * @param pJson 消息内容
 * @param session 当前连接会话连接属性
 * */
func SendToClient(iType int, pJson *sjson.Json, session *TcpSession) {
	sslog.LoggerDebug("SendToClient  SendCh")
	responseJson, _ := sjson.NewJson([]byte(`{}`))
	responseJson.Set("cid", -1) //cid -1单发回复
	responseJson.SetPath([]string{"sendbuf", "m"}, iType)
	responseJson.SetPath([]string{"sendbuf", "p"}, pJson)
	session.SendCh <- responseJson
}

/**
 * 通知与会所有成员
 * */
func BroadCastToClients(cid int, iType int, pJson *sjson.Json, session *TcpSession) {
	sslog.LoggerDebug("BroadCastToClients SendCh ")
	informJson, _ := sjson.NewJson([]byte(`{}`))
	informJson.Set("cid", cid)
	informJson.SetPath([]string{"sendbuf", "m"}, iType)
	informJson.SetPath([]string{"sendbuf", "p"}, pJson)
	session.SendCh <- informJson

}

func GetConference(cid int) *ConferenceDB {
	ConferenceListRWMutex.RLock()
	cF := ConferenceList[cid]
	ConferenceListRWMutex.RUnlock()
	return cF
}

func CheckConference(cid int) bool {
	ConferenceListRWMutex.RLock()
	_, ok := ConferenceList[cid]
	ConferenceListRWMutex.RUnlock()
	return ok
}

func LoginNewConference(uDB *UserDB) {
	ConferenceListRWMutex.Lock()

	uDB.Cid = uDB.Cid * len(ConferenceList)
	uDB.Character = 1

	tmpUserDB := make(map[string]*UserDB)
	tmpUserDB[uDB.Userid] = uDB
	SessionList[uDB.Cid] = tmpUserDB

	cDB := &ConferenceDB{uDB.Cid, 1, true, true, true, true, 0, "", 0, nil, nil, 0, nil, 0, nil}
	ConferenceList[cDB.Cid] = cDB
	ConferenceListRWMutex.Unlock()
}

func GetUserByUid(cid int, uid string) (*UserDB, bool) {
	SessionListRWMutex.RLock()
	defer SessionListRWMutex.RUnlock()

	uDB, ok := SessionList[cid][uid]
	if ok == false {
		return &UserDB{}, ok
	}

	return uDB, ok
}

func InsertUserDB(uDB *UserDB) {
	SessionListRWMutex.Lock()
	SessionList[uDB.Cid][uDB.Userid] = uDB
	SessionListRWMutex.Unlock()
}
