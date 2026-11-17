//nolint:testpackage // Tests exercise internal storage and inspection boundaries.
package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecentAuthoredReports(t *testing.T) {
	now := time.Now()
	store := &MessageStore{data: messageStoreData{Messages: []MessageRecord{
		{
			DashboardID:     "dash",
			SenderSessionID: "worker",
			Replies:         []MessageReply{{ID: "received", SenderSessionID: "other", CreatedAt: now}},
		},
		{DashboardID: "foreign", Replies: []MessageReply{{ID: "foreign", SenderSessionID: "worker", CreatedAt: now}}},
		{DashboardID: "dash", Replies: []MessageReply{
			{ID: "old", SenderSessionID: "worker", SenderActivationID: "first", CreatedAt: now.Add(-time.Second), Sequence: 9},
			{
				ID:                 "b",
				SenderSessionID:    "worker",
				SenderActivationID: "second",
				CreatedAt:          now,
				Sequence:           10,
				Kind:               "progress",
				Notification:       &ReplyNotification{},
			},
			{ID: "a", SenderSessionID: "worker", CreatedAt: now, Sequence: 10},
		}},
	}}}
	got, err := store.RecentAuthoredReports(context.Background(), "dash", "worker", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 2 || got.Items[0].ID != "a" || got.Items[1].ID != "b" || got.Omitted != 1 {
		t.Fatalf("unexpected projection: %#v", got)
	}
	if got.Items[1].Notification != nil {
		t.Fatal("notification leaked into projection")
	}
	got.Items[0].Text = "changed"
	if store.data.Messages[2].Replies[2].Text != "" {
		t.Fatal("projection aliases store")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, queryErr := store.RecentAuthoredReports(ctx, "dash", "worker", 2); !errors.Is(queryErr, context.Canceled) {
		t.Fatalf("canceled query: %v", queryErr)
	}
	if _, queryErr := store.RecentAuthoredReports(context.Background(), "dash", "worker", 0); queryErr == nil {
		t.Fatal("accepted invalid limit")
	}
}

func TestRecentAuthoredReportsBoundsAndReload(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "messaging")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	record := MessageRecord{ID: "task", DashboardID: "dashboard", Replies: make([]MessageReply, 0, MaxAuthoredReports+1)}
	for i := range MaxAuthoredReports + 1 {
		record.Replies = append(
			record.Replies,
			MessageReply{
				ID:                 fmt.Sprintf("%03d", i),
				MessageID:          "task",
				SenderSessionID:    "author",
				SenderActivationID: fmt.Sprintf("activation-%d", i%2),
				Text:               "text",
				Kind:               "progress",
				CreatedAt:          now,
				Sequence:           uint64(i),
			},
		)
	}
	data, err := json.Marshal(messageStoreData{Messages: []MessageRecord{record}})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "history.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	store, err := NewMessageStore(root)
	if err != nil {
		t.Fatal(err)
	}
	page, err := store.RecentAuthoredReports(context.Background(), "dashboard", "author", MaxAuthoredReports+50)
	if err != nil || len(page.Items) != MaxAuthoredReports || page.Omitted != 1 || page.Items[0].Sequence != uint64(MaxAuthoredReports) {
		t.Fatalf("reloaded projection: %d omitted=%d err=%v", len(page.Items), page.Omitted, err)
	}
	if page.Items[0].SenderActivationID == page.Items[1].SenderActivationID {
		t.Fatal("expected reports across activations")
	}
	if len(store.Notifications()) != 0 {
		t.Fatal("projection created notifications")
	}
}
