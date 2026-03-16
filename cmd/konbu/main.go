package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/krtw00/konbu/internal/client"
	"github.com/krtw00/konbu/internal/mcp"
	"github.com/spf13/cobra"
)

var (
	apiURL string
	apiKey string
	jsonOut bool
)

func cli() *client.Client {
	if u := os.Getenv("KONBU_API"); u != "" {
		apiURL = u
	}
	if k := os.Getenv("KONBU_API_KEY"); k != "" {
		apiKey = k
	}
	return client.New(apiURL, apiKey)
}

// resolveID resolves short ID prefix to full UUID by listing items
func resolveMemoID(prefix string) string {
	if len(prefix) >= 36 {
		return prefix
	}
	memos, err := cli().ListMemos()
	if err != nil {
		return prefix
	}
	for _, m := range memos {
		if strings.HasPrefix(m.ID, prefix) {
			return m.ID
		}
	}
	return prefix
}

func resolveTodoID(prefix string) string {
	if len(prefix) >= 36 {
		return prefix
	}
	todos, err := cli().ListTodos()
	if err != nil {
		return prefix
	}
	for _, t := range todos {
		if strings.HasPrefix(t.ID, prefix) {
			return t.ID
		}
	}
	return prefix
}

func resolveEventID(prefix string) string {
	if len(prefix) >= 36 {
		return prefix
	}
	events, err := cli().ListEvents()
	if err != nil {
		return prefix
	}
	for _, e := range events {
		if strings.HasPrefix(e.ID, prefix) {
			return e.ID
		}
	}
	return prefix
}

func resolveToolID(prefix string) string {
	if len(prefix) >= 36 {
		return prefix
	}
	tools, err := cli().ListTools()
	if err != nil {
		return prefix
	}
	for _, t := range tools {
		if strings.HasPrefix(t.ID, prefix) {
			return t.ID
		}
	}
	return prefix
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func main() {
	root := &cobra.Command{
		Use:   "konbu",
		Short: "konbu CLI — personal workspace",
	}

	root.PersistentFlags().StringVar(&apiURL, "api", "http://localhost:8080", "API base URL (or KONBU_API env)")
	root.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key (or KONBU_API_KEY env)")
	root.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output as JSON")

	root.AddCommand(memoCmd(), todoCmd(), eventCmd(), toolCmd(), tagCmd(), searchCmd(), exportCmd(), importCmd(), apiKeyCmd(), mcpCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// --- memo ---

func memoCmd() *cobra.Command {
	memo := &cobra.Command{Use: "memo", Short: "Manage memos"}

	memo.AddCommand(&cobra.Command{
		Use: "list", Short: "List memos",
		RunE: func(cmd *cobra.Command, args []string) error {
			memos, err := cli().ListMemos()
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(memos)
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE\tTAGS\tUPDATED")
			for _, m := range memos {
				tags := make([]string, len(m.Tags))
				for i, t := range m.Tags {
					tags[i] = t.Name
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.ID[:8], m.Title, strings.Join(tags, ","), m.UpdatedAt[:10])
			}
			w.Flush()
			return nil
		},
	})

	memo.AddCommand(&cobra.Command{
		Use: "show [id]", Short: "Show memo content",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveMemoID(args[0])
			m, err := cli().GetMemo(id)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(m)
				return nil
			}
			tags := make([]string, len(m.Tags))
			for i, t := range m.Tags {
				tags[i] = t.Name
			}
			fmt.Printf("# %s\n\n", m.Title)
			if len(tags) > 0 {
				fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
			}
			fmt.Printf("Updated: %s\n\n", m.UpdatedAt[:19])
			fmt.Println(m.Content)
			return nil
		},
	})

	add := &cobra.Command{
		Use: "add [title]", Short: "Create a memo",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := strings.Join(args, " ")
			content, _ := cmd.Flags().GetString("content")
			tagsStr, _ := cmd.Flags().GetString("tags")
			var tags []string
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
			}
			if content == "-" {
				b, err := os.ReadFile("/dev/stdin")
				if err != nil {
					return err
				}
				content = string(b)
			}
			m, err := cli().CreateMemo(title, content, tags)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(m)
				return nil
			}
			fmt.Printf("Created: %s (%s)\n", m.Title, m.ID[:8])
			return nil
		},
	}
	add.Flags().StringP("content", "c", "", "Memo content (use '-' for stdin)")
	add.Flags().StringP("tags", "t", "", "Comma-separated tags")
	memo.AddCommand(add)

	edit := &cobra.Command{
		Use: "edit [id]", Short: "Update a memo",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveMemoID(args[0])
			fields := map[string]any{}
			if t, _ := cmd.Flags().GetString("title"); t != "" {
				fields["title"] = t
			}
			if c, _ := cmd.Flags().GetString("content"); c != "" {
				if c == "-" {
					b, err := os.ReadFile("/dev/stdin")
					if err != nil {
						return err
					}
					c = string(b)
				}
				fields["content"] = c
			}
			if t, _ := cmd.Flags().GetString("tags"); t != "" {
				fields["tags"] = strings.Split(t, ",")
			}
			m, err := cli().UpdateMemo(id, fields)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(m)
				return nil
			}
			fmt.Printf("Updated: %s\n", m.Title)
			return nil
		},
	}
	edit.Flags().String("title", "", "New title")
	edit.Flags().StringP("content", "c", "", "New content (use '-' for stdin)")
	edit.Flags().StringP("tags", "t", "", "Comma-separated tags")
	memo.AddCommand(edit)

	memo.AddCommand(&cobra.Command{
		Use: "rm [id]", Short: "Delete a memo",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveMemoID(args[0])
			if err := cli().DeleteMemo(id); err != nil {
				return err
			}
			fmt.Println("Deleted.")
			return nil
		},
	})

	return memo
}

// --- todo ---

func todoCmd() *cobra.Command {
	todo := &cobra.Command{Use: "todo", Short: "Manage todos"}

	todo.AddCommand(&cobra.Command{
		Use: "list", Short: "List todos",
		RunE: func(cmd *cobra.Command, args []string) error {
			todos, err := cli().ListTodos()
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(todos)
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tSTATUS\tTITLE\tDUE\tTAGS")
			for _, t := range todos {
				mark := "○"
				if t.Status == "done" {
					mark = "✓"
				}
				due := "-"
				if t.DueDate != "" {
					due = t.DueDate[:10]
				}
				tags := make([]string, len(t.Tags))
				for i, tag := range t.Tags {
					tags[i] = tag.Name
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", t.ID[:8], mark, t.Title, due, strings.Join(tags, ","))
			}
			w.Flush()
			return nil
		},
	})

	todo.AddCommand(&cobra.Command{
		Use: "show [id]", Short: "Show todo details",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveTodoID(args[0])
			t, err := cli().GetTodo(id)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(t)
				return nil
			}
			mark := "○ open"
			if t.Status == "done" {
				mark = "✓ done"
			}
			tags := make([]string, len(t.Tags))
			for i, tag := range t.Tags {
				tags[i] = tag.Name
			}
			fmt.Printf("[%s] %s\n", mark, t.Title)
			if t.DueDate != "" {
				fmt.Printf("Due: %s\n", t.DueDate[:10])
			}
			if len(tags) > 0 {
				fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
			}
			if t.Description != "" {
				fmt.Printf("\n%s\n", t.Description)
			}
			return nil
		},
	})

	add := &cobra.Command{
		Use: "add [title]", Short: "Create a todo",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := strings.Join(args, " ")
			tagsStr, _ := cmd.Flags().GetString("tags")
			due, _ := cmd.Flags().GetString("due")
			desc, _ := cmd.Flags().GetString("desc")
			body := map[string]any{"title": title}
			if tagsStr != "" {
				body["tags"] = strings.Split(tagsStr, ",")
			}
			if due != "" {
				body["due_date"] = due
			}
			if desc != "" {
				body["description"] = desc
			}
			t, err := cli().CreateTodo(body)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(t)
				return nil
			}
			fmt.Printf("Created: %s (%s)\n", t.Title, t.ID[:8])
			return nil
		},
	}
	add.Flags().StringP("tags", "t", "", "Comma-separated tags")
	add.Flags().StringP("due", "d", "", "Due date (YYYY-MM-DD)")
	add.Flags().String("desc", "", "Description/notes")
	todo.AddCommand(add)

	edit := &cobra.Command{
		Use: "edit [id]", Short: "Update a todo",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveTodoID(args[0])
			fields := map[string]any{}
			if v, _ := cmd.Flags().GetString("title"); v != "" {
				fields["title"] = v
			}
			if v, _ := cmd.Flags().GetString("desc"); v != "" {
				fields["description"] = v
			}
			if v, _ := cmd.Flags().GetString("due"); v != "" {
				fields["due_date"] = v
			}
			if v, _ := cmd.Flags().GetString("tags"); v != "" {
				fields["tags"] = strings.Split(v, ",")
			}
			t, err := cli().UpdateTodo(id, fields)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(t)
				return nil
			}
			fmt.Printf("Updated: %s\n", t.Title)
			return nil
		},
	}
	edit.Flags().String("title", "", "New title")
	edit.Flags().String("desc", "", "New description/notes")
	edit.Flags().StringP("due", "d", "", "Due date (YYYY-MM-DD)")
	edit.Flags().StringP("tags", "t", "", "Comma-separated tags")
	todo.AddCommand(edit)

	todo.AddCommand(&cobra.Command{
		Use: "done [id]", Short: "Mark todo as done",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveTodoID(args[0])
			if err := cli().DoneTodo(id); err != nil {
				return err
			}
			fmt.Println("Done.")
			return nil
		},
	})

	todo.AddCommand(&cobra.Command{
		Use: "reopen [id]", Short: "Reopen a todo",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveTodoID(args[0])
			if err := cli().ReopenTodo(id); err != nil {
				return err
			}
			fmt.Println("Reopened.")
			return nil
		},
	})

	todo.AddCommand(&cobra.Command{
		Use: "rm [id]", Short: "Delete a todo",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveTodoID(args[0])
			if err := cli().DeleteTodo(id); err != nil {
				return err
			}
			fmt.Println("Deleted.")
			return nil
		},
	})

	return todo
}


// --- event ---

func eventCmd() *cobra.Command {
	event := &cobra.Command{Use: "event", Short: "Manage calendar events"}

	event.AddCommand(&cobra.Command{
		Use: "list", Short: "List events",
		RunE: func(cmd *cobra.Command, args []string) error {
			events, err := cli().ListEvents()
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(events)
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE\tSTART\tEND\tALL_DAY")
			for _, e := range events {
				start := e.StartAt[:16]
				end := "-"
				allDay := ""
				if e.AllDay {
					allDay = "✓"
					start = e.StartAt[:10]
				}
				if e.EndAt != nil && *e.EndAt != "" {
					if e.AllDay {
						end = (*e.EndAt)[:10]
					} else {
						end = (*e.EndAt)[:16]
					}
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", e.ID[:8], e.Title, start, end, allDay)
			}
			w.Flush()
			return nil
		},
	})

	event.AddCommand(&cobra.Command{
		Use: "show [id]", Short: "Show event details",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveEventID(args[0])
			e, err := cli().GetEvent(id)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(e)
				return nil
			}
			fmt.Printf("# %s\n\n", e.Title)
			fmt.Printf("Start: %s\n", e.StartAt[:19])
			if e.EndAt != nil && *e.EndAt != "" {
				fmt.Printf("End:   %s\n", (*e.EndAt)[:19])
			}
			if e.AllDay {
				fmt.Println("All day: yes")
			}
			if e.RecurrenceRule != nil && *e.RecurrenceRule != "" {
				fmt.Printf("Recurrence: %s\n", *e.RecurrenceRule)
			}
			tags := make([]string, len(e.Tags))
			for i, t := range e.Tags {
				tags[i] = t.Name
			}
			if len(tags) > 0 {
				fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
			}
			if e.Description != "" {
				fmt.Printf("\n%s\n", e.Description)
			}
			return nil
		},
	})

	add := &cobra.Command{
		Use: "add [title]", Short: "Create an event",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := strings.Join(args, " ")
			startStr, _ := cmd.Flags().GetString("start")
			endStr, _ := cmd.Flags().GetString("end")
			desc, _ := cmd.Flags().GetString("desc")
			allDay, _ := cmd.Flags().GetBool("all-day")

			if startStr == "" {
				startStr = time.Now().Format(time.RFC3339)
			}

			fields := map[string]any{
				"title":       title,
				"description": desc,
				"start_at":    startStr,
				"all_day":     allDay,
				"tags":        []string{},
			}
			if endStr != "" {
				fields["end_at"] = endStr
			}

			e, err := cli().CreateEvent(fields)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(e)
				return nil
			}
			fmt.Printf("Created: %s (%s)\n", e.Title, e.ID[:8])
			return nil
		},
	}
	add.Flags().StringP("start", "s", "", "Start time (RFC3339, default: now)")
	add.Flags().StringP("end", "e", "", "End time (RFC3339)")
	add.Flags().StringP("desc", "d", "", "Description")
	add.Flags().Bool("all-day", false, "All day event")
	event.AddCommand(add)

	edit := &cobra.Command{
		Use: "edit [id]", Short: "Update an event",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveEventID(args[0])
			fields := map[string]any{}
			if v, _ := cmd.Flags().GetString("title"); v != "" {
				fields["title"] = v
			}
			if v, _ := cmd.Flags().GetString("start"); v != "" {
				fields["start_at"] = v
			}
			if v, _ := cmd.Flags().GetString("end"); v != "" {
				fields["end_at"] = v
			}
			if v, _ := cmd.Flags().GetString("desc"); v != "" {
				fields["description"] = v
			}
			if cmd.Flags().Changed("all-day") {
				v, _ := cmd.Flags().GetBool("all-day")
				fields["all_day"] = v
			}
			e, err := cli().UpdateEvent(id, fields)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(e)
				return nil
			}
			fmt.Printf("Updated: %s\n", e.Title)
			return nil
		},
	}
	edit.Flags().String("title", "", "New title")
	edit.Flags().StringP("start", "s", "", "Start time (RFC3339)")
	edit.Flags().StringP("end", "e", "", "End time (RFC3339)")
	edit.Flags().StringP("desc", "d", "", "Description")
	edit.Flags().Bool("all-day", false, "All day event")
	event.AddCommand(edit)

	event.AddCommand(&cobra.Command{
		Use: "rm [id]", Short: "Delete an event",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveEventID(args[0])
			if err := cli().DeleteEvent(id); err != nil {
				return err
			}
			fmt.Println("Deleted.")
			return nil
		},
	})

	return event
}

// --- tool ---

func toolCmd() *cobra.Command {
	tool := &cobra.Command{Use: "tool", Short: "Manage tools/bookmarks"}

	tool.AddCommand(&cobra.Command{
		Use: "list", Short: "List tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			tools, err := cli().ListTools()
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(tools)
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tURL\tCATEGORY")
			for _, t := range tools {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.ID[:8], t.Name, t.URL, t.Category)
			}
			w.Flush()
			return nil
		},
	})

	add := &cobra.Command{
		Use: "add [name] [url]", Short: "Add a tool",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cat, _ := cmd.Flags().GetString("category")
			t, err := cli().CreateTool(args[0], args[1], cat)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(t)
				return nil
			}
			fmt.Printf("Created: %s (%s)\n", t.Name, t.ID[:8])
			return nil
		},
	}
	add.Flags().String("category", "", "Category")
	tool.AddCommand(add)

	edit := &cobra.Command{
		Use: "edit [id]", Short: "Update a tool",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveToolID(args[0])
			fields := map[string]any{}
			if v, _ := cmd.Flags().GetString("name"); v != "" {
				fields["name"] = v
			}
			if v, _ := cmd.Flags().GetString("url"); v != "" {
				fields["url"] = v
			}
			if v, _ := cmd.Flags().GetString("icon"); v != "" {
				fields["icon"] = v
			}
			if v, _ := cmd.Flags().GetString("category"); v != "" {
				fields["category"] = v
			}
			t, err := cli().UpdateTool(id, fields)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(t)
				return nil
			}
			fmt.Printf("Updated: %s\n", t.Name)
			return nil
		},
	}
	edit.Flags().String("name", "", "New name")
	edit.Flags().String("url", "", "New URL")
	edit.Flags().String("icon", "", "Icon (emoji or letter)")
	edit.Flags().String("category", "", "Category")
	tool.AddCommand(edit)

	tool.AddCommand(&cobra.Command{
		Use: "rm [id]", Short: "Delete a tool",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveToolID(args[0])
			if err := cli().DeleteTool(id); err != nil {
				return err
			}
			fmt.Println("Deleted.")
			return nil
		},
	})

	return tool
}

// --- tag ---

func tagCmd() *cobra.Command {
	tag := &cobra.Command{Use: "tag", Short: "Manage tags"}

	tag.AddCommand(&cobra.Command{
		Use: "list", Short: "List tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			tags, err := cli().ListTags()
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(tags)
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME")
			for _, t := range tags {
				fmt.Fprintf(w, "%s\t%s\n", t.ID[:8], t.Name)
			}
			w.Flush()
			return nil
		},
	})

	tag.AddCommand(&cobra.Command{
		Use: "rm [id]", Short: "Delete a tag",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli().DeleteTag(args[0])
		},
	})

	return tag
}

// --- search ---

func searchCmd() *cobra.Command {
	return &cobra.Command{
		Use: "search [query]", Short: "Search memos, todos, events",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			results, err := cli().Search(query)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(results)
				return nil
			}
			if len(results) == 0 {
				fmt.Println("No results found.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TYPE\tID\tTITLE\tSNIPPET")
			for _, r := range results {
				snippet := r.Snippet
				if len(snippet) > 60 {
					snippet = snippet[:60] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Type, r.ID[:8], r.Title, snippet)
			}
			w.Flush()
			return nil
		},
	}
}

// --- api-key ---

func apiKeyCmd() *cobra.Command {
	ak := &cobra.Command{Use: "api-key", Short: "Manage API keys"}

	ak.AddCommand(&cobra.Command{
		Use: "list", Short: "List API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			keys, err := cli().ListAPIKeys()
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(keys)
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tCREATED")
			for _, k := range keys {
				fmt.Fprintf(w, "%s\t%s\t%s\n", k.ID[:8], k.Name, k.CreatedAt[:10])
			}
			w.Flush()
			return nil
		},
	})

	ak.AddCommand(&cobra.Command{
		Use: "create [name]", Short: "Create an API key",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			k, err := cli().CreateAPIKey(args[0])
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(k)
				return nil
			}
			fmt.Printf("Key: %s\n", k.Key)
			fmt.Println("Save this key — it won't be shown again.")
			return nil
		},
	})

	ak.AddCommand(&cobra.Command{
		Use: "rm [id]", Short: "Delete an API key",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli().DeleteAPIKey(args[0]); err != nil {
				return err
			}
			fmt.Println("Deleted.")
			return nil
		},
	})

	return ak
}

// --- export ---

func exportCmd() *cobra.Command {
	export := &cobra.Command{Use: "export", Short: "Export data"}

	jsonCmd := &cobra.Command{
		Use: "json", Short: "Export all data as JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, _ := cmd.Flags().GetString("output")
			if out == "" {
				out = fmt.Sprintf("konbu-export-%s.json", time.Now().Format("20060102-150405"))
			}
			if err := cli().ExportJSON(out); err != nil {
				return err
			}
			fmt.Printf("Exported to %s\n", out)
			return nil
		},
	}
	jsonCmd.Flags().StringP("output", "o", "", "Output file path")
	export.AddCommand(jsonCmd)

	mdCmd := &cobra.Command{
		Use: "markdown", Short: "Export all data as Markdown (ZIP)",
		Aliases: []string{"md"},
		RunE: func(cmd *cobra.Command, args []string) error {
			out, _ := cmd.Flags().GetString("output")
			if out == "" {
				out = fmt.Sprintf("konbu-export-%s.zip", time.Now().Format("20060102-150405"))
			}
			if err := cli().ExportMarkdown(out); err != nil {
				return err
			}
			fmt.Printf("Exported to %s\n", out)
			return nil
		},
	}
	mdCmd.Flags().StringP("output", "o", "", "Output file path")
	export.AddCommand(mdCmd)

	return export
}

// --- import ---

func importCmd() *cobra.Command {
	imp := &cobra.Command{Use: "import", Short: "Import data"}

	imp.AddCommand(&cobra.Command{
		Use: "ical [file]", Short: "Import iCal (.ics) file",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := cli().ImportICal(args[0])
			if err != nil {
				return err
			}
			fmt.Println("Import complete.")
			return nil
		},
	})

	return imp
}

// --- mcp ---

func mcpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP (Model Context Protocol) server on stdio",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mcp.Run(cli())
		},
	}
}
