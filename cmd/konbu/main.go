package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/krtw00/konbu/internal/client"
	"github.com/spf13/cobra"
)

var apiURL = "http://localhost:8080"

func cli() *client.Client {
	if u := os.Getenv("KONBU_API"); u != "" {
		apiURL = u
	}
	return client.New(apiURL)
}

func main() {
	root := &cobra.Command{
		Use:   "konbu",
		Short: "konbu CLI — memo, todo, tool management",
	}

	root.PersistentFlags().StringVar(&apiURL, "api", apiURL, "API base URL (or KONBU_API env)")

	// --- memo ---
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

	memoAdd := &cobra.Command{
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
			// If content is "-", read from stdin
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
	memoAdd.Flags().StringP("content", "c", "", "Memo content (use '-' for stdin)")
	memoAdd.Flags().StringP("tags", "t", "", "Comma-separated tags")
	memo.AddCommand(memoAdd)

	memo.AddCommand(&cobra.Command{
		Use: "show [id]", Short: "Show memo content",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Support partial ID match
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

	memo.AddCommand(&cobra.Command{
		Use: "rm [id]", Short: "Delete a memo",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli().DeleteMemo(args[0])
		},
	})

	// --- todo ---
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

	todoAdd := &cobra.Command{
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
	todoAdd.Flags().StringP("tags", "t", "", "Comma-separated tags")
	todo.AddCommand(todoAdd)

	todo.AddCommand(&cobra.Command{
		Use: "done [id]", Short: "Mark todo as done",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli().UpdateTodo(args[0], map[string]any{"status": "done"})
		},
	})

	todo.AddCommand(&cobra.Command{
		Use: "rm [id]", Short: "Delete a todo",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli().DeleteTodo(args[0])
		},
	})

	// --- tool ---
	tool := &cobra.Command{Use: "tool", Short: "Manage tools/bookmarks"}

	tool.AddCommand(&cobra.Command{
		Use: "list", Short: "List tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			tools, err := cli().ListTools()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tURL")
			for _, t := range tools {
				fmt.Fprintf(w, "%s\t%s\t%s\n", t.ID[:8], t.Name, t.URL)
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

	root.AddCommand(memo, todo, tool)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
