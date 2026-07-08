package migrations

import "github.com/zhenruyan/postgrebase/dbx"

func init() {
	AppMigrations.Register(func(db dbx.Builder) error {
		driver := db.DriverName()

		sessionTable := `
			CREATE TABLE IF NOT EXISTS {{_pb_agent_sessions_}} (
				[[id]]           ` + agentIdType(driver) + ` NOT NULL PRIMARY KEY,
				[[project_id]]   ` + agentStringType(driver) + ` NOT NULL,
				[[name]]         ` + agentStringType(driver) + ` NOT NULL,
				[[provider]]     ` + agentStringType(driver) + ` NOT NULL,
				[[model]]        ` + agentStringType(driver) + ` NOT NULL,
				[[name_locked]]  ` + agentBoolType(driver) + ` NOT NULL DEFAULT ` + agentBoolDefault(driver) + `,
				[[last_message]] ` + agentLongTextType(driver) + ` NOT NULL,
				[[created]]      ` + agentTsType(driver) + ` NOT NULL,
				[[updated]]      ` + agentTsType(driver) + ` NOT NULL
			);`

		messageTable := `
			CREATE TABLE IF NOT EXISTS {{_pb_agent_messages_}} (
				[[id]]         ` + agentIdType(driver) + ` NOT NULL PRIMARY KEY,
				[[session_id]] ` + agentStringType(driver) + ` NOT NULL,
				[[role]]       ` + agentStringType(driver) + ` NOT NULL,
				[[content]]    ` + agentLongTextType(driver) + ` NOT NULL,
				[[images]]     ` + agentJsonType(driver) + `,
				[[created]]    ` + agentTsType(driver) + ` NOT NULL,
				[[updated]]    ` + agentTsType(driver) + ` NOT NULL
			);`

		auditTable := `
			CREATE TABLE IF NOT EXISTS {{_pb_agent_audit_}} (
				[[id]]             ` + agentIdType(driver) + ` NOT NULL PRIMARY KEY,
				[[session_id]]     ` + agentStringType(driver) + ` NOT NULL,
				[[project_id]]     ` + agentStringType(driver) + ` NOT NULL,
				[[actor]]          ` + agentStringType(driver) + ` NOT NULL DEFAULT '',
				[[tool]]           ` + agentStringType(driver) + ` NOT NULL,
				[[category]]       ` + agentStringType(driver) + ` NOT NULL DEFAULT '',
				[[risk]]           ` + agentStringType(driver) + ` NOT NULL DEFAULT '',
				[[audit_category]] ` + agentStringType(driver) + ` NOT NULL DEFAULT '',
				[[decision]]       ` + agentStringType(driver) + ` NOT NULL DEFAULT '',
				[[reason]]         ` + agentLongTextType(driver) + ` NOT NULL,
				[[status]]         ` + agentStringType(driver) + ` NOT NULL DEFAULT '',
				[[error_msg]]      ` + agentLongTextType(driver) + ` NOT NULL,
				[[created]]        ` + agentTsType(driver) + ` NOT NULL,
				[[updated]]        ` + agentTsType(driver) + ` NOT NULL
			);`

		stmts := []string{sessionTable, messageTable, auditTable}
		stmts = append(stmts,
			`CREATE TABLE IF NOT EXISTS {{_pb_agent_project_configs_}} (
				[[id]]                  `+agentIdType(driver)+` NOT NULL PRIMARY KEY,
				[[project_id]]          `+agentStringType(driver)+` NOT NULL,
				[[default_provider]]    `+agentStringType(driver)+` NOT NULL DEFAULT '',
				[[default_model]]       `+agentStringType(driver)+` NOT NULL DEFAULT '',
				[[allowed_tools]]       `+agentJsonType(driver)+`,
				[[allow_schema_change]] `+agentStringType(driver)+` NOT NULL DEFAULT 'inherit',
				[[approval_policy]]     `+agentStringType(driver)+` NOT NULL DEFAULT 'inherit',
				[[created]]             `+agentTsType(driver)+` NOT NULL,
				[[updated]]             `+agentTsType(driver)+` NOT NULL
			);`,
		)

		for _, stmt := range stmts {
			if _, err := db.NewQuery(stmt).Execute(); err != nil {
				return err
			}
		}

		indexes := []agentIndex{
			{name: "idx_agent_sessions_project", table: "_pb_agent_sessions_", columns: []string{"project_id"}},
			{name: "idx_agent_messages_session", table: "_pb_agent_messages_", columns: []string{"session_id"}},
			{name: "idx_agent_audit_session", table: "_pb_agent_audit_", columns: []string{"session_id"}},
			{name: "idx_agent_project_config", table: "_pb_agent_project_configs_", columns: []string{"project_id"}, unique: true},
		}
		for _, idx := range indexes {
			if err := createAgentIndex(db, idx); err != nil {
				return err
			}
		}

		return nil
	}, func(db dbx.Builder) error {
		for _, table := range []string{"_pb_agent_project_configs_", "_pb_agent_audit_", "_pb_agent_messages_", "_pb_agent_sessions_"} {
			if _, err := db.NewQuery("DROP TABLE IF EXISTS {{" + table + "}}").Execute(); err != nil {
				return err
			}
		}
		return nil
	})
}

func agentIdType(driver string) string {
	if driver == "mysql" {
		return "VARCHAR(36)"
	}
	return "text"
}

func agentStringType(driver string) string {
	if driver == "mysql" {
		return "VARCHAR(255)"
	}
	return "text"
}

func agentLongTextType(driver string) string {
	if driver == "mysql" {
		return "LONGTEXT"
	}
	return "text"
}

func agentJsonType(driver string) string {
	if driver == "mysql" {
		return "JSON"
	}
	return "text"
}

func agentBoolType(driver string) string {
	switch driver {
	case "mysql":
		return "TINYINT(1)"
	case "sqlite", "sqlite3":
		return "INTEGER"
	default:
		return "BOOLEAN"
	}
}

func agentBoolDefault(driver string) string {
	switch driver {
	case "mysql", "sqlite", "sqlite3":
		return "0"
	default:
		return "false"
	}
}

func agentTsType(driver string) string {
	if driver == "mysql" {
		return "DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3)"
	}
	if driver == "sqlite" || driver == "sqlite3" {
		return "TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%f', 'now'))"
	}
	return "timestamp DEFAULT now()::TIMESTAMP"
}

type agentIndex struct {
	name    string
	table   string
	columns []string
	unique  bool
}

func createAgentIndex(db dbx.Builder, idx agentIndex) error {
	if db.DriverName() == "mysql" {
		var count int
		err := db.NewQuery(`
			SELECT COUNT(1)
			FROM information_schema.statistics
			WHERE table_schema = DATABASE()
				AND table_name = {:table}
				AND index_name = {:index}
		`).Bind(dbx.Params{
			"table": idx.table,
			"index": idx.name,
		}).Row(&count)
		if err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
	}

	optional := " IF NOT EXISTS"
	if db.DriverName() == "mysql" {
		optional = ""
	}
	unique := ""
	if idx.unique {
		unique = "UNIQUE "
	}
	cols := ""
	for i, col := range idx.columns {
		if i > 0 {
			cols += ", "
		}
		cols += "[[" + col + "]]"
	}

	stmt := "CREATE " + unique + "INDEX" + optional + " [[" + idx.name + "]] ON {{" + idx.table + "}} (" + cols + ")"
	_, err := db.NewQuery(stmt).Execute()
	return err
}
