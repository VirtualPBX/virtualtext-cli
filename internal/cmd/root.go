package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/VirtualPBX/virtualtext-cli/internal/auth"
	"github.com/VirtualPBX/virtualtext-cli/internal/client"
	"github.com/VirtualPBX/virtualtext-cli/internal/config"
	"github.com/spf13/cobra"
)

//go:embed skill.md
var skillMarkdown []byte

func Execute() error {
	return rootCmd().Execute()
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "vt",
		Short: "VirtualText CLI for ops metrics, observations, and conversations",
	}
	root.AddCommand(authCmd(), metricsCmd(), conversationsCmd(), conversationCmd(), messagesCmd(), observationsCmd(), observationCmd(), notesCmd(), huddlesCmd(), skillCmd())
	return root
}

func mustClient() *client.Client {
	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}
	return client.New(cfg.Host, cfg.Token)
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fail(err)
	}
}

func printRaw(raw json.RawMessage) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		fmt.Println(string(raw))
		return
	}
	printJSON(v)
}

func fail(err error) {
	printJSON(map[string]string{"error": err.Error()})
	os.Exit(1)
}

func authCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Login and status"}
	var host, token string
	login := &cobra.Command{
		Use:   "login",
		Short: "Authorize the CLI and create an ops API key",
		Run: func(_ *cobra.Command, _ []string) {
			if host == "" {
				fail(fmt.Errorf("--host is required, e.g. https://use.virtualtext.app"))
			}
			if err := auth.ValidateHost(host); err != nil {
				fail(err)
			}
			if token == "" {
				var err error
				token, err = auth.Login(host, auth.RandomState(), 5*time.Minute)
				if err != nil {
					fail(err)
				}
			}
			if err := config.Save(&config.File{Host: host, Token: token}); err != nil {
				fail(err)
			}
			printJSON(map[string]string{"ok": "logged in", "host": host, "config": config.Path()})
		},
	}
	login.Flags().StringVar(&host, "host", os.Getenv("VIRTUALTEXT_HOST"), "VirtualText host")
	login.Flags().StringVar(&token, "token", "", "Paste an existing API key instead of browser login")
	status := &cobra.Command{
		Use:   "status",
		Short: "Show whether a host and token are configured",
		Run: func(_ *cobra.Command, _ []string) {
			cfg, err := config.Load()
			if err != nil {
				fail(err)
			}
			printJSON(map[string]any{
				"host":          cfg.Host,
				"token_present": cfg.Token != "",
				"config":        config.Path(),
			})
		},
	}
	cmd.AddCommand(login, status)
	return cmd
}

func metricsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "metrics", Short: "Response-time metrics"}
	var since, kind string
	response := &cobra.Command{
		Use:   "response",
		Short: "First and average response times",
		Run: func(_ *cobra.Command, _ []string) {
			q := url.Values{}
			if since != "" {
				q.Set("since", since)
			}
			if kind != "" {
				q.Set("kind", kind)
			}
			raw, err := mustClient().Get("/api/ops/metrics/response", q)
			if err != nil {
				fail(err)
			}
			printRaw(raw)
		},
	}
	response.Flags().StringVar(&since, "since", "24h", "Window (24h, 7d, or ISO8601)")
	response.Flags().StringVar(&kind, "kind", "", "human or ai_agent")
	cmd.AddCommand(response)
	return cmd
}

func conversationsCmd() *cobra.Command {
	var waiting bool
	var since string
	cmd := &cobra.Command{Use: "conversations", Short: "List conversations"}
	list := &cobra.Command{
		Use: "list",
		Run: func(_ *cobra.Command, _ []string) {
			q := url.Values{}
			if waiting {
				q.Set("waiting", "true")
			}
			if since != "" {
				q.Set("since", since)
			}
			raw, err := mustClient().Get("/api/ops/conversations", q)
			if err != nil {
				fail(err)
			}
			printRaw(raw)
		},
	}
	list.Flags().BoolVar(&waiting, "waiting", false, "Only conversations waiting on a reply")
	list.Flags().StringVar(&since, "since", "", "Updated since 24h/7d/ISO8601")
	cmd.AddCommand(list)
	return cmd
}

func conversationCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "conversation", Short: "Show one conversation"}
	show := &cobra.Command{
		Use:  "show",
		Args: cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			raw, err := mustClient().Get("/api/ops/conversations/"+args[0], nil)
			if err != nil {
				fail(err)
			}
			printRaw(raw)
		},
	}
	cmd.AddCommand(show)
	return cmd
}

func messagesCmd() *cobra.Command {
	var conversationID string
	cmd := &cobra.Command{Use: "messages", Short: "List messages"}
	list := &cobra.Command{
		Use: "list",
		Run: func(_ *cobra.Command, _ []string) {
			if conversationID == "" {
				fail(fmt.Errorf("--conversation is required"))
			}
			raw, err := mustClient().Get("/api/ops/conversations/"+conversationID+"/messages", nil)
			if err != nil {
				fail(err)
			}
			printRaw(raw)
		},
	}
	list.Flags().StringVar(&conversationID, "conversation", "", "Conversation ID")
	cmd.AddCommand(list)
	return cmd
}

func observationsCmd() *cobra.Command {
	var status, severity string
	cmd := &cobra.Command{Use: "observations", Short: "List observations"}
	list := &cobra.Command{
		Use: "list",
		Run: func(_ *cobra.Command, _ []string) {
			q := url.Values{}
			if status != "" {
				q.Set("status", status)
			}
			if severity != "" {
				q.Set("severity", severity)
			}
			raw, err := mustClient().Get("/api/ops/observations", q)
			if err != nil {
				fail(err)
			}
			printRaw(raw)
		},
	}
	list.Flags().StringVar(&status, "status", "", "pending, reviewed, resolved, dismissed")
	list.Flags().StringVar(&severity, "severity", "", "low, medium, high")
	cmd.AddCommand(list)
	return cmd
}

func observationCmd() *cobra.Command {
	var status, note string
	cmd := &cobra.Command{Use: "observation", Short: "Update an observation"}
	update := &cobra.Command{
		Use:  "update",
		Args: cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			if status == "" {
				fail(fmt.Errorf("--status is required"))
			}
			body := map[string]string{"status": status}
			if note != "" {
				body["note"] = note
			}
			raw, err := mustClient().Patch("/api/ops/observations/"+args[0], body)
			if err != nil {
				fail(err)
			}
			printRaw(raw)
		},
	}
	update.Flags().StringVar(&status, "status", "", "reviewed, resolved, dismissed, reopened")
	update.Flags().StringVar(&note, "note", "", "Optional note")
	cmd.AddCommand(update)
	return cmd
}

func notesCmd() *cobra.Command {
	var conversationID, content string
	cmd := &cobra.Command{Use: "notes", Short: "Create a note"}
	create := &cobra.Command{
		Use: "create",
		Run: func(_ *cobra.Command, _ []string) {
			if conversationID == "" || content == "" {
				fail(fmt.Errorf("--conversation and --content are required"))
			}
			raw, err := mustClient().Post("/api/ops/notes", map[string]string{
				"conversation_id": conversationID,
				"content":         content,
			})
			if err != nil {
				fail(err)
			}
			printRaw(raw)
		},
	}
	create.Flags().StringVar(&conversationID, "conversation", "", "Conversation ID")
	create.Flags().StringVar(&content, "content", "", "Note text")
	cmd.AddCommand(create)
	return cmd
}

func huddlesCmd() *cobra.Command {
	var active bool
	cmd := &cobra.Command{Use: "huddles", Short: "List huddles"}
	list := &cobra.Command{
		Use: "list",
		Run: func(_ *cobra.Command, _ []string) {
			q := url.Values{}
			if active {
				q.Set("active", "true")
			}
			raw, err := mustClient().Get("/api/ops/huddles", q)
			if err != nil {
				fail(err)
			}
			printRaw(raw)
		},
	}
	list.Flags().BoolVar(&active, "active", false, "Only active huddles")
	cmd.AddCommand(list)
	return cmd
}

func skillCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "skill", Short: "Install the Grok skill"}
	install := &cobra.Command{
		Use:   "install",
		Short: "Copy the VirtualText skill into ~/.grok/skills/virtualtext",
		Run: func(_ *cobra.Command, _ []string) {
			home, err := os.UserHomeDir()
			if err != nil {
				fail(err)
			}
			dir := filepath.Join(home, ".grok", "skills", "virtualtext")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				fail(err)
			}
			dest := filepath.Join(dir, "SKILL.md")
			if err := os.WriteFile(dest, skillMarkdown, 0o644); err != nil {
				fail(err)
			}
			printJSON(map[string]string{"ok": "installed", "path": dest})
		},
	}
	cmd.AddCommand(install)
	return cmd
}
