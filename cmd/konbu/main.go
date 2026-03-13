package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/krtw00/konbu/internal/client"
	"github.com/spf13/cobra"
)

var (
	apiURL string
	apiKey string
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

func resolveID(items []struct{ id string }, prefix string) string {
	for _, item := range items {
		if strings.HasPrefix(item.id, prefix) {
			return item.id
		}
	}
	return prefix
}

func main() {
	root := &cobra.Command{
		Use:   "konbu",
		Short: "konbu CLI — personal workspace",
	}

	root.PersistentFlags().StringVar(&apiURL, "api", "http://localhost:8080", "API base URL (or KONBU_API env)")
	root.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key (or KONBU_API_KEY env)")

	root.AddCommand(memoCmd(), todoCmd(), eventCmd(), toolCmd(), searchCmd(), exportCmd(), importCmd())

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
			fmt.Printf("Created: %s (%s)\n", m.Title, m.ID[:8])
			return nil
		},
	}
	add.Flags().StringP("content", "c", "", "Memo content (use '-' for stdin)")
	add.Flags().StringP("tags", "t", "", "Comma-separated tags")
	memo.AddCommand(add)

	memo.AddCommand(&cobra.Command{
		Use: "show [id]", Short: "Show memo content",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if len(id) < 36 {
				memos, err := cli().ListMemos()
				if err != nil {
					return err
				}
				for _, m := range memos {
					if strings.HasPrefix(m.ID, id) {
						id = m.ID
						break
					}
				}
			}
			m, err := cli().GetMemo(id)
			if err != nil {
				return err
			}
			fmt.Printf("# %s\n\n%s\n", m.Title, m.Content)
			return nil
		},
	})

	edit := &cobra.Command{
		Use: "edit [id]", Short: "Update a memo",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
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
			return cli().DeleteMemo(args[0])
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
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tSTATUS\tTITLE\tDUE")
			for _, t := range todos {
				mark := "○"
				if t.Status == "done" {
					mark = "✓"
				}
				due := "-"
				if t.DueDate != "" {
					due = t.DueDate[:10]
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.ID[:8], mark, t.Title, due)
			}
			w.Flush()
			return nil
		},
	})

	add := &cobra.Command{
		Use: "add [title]", Short: "Create a todo",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := strings.Join(args, " ")
			tagsStr, _ := cmd.Flags().GetString("tags")
			var tags []string
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
			}
			t, err := cli().CreateTodo(title, tags)
			if err != nil {
				return err
			}
			fmt.Printf("Created: %s (%s)\n", t.Title, t.ID[:8])
			return nil
		},
	}
	add.Flags().StringP("tags", "t", "", "Comma-separated tags")
	todo.AddCommand(add)

	todo.AddCommand(&cobra.Command{
		Use: "done [id]", Short: "Mark todo as done",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli().DoneTodo(args[0])
		},
	})

	todo.AddCommand(&cobra.Command{
		Use: "reopen [id]", Short: "Reopen a todo",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli().ReopenTodo(args[0])
		},
	})

	todo.AddCommand(&cobra.Command{
		Use: "rm [id]", Short: "Delete a todo",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli().DeleteTodo(args[0])
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
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE\tSTART\tALL_DAY")
			for _, e := range events {
				start := e.StartAt[:16]
				allDay := ""
				if e.AllDay {
					allDay = "✓"
					start = e.StartAt[:10]
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.ID[:8], e.Title, start, allDay)
			}
			w.Flush()
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
			fmt.Printf("Created: %s (%s)\n", e.Title, e.ID[:8])
			return nil
		},
	}
	add.Flags().StringP("start", "s", "", "Start time (RFC3339, default: now)")
	add.Flags().StringP("end", "e", "", "End time (RFC3339)")
	add.Flags().StringP("desc", "d", "", "Description")
	add.Flags().Bool("all-day", false, "All day event")
	event.AddCommand(add)

	event.AddCommand(&cobra.Command{
		Use: "rm [id]", Short: "Delete an event",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli().DeleteEvent(args[0])
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
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tURL\tCATEGORY")
			for _, t := range tools {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.ID[:8], t.Name, t.URL, t.Category)
			}
			w.Flush()
			return nil
		},
	})

	tool.AddCommand(&cobra.Command{
		Use: "add [name] [url]", Short: "Add a tool",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := cli().CreateTool(args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Printf("Created: %s (%s)\n", t.Name, t.ID[:8])
			return nil
		},
	})

	tool.AddCommand(&cobra.Command{
		Use: "rm [id]", Short: "Delete a tool",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli().DeleteTool(args[0])
		},
	})

	return tool
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
