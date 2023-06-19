package interaction

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/cmd/Open-IM-SDK-Core/sdk_struct"
	"Open_IM/pkg/common/log"
	sdk "Open_IM/pkg/proto/sdk_ws"
	"errors"
	"sync"

	"github.com/golang/protobuf/proto"
)

type ReadDiffusionGroupMsgSync struct {
	*db.DataBase
	*Ws
	loginUserID              string
	conversationCh           chan common.Cmd2Value
	superGroupMtx            sync.Mutex
	Group2SeqMaxNeedSync     map[string]uint32
	Group2SeqMaxSynchronized map[string]uint32
	SuperGroupIDList         []string
	joinedSuperGroupCh       chan common.Cmd2Value
	Group2SyncMsgFinished    map[string]bool
}

func NewReadDiffusionGroupMsgSync(dataBase *db.DataBase, ws *Ws, loginUserID string, conversationCh chan common.Cmd2Value, joinedSuperGroupCh chan common.Cmd2Value) *ReadDiffusionGroupMsgSync {
	p := &ReadDiffusionGroupMsgSync{DataBase: dataBase, Ws: ws, loginUserID: loginUserID, conversationCh: conversationCh, joinedSuperGroupCh: joinedSuperGroupCh}
	p.Group2SeqMaxNeedSync = make(map[string]uint32, 0)
	p.Group2SeqMaxSynchronized = make(map[string]uint32, 0)
	p.Group2SyncMsgFinished = make(map[string]bool, 0)
	// go p.updateJoinedSuperGroup()
	return p
}

// 协程方式加锁获取读扩散群列表 ok
func (m *ReadDiffusionGroupMsgSync) updateJoinedSuperGroup() {
	for {
		select {
		case cmd := <-m.joinedSuperGroupCh:
			operationID := cmd.Value.(sdk_struct.CmdJoinedSuperGroup).OperationID
			log.Info(operationID, "updateJoinedSuperGroup cmd: ", cmd)
			g, err := m.GetJoinedGroupIDList()
			if err == nil {
				log.Info(operationID, "GetReadDiffusionGroupIDList, group id list: ", g)
				m.superGroupMtx.Lock()
				m.SuperGroupIDList = g
				m.superGroupMtx.Unlock()
				m.compareSeq(operationID)
			} else {
				log.Error(operationID, "GetReadDiffusionGroupIDList failed ", err.Error())
			}
		}
	}
}

// 读取所有的读扩散群id，并加载seq到map中，初始化调用一次， 群列表变化时调用一次  ok
// this is for real time msg compare, invited group sync not checked here
func (m *ReadDiffusionGroupMsgSync) compareSeq(operationID string) {
	g, err := m.GetJoinedGroupIDList()
	log.Debug(operationID, "compareSeq GetJoinedGroupIDList ", g)

	if err != nil {
		log.Error(operationID, "GetReadDiffusionGroupIDList failed ", err.Error())
		return
	}
	m.superGroupMtx.Lock()
	m.SuperGroupIDList = m.SuperGroupIDList[0:0]
	m.SuperGroupIDList = g
	m.superGroupMtx.Unlock()
	log.Debug(operationID, "compareSeq load groupID list ", m.SuperGroupIDList)

	m.superGroupMtx.Lock()

	defer m.superGroupMtx.Unlock()
	for _, v := range m.SuperGroupIDList {

		n, err := m.GetSuperGroupNormalMsgSeq(v)
		if err != nil {
			log.Error(operationID, "GetSuperGroupNormalMsgSeq failed ", err.Error(), v)
		}
		a, err := m.GetSuperGroupAbnormalMsgSeq(v)
		if err != nil {
			log.Error(operationID, "GetSuperGroupAbnormalMsgSeq failed ", err.Error(), v)
		}
		log.Debug(operationID, "GetSuperGroupNormalMsgSeq GetSuperGroupAbnormalMsgSeq ", n, a)
		var seqMaxSynchronized uint32
		if n > a {
			seqMaxSynchronized = n
		} else {
			seqMaxSynchronized = a
		}
		if seqMaxSynchronized > m.Group2SeqMaxNeedSync[v] {
			m.Group2SeqMaxNeedSync[v] = seqMaxSynchronized
		}
		if seqMaxSynchronized > m.Group2SeqMaxSynchronized[v] {
			m.Group2SeqMaxSynchronized[v] = seqMaxSynchronized
		}
		//log.NewError(operationID, "load seq, normal, abnormal, ", n, a, m.Group2SeqMaxNeedSync[v], m.Group2SeqMaxSynchronized[v], v)
	}
}

// 处理最大seq消息
func (m *ReadDiffusionGroupMsgSync) doMaxSeq(cmd common.Cmd2Value) {
	operationID := cmd.Value.(sdk_struct.CmdMaxSeqToMsgSync).OperationID
	//同步最新消息，内部保证只调用一次

	//更新需要同步的最大seq
	for groupID, MinMaxSeqOnSvr := range cmd.Value.(sdk_struct.CmdMaxSeqToMsgSync).GroupID2MinMaxSeqOnSvr {

		if MinMaxSeqOnSvr.MinSeq > MinMaxSeqOnSvr.MaxSeq {
			log.Warn(operationID, "MinMaxSeqOnSvr.MinSeq > MinMaxSeqOnSvr.MaxSeq", MinMaxSeqOnSvr.MinSeq, MinMaxSeqOnSvr.MaxSeq)
			return
		}
		if MinMaxSeqOnSvr.MaxSeq > m.Group2SeqMaxNeedSync[groupID] {
			m.Group2SeqMaxNeedSync[groupID] = MinMaxSeqOnSvr.MaxSeq
		}
		//logString := "SandmanO Sync User ID" + m.loginUserID + "GroupID " + groupID + " Min Seq Loal " + fmt.Sprint(m.Group2SeqMaxSynchronized[groupID]) + " Min Sync " + fmt.Sprint(MinMaxSeqOnSvr.MinSeq)
		//log.NewError(logString)
		// if m.Group2SeqMaxSynchronized[groupID] == 1 {
		// 	m.Group2SeqMaxSynchronized[groupID] = MinMaxSeqOnSvr.MinSeq - 1
		// } else
		if MinMaxSeqOnSvr.MinSeq > m.Group2SeqMaxSynchronized[groupID] {
			m.Group2SeqMaxSynchronized[groupID] = MinMaxSeqOnSvr.MinSeq - 1
		}
		//logString = "SandmanO Sync User ID" + m.loginUserID + "GroupID " + groupID + " Min Seq Loal " + fmt.Sprint(m.Group2SeqMaxSynchronized[groupID]) + " Min Sync " + fmt.Sprint(MinMaxSeqOnSvr.MinSeq)
		//log.NewError(logString)
	}

	m.syncLatestMsg(operationID)
	//同步所有群的新消息
	m.syncMsgFroAllGroup(operationID)
	//log.NewError(operationID, utils.GetSelfFuncName(), "end!!")
}

// 在获取最大seq后同步最新消息，只调用一次 ok
func (m *ReadDiffusionGroupMsgSync) syncLatestMsg(operationID string) {
	m.superGroupMtx.Lock()
	flag := 0
	for _, v := range m.SuperGroupIDList {
		if m.Group2SyncMsgFinished[v] == false {
			flag = 1
			break
		}
	}

	if flag == 1 {
		log.Info(operationID, "sync latest msg begin for read diffusion group: ", m.SuperGroupIDList)
		m.TriggerCmdNewMsgCome(nil, operationID, constant.MsgSyncBegin)
		for _, v := range m.SuperGroupIDList {
			m.syncLatestMsgForGroup(v, operationID, constant.MsgSyncProcessing)
		}
		m.TriggerCmdNewMsgCome(nil, operationID, constant.MsgSyncEnd)
		//log.NewError(operationID, "sync latest msg end for read diffusion group: ", m.SuperGroupIDList)
	} else {
		//log.NewError(operationID, "do nothing ")
		for _, v := range m.SuperGroupIDList {
			m.syncLatestMsgForGroup(v, operationID, 0)
		}
	}

	//end
	m.superGroupMtx.Unlock()
}

// 获取某个群的最新消息，只调用一次
func (m *ReadDiffusionGroupMsgSync) syncLatestMsgForGroup(groupID, operationID string, loginSync int) {
	log.Debug(operationID, utils.GetSelfFuncName(), "syncLatestMsgForGroup start", groupID, loginSync, m.Group2SyncMsgFinished[groupID], m.Group2SeqMaxNeedSync[groupID], m.Group2SeqMaxSynchronized[groupID])
	if !m.Group2SyncMsgFinished[groupID] {
		need := m.Group2SeqMaxNeedSync[groupID]
		synchronized := m.Group2SeqMaxSynchronized[groupID]
		begin := synchronized + 1
		//if int64(need)-int64(synchronized) > int64(constant.PullMsgNumForReadDiffusion) {
		//	begin = need - uint32(constant.PullMsgNumForReadDiffusion) + 1
		//}
		//log.NewError(operationID, "syncLatestMsgForGroup seq: ", need, synchronized, begin, groupID)
		if begin > need {
			//log.Debug(operationID, "do nothing syncLatestMsgForGroup seq: ", need, synchronized, begin)
			return
		}
		m.syncMsgFromServer(begin, need, groupID, operationID, loginSync)
		m.Group2SyncMsgFinished[groupID] = true
		m.Group2SeqMaxSynchronized[groupID] = begin
	}
}

func (m *ReadDiffusionGroupMsgSync) doPushMsg(cmd common.Cmd2Value) {
	msg := cmd.Value.(sdk_struct.CmdPushMsgToMsgSync).Msg
	operationID := cmd.Value.(sdk_struct.CmdPushMsgToMsgSync).OperationID
	log.Debug(operationID, "recv super group push msg, doPushMsg ", msg.Seq, msg.ServerMsgID, msg.ClientMsgID, msg.GroupID, msg.SessionType, m.Group2SeqMaxSynchronized[msg.GroupID], m.Group2SeqMaxNeedSync[msg.GroupID], m.loginUserID)
	if msg.Seq == 0 {
		m.TriggerCmdNewMsgCome([]*sdk.MsgData{msg}, operationID, 0)
		return
	}
	if msg.Seq == m.Group2SeqMaxSynchronized[msg.GroupID]+1 {
		log.Debug(operationID, "TriggerCmdNewMsgCome ", msg.ServerMsgID, msg.ClientMsgID, msg.Seq)
		m.TriggerCmdNewMsgCome([]*sdk.MsgData{msg}, operationID, 0)
		m.Group2SeqMaxSynchronized[msg.GroupID] = msg.Seq
	}

	if msg.Seq > m.Group2SeqMaxNeedSync[msg.GroupID] {
		m.Group2SeqMaxNeedSync[msg.GroupID] = msg.Seq
	}

	/*
	* Below line of code is for fetch MinSeq from redis, So we sync messages from that MinSeq to MaxSeq
	* This works only When user is online and some one add that user in pre-exsisting converstation.
	* If user is offline MinSeq set by Heartbeat (which also fetch MinSeq from redis, same point)
	 */
	if m.Group2SeqMaxSynchronized[msg.GroupID] == 0 || msg.ContentType == constant.MemberInvitedNotification {
		groupIDList := make([]string, 0)
		groupIDList = append(groupIDList, msg.GroupID)
		localSDKGroupMaxSeq := make(map[string]uint32)
		localSDKGroupMaxSeq[msg.GroupID] = 0
		resp, err := m.SendReqWaitResp(&sdk.GetMaxAndMinSeqReq{UserID: m.loginUserID, GroupIDList: groupIDList, SdkMaxSeq: uint32(0), GroupSdkMaxSeq: localSDKGroupMaxSeq}, constant.WSGetNewestSeq, 200, 1, m.loginUserID, operationID)
		if err == nil && resp != nil {
			var wsSeqResp sdk.GetMaxAndMinSeqResp
			err = proto.Unmarshal(resp.Data, &wsSeqResp)
			if err == nil && wsSeqResp.GroupMaxAndMinSeq != nil {
				m.Group2SeqMaxSynchronized[msg.GroupID] = wsSeqResp.GroupMaxAndMinSeq[msg.GroupID].MinSeq
			} else {
				m.Group2SeqMaxSynchronized[msg.GroupID] = 0
			}

		} else if err != nil {
			if !errors.Is(err, constant.WsRecvConnSame) && !errors.Is(err, constant.WsRecvConnDiff) {
				log.Error(operationID, "other err,  close conn", err.Error())
				//u.CloseConn()
			}
		}

	}
	//else {
	//	localseq := m.DataBase.GetGroupLocalMaxSeq(msg.GroupID)
	//	log.Error(operationID, "Local Seq", localseq)
	//}
	//log.NewError(operationID, "syncMsgFromServer ", m.Group2SeqMaxSynchronized[msg.GroupID]+1, m.Group2SeqMaxNeedSync[msg.GroupID])
	//获取此群最新消息，内部保证只调用一次
	m.syncLatestMsgForGroup(msg.GroupID, operationID, 0)
	//同步此群新消息
	m.syncMsgForOneGroup(operationID, msg.GroupID)
}

// 同步所有群的新消息 ok
func (m *ReadDiffusionGroupMsgSync) syncMsgFroAllGroup(operationID string) {
	m.superGroupMtx.Lock()
	//log.NewError(operationID, "syncMsgFroAllGroup groupid list ", m.SuperGroupIDList)
	for _, v := range m.SuperGroupIDList {
		if !m.Group2SyncMsgFinished[v] {
			continue
		}
		seqMaxNeedSync := m.Group2SeqMaxNeedSync[v]
		seqMaxSynchronized := m.Group2SeqMaxSynchronized[v]
		//log.NewError(operationID, "do syncMsgFromServer ", seqMaxSynchronized, seqMaxNeedSync, v)
		if seqMaxNeedSync > seqMaxSynchronized {
			//log.NewError(operationID, "do syncMsgFromServer ", seqMaxSynchronized+1, seqMaxNeedSync, v)
			m.syncMsgFromServer(seqMaxSynchronized+1, seqMaxNeedSync, v, operationID, 0)
			m.Group2SeqMaxSynchronized[v] = seqMaxNeedSync
		} else {
			//log.NewError(operationID, "do nothing ", seqMaxSynchronized+1, seqMaxNeedSync, v)
		}
	}
	m.superGroupMtx.Unlock()
}

// 同步某个群的新消息 ok
func (m *ReadDiffusionGroupMsgSync) syncMsgForOneGroup(operationID string, groupID string) {
	log.Debug(operationID, utils.GetSelfFuncName(), "syncMsgForOneGroup start", groupID)
	m.superGroupMtx.Lock()
	for _, v := range m.SuperGroupIDList {
		if groupID != "" && v != groupID {
			continue
		}
		seqMaxNeedSync := m.Group2SeqMaxNeedSync[v]
		seqMaxSynchronized := m.Group2SeqMaxSynchronized[v]
		if seqMaxNeedSync > seqMaxSynchronized {
			//log.NewError(operationID, "do syncMsg ", seqMaxSynchronized+1, seqMaxNeedSync)
			m.syncMsgFromServer(seqMaxSynchronized+1, seqMaxNeedSync, v, operationID, 0)
			m.Group2SeqMaxSynchronized[v] = seqMaxNeedSync
		} else {
			//log.NewError(operationID, "msg not sync", seqMaxNeedSync, seqMaxSynchronized)
		}
		break
	}
	m.superGroupMtx.Unlock()
	//log.NewError(operationID, utils.GetSelfFuncName(), "syncMsgForOneGroup end", groupID)
}

func (m *ReadDiffusionGroupMsgSync) syncMsgFromServer(beginSeq, endSeq uint32, groupID, operationID string, loginSync int) {
	log.Debug(operationID, utils.GetSelfFuncName(), "args: ", beginSeq, endSeq, groupID)
	if beginSeq > endSeq {
		log.Debug(operationID, "beginSeq > endSeq", beginSeq, endSeq)
		return
	}

	var needSyncSeqList []uint32
	for i := beginSeq; i <= endSeq; i++ {
		needSyncSeqList = append(needSyncSeqList, i)
	}
	var SPLIT = constant.SplitPullMsgNum
	for i := 0; i < len(needSyncSeqList)/SPLIT; i++ {
		m.syncMsgFromServerSplit(needSyncSeqList[i*SPLIT:(i+1)*SPLIT], groupID, operationID, loginSync)
	}
	m.syncMsgFromServerSplit(needSyncSeqList[SPLIT*(len(needSyncSeqList)/SPLIT):], groupID, operationID, loginSync)
}

func (m *ReadDiffusionGroupMsgSync) syncMsgFromServerSplit(needSyncSeqList []uint32, groupID, operationID string, loginSync int) {
	var pullMsgReq sdk.PullMessageBySeqListReq
	pullMsgReq.UserID = m.loginUserID
	pullMsgReq.GroupSeqList = make(map[string]*sdk.SeqList, 0)
	pullMsgReq.GroupSeqList[groupID] = &sdk.SeqList{SeqList: needSyncSeqList}

	for {
		pullMsgReq.OperationID = operationID
		log.Debug(operationID, "read diffusion group pull message, req: ", pullMsgReq.String())
		resp, err := m.SendReqWaitResp(&pullMsgReq, constant.WSPullMsgBySeqList, 30, 2, m.loginUserID, operationID)
		if err != nil {
			log.Error(operationID, "SendReqWaitResp failed ", err.Error(), constant.WSPullMsgBySeqList, 30, 2, m.loginUserID)
			continue
		}

		var pullMsgResp sdk.PullMessageBySeqListResp
		err = proto.Unmarshal(resp.Data, &pullMsgResp)
		if err != nil {
			log.Error(operationID, "pullMsgResp Unmarshal failed ", err.Error())
			return
		}
		log.Debug(operationID, "syncMsgFromServerSplit pull msg ", pullMsgReq.String(), pullMsgResp.String())
		for _, v := range pullMsgResp.GroupMsgDataList {
			log.Debug(operationID, utils.GetSelfFuncName(), "TriggerCmdNewMsgCome ", len(v.MsgDataList))
			m.TriggerCmdNewMsgCome(v.MsgDataList, operationID, loginSync)
		}
		return
	}
}

func (m *ReadDiffusionGroupMsgSync) TriggerCmdNewMsgCome(msgList []*sdk.MsgData, operationID string, loginSync int) {
	for {
		err := common.TriggerCmdSuperGroupMsgCome(sdk_struct.CmdNewMsgComeToConversation{MsgList: msgList, OperationID: operationID, SyncFlag: loginSync}, m.conversationCh)
		if err != nil {
			log.Warn(operationID, "TriggerCmdSuperGroupMsgCome failed ", err.Error(), m.loginUserID)
			continue
		}
		log.Info(operationID, "TriggerCmdSuperGroupMsgCome ok ", m.loginUserID)
		return
	}
}
