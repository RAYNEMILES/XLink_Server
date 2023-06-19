package interaction

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db"
	"Open_IM/pkg/common/log"

	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/cmd/Open-IM-SDK-Core/sdk_struct"
)

type SeqPair struct {
	BeginSeq uint32
	EndSeq   uint32
}

type MsgSync struct {
	*db.DataBase
	*Ws
	LoginUserID        string
	conversationCh     chan common.Cmd2Value
	PushMsgAndMaxSeqCh chan common.Cmd2Value

	SelfMsgSync *SelfMsgSync
	//selfMsgSyncLatestModel *SelfMsgSyncLatestModel
	//superGroupMsgSync *SuperGroupMsgSync

	ReadDiffusionGroupMsgSync *ReadDiffusionGroupMsgSync
}

// compareSeq comapre SQ number for Private Messages and Group Messages.
// New Groups sync from some else point for first time. After 1#st sync new messages in group sync from here
func (m *MsgSync) compareSeq() {
	operationID := utils.OperationIDGenerator()
	m.SelfMsgSync.compareSeq(operationID)
	m.ReadDiffusionGroupMsgSync.compareSeq(operationID)
}

// New Groups sync from some else point for first time. After 1#st sync new messages in group sync from here
func (m *MsgSync) CompareGroupSeq() {
	operationID := utils.OperationIDGenerator()
	m.ReadDiffusionGroupMsgSync.compareSeq(operationID)
}

func (m *MsgSync) doMaxSeq(cmd common.Cmd2Value) {
	m.ReadDiffusionGroupMsgSync.doMaxSeq(cmd)
	m.SelfMsgSync.doMaxSeq(cmd)
}

func (m *MsgSync) doPushMsg(cmd common.Cmd2Value) {
	msg := cmd.Value.(sdk_struct.CmdPushMsgToMsgSync).Msg
	switch msg.SessionType {
	case constant.GroupChatType:
		m.ReadDiffusionGroupMsgSync.doPushMsg(cmd)
	default:
		m.SelfMsgSync.doPushMsg(cmd)
	}
}

func (m *MsgSync) Work(cmd common.Cmd2Value) {
	switch cmd.Cmd {
	case constant.CmdPushMsg:
		m.doPushMsg(cmd)
	case constant.CmdMaxSeq:
		m.doMaxSeq(cmd)
	default:
		log.Error("", "cmd failed ", cmd.Cmd)
	}
}

func (m *MsgSync) GetCh() chan common.Cmd2Value {
	return m.PushMsgAndMaxSeqCh
}

func NewMsgSync(dataBase *db.DataBase, ws *Ws, loginUserID string, ch chan common.Cmd2Value, pushMsgAndMaxSeqCh chan common.Cmd2Value, joinedSuperGroupCh chan common.Cmd2Value) *MsgSync {
	p := &MsgSync{DataBase: dataBase, Ws: ws, LoginUserID: loginUserID, conversationCh: ch, PushMsgAndMaxSeqCh: pushMsgAndMaxSeqCh}
	//	p.superGroupMsgSync = NewSuperGroupMsgSync(dataBase, ws, loginUserID, ch, joinedSuperGroupCh)
	p.SelfMsgSync = NewSelfMsgSync(dataBase, ws, loginUserID, ch)
	p.ReadDiffusionGroupMsgSync = NewReadDiffusionGroupMsgSync(dataBase, ws, loginUserID, ch, joinedSuperGroupCh)
	//	p.selfMsgSync = NewSelfMsgSyncLatestModel(dataBase, ws, loginUserID, ch)
	p.compareSeq()
	go common.DoListener(p)
	return p
}
