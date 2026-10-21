package wechat_official

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/stretchr/testify/require"
)

const testOfficialAccountToUser = "gh_bot"

func TestParseInboundMessage_LocationSetsLocationContent(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(WechatOfficialMessage{
		MsgType:      "location",
		MsgID:        42,
		FromUserName: "openid-1",
		ToUserName:   testOfficialAccountToUser,
		CreateTime:   1700000000,
		LocationX:    "31.2304",
		LocationY:    "121.4737",
		Label:        "Shanghai",
	})
	require.NoError(t, err)

	msg, err := NewWechatOfficialAdapter().ParseInboundMessage(context.Background(), body, nil)
	require.NoError(t, err)
	require.Equal(t, entity.MessageTypeLocation, msg.MessageType)
	require.NotNil(t, msg.LocationContent)
	require.InDelta(t, 31.2304, msg.LocationContent.Latitude, 0.0001)
	require.InDelta(t, 121.4737, msg.LocationContent.Longitude, 0.0001)
	require.Equal(t, "Shanghai", msg.LocationContent.Address)

	unified := msg.BuildUnifiedContent()
	require.Contains(t, unified, "location")
	require.NotContains(t, unified, "original_message")
}

func TestParseInboundMessage_EventSetsEventContent(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(WechatOfficialMessage{
		MsgType:      "event",
		MsgID:        43,
		FromUserName: "openid-2",
		ToUserName:   testOfficialAccountToUser,
		CreateTime:   1700000001,
		Event:        "subscribe",
		EventKey:     "qrscene_1",
	})
	require.NoError(t, err)

	msg, err := NewWechatOfficialAdapter().ParseInboundMessage(context.Background(), body, nil)
	require.NoError(t, err)
	require.Equal(t, entity.MessageTypeEvent, msg.MessageType)
	require.NotNil(t, msg.EventContent)
	require.Equal(t, entity.EventTypeUserSubscribe, msg.EventContent.EventType)
	require.Equal(t, "qrscene_1", msg.EventContent.EventKey)

	unified := msg.BuildUnifiedContent()
	require.Contains(t, unified, "event")
	_, hasRawDump := unified["original_message"]
	require.False(t, hasRawDump)
}
