package main

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

func runMigrations(db *sql.DB, dir string) error {
    if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil { return err }
    entries, err := os.ReadDir(dir); if err != nil { return err }
    var files []string
    for _, e := range entries { if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") { files = append(files, e.Name()) } }
    sort.Strings(files)
    for _, name := range files {
        var exists bool
        if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version=$1)`, name).Scan(&exists); err != nil { return err }
        if exists { continue }
        b, err := os.ReadFile(filepath.Join(dir, name)); if err != nil { return err }
        tx, err := db.Begin(); if err != nil { return err }
        if _, err = tx.Exec(string(b)); err != nil { _ = tx.Rollback(); return fmt.Errorf("migration %s: %w", name, err) }
        if _, err = tx.Exec(`INSERT INTO schema_migrations(version) VALUES($1)`, name); err != nil { _ = tx.Rollback(); return fmt.Errorf("record migration %s: %w", name, err) }
        if err = tx.Commit(); err != nil { return fmt.Errorf("commit migration %s: %w", name, err) }
    }
    return nil
}
