package db

import (
	"database/sql"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
)

type HiddenType int32

const (
	Normal HiddenType = iota
	Hidden
	GeneratedVirtual
	GeneratedStored
)

var _ sql.Scanner = (*HiddenType)(nil)
var _ huma.SchemaProvider = Normal

func (h *HiddenType) Scan(src any) error {
	if src == nil {
		*h = 0
		return nil
	}

	var num HiddenType
	switch s := src.(type) {
	case int32:
		num = HiddenType(s)
	default:
		return fmt.Errorf("unsupported type %T for Variant", src)
	}

	switch num {
	case Normal, Hidden, GeneratedVirtual, GeneratedStored:
		*h = num
	default:
		return fmt.Errorf("invalid database variant value: %q", num)
	}
	return nil
}

func (HiddenType) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: huma.TypeString,
		Enum: []any{
			Normal,
			Hidden,
			GeneratedVirtual,
			GeneratedStored,
		},
	}
}

type ForeignKeyAction string

const (
	NoAction   ForeignKeyAction = "NoAction"
	Restrict   ForeignKeyAction = "Restrict"
	SetNull    ForeignKeyAction = "SetNull"
	SetDefault ForeignKeyAction = "SetDefault"
	Cascade    ForeignKeyAction = "Cascade"
)

var _ sql.Scanner = (*ForeignKeyAction)(nil)
var _ huma.SchemaProvider = NoAction

func (f *ForeignKeyAction) Scan(src any) error {
	if src == nil {
		*f = ""
		return nil
	}

	var str string
	switch s := src.(type) {
	case []byte:
		str = string(s)
	case string:
		str = s
	default:
		return fmt.Errorf("unsupported type %T for Variant", src)
	}

	switch str {
	case "NO ACTION":
		*f = NoAction
	case "RESTRICT":
		*f = Restrict
	case "SET NULL":
		*f = SetNull
	case "SET DEFAULT":
		*f = SetDefault
	case "CASCADE":
		*f = Cascade
	default:
		return fmt.Errorf("invalid database variant value: %q", str)
	}

	return nil
}

func (ForeignKeyAction) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: huma.TypeString,
		Enum: []any{
			NoAction,
			Restrict,
			SetNull,
			SetDefault,
			Cascade,
		},
	}
}

const schemaDiscoveryQuery = `
	WITH unique_columns AS (
		SELECT
			sm.name AS table_name,
			ii.name AS column_name
		FROM sqlite_schema AS sm
		JOIN pragma_index_list(sm.name) AS il
		JOIN pragma_index_info(il.name) AS ii
		WHERE il."unique" = 1
			AND il.origin != 'pk'
			AND il.partial = 0
		GROUP BY
			sm.name,
			il.name
		HAVING COUNT(*) = 1
	)

	SELECT
		s.name AS tableName,
		c."cid" AS columnId,
		c.name AS columnName,
		c."type" AS columnType,

		CASE
			WHEN c."notnull" = 1 THEN 0
			ELSE 1
		END AS nullable,

		c."dflt_value" AS defaultValue,

		c."pk" AS primaryKeyPosition,

		CASE
			WHEN c."pk" > 0
				AND UPPER(TRIM(c."type")) = 'INTEGER'
				AND s.sql LIKE '%AUTOINCREMENT%'
			THEN 1
			ELSE 0
		END AS isAutoIncrement,

		CASE
			WHEN c."pk" > 0 THEN 1
			WHEN uc.column_name IS NOT NULL THEN 1
			ELSE 0
		END AS isUnique,

		c."hidden" AS hiddenType,

		fk."table" AS referencesTable,
		fk."to" AS referencesColumn,
		fk.on_delete AS onDeleteAction,
		fk.on_update AS onUpdateAction,
		fk.match AS matchType

	FROM sqlite_schema AS s

	JOIN pragma_table_xinfo(s.name) AS c

	LEFT JOIN pragma_foreign_key_list(s.name) AS fk
		ON fk."from" = c.name

	LEFT JOIN unique_columns AS uc
		ON uc.table_name = s.name
		AND uc.column_name = c.name

	WHERE s.type = 'table'
		AND s.name NOT LIKE 'sqlite_%'

	ORDER BY
		s.name,
		c."cid";
`
