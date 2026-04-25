package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/mikkel-andersen/nota/internal/store"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nota [note text]",
	Short: "Sticky notes pinned to your current directory",
	Long: `nota — directory-scoped sticky notes.

Running nota with no arguments lists notes for the current directory.
Pass text directly to add a new note.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return addNote(strings.Join(args, " "))
		}
		return listNotes()
	},
}

var deleteCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "Delete a note by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid id: %s", args[0])
		}
		s, err := getStore()
		if err != nil {
			return err
		}
		if err := s.Delete(id); err != nil {
			return err
		}
		fmt.Printf("deleted note #%d\n", id)
		return nil
	},
}

var doneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a note as done",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid id: %s", args[0])
		}
		s, err := getStore()
		if err != nil {
			return err
		}
		if err := s.MarkDone(id); err != nil {
			return err
		}
		fmt.Printf("note #%d done\n", id)
		return nil
	},
}

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a note in $EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid id: %s", args[0])
		}
		s, err := getStore()
		if err != nil {
			return err
		}
		notes, err := s.List()
		if err != nil {
			return err
		}
		var current string
		for _, n := range notes {
			if n.ID == id {
				current = n.Body
				break
			}
		}
		if current == "" {
			return fmt.Errorf("note #%d not found", id)
		}

		f, err := os.CreateTemp("", "nota-*.txt")
		if err != nil {
			return err
		}
		defer os.Remove(f.Name())
		if _, err := f.WriteString(current); err != nil {
			return err
		}
		f.Close()

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		c := exec.Command(editor, f.Name())
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}

		updated, err := os.ReadFile(f.Name())
		if err != nil {
			return err
		}
		body := strings.TrimSpace(string(updated))
		if body == "" || body == current {
			fmt.Println("no changes")
			return nil
		}
		if err := s.Update(id, body); err != nil {
			return err
		}
		fmt.Printf("updated note #%d\n", id)
		return nil
	},
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List notes (--all to show all directories)",
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		if !all {
			return listNotes()
		}
		baseDir := filepath.Join(xdg.DataHome, "nota")
		groups, err := store.AllNotes(baseDir)
		if err != nil {
			return err
		}
		if len(groups) == 0 {
			fmt.Println("no notes found")
			return nil
		}
		paths := make([]string, 0, len(groups))
		for p := range groups {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		for _, p := range paths {
			fmt.Printf("\033[1m%s\033[0m\n", p)
			fmt.Println(strings.Repeat("─", 48))
			for _, n := range groups[p] {
				age := formatAge(n.CreatedAt)
				fmt.Printf("  \033[2m#%d\033[0m  %s  \033[2m(%s)\033[0m\n", n.ID, n.Body, age)
			}
			fmt.Println()
		}
		return nil
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Delete all notes for the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStore()
		if err != nil {
			return err
		}
		if err := s.Clear(); err != nil {
			return err
		}
		fmt.Println("cleared all notes")
		return nil
	},
}

func getStore() (*store.Store, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// hash the cwd into a safe directory name by replacing separators
	safe := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(cwd)
	dir := filepath.Join(xdg.DataHome, "nota", safe)
	return store.New(dir)
}

func addNote(body string) error {
	s, err := getStore()
	if err != nil {
		return err
	}
	n, err := s.Add(body)
	if err != nil {
		return err
	}
	fmt.Printf("added note #%d\n", n.ID)
	return nil
}

func listNotes() error {
	s, err := getStore()
	if err != nil {
		return err
	}
	notes, err := s.List()
	if err != nil {
		return err
	}
	if len(notes) == 0 {
		cwd, _ := os.Getwd()
		fmt.Printf("no notes for %s\n", cwd)
		return nil
	}

	cwd, _ := os.Getwd()
	fmt.Printf("\033[1m%s\033[0m\n", cwd)
	fmt.Println(strings.Repeat("─", 48))
	for _, n := range notes {
		age := formatAge(n.CreatedAt)
		if n.Done {
			fmt.Printf("  \033[2m#%d\033[0m  \033[9m%s\033[0m  \033[2m✓ (%s)\033[0m\n", n.ID, n.Body, age)
		} else {
			fmt.Printf("  \033[2m#%d\033[0m  %s  \033[2m(%s)\033[0m\n", n.ID, n.Body, age)
		}
	}
	return nil
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func Execute() {
	lsCmd.Flags().BoolP("all", "a", false, "show notes from all directories")
	rootCmd.AddCommand(deleteCmd, clearCmd, doneCmd, editCmd, lsCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
