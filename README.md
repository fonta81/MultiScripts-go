# MultiScripts

MultiScripts is a TUI-based (Terminal User Interface) command-line tool designed to manage and execute scripts interactively. It is built in Go using the `gocui` library.

[Leer en español](README.es.md)

## Features

- **Script Catalog:** Centralized management of executable scripts.
- **TUI Interface:** Intuitive navigation directly in the terminal.
- **Details:** View metadata (Category, Author, Description, Command).
- **Code Preview:** Preview source code directly in the interface.
- **Interactive Execution:** Run scripts and view their output without leaving the tool.

## Getting Started

1.  Ensure you have [Go](https://golang.org/) installed.
2.  Clone this repository.
3.  Run the application:
    ```bash
    go run main.go
    ```

## Adding New Scripts

To add new scripts, simply instantiate a new `Script` element in the global `scriptsCatalog` slice defined in `main.go`:

```go
{
    Name:        "name.sh",
    Description: "Brief description.",
    Category:    "Category",
    Author:      "Author",
    FilePath:    "scripts/name.sh",
    Command:     []string{"bash", "scripts/name.sh"},
},
```

The visual layout will update automatically.

## Included Scripts

- `hello.sh`: Interactive welcome script.
- `sysinfo.sh`: Displays basic system information.
- `backup.sh`: Simulation of a backup process.

## Controls

- `↑ / ↓`: Navigate the list.
- `Enter`: Execute the selected script.
- `q / Ctrl+C`: Quit the application.
