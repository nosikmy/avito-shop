package e2e

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func GetProjectPath(projectName string) (string, error) {
	currDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("can't get working directory: %w", err)
	}

	splitDir := strings.Split(currDir, string(os.PathSeparator))
	projectIndex := slices.Index(splitDir, projectName)
	if projectIndex == -1 {
		return "", fmt.Errorf("path doesn't contain provided project name: %s", projectName)
	}

	suffix := string(os.PathSeparator) + filepath.Join(splitDir[projectIndex+1:]...)
	newPath, found := strings.CutSuffix(currDir, suffix)
	if !found {
		return "", fmt.Errorf("can't find suffix %s in %s", suffix, currDir)
	}

	return newPath, nil
}

// CreatePostgresDB returns db conn, function to remove docker container and error
func CreatePostgresDB() (*sqlx.DB, func() error, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("can't create pool: %w", err)
	}

	if err := pool.Client.Ping(); err != nil {
		return nil, nil, fmt.Errorf("can't ping docker client: %w", err)
	}

	currDir, err := GetProjectPath("avito-shop")
	if err != nil {
		return nil, nil, fmt.Errorf("can't get project path: %w", err)
	}

	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13",
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432/tcp": {
				{HostPort: "5443"},
			},
		},
		ExposedPorts: []string{"5432/tcp"},
		Mounts: []string{
			fmt.Sprintf("%s/migrations/init.sql:/docker-entrypoint-initdb.d/init.sql", currDir),
		},
		Env: []string{
			fmt.Sprintf("POSTGRES_USER=%s", os.Getenv("DATABASE_USER")),
			fmt.Sprintf("POSTGRES_PASSWORD=%s", os.Getenv("DATABASE_PASSWORD")),
			fmt.Sprintf("POSTGRES_DB=%s", os.Getenv("DATABASE_NAME")),
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, nil, fmt.Errorf("can't start resource: %w", err)
	}

	hostAndPort := res.GetHostPort("5432/tcp")
	dbUrl := getConnectionString(hostAndPort)

	log.Printf("trying to connect to: %s\n", dbUrl)

	var db *sqlx.DB
	pool.MaxWait = 2 * time.Minute
	if err := pool.Retry(func() error {
		db, err = sqlx.Open("postgres", dbUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		return nil, nil, fmt.Errorf("can't connect to docker db: %w", err)
	}

	purgeFunc := func() error {
		return pool.Purge(res)
	}

	return db, purgeFunc, nil
}

func getConnectionString(hostAndPort string) string {
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	name := os.Getenv("DATABASE_NAME")

	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, hostAndPort, name)
}
