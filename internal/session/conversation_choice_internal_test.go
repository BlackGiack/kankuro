package session

import "testing"

func TestConversationChoiceFindsMatchingDriverAfterSameID(t *testing.T) {
	for _, bindings := range [][]ConversationBinding{
		{
			{ID: "same-id", Name: "Old", Driver: "old-driver"},
			{ID: "same-id", Name: "Current", Driver: "current-driver"},
		},
		{
			{ID: "same-id", Name: "Current", Driver: "current-driver"},
			{ID: "same-id", Name: "Old", Driver: "old-driver"},
		},
	} {
		s := &Session{Conversations: map[string][]ConversationBinding{"card": bindings}}
		got, err := validateConversationChoice(s, "card", "current-driver", ConversationChoice{ID: "same-id"})
		if err != nil || got != "same-id" {
			t.Fatalf("choice with identical IDs in different drivers = %q, %v", got, err)
		}
	}
}
