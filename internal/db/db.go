package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "modernc.org/sqlite" // Pure-Go SQLite driver
)

//go:embed schema.sql
var schemaFS embed.FS

// InitDB はデータベース接続を初期化し、スキーマを適用します。
// database/sql の標準インターフェースを使用します。
func InitDB(dbPath string) (*sql.DB, error) {
	// データベースファイルに接続します。ファイルが存在しない場合は作成されます。
	// URI形式のDSNとmode=rwcを指定して、読み書き可能・作成モードでデータベースを開きます。
	dsn := fmt.Sprintf("file:%s?mode=rwc", dbPath)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// 接続を確認します。
	if err = conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// スキーマファイルを読み込みます。
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read schema.sql: %w", err)
	}

	// スキーマを実行してテーブルを作成します。
	// トランザクション内で実行することもできますが、ここでは単純にExecします。
	if _, err = conn.Exec(string(schema)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to execute schema: %w", err)
	}

	fmt.Println("Database initialized successfully with modernc.org/sqlite.")
	return conn, nil
}