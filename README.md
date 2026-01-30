# qsave

<p align="center">
  <img src="assets/qsave.png" width="300" alt="qsave Logo">
</p>

`qsave` is a simple command-line interface (CLI) tool written in Go that allows you to easily save, list, search, edit, and manage frequently used SQL queries or any text snippets. It stores your queries in a local SQLite database, making them readily accessible from your terminal.

## Features

*   **Add Queries:** Save new queries with a given name. The content is edited using your default system editor (`EDITOR` environment variable, `vim` for Unix-like, `notepad` for Windows).
*   **List Queries:** Display the names of all saved queries.
*   **Show Query:** View the content of a specific saved query by its name.
*   **Search Queries:** Find queries containing a specific text token within their body.
*   **Edit Queries:** Modify the content of an existing query.
*   **Delete Queries:** Remove saved queries.

## Installation

To install `qsave`, make sure you have Go installed. Then, run the following command:

```bash
go install github.com/machin0r/qsave@latest
```

This will install the `qsave` executable in your Go bin directory (usually `$GOPATH/bin`). Make sure this directory is in your system's `PATH`.

## Usage

Once installed, you can use `qsave` from your terminal.

### Add a new query

This will open your default editor. Type or paste your query, save, and exit the editor.

```bash
qsave add <query-name>
```

Example:
```bash
qsave add my_complex_join
```

### List all saved queries

```bash
qsave list
```

### Show a specific query

```bash
qsave show <query-name>
```

Example:
```bash
qsave show my_complex_join
```

### Search for queries

```bash
qsave search <search-token>
```

Example:
```bash
qsave search "CREATE TABLE"
```

### Edit an existing query

This will open your default editor with the current content of the query. Edit, save, and exit.

```bash
qsave edit <query-name>
```

Example:
```bash
qsave edit my_complex_join
```

### Delete a query

```bash
qsave delete <query-name>
```

Example:
```bash
qsave delete my_old_query
```

## Database Location

`qsave` stores its data in an SQLite database file named `qsave.db` located in your user's home directory.
