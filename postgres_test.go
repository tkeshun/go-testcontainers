package test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDatabaseOperations(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer postgresContainer.Terminate(ctx)

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatal(err)
	}

	dsn := fmt.Sprintf("postgres://testuser:testpassword@%s:%s/testdb", host, port.Port())
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)

	// テーブル作成
	_, err = conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT NOT NULL)")
	if err != nil {
		t.Fatal(err)
	}

	log.Println("++++テーブル作成完了++++")

	// INSERTテスト
	_, err = conn.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "TEST")
	if err != nil {
		t.Fatal(err)
	}

	log.Println("++++データ挿入完了++++")

	// SELECTテスト
	var name string
	err = conn.QueryRow(ctx, "SELECT name FROM users WHERE name=$1", "TEST").Scan(&name)
	if err != nil {
		t.Fatal(err)
	}

	if name != "TEST" {
		t.Fatalf("expected name to be 'Alice', got '%s'", name)
	}

	log.Println("++++データ検索完了++++")

	// DELETEテスト
	_, err = conn.Exec(ctx, "DELETE FROM users WHERE name=$1", "Alice")
	if err != nil {
		t.Fatal(err)
	}

	// 削除確認
	err = conn.QueryRow(ctx, "SELECT name FROM users WHERE name=$1", "Alice").Scan(&name)
	if err == nil {
		t.Fatal("expected no rows to be returned")
	}

	log.Println("++++データ削除完了++++")

	log.Println("Database operations completed successfully!")
}
