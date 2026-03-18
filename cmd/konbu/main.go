package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/krtw00/konbu/internal/client"
	"github.com/krtw00/konbu/internal/mcp"
	"github.com/spf13/cobra"
)

var (
	apiURL  string
	apiKey  string
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
	events, err := cli().ListEvents("")
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

func resolveCalendarID(prefix string) string {
	if len(prefix) >= 36 {
		return prefix
	}
	calendars, err := cli().ListCalendars()
	if err != nil {
		return prefix
	}
	for _, c := range calendars {
		if strings.HasPrefix(c.ID, prefix) {
			return c.ID
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

func parseRowData(data string) (map[string]string, error) {
	if strings.TrimSpace(data) == "" {
		return nil, fmt.Errorf("--data is required")
	}
	rowData := map[string]string{}
	for _, pair := range strings.Split(data, ",") {
		pair = strings.TrimSpace(pair)
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			return nil, fmt.Errorf("invalid --data entry: %s", pair)
		}
		rowData[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	if len(rowData) == 0 {
		return nil, fmt.Errorf("--data is required")
	}
	return rowData, nil
}

func formatRowData(rowData map[string]string) string {
	b, err := json.Marshal(rowData)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func resolveResourceID(resourceType, prefix string) string {
	switch resourceType {
	case "memo":
		return resolveMemoID(prefix)
	case "todo":
		return resolveTodoID(prefix)
	case "calendar":
		return resolveCalendarID(prefix)
	case "event":
		return resolveEventID(prefix)
	default:
		return prefix
	}
}

func normalizeResourceType(resourceType string) string {
	switch strings.ToLower(resourceType) {
	case "memo", "todo", "calendar", "event":
		return strings.ToLower(resourceType)
	default:
		return ""
	}
}

func main() {
	root := &cobra.Command{
		Use:   "konbu",
		Short: "konbu CLI — personal workspace",
	}

	root.PersistentFlags().StringVar(&apiURL, "api", "http://localhost:8080", "API base URL (or KONBU_API env)")
	root.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key (or KONBU_API_KEY env)")
	root.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output as JSON")

	root.AddCommand(memoCmd(), todoCmd(), calendarCmd(), eventCmd(), toolCmd(), shareCmd(), legacyPublicCmd(), tagCmd(), searchCmd(), exportCmd(), importCmd(), apiKeyCmd(), mcpCmd())

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

	rows := &cobra.Command{
		Use:   "rows",
		Short: "Manage table memo rows",
	}

	rows.AddCommand(&cobra.Command{
		Use:   "list [memo_id]",
		Short: "List memo rows",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			memoID := resolveMemoID(args[0])
			rows, err := cli().ListMemoRows(memoID)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(rows)
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tROW_DATA")
			for _, row := range rows {
				fmt.Fprintf(w, "%s\t%s\n", row.ID[:8], formatRowData(row.RowData))
			}
			w.Flush()
			return nil
		},
	})

	addRow := &cobra.Command{
		Use:   "add [memo_id]",
		Short: "Add a memo row",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			memoID := resolveMemoID(args[0])
			dataArg, _ := cmd.Flags().GetString("data")
			rowData, err := parseRowData(dataArg)
			if err != nil {
				return err
			}
			row, err := cli().CreateMemoRow(memoID, rowData)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(row)
				return nil
			}
			fmt.Printf("Created: %s\n", row.ID[:8])
			return nil
		},
	}
	addRow.Flags().String("data", "", "Row data (col_id=value,col_id2=value2)")
	_ = addRow.MarkFlagRequired("data")
	rows.AddCommand(addRow)

	rows.AddCommand(&cobra.Command{
		Use:   "rm [memo_id] [row_id]",
		Short: "Delete a memo row",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			memoID := resolveMemoID(args[0])
			rowID := args[1]
			if err := cli().DeleteMemoRow(memoID, rowID); err != nil {
				return err
			}
			if jsonOut {
				printJSON(map[string]string{"status": "deleted", "memo_id": memoID, "row_id": rowID})
				return nil
			}
			fmt.Println("Deleted.")
			return nil
		},
	})

	exportRows := &cobra.Command{
		Use:   "export [memo_id]",
		Short: "Export memo rows as CSV",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			memoID := resolveMemoID(args[0])
			csvData, err := cli().ExportMemoRowsCSV(memoID)
			if err != nil {
				return err
			}
			out, _ := cmd.Flags().GetString("output")
			if out != "" {
				if err := os.WriteFile(out, []byte(csvData), 0o644); err != nil {
					return err
				}
				if jsonOut {
					printJSON(map[string]string{"memo_id": memoID, "output": out})
					return nil
				}
				fmt.Printf("Exported to %s\n", out)
				return nil
			}
			if jsonOut {
				printJSON(map[string]string{"memo_id": memoID, "csv": csvData})
				return nil
			}
			fmt.Print(csvData)
			return nil
		},
	}
	exportRows.Flags().StringP("output", "o", "", "Output file path")
	rows.AddCommand(exportRows)

	memo.AddCommand(rows)

	memo.AddCommand(&cobra.Command{
		Use: "attach [memo_id] [image_path]", Short: "Attach an image to a memo",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveMemoID(args[0])
			imagePath := args[1]
			attached, err := cli().UploadAttachment(imagePath)
			if err != nil {
				return err
			}
			m, err := cli().GetMemo(id)
			if err != nil {
				return err
			}
			newContent := m.Content + "\n![" + filepath.Base(imagePath) + "](" + attached.URL + ")"
			updated, err := cli().UpdateMemo(id, map[string]any{"content": newContent})
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(map[string]any{"url": attached.URL, "memo": updated})
				return nil
			}
			fmt.Printf("Attached: %s\n", attached.URL)
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

	list := &cobra.Command{
		Use: "list", Short: "List events",
		RunE: func(cmd *cobra.Command, args []string) error {
			calendarID, _ := cmd.Flags().GetString("calendar")
			if calendarID != "" {
				calendarID = resolveCalendarID(calendarID)
			}
			events, err := cli().ListEvents(calendarID)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(events)
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE\tSTART\tEND\tCAL\tALL_DAY\tRECUR")
			for _, e := range events {
				start := e.StartAt[:16]
				end := "-"
				calendar := "-"
				allDay := ""
				recur := ""
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
				if e.CalendarID != nil && *e.CalendarID != "" {
					calendar = (*e.CalendarID)[:8]
				}
				if e.RecurrenceRule != nil && *e.RecurrenceRule != "" {
					recur = *e.RecurrenceRule
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", e.ID[:8], e.Title, start, end, calendar, allDay, recur)
			}
			w.Flush()
			return nil
		},
	}
	list.Flags().String("calendar", "", "Calendar ID")
	event.AddCommand(list)

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
			if calendarID, _ := cmd.Flags().GetString("calendar"); calendarID != "" {
				fields["calendar_id"] = resolveCalendarID(calendarID)
			}
			if endStr != "" {
				fields["end_at"] = endStr
			}
			if recurrence, _ := cmd.Flags().GetString("recurrence"); recurrence != "" {
				fields["recurrence_rule"] = recurrence
			}
			if recurrenceEnd, _ := cmd.Flags().GetString("recurrence-end"); recurrenceEnd != "" {
				fields["recurrence_end"] = recurrenceEnd
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
	add.Flags().String("calendar", "", "Calendar ID")
	add.Flags().StringP("recurrence", "r", "", "Recurrence rule (daily, weekly, monthly, yearly)")
	add.Flags().String("recurrence-end", "", "Recurrence end date (YYYY-MM-DD or RFC3339)")
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
			if cmd.Flags().Changed("calendar") {
				v, _ := cmd.Flags().GetString("calendar")
				if v == "" || v == "none" {
					fields["calendar_id"] = nil
				} else {
					fields["calendar_id"] = resolveCalendarID(v)
				}
			}
			if cmd.Flags().Changed("all-day") {
				v, _ := cmd.Flags().GetBool("all-day")
				fields["all_day"] = v
			}
			if cmd.Flags().Changed("recurrence") {
				v, _ := cmd.Flags().GetString("recurrence")
				if v == "none" {
					fields["recurrence_rule"] = nil
				} else {
					fields["recurrence_rule"] = v
				}
			}
			if cmd.Flags().Changed("recurrence-end") {
				v, _ := cmd.Flags().GetString("recurrence-end")
				if v == "" {
					fields["recurrence_end"] = nil
				} else {
					fields["recurrence_end"] = v
				}
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
	edit.Flags().String("calendar", "", "Calendar ID (or 'none' to remove)")
	edit.Flags().Bool("all-day", false, "All day event")
	edit.Flags().StringP("recurrence", "r", "", "Recurrence rule (daily, weekly, monthly, yearly, none to remove)")
	edit.Flags().String("recurrence-end", "", "Recurrence end date (YYYY-MM-DD or RFC3339)")
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

// --- calendar ---

func calendarCmd() *cobra.Command {
	calendar := &cobra.Command{Use: "calendar", Short: "Manage calendars"}

	calendar.AddCommand(&cobra.Command{
		Use: "list", Short: "List calendars",
		RunE: func(cmd *cobra.Command, args []string) error {
			calendars, err := cli().ListCalendars()
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(calendars)
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tCOLOR\tMEMBERS\tDEFAULT\tSHARED")
			for _, c := range calendars {
				def := ""
				shared := ""
				if c.IsDefault {
					def = "✓"
				}
				if c.ShareToken != nil && *c.ShareToken != "" {
					shared = "✓"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n", c.ID[:8], c.Name, c.Color, c.MemberCount, def, shared)
			}
			w.Flush()
			return nil
		},
	})

	calendar.AddCommand(&cobra.Command{
		Use: "show [id]", Short: "Show calendar details",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveCalendarID(args[0])
			detail, err := cli().GetCalendar(id)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(detail)
				return nil
			}
			fmt.Printf("# %s\n\n", detail.Name)
			fmt.Printf("ID: %s\nColor: %s\nMembers: %d\n", detail.ID, detail.Color, len(detail.Members))
			if detail.IsDefault {
				fmt.Println("Default: yes")
			}
			if detail.ShareToken != nil && *detail.ShareToken != "" {
				fmt.Printf("Share token: %s\n", *detail.ShareToken)
			}
			if len(detail.Members) > 0 {
				fmt.Println("\nMembers:")
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "USER_ID\tNAME\tEMAIL\tROLE\tCOLOR")
				for _, m := range detail.Members {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", m.UserID[:8], m.UserName, m.UserEmail, m.Role, m.Color)
				}
				w.Flush()
			}
			return nil
		},
	})

	add := &cobra.Command{
		Use: "add [name]", Short: "Create a calendar",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			color, _ := cmd.Flags().GetString("color")
			cal, err := cli().CreateCalendar(strings.Join(args, " "), color)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(cal)
				return nil
			}
			fmt.Printf("Created: %s (%s)\n", cal.Name, cal.ID[:8])
			return nil
		},
	}
	add.Flags().String("color", "", "Calendar color (hex)")
	calendar.AddCommand(add)

	edit := &cobra.Command{
		Use: "edit [id]", Short: "Update a calendar",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveCalendarID(args[0])
			fields := map[string]any{}
			if v, _ := cmd.Flags().GetString("name"); v != "" {
				fields["name"] = v
			}
			if v, _ := cmd.Flags().GetString("color"); v != "" {
				fields["color"] = v
			}
			cal, err := cli().UpdateCalendar(id, fields)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(cal)
				return nil
			}
			fmt.Printf("Updated: %s\n", cal.Name)
			return nil
		},
	}
	edit.Flags().String("name", "", "New name")
	edit.Flags().String("color", "", "New color (hex)")
	calendar.AddCommand(edit)

	calendar.AddCommand(&cobra.Command{
		Use: "rm [id]", Short: "Delete a calendar",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveCalendarID(args[0])
			if err := cli().DeleteCalendar(id); err != nil {
				return err
			}
			fmt.Println("Deleted.")
			return nil
		},
	})

	calendar.AddCommand(&cobra.Command{
		Use: "join [token]", Short: "Join a calendar by invitation token",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cal, err := cli().JoinCalendar(args[0])
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(cal)
				return nil
			}
			fmt.Printf("Joined: %s (%s)\n", cal.Name, cal.ID[:8])
			return nil
		},
	})

	calendar.AddCommand(&cobra.Command{
		Use: "share-link [id]", Short: "Create or rotate an invite share token",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveCalendarID(args[0])
			token, err := cli().CreateCalendarShareLink(id)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(map[string]string{"share_token": token})
				return nil
			}
			fmt.Println(token)
			return nil
		},
	})

	calendar.AddCommand(&cobra.Command{
		Use: "revoke-share-link [id]", Short: "Delete the invite share token",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := resolveCalendarID(args[0])
			if err := cli().DeleteCalendarShareLink(id); err != nil {
				return err
			}
			fmt.Println("Revoked.")
			return nil
		},
	})

	member := &cobra.Command{Use: "member", Short: "Manage calendar members"}

	memberAdd := &cobra.Command{
		Use: "add [calendar_id] [email]", Short: "Add a calendar member",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			role, _ := cmd.Flags().GetString("role")
			calendarID := resolveCalendarID(args[0])
			m, err := cli().AddCalendarMember(calendarID, args[1], role)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(m)
				return nil
			}
			fmt.Printf("Added: %s (%s)\n", m.UserEmail, m.Role)
			return nil
		},
	}
	memberAdd.Flags().String("role", "editor", "Role (admin, editor, viewer)")
	member.AddCommand(memberAdd)

	memberEdit := &cobra.Command{
		Use: "edit [calendar_id] [user_id]", Short: "Update a calendar member",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fields := map[string]any{}
			if v, _ := cmd.Flags().GetString("role"); v != "" {
				fields["role"] = v
			}
			if v, _ := cmd.Flags().GetString("color"); v != "" {
				fields["color"] = v
			}
			return cli().UpdateCalendarMember(resolveCalendarID(args[0]), args[1], fields)
		},
	}
	memberEdit.Flags().String("role", "", "Role (admin, editor, viewer)")
	memberEdit.Flags().String("color", "", "Member color")
	member.AddCommand(memberEdit)

	member.AddCommand(&cobra.Command{
		Use: "rm [calendar_id] [user_id]", Short: "Remove a calendar member",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli().RemoveCalendarMember(resolveCalendarID(args[0]), args[1]); err != nil {
				return err
			}
			fmt.Println("Removed.")
			return nil
		},
	})

	calendar.AddCommand(member)

	return calendar
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
		Use: "reorder [id1] [id2] ...", Short: "Reorder tools",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tools, err := cli().ListTools()
			if err != nil {
				return err
			}
			seen := map[string]bool{}
			completeOrder := make([]string, 0, len(tools))
			for _, arg := range args {
				id := resolveToolID(arg)
				if seen[id] {
					continue
				}
				seen[id] = true
				completeOrder = append(completeOrder, id)
			}
			for _, t := range tools {
				if seen[t.ID] {
					continue
				}
				completeOrder = append(completeOrder, t.ID)
			}
			if err := cli().ReorderTools(completeOrder); err != nil {
				return err
			}
			fmt.Println("Reordered.")
			return nil
		},
	})

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

// --- share ---

func shareCmd() *cobra.Command {
	share := &cobra.Command{Use: "share", Short: "Manage share links"}

	share.AddCommand(&cobra.Command{
		Use: "get [resource_type] [id]", Short: "Get share link for a resource",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceType := normalizeResourceType(args[0])
			if resourceType == "" {
				return fmt.Errorf("unsupported resource type: %s", args[0])
			}
			share, err := cli().GetPublicShare(resourceType, resolveResourceID(resourceType, args[1]))
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(share)
				return nil
			}
			if share == nil {
				fmt.Println("No share link.")
				return nil
			}
			fmt.Println(cli().PublicURL(share.Token))
			return nil
		},
	})

	share.AddCommand(&cobra.Command{
		Use: "create [resource_type] [id]", Short: "Create a share link for a resource",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceType := normalizeResourceType(args[0])
			if resourceType == "" {
				return fmt.Errorf("unsupported resource type: %s", args[0])
			}
			share, err := cli().CreatePublicShare(resourceType, resolveResourceID(resourceType, args[1]))
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(map[string]any{"share": share, "url": cli().PublicURL(share.Token)})
				return nil
			}
			fmt.Println(cli().PublicURL(share.Token))
			return nil
		},
	})

	share.AddCommand(&cobra.Command{
		Use: "rm [resource_type] [id]", Short: "Delete a share link for a resource",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceType := normalizeResourceType(args[0])
			if resourceType == "" {
				return fmt.Errorf("unsupported resource type: %s", args[0])
			}
			if err := cli().DeletePublicShare(resourceType, resolveResourceID(resourceType, args[1])); err != nil {
				return err
			}
			fmt.Println("Deleted.")
			return nil
		},
	})

	return share
}

func legacyPublicCmd() *cobra.Command {
	public := shareCmd()
	public.Use = "public"
	public.Hidden = true
	public.Deprecated = "use 'konbu share' instead"
	return public
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
			fmt.Fprintln(w, "TYPE\tID\tTITLE\tTAGS\tSNIPPET")
			for _, r := range results {
				snippet := r.Snippet
				if len(snippet) > 60 {
					snippet = snippet[:60] + "..."
				}
				tags := strings.Join(r.Tags, ",")
				id := r.ID
				if len(id) > 8 {
					id = id[:8]
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.Type, id, r.Title, tags, snippet)
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

	icalCmd := &cobra.Command{
		Use: "ical", Short: "Export calendar as iCal",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, _ := cmd.Flags().GetString("output")
			if err := cli().ExportICal(out); err != nil {
				return err
			}
			if out != "" {
				fmt.Printf("Exported to %s\n", out)
			}
			return nil
		},
	}
	icalCmd.Flags().StringP("output", "o", "", "Output file path")
	export.AddCommand(icalCmd)

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
