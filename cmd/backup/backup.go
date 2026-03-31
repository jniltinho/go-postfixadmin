package backup

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

var skipDatabases = map[string]bool{
	"Database":           true,
	"information_schema": true,
	"test":               true,
	"performance_schema": true,
}

func mysqlCmdError(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
		return fmt.Errorf("%w\n%s", err, strings.TrimSpace(string(exitErr.Stderr)))
	}
	return err
}

// mysqlEnv returns the environment with MYSQL_PWD set to avoid exposing the
// password via -p flag (which shows up in process listings and logs).
func mysqlEnv(pass string) []string {
	return append(os.Environ(), "MYSQL_PWD="+pass)
}

func mysqlDatabases(cfg Config) ([]string, error) {
	cmd := exec.Command("mysql",
		"-u", cfg.MySQLUser,
		"-h", cfg.MySQLHost, "-P", cfg.MySQLPort, "--silent", "-N",
		"-e", "show databases",
	)
	cmd.Env = mysqlEnv(cfg.MySQLPass)
	out, err := cmd.Output()
	if err != nil {
		return nil, mysqlCmdError(err)
	}

	var dbs []string
	for _, db := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		db = strings.TrimSpace(db)
		if db != "" && !skipDatabases[db] {
			dbs = append(dbs, db)
		}
	}
	return dbs, nil
}

// ExecBackup creates compressed .sql.gz backups for every non-system database.
func ExecBackup(cfg Config) {
	if err := os.MkdirAll(cfg.BackupDir, 0755); err != nil {
		LogMsg(fmt.Sprintf("ERROR creating backup directory: %s", err))
		return
	}

	dbs, err := mysqlDatabases(cfg)
	if err != nil {
		LogMsg(fmt.Sprintf("ERROR listing databases: %s", err))
		return
	}

	for _, db := range dbs {
		filename := filepath.Join(cfg.BackupDir, fmt.Sprintf("%s-%s.sql.gz", db, Filestamp))
		LogMsg(fmt.Sprintf("BACKUP database: %s file: %s", db, filename))

		shellCmd := fmt.Sprintf(
			"mysqldump -u %s -h %s -P %s -e --opt -B -R -c %s | gzip -c > %s",
			cfg.MySQLUser, cfg.MySQLHost, cfg.MySQLPort, db, filename,
		)
		cmd := exec.Command("bash", "-c", shellCmd)
		cmd.Env = mysqlEnv(cfg.MySQLPass)
		result, _ := cmd.CombinedOutput()
		if len(result) > 0 {
			LogMsg(string(result))
		}
	}
}

func mysqlDatabaseSizes(cfg Config) (map[string]string, error) {
	query := "SELECT table_schema, ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) " +
		"FROM information_schema.TABLES GROUP BY table_schema;"
	cmd := exec.Command("mysql",
		"-u", cfg.MySQLUser,
		"-h", cfg.MySQLHost, "-P", cfg.MySQLPort, "--silent", "-N",
		"-e", query,
	)
	cmd.Env = mysqlEnv(cfg.MySQLPass)
	out, err := cmd.Output()
	if err != nil {
		return nil, mysqlCmdError(err)
	}

	sizes := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.Split(strings.TrimSpace(line), "\t")
		if len(parts) == 2 {
			sizes[parts[0]] = parts[1] + " MB"
		}
	}
	return sizes, nil
}

// ListDatabases prints a formatted table of all non-system databases and their sizes.
func ListDatabases(cfg Config) {
	dbs, err := mysqlDatabases(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR listing databases: %s\n", err)
		return
	}

	sizes, err := mysqlDatabaseSizes(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR fetching database sizes: %s\n", err)
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Database", "Size"})
	for i, db := range dbs {
		t.AppendRow(table.Row{i + 1, db, sizes[db]})
	}
	t.Render()
}
