package dingtalk_enterprise

import (
	"time"

	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
)

// DingtalkWebhookData 钉钉企业应用回调数据
type DingtalkWebhookData struct {
	EventType    string   `json:"EventType"`
	EventTime    int64    `json:"EventTime"`
	EventId      string   `json:"EventId"`
	CorpId       string   `json:"CorpId"`
	UserId       string   `json:"UserId"`
	UserName     string   `json:"UserName"`
	DeptId       []int64  `json:"DeptId"`
	OrgDeptOwner []string `json:"OrgDeptOwner"`
	IsAdmin      bool     `json:"IsAdmin"`
	Level        int      `json:"Level"`
	// 消息相关
	MsgType  string `json:"msgtype"`
	MsgId    string `json:"msgId"`
	Content  string `json:"content"`
	CreateAt int64  `json:"createAt"`
	// 扩展数据
	ExtendData map[string]interface{} `json:"ExtendData,omitempty"`
}

// DingtalkUserInfo 钉钉用户信息
type DingtalkUserInfo struct {
	UserId     string  `json:"userid"`
	Name       string  `json:"name"`
	Mobile     string  `json:"mobile"`
	Department []int64 `json:"department"`
	Position   string  `json:"position"`
	Email      string  `json:"email"`
	Avatar     string  `json:"avatar"`
	JobNumber  string  `json:"jobnumber"`
	UnionId    string  `json:"unionid"`
}

// DingtalkDepartment 钉钉部门信息
type DingtalkDepartment struct {
	DeptId          int64  `json:"dept_id"`
	Name            string `json:"name"`
	ParentId        int64  `json:"parent_id"`
	CreateDeptGroup bool   `json:"createDeptGroup"`
	Order           int64  `json:"order"`
}

// DingtalkWorkNoticeRequest 工作通知请求
type DingtalkWorkNoticeRequest struct {
	AgentId    int64                  `json:"agent_id"`
	UseridList string                 `json:"userid_list,omitempty"`
	DeptIdList string                 `json:"dept_id_list,omitempty"`
	ToAllUser  bool                   `json:"to_all_user,omitempty"`
	Msg        map[string]interface{} `json:"msg"`
}

// DingtalkWorkNoticeResponse 工作通知响应
type DingtalkWorkNoticeResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	TaskId    int64  `json:"task_id"`
	RequestId string `json:"request_id"`
}

// DingtalkGroupMessageRequest 群消息请求
type DingtalkGroupMessageRequest struct {
	ChatId string                 `json:"chatid"`
	Msg    map[string]interface{} `json:"msg"`
}

// DingtalkMediaUploadResponse 媒体文件上传响应
type DingtalkMediaUploadResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	Type    string `json:"type"`
	MediaId string `json:"media_id"`
	Created int64  `json:"created_at"`
}

// ToUnifiedMessage 转换为统一消息格式
func (d *DingtalkWebhookData) ToUnifiedMessage() *dto.UnifiedMessage {
	msg := &dto.UnifiedMessage{
		MessageID:         d.MsgId,
		MessageType:       d.MsgType,
		SenderID:          d.UserId,
		SenderName:        d.UserName,
		SenderType:        entity.SenderTypeUser,
		Content:           d.Content,
		PlatformMessageID: d.MsgId,
		PlatformTimestamp: time.Unix(d.CreateAt/1000, 0),
		RawContent:        make(map[string]interface{}),
	}

	// 处理事件类型
	if d.EventType != "" {
		msg.MessageType = entity.MessageTypeEvent
		msg.EventContent = &entity.EventMessage{
			EventType: mapEventType(d.EventType),
			EventKey:  d.EventType,
			EventData: d.ExtendData,
		}
	}

	// 存储原始数据
	msg.RawContent["corp_id"] = d.CorpId
	msg.RawContent["dept_id"] = d.DeptId
	msg.RawContent["is_admin"] = d.IsAdmin
	msg.RawContent["level"] = d.Level
	msg.RawContent["event_id"] = d.EventId

	return msg
}

// mapEventType 映射钉钉事件类型到统一事件类型
func mapEventType(eventType string) string {
	switch eventType {
	case "user_add_org":
		return entity.EventTypeUserSubscribe
	case "user_leave_org":
		return entity.EventTypeUserUnsubscribe
	case "user_modify_org":
		return entity.EventTypeUserUpdate
	case "org_dept_create":
		return entity.EventTypeGroupCreate
	case "org_dept_modify":
		return entity.EventTypeGroupUpdate
	case "org_dept_remove":
		return entity.EventTypeGroupDelete
	case "check_in":
		return entity.EventTypeCheckIn
	case "bpms_task_change", "bpms_instance_change":
		return entity.EventTypeApproval
	default:
		return entity.EventTypeCustom
	}
}

// DingtalkEventType 钉钉企业应用事件类型
type DingtalkEventType string

const (
	// 组织架构事件
	EventTypeUserAddOrg    DingtalkEventType = "user_add_org"    // 用户加入企业
	EventTypeUserLeaveOrg  DingtalkEventType = "user_leave_org"  // 用户离职
	EventTypeUserModifyOrg DingtalkEventType = "user_modify_org" // 用户信息变更
	EventTypeOrgDeptCreate DingtalkEventType = "org_dept_create" // 部门创建
	EventTypeOrgDeptModify DingtalkEventType = "org_dept_modify" // 部门修改
	EventTypeOrgDeptRemove DingtalkEventType = "org_dept_remove" // 部门删除

	// 应用事件
	EventTypeAppUserAdd    DingtalkEventType = "app_user_add"    // 用户激活应用
	EventTypeAppUserRemove DingtalkEventType = "app_user_remove" // 用户停用应用

	// 审批事件
	EventTypeBpmsTaskChange     DingtalkEventType = "bpms_task_change"     // 审批任务变更
	EventTypeBpmsInstanceChange DingtalkEventType = "bpms_instance_change" // 审批实例变更

	// 考勤事件
	EventTypeCheckIn         DingtalkEventType = "check_in"         // 员工打卡
	EventTypeAttendanceCheck DingtalkEventType = "attendance_check" // 考勤结果

	// 日程事件
	EventTypeCalendarEvent DingtalkEventType = "calendar_event" // 日程事件
)

// DingtalkMessageMode 钉钉消息发送模式
type DingtalkMessageMode string

const (
	MessageModeWorkNotice DingtalkMessageMode = "work_notice" // 工作通知
	MessageModeGroupChat  DingtalkMessageMode = "group_chat"  // 群消息
	MessageModeDing       DingtalkMessageMode = "ding"        // DING消息
)

// DingtalkEnterpriseConfig 钉钉企业应用配置
type DingtalkEnterpriseConfig struct {
	AppKey       string `json:"app_key" validate:"required"`
	AppSecret    string `json:"app_secret" validate:"required"`
	AgentId      string `json:"agent_id" validate:"required"`
	CorpId       string `json:"corp_id"`
	MessageMode  string `json:"message_mode"` // 默认消息发送模式
	WebhookToken string `json:"webhook_token"`
	AESKey       string `json:"aes_key"`
}

// ParseEnterpriseConfig 解析企业应用配置
func ParseEnterpriseConfig(config map[string]interface{}) (*DingtalkEnterpriseConfig, error) {
	cfg := &DingtalkEnterpriseConfig{}

	if v, ok := config["app_key"].(string); ok {
		cfg.AppKey = v
	}
	if v, ok := config["app_secret"].(string); ok {
		cfg.AppSecret = v
	}
	if v, ok := config["agent_id"].(string); ok {
		cfg.AgentId = v
	}
	if v, ok := config["corp_id"].(string); ok {
		cfg.CorpId = v
	}
	if v, ok := config["message_mode"].(string); ok {
		cfg.MessageMode = v
	} else {
		cfg.MessageMode = string(MessageModeWorkNotice)
	}
	if v, ok := config["webhook_token"].(string); ok {
		cfg.WebhookToken = v
	}
	if v, ok := config["aes_key"].(string); ok {
		cfg.AESKey = v
	}

	return cfg, nil
}
