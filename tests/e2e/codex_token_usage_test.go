//go:build e2e && token_usage

package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	tuitest "github.com/microsoft/tui-test/bindings/go"
	"github.com/microsoft/tui-test/bindings/go/tuitesttest"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/configstartup"
	"github.com/BlackGiack/kankuro/internal/session"
)

const (
	lunaModel               = "gpt-6-luna"
	lunaInputLimit          = 60000
	lunaOutputLimit         = 2000
	lunaRecallInputReserve  = 16000
	lunaRecallOutputReserve = 400
)

func TestNativeAgentJourney(t *testing.T) {
	if os.Getenv("KANKURO_E2E_TOKEN_USAGE") != "1" {
		t.Skip("use ./tests/e2e/run token-usage after approving the provider budget")
	}
	j, sender, codexHome := newCodexTokenJourney(t)
	first := j.launch()
	expectIdleCodex(t, j.terminal)
	j.expect("GPT-6-Luna")
	j.detach()
	j.expect("Directory:")
	j.expect("worker-e2e.txt")
	j.capture("workspace-before-report", "The selected Codex session initially displays its saved workspace directory.")
	controller := j.launchFixtureSender(sender)
	conn := connectFixtureSender(t, controller.WorkingDir)
	j.press("Down")
	j.expect("worker-e2e.txt")
	nonce := uuid.NewString()[:12]
	reportText := "LUNA-REPORT-" + uuid.NewString()[:8]
	messageID := j.sendLunaTask(conn, first, controller.ID, nonce, reportText)
	j.expect(reportText)
	j.expect("[progress]")
	j.capture("live-report", "A live Codex progress report replaces the selected session’s workspace listing.")
	replies := callNativeTool[session.ReplyPage](t, conn, "get_replies", map[string]any{"message_id": messageID})
	if len(replies.Items) != 1 || replies.Items[0].SenderSessionID != first.ID || replies.Items[0].Kind != "progress" || replies.Items[0].Text != reportText {
		t.Fatal("the correlated report does not match the selected session and authored text")
	}
	registered := waitSession(t, j.configDir, 10*time.Second, func(data session.Data) bool { return data.ID == first.ID && len(data.Conversations[j.card.ID]) == 1 })
	binding := registered.Conversations[j.card.ID][0]
	if binding.ID == "" || binding.Driver != "codex" || binding.Name != "Luna E2E" {
		t.Fatal("Codex did not persist the expected conversation binding")
	}
	j.run.ConversationID = binding.ID
	waitNativeTurn(t, codexHome, binding.ID, 1)
	j.run.Assertions = append(j.run.Assertions, "MCP attributed the correlated progress report to the Codex session and Kankuro persisted its named conversation binding.")
	j.resumeLuna(first, reportText, conn)

	j.run.Stage = "recall budget"
	before := nativeTokenUsage(t, codexHome, binding.ID)
	if reason := lunaRecallBudgetReason(before); reason != "" {
		j.run.TokenUsage = &before
		j.finish()
		j.run.Status = "budget-limited"
		j.run.Stage = "recall budget"
		j.run.SkipReason = reason
		t.Skip(reason)
	}
	j.recallLuna(nonce, codexHome)
	j.finish()
	usage := nativeTokenUsage(t, codexHome, binding.ID)
	t.Logf("Codex rollout token usage: input=%d cached_input=%d output=%d", usage.Input, usage.Cached, usage.Output)
	j.run.TokenUsage = &usage
	if usage.Input > lunaInputLimit || usage.Output > lunaOutputLimit {
		t.Fatalf("approved token budget exceeded: %+v", usage)
	}
}

func lunaRecallBudgetReason(usage tokenUsage) string {
	inputThreshold := lunaInputLimit - lunaRecallInputReserve
	outputThreshold := lunaOutputLimit - lunaRecallOutputReserve
	var limits []string
	if usage.Input > inputThreshold {
		limits = append(limits, fmt.Sprintf("input %d > %d (%d cached)", usage.Input, inputThreshold, usage.Cached))
	}
	if usage.Output > outputThreshold {
		limits = append(limits, fmt.Sprintf("output %d > %d", usage.Output, outputThreshold))
	}
	if len(limits) == 0 {
		return ""
	}
	return "recall skipped to preserve token reserve: " + strings.Join(limits, ", ")
}

func newCodexTokenJourney(t *testing.T) (*nativeJourney, config.IDCard, string) {
	t.Helper()
	for _, name := range []string{"codex", "cat", "mcp-cli"} {
		_, err := exec.LookPath(name)
		must(t, err)
	}
	if os.Getenv("TUI_TEST_RECORDING_FONT_FAMILIES") == "" {
		t.Setenv("TUI_TEST_RECORDING_FONT_FAMILIES", "DejaVu Sans Mono,DejaVu Sans")
	}
	verifyNativeCapture(t)
	root, err := filepath.Abs("../..")
	must(t, err)
	agent := nativeAgent{name: "codex"}
	run, artifacts := newJourneyRun(t, root, agent)
	run.Model = lunaModel
	sandbox := t.TempDir()
	env := isolatedAgentEnvironment(t, sandbox, "codex")
	codexHome := env["CODEX_HOME"]
	// Leave user instructions, hooks, third-party tools, and model overrides out of the fixture.
	for _, name := range []string{"config.toml", "instructions.md", "rules"} {
		must(t, os.RemoveAll(filepath.Join(codexHome, name)))
	}
	must(t, os.MkdirAll(codexHome, 0700))
	settings := fmt.Sprintf("model = %q\nmodel_reasoning_effort = \"low\"\ncheck_for_update_on_startup = false\napproval_policy = \"never\"\nsandbox_mode = \"workspace-write\"\n[sandbox_workspace_write]\nnetwork_access = true\n", lunaModel)
	must(t, os.WriteFile(filepath.Join(codexHome, "config.toml"), []byte(settings), 0600))
	t.Cleanup(func() {
		defer stopCodexUpdater(t, codexHome)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "codex", "app-server", "daemon", "stop")
		cmd.Env = append(os.Environ(), "CODEX_HOME="+codexHome, "HOME="+env["HOME"])
		if err := cmd.Run(); err != nil {
			t.Errorf("stop isolated Codex daemon: %v", err)
		}
	})
	configDir := filepath.Join(sandbox, "config")
	must(t, os.MkdirAll(configDir, 0700))
	must(t, configstartup.PrepareStartup(configDir))
	repo, err := config.NewFileRepository(configDir)
	must(t, err)
	workspace := filepath.Join(sandbox, "worker")
	senderDir := filepath.Join(sandbox, "sender")
	for _, dir := range []string{workspace, senderDir} {
		must(t, os.MkdirAll(dir, 0700))
	}
	must(t, os.WriteFile(filepath.Join(workspace, "worker-e2e.txt"), nil, 0600))
	must(t, repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}, {Path: senderDir}}))
	workspaces, err := repo.ListWorkspaces()
	must(t, err)
	ids := map[string]string{}
	for _, w := range workspaces {
		ids[w.Path] = w.ID
	}
	card, err := repo.CreateIDCard(config.IDCard{DashboardID: "default", Name: "Native E2E", Driver: "codex", WorkspaceID: ids[workspace], Puppet: "kankuro", MCPIDs: []string{config.BuiltinMCPKankuro}})
	must(t, err)
	sender, err := repo.CreateIDCard(config.IDCard{DashboardID: "default", Name: "Sender E2E", Driver: "cat", WorkspaceID: ids[senderDir], Puppet: "kankuro", DashboardControl: true, MCPIDs: []string{config.BuiltinMCPKankuro}})
	must(t, err)
	binary := filepath.Join(sandbox, "kankuro")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, ".")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build Kankuro: %v\n%s", err, output)
	}
	env["KANKURO_CONFIG_DIR"] = configDir
	env["KANKURO_TMP_DIR"] = filepath.Join(sandbox, "tmp")
	terminal := tuitesttest.New(t, tuitesttest.Options{Program: binary, Prefix: "kankuro-luna", Client: tuitest.ClientOptions{Recording: &tuitest.AutomaticRecording{Mode: tuitest.RecordingDisabled}}, Spawn: tuitest.SpawnOptions{Cols: new(uint16(120)), Rows: new(uint16(40)), Cwd: workspace, Env: env, WaitReady: new(false)}})
	registerTerminalCleanup(t, terminal, run, artifacts)
	t.Cleanup(func() {
		if run.TokenUsage != nil {
			return
		}
		_, usage := readNativeUsage(codexHome, run.ConversationID)
		if usage.Input > 0 {
			run.TokenUsage = &usage
			t.Logf("Codex rollout token usage at shutdown: input=%d cached_input=%d output=%d", usage.Input, usage.Cached, usage.Output)
		}
	})
	return &nativeJourney{t: t, root: root, agent: agent, artifacts: artifacts, configDir: configDir, card: card, terminal: terminal, run: run}, sender, codexHome
}

func (j *nativeJourney) launchFixtureSender(card config.IDCard) session.Data {
	j.run.Stage = "fixture sender"
	j.press("Tab")
	j.press("Down")
	j.press("Enter")
	j.expect("Sub Workspaces:")
	j.press("Enter")
	j.expect("Session launched successfully!")
	sender := waitSession(j.t, j.configDir, 30*time.Second, func(d session.Data) bool { return d.CardID == card.ID && d.PID > 0 })
	j.press("Enter")
	timeout := 10 * time.Second
	for _, text := range []string{"Active Sessions", "Session launched successfully!"} {
		must(j.t, j.terminal.GetByText(text, tuitest.TextSelectorOptions{}).Wait(tuitest.LocatorWaitOptions{State: tuitest.Hidden, Timeout: &timeout}))
	}
	j.press("Ctrl+K")
	j.expect("Active Sessions")
	return sender
}

type nativeAuthTransport struct {
	endpoint      string
	authorization string
}

func (rt nativeAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host != rt.endpoint {
		return nil, fmt.Errorf("unexpected MCP destination")
	}
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", rt.authorization)
	return http.DefaultTransport.RoundTrip(clone)
}

func connectFixtureSender(t *testing.T, workspace string) *sdk.ClientSession {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(workspace, ".kankuro", "mcp-servers.json"))
	must(t, err)
	var cfg struct {
		Servers map[string]struct {
			URL     string            `json:"url"`
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	must(t, json.Unmarshal(raw, &cfg))
	server := cfg.Servers["kankuro"]
	address, err := url.Parse(server.URL)
	must(t, err)
	if address.Scheme != "http" || (address.Hostname() != "127.0.0.1" && address.Hostname() != "localhost" && address.Hostname() != "::1") || server.Headers["Authorization"] == "" {
		t.Fatal("expected an authenticated loopback Kankuro MCP")
	}
	client := sdk.NewClient(&sdk.Implementation{Name: "kankuro-native-e2e", Version: "test"}, nil)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	conn, err := client.Connect(ctx, &sdk.StreamableClientTransport{Endpoint: server.URL, HTTPClient: &http.Client{Transport: nativeAuthTransport{endpoint: address.Host, authorization: server.Headers["Authorization"]}}}, nil)
	must(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func callNativeTool[T any](t *testing.T, conn *sdk.ClientSession, name string, args map[string]any) T {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	result, err := conn.CallTool(ctx, &sdk.CallToolParams{Name: name, Arguments: args})
	must(t, err)
	if result.IsError {
		t.Fatalf("%s failed: %#v", name, result.Content)
	}
	raw, err := json.Marshal(result.StructuredContent)
	must(t, err)
	var out T
	must(t, json.Unmarshal(raw, &out))
	return out
}

func (j *nativeJourney) sendLunaTask(conn *sdk.ClientSession, worker session.Data, sender, nonce, report string) string {
	t := j.t
	j.run.Stage = "one live report and registration"
	deadline := time.Now().Add(20 * time.Second)
	var item api.SessionInspection
	for {
		res := callNativeTool[struct {
			Items api.SessionInspection `json:"items"`
		}](t, conn, "get_session", map[string]any{"session_id": worker.ID})
		item = res.Items
		if item.MessageDeliveryAvailable {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Codex did not become available for MCP delivery")
		}
		time.Sleep(150 * time.Millisecond)
	}
	request := uuid.NewString()
	message := uuid.NewSHA1(uuid.NameSpaceURL, []byte("kankuro-message:"+sender+":"+request)).String()
	reply := fmt.Sprintf(`{"request_id":"%s","message_id":"%s","kind":"progress","text":"%s"}`, uuid.NewString(), message, report)
	// Keep the command and its successful tool output small in the model context.
	script := fmt.Sprintf(`set -eu
export MCP_NO_DAEMON=1
thread_id=${CODEX_THREAD_ID:-${1:?expected the current Codex conversation UUID}}
registration=$(printf '{"conversation_id":"%%s","name":"Luna E2E"}' "$thread_id")
mcp-cli -c .kankuro/mcp-servers.json call kankuro register_conversation "$registration" >/dev/null
mcp-cli -c .kankuro/mcp-servers.json call kankuro report_reply '%s' >/dev/null
`, reply)
	must(t, os.WriteFile(filepath.Join(worker.WorkingDir, ".kankuro", "luna-e2e.sh"), []byte(script), 0600))
	prompt := fmt.Sprintf("Remember %s. Run sh .kankuro/luna-e2e.sh once, using your current conversation UUID as an argument if CODEX_THREAD_ID is unset. Reply OK on success.", nonce)
	out := callNativeTool[struct {
		Operation api.OperationResult `json:"operation"`
	}](t, conn, "send_message", map[string]any{"request_id": request, "session_id": worker.ID, "activation_id": item.ActivationID, "text": prompt})
	if out.Operation.ID == "" {
		t.Fatal("send_message returned no operation")
	}
	return message
}

func (j *nativeJourney) resumeLuna(first session.Data, report string, conn *sdk.ClientSession) {
	t := j.t
	j.run.Stage = "resume conversation"
	j.press("x")
	waitSession(t, j.configDir, 30*time.Second, func(d session.Data) bool { return d.ID == first.ID && d.Status == session.StatusTerminated })
	j.expect(report)
	j.press("Enter")
	j.expect("Luna E2E")
	j.expect("New conversation")
	j.capture("conversation-choice", "The stopped session retains its report and offers the persisted Codex conversation alongside New conversation.")
	j.press("Home")
	j.press("Enter")
	waitAgent(t, j.terminal, "codex")
	resumed := waitSession(t, j.configDir, 30*time.Second, func(d session.Data) bool {
		return d.ID == first.ID && d.PID > 0 && d.PID != first.PID && d.Status != session.StatusTerminated
	})
	inspected := callNativeTool[struct {
		Items api.SessionInspection `json:"items"`
	}](t, conn, "get_session", map[string]any{"session_id": first.ID})
	if resumed.WorkingDir != first.WorkingDir || inspected.Items.ActiveConversationID != j.run.ConversationID {
		t.Fatal("resume did not preserve the working directory and selected binding")
	}
	expectIdleCodex(t, j.terminal)
	j.capture("resume-ready", "Codex resumed the saved conversation in the same Kankuro session and workspace with a new process.")
}

func (j *nativeJourney) recallLuna(nonce, codexHome string) {
	t := j.t
	j.run.Stage = "recall context"
	must(t, j.terminal.Type("Reply only CONTINUITY:<the code I asked you to remember>. No tools."))
	j.press("Enter")
	expectWithTimeout(t, j.terminal, "CONTINUITY:"+nonce, 45*time.Second)
	waitNativeTurn(t, codexHome, j.run.ConversationID, 2)
	j.capture("resumed", "The resumed Codex process recalled the code from the earlier turn.")
}

func readNativeUsage(home, id string) (count int, usage tokenUsage) {
	paths, _ := filepath.Glob(filepath.Join(home, "sessions", "*", "*", "*", "*"+id+".jsonl"))
	for _, p := range paths {
		body, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(body), "\n") {
			var entry struct {
				Type    string `json:"type"`
				Payload struct {
					Type string `json:"type"`
					Info *struct {
						Total tokenUsage `json:"total_token_usage"`
					} `json:"info"`
				} `json:"payload"`
			}
			if json.Unmarshal([]byte(line), &entry) != nil || entry.Type != "event_msg" {
				continue
			}
			switch entry.Payload.Type {
			case "task_complete", "turn_complete":
				count++
			case "token_count":
				if entry.Payload.Info != nil {
					usage = entry.Payload.Info.Total
				}
			}
		}
	}
	return
}
func waitNativeTurn(t *testing.T, home, id string, n int) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for {
		complete, usage := readNativeUsage(home, id)
		if usage.Input > lunaInputLimit || usage.Output > lunaOutputLimit {
			t.Fatalf("rollout exceeded approved budget: %+v", usage)
		}
		if complete >= n {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("Codex rollout did not record completion for turn %d", n)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
func nativeTokenUsage(t *testing.T, home, id string) tokenUsage {
	t.Helper()
	_, usage := readNativeUsage(home, id)
	if usage.Input == 0 || usage.Output == 0 {
		t.Fatal("Codex did not record input/output token usage")
	}
	return usage
}
