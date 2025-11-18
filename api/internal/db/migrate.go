package db

import (
    "database/sql"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"

    "gorm.io/gorm"
    _ "github.com/lib/pq"
)

func RunMigrations(gdb *gorm.DB, dir string) error {
    sqlDB, err := gdb.DB()
    if err != nil {
        return err
    }
    return applySQLFiles(sqlDB, dir)
}

func applySQLFiles(db *sql.DB, dir string) error {
    var files []string
    err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
        if err != nil { return err }
        if !d.IsDir() && filepath.Ext(path) == ".sql" {
            files = append(files, path)
        }
        return nil
    })
    if err != nil { return err }

    for _, f := range files {
        content, err := os.ReadFile(f)
        if err != nil { return err }
        if _, err := db.Exec(string(content)); err != nil {
            return fmt.Errorf("migration %s failed: %w", f, err)
        }
    }
    return nil
}

