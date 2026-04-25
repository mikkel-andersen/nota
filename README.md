# nota

Directory-scoped sticky notes for your terminal. Leave notes pinned to a project folder and see them every time you're working there.

```
$ cd ~/workspace/my-api
$ nota "fix rate limiting before deploy"
added note #1
$ nota "ask @john about the auth flow"
added note #2

$ nota
/Users/mikkel/workspace/my-api
────────────────────────────────────────────────
  #1  fix rate limiting before deploy  (2h ago)
  #2  ask @john about the auth flow    (5m ago)
```

## Install

Requires [Go 1.18+](https://go.dev/dl/).

```sh
go install github.com/mikkel-andersen/nota@latest
```

Make sure `~/go/bin` is in your PATH. Add this to your `~/.zshrc` or `~/.bashrc`:

```sh
export PATH=$PATH:$HOME/go/bin
```

## Usage

```
nota [note text]     Add a note for the current directory
nota                 List all notes for the current directory
nota rm <id>         Delete a note by ID
nota clear           Delete all notes for the current directory
```

### Examples

```sh
# Add a note
nota "don't forget to run migrations"

# List notes
nota

# Remove a specific note
nota rm 3

# Clear everything for this directory
nota clear
```

## How it works

Notes are stored as JSON files in `~/.local/share/nota/`, scoped to each directory. Switching directories gives you a completely separate set of notes — nothing bleeds between projects.

```
~/.local/share/nota/
  _Users_mikkel_workspace_my-api/notes.json
  _Users_mikkel_workspace_other-project/notes.json
```

## License

MIT
