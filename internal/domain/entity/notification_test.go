package entity_test

import (
	"testing"

	"github.com/Doratou/Backend/internal/domain/broker/payload"
	"github.com/Doratou/Backend/internal/domain/entity"
)

func TestNotification_SetAndUnmarshalData(t *testing.T) {
	notif := &entity.Notification{
		RecipientID: 10,
		Event:       payload.WSEventFriendRequestReceived,
	}

	dataPayload := payload.FriendRequestReceivedData{
		RequestID:  99,
		SenderID:   "user-12",
		SenderName: "Ali",
	}

	if err := notif.SetData(dataPayload); err != nil {
		t.Fatalf("failed to set data: %v", err)
	}

	if len(notif.Data) == 0 {
		t.Fatal("expected non-empty data json")
	}

	var parsed payload.FriendRequestReceivedData
	if err := notif.UnmarshalData(&parsed); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if parsed.RequestID != 99 || parsed.SenderName != "Ali" {
		t.Fatalf("unexpected unmarshaled data: %+v", parsed)
	}
}

func TestNotification_SetData_Nil(t *testing.T) {
	notif := &entity.Notification{}
	if err := notif.SetData(nil); err != nil {
		t.Fatalf("unexpected error setting nil data: %v", err)
	}
	if string(notif.Data) != "{}" {
		t.Fatalf("expected empty json object '{}', got %s", string(notif.Data))
	}
}
