package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mikkel-andersen/nota/domain"
	"github.com/mikkel-andersen/nota/usecase"
	"github.com/spf13/cobra"
)

func NewRootCmd(svc *usecase.NoteService) *cobra.Command {
	root := &cobra.Command{
		Use:   "nota [note text]",
		Short: "Sticky notes pinned to your current directory",
		Long: `nota — directory-scoped sticky notes.

Running nota with no arguments lists notes for the current directory.
Pass text directly to add a new note.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return addNote(svc, strings.Join(args, " "))
			}
			plain, _ := cmd.Flags().GetBool("plain")
			asJSON, _ := cmd.Flags().GetBool("json")
			return listNotes(svc, plain, asJSON)
		},
	}
	root.Flags().BoolP("plain", "p", false, "plain output without ANSI codes")
	root.Flags().BoolP("json", "j", false, "JSON output")

	root.AddCommand(
		newRmCmd(svc),
		newDoneCmd(svc),
		newEditCmd(svc),
		newClearCmd(svc),
		newLsCmd(svc),
	)
	return root
}

func Execute(svc *usecase.NoteService) {
	if err := NewRootCmd(svc).Execute(); err != nil {
		os.Exit(1)
	}
}

func newRmCmd(svc *usecase.NoteService) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <id>",
		Short: "Delete a note by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id: %s", args[0])
			}
			if err := svc.Delete(id); err != nil {
				return err
			}
			fmt.Printf("deleted note #%d\n", id)
			return nil
		},
	}
}

func newDoneCmd(svc *usecase.NoteService) *cobra.Command {
	return &cobra.Command{
		Use:   "done <id>",
		Short: "Mark a note as done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id: %s", args[0])
			}
			if err := svc.MarkDone(id); err != nil {
				return err
			}
			fmt.Printf("note #%d done\n", id)
			return nil
		},
	}
}

func newEditCmd(svc *usecase.NoteService) *cobra.Command {
	return &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a note in $EDITOR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id: %s", args[0])
			}
			err = svc.Edit(id)
			if errors.Is(err, usecase.ErrNoChanges) {
				fmt.Println("no changes")
				return nil
			}
			if err != nil {
				return err
			}
			fmt.Printf("updated note #%d\n", id)
			return nil
		},
	}
}

func newClearCmd(svc *usecase.NoteService) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Delete all notes for the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := svc.Clear(); err != nil {
				return err
			}
			fmt.Println("cleared all notes")
			return nil
		},
	}
}

func newLsCmd(svc *usecase.NoteService) *cobra.Command {
	ls := &cobra.Command{
		Use:   "ls",
		Short: "List notes (--all to show all directories)",
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			if !all {
				plain, _ := cmd.Flags().GetBool("plain")
				asJSON, _ := cmd.Flags().GetBool("json")
				return listNotes(svc, plain, asJSON)
			}
			groups, err := svc.AllNotes()
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
					fmt.Printf("  \033[2m#%d\033[0m  %s  \033[2m(%s)\033[0m\n", n.ID, n.Body, formatAge(n.CreatedAt))
				}
				fmt.Println()
			}
			return nil
		},
	}
	ls.Flags().BoolP("all", "a", false, "show notes from all directories")
	ls.Flags().BoolP("plain", "p", false, "plain output without ANSI codes")
	ls.Flags().BoolP("json", "j", false, "JSON output")
	return ls
}

func addNote(svc *usecase.NoteService, body string) error {
	n, err := svc.Add(body)
	if err != nil {
		return err
	}
	fmt.Printf("added note #%d\n", n.ID)
	return nil
}

func listNotes(svc *usecase.NoteService, plain, asJSON bool) error {
	notes, err := svc.List()
	if err != nil {
		return err
	}
	if asJSON {
		return renderJSON(notes)
	}
	if len(notes) == 0 {
		if !plain {
			cwd, _ := os.Getwd()
			fmt.Printf("no notes for %s\n", cwd)
		}
		return nil
	}
	if plain {
		renderPlain(notes)
		return nil
	}
	cwd, _ := os.Getwd()
	fmt.Printf("\033[1m%s\033[0m\n", cwd)
	fmt.Println(strings.Repeat("─", 48))
	for _, n := range notes {
		if n.Done {
			fmt.Printf("  \033[2m#%d\033[0m  \033[9m%s\033[0m  \033[2m✓ (%s)\033[0m\n", n.ID, n.Body, formatAge(n.CreatedAt))
		} else {
			fmt.Printf("  \033[2m#%d\033[0m  %s  \033[2m(%s)\033[0m\n", n.ID, n.Body, formatAge(n.CreatedAt))
		}
	}
	return nil
}

func renderPlain(notes []domain.Note) {
	for _, n := range notes {
		fmt.Printf("#%d  %s  (%s)\n", n.ID, n.Body, formatAge(n.CreatedAt))
	}
}

func renderJSON(notes []domain.Note) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(notes)
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
