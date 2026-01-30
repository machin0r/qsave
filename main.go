package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"

	_ "modernc.org/sqlite"
)

func getOrCreateDB() (*sql.DB, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	dbDir := filepath.Join(dirname, "qsave.db")

	db, err := sql.Open("sqlite", dbDir)
	if err != nil {
		log.Fatal(err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS queries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		body TEXT NOT NULL
	);`

	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	return db, nil
}

func getEditor() string {
	if envEditor := os.Getenv("EDITOR"); envEditor != "" {
		return envEditor
	}

	switch runtime.GOOS {
	case "windows":
		return "notepad"
	default:
		return "vim"
	}
}

func openEditor(initialContent string) (string, error) {
	tmpFile, err := os.CreateTemp("", "example-*.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	if initialContent != "" {
		if _, err := tmpFile.WriteString(initialContent); err != nil {
			return "", err
		}
	}
	tmpFile.Close()

	editorFull := getEditor()
	parts := strings.Fields(editorFull)
	var filteredArgs []string
	for _, arg := range parts[1:] {
		if arg != "--wait" && arg != "-w" {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	executable := parts[0]
	args := append(filteredArgs, tmpFile.Name())

	cmd := exec.Command(executable, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	updatedContent, err := os.ReadFile(tmpFile.Name())
	return string(updatedContent), err

}

func onSaveSuccess(queryName string, body string, db *sql.DB) error {
	if body == "" {
		fmt.Println("Editor was empty, please type a query to save it")
		return nil
	}

	insertSQL := `INSERT INTO queries (name, body) VALUES (?, ?)`
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("could not create SQL statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(queryName, body)
	if err != nil {
		return fmt.Errorf("could not execute insert: %w", err)
	}

	fmt.Println("Query saved successfully!")
	return nil
}

func saveQuery(queryName string, db *sql.DB) error {
	content, err := openEditor("")
	if err != nil {
		return fmt.Errorf("Could not edit query, error: %w", err)
	}

	onSaveSuccess(queryName, content, db)
	return nil
}

func updateQuery(name string, newBody string, db *sql.DB) error {
	if newBody == "" {
		fmt.Println("Query body was empty, if you want to delete a query use the delete command")
		return nil
	}

	_, err := db.Exec("UPDATE queries SET body = ? WHERE name = ?", newBody, name)
	if err != nil {
		return fmt.Errorf("could not update query: %w", err)
	}
	fmt.Println("Query updated successfully!")
	return nil
}

func editQuery(queryName string, db *sql.DB) error {
	var name, body string
	err := db.QueryRow("SELECT name, body FROM queries WHERE name = ?", queryName).Scan(&name, &body)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no query found with name %s", queryName)
		}
		return err
	}

	content, err := openEditor(body)
	if err != nil {
		return fmt.Errorf("Could not edit query, error: %w", err)
	}

	updateQuery(queryName, content, db)
	return nil
}

func deleteQuery(queryName string, db *sql.DB) error {
	deleteSQL := `DELETE FROM queries WHERE name = ?`
	stmt, err := db.Prepare(deleteSQL)
	if err != nil {
		return fmt.Errorf("could not create SQL statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(queryName)
	if err != nil {
		return fmt.Errorf("could not execute delete: %w", err)
	}

	fmt.Println("Query deleted successfully!")
	return nil
}

func searchQuery(searchToken string, db *sql.DB) {
	rows, err := db.Query("SELECT name, body FROM queries WHERE body LIKE ?", "%"+searchToken+"%")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, body string
		rows.Scan(&name, &body)
		fmt.Printf("--- NAME: %s ---\n%s\n\n", name, body)
	}
}

func listQueries(db *sql.DB) {
	rows, err := db.Query("SELECT name FROM queries")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Saved Queries:")
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Printf("  - %s\n", name)
	}
}

func showQuery(queryName string, db *sql.DB) error {
	var name, body string
	err := db.QueryRow("SELECT name, body FROM queries WHERE name = ?", queryName).Scan(&name, &body)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no query found with name %s", queryName)
		}
		return err
	}
	fmt.Printf("--- NAME: %s ---\n%s\n\n", name, body)

	err = clipboard.WriteAll(body)
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}
	fmt.Println("Query copied to clipboard!")

	return nil
}

func main() {
	db, err := getOrCreateDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if len(os.Args) < 2 {
		fmt.Println("Usage: qsave [add|list|search|delete] [args]")
		return
	}

	command := os.Args[1]

	switch command {
	case "add":
		if len(os.Args) < 3 {
			log.Fatal("Please provide a name for the query")
		}
		saveQuery(os.Args[2], db)
	case "search":
		if len(os.Args) < 3 {
			log.Fatal("Please provide a search term")
		}
		searchQuery(os.Args[2], db)
	case "delete":
		if len(os.Args) < 3 {
			log.Fatal("Please provide a name for the query to delete")
		}
		deleteQuery(os.Args[2], db)
	case "edit":
		if len(os.Args) < 3 {
			log.Fatal("Please provide a name for the query to edit")
		}
		editQuery(os.Args[2], db)
	case "list":
		listQueries(db)
	case "show":
		if len(os.Args) < 3 {
			log.Fatal("Please provide a name for the query to show")
		}
		showQuery(os.Args[2], db)
	}
}
