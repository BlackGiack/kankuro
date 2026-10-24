package session_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

func TestNaturalRecipientExitWakesMessagingWaiters(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	if err = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("list workspaces: %v", err)
	}
	manager := newTestManager(t, repo.Dir(), repo)
	manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	card, err := repo.CreateIDCard(config.IDCard{ID: "natural-exit", Name: "natural exit", Driver: "cat",
		WorkspaceID: workspaces[0].ID, MCPIDs: []string{config.BuiltinMCPKankuro}})
	if err != nil {
		t.Fatal(err)
	}
	sessionID, err := manager.Launch(card, nil, nil, config.BuiltinMCPs())
	if err != nil {
		t.Fatal(err)
	}
	snapshot, found := manager.GetSession(sessionID)
	if !found || snapshot.PID <= 0 || snapshot.ActivationID == "" {
		t.Fatalf("recipient snapshot = %#v, found=%v", snapshot, found)
	}
	messageID := uuid.NewString()
	if _, err = manager.Messages().CreateIntent(session.MessageRecord{ID: messageID, DashboardID: "default",
		SenderSessionID: "sender", RecipientSessionID: sessionID, RecipientActivationID: snapshot.ActivationID, Text: "task"},
		uuid.NewString(), session.SendFingerprint(sessionID, snapshot.ActivationID, "task")); err != nil {
		t.Fatal(err)
	}
	changed := manager.Messages().Changed()
	process, err := os.FindProcess(snapshot.PID)
	if err != nil {
		t.Fatal(err)
	}
	if err = process.Kill(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-changed:
	case <-time.After(5 * time.Second):
		t.Fatal("natural process termination did not wake messaging waiters")
	}
	live, err := manager.MessageRecipientLive(messageID)
	if err != nil || live {
		t.Fatalf("recipient live=%v, err=%v after natural exit", live, err)
	}
}

//nolint:gocognit // Stop and shutdown must each release a pending manager reservation.
func TestPendingMessageReservationDoesNotBlockStopOrShutdown(t *testing.T) {
	for _, lifecycle := range []string{"stop", "shutdown"} {
		t.Run(lifecycle, func(t *testing.T) {
			repo, err := config.NewFileRepository(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			workspace := t.TempDir()
			if err = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
				t.Fatal(err)
			}
			workspaces, err := repo.ListWorkspaces()
			if err != nil || len(workspaces) != 1 {
				t.Fatalf("list workspaces: %v", err)
			}
			manager := newTestManager(t, repo.Dir(), repo)
			manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
			controller, err := repo.CreateIDCard(config.IDCard{ID: "admission-controller", Name: "controller", Driver: "cat",
				WorkspaceID: workspaces[0].ID, DashboardControl: true, MCPIDs: []string{config.BuiltinMCPKankuro}})
			if err != nil {
				t.Fatal(err)
			}
			recipient, err := repo.CreateIDCard(config.IDCard{ID: "admission-recipient", Name: "recipient", Driver: "cat",
				WorkspaceID: workspaces[0].ID, MCPIDs: []string{config.BuiltinMCPKankuro}})
			if err != nil {
				t.Fatal(err)
			}
			controllerID, err := manager.Launch(controller, nil, nil, config.BuiltinMCPs())
			if err != nil {
				t.Fatal(err)
			}
			controllerWorkspace := filepath.Join(workspace, ".kankuro", "mcp-servers.json")
			token := readRelaunchToken(t, controllerWorkspace)
			recipientID, err := manager.Launch(recipient, nil, nil, config.BuiltinMCPs())
			if err != nil {
				t.Fatal(err)
			}
			recipientSession, found := manager.GetSession(recipientID)
			if !found {
				t.Fatal("recipient was not registered")
			}
			admission, err := manager.ReserveMessageAdmission(token, recipientID, recipientSession.ActivationID)
			if err != nil {
				t.Fatal(err)
			}
			messageID := uuid.NewString()
			if _, err = manager.Messages().CreateIntent(session.MessageRecord{ID: messageID, DashboardID: "default",
				SenderSessionID: controllerID, SenderActivationID: token, RecipientSessionID: recipientID,
				RecipientActivationID: recipientSession.ActivationID, Text: "queued"}, uuid.NewString(),
				session.SendFingerprint(recipientID, recipientSession.ActivationID, "queued")); err != nil {
				t.Fatal(err)
			}

			completed := make(chan struct{})
			go func() {
				if lifecycle == "stop" {
					_ = manager.Kill(recipientID)
				} else {
					manager.Shutdown()
				}
				close(completed)
			}()
			select {
			case <-completed:
			case <-time.After(5 * time.Second):
				t.Fatal("lifecycle operation waited on an unstarted message reservation")
			}
			if lifecycle == "stop" {
				defer admission.Release()
				_, submitErr := manager.SubmitMessage(context.Background(), admission, token, messageID,
					recipientID, recipientSession.ActivationID, "queued")
				if submitErr == nil {
					t.Fatal("invalidated queued message submission unexpectedly succeeded")
				}
				got, getErr := manager.Messages().GetMessage(messageID)
				if getErr != nil || got.Delivery != session.MessageNotSubmitted {
					t.Fatalf("queued message after stop = %#v, %v", got, getErr)
				}
			}
		})
	}
}
