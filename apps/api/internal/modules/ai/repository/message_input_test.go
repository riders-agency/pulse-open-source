package repository

import (
	"encoding/json"
	"testing"
)

func TestMessageTimestampCannotBeSuppliedByClient(t *testing.T) {
	var input MessageInput
	if err := json.Unmarshal([]byte(`{"body":"Да, создавай","ReceivedAt":"2000-01-01T00:00:00Z","received_at":"2000-01-01T00:00:00Z"}`), &input); err != nil {
		t.Fatal(err)
	}
	if !input.ReceivedAt.IsZero() {
		t.Fatal("client was able to override the server receipt time")
	}
}
