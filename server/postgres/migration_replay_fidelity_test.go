package postgres

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// schemaSnapshot fingerprints one schema's tables so a replay divergence is
// reported as concrete missing or mismatched objects.
type schemaSnapshot map[string]schemaTable

type schemaTable struct {
	columns     []string
	constraints []string
	indexes     []string
}

// The namespace predicate selects either the transaction's temp schema or the
// goose-applied public schema. Table names are matched by name instead of
// regclass so temp tables never shadow the schema being read.
const (
	schemaColumnsQuery = `SELECT a.attname || '|' || format_type(a.atttypid, a.atttypmod) || '|' ||
    CASE WHEN a.attnotnull THEN 'not null' ELSE 'null' END || '|' ||
    COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '')
FROM pg_attribute a
JOIN pg_class c ON c.oid = a.attrelid
LEFT JOIN pg_attrdef ad ON ad.adrelid = a.attrelid AND ad.adnum = a.attnum
WHERE c.relnamespace = %[1]s AND c.relname = $1 AND a.attnum > 0 AND NOT a.attisdropped
ORDER BY a.attname`

	schemaConstraintsQuery = `SELECT con.conname || '|' || con.contype::text || '|' ||
    CASE WHEN con.convalidated THEN 'validated' ELSE 'not valid' END || '|' ||
    pg_get_constraintdef(con.oid)
FROM pg_constraint con
JOIN pg_class c ON c.oid = con.conrelid
WHERE c.relnamespace = %[1]s AND c.relname = $1
ORDER BY con.conname`

	schemaIndexesQuery = `SELECT ic.relname || '|' || pg_get_indexdef(i.indexrelid)
FROM pg_index i
JOIN pg_class c ON c.oid = i.indrelid
JOIN pg_class ic ON ic.oid = i.indexrelid
WHERE c.relnamespace = %[1]s AND c.relname = $1
ORDER BY ic.relname`
)

// TestMigrationReplayMatchesAppliedSchema proves the temp-table replay in
// migration_helpers_test.go builds the same tables, columns, constraints, and
// indexes goose applied. The literal "CREATE TABLE " replacement can drift from
// the migrations it replays, and a drifted replay would make every
// server/postgres test exercise a schema production does not have. The two
// schemas are read from separate sessions because temp tables shadow the public
// schema for name resolution.
func TestMigrationReplayMatchesAppliedSchema(t *testing.T) {
	ctx, _, replayTx := newMigrationTestRepository(t)
	_, appliedTx := newMigrationTestTransaction(t)

	replayed := readSchemaSnapshot(t, ctx, replayTx, true)
	applied := readSchemaSnapshot(t, ctx, appliedTx, false)

	if differences := diffSchemaSnapshots(applied, replayed); len(differences) > 0 {
		t.Fatalf("temp replay diverges from the goose-applied schema:\n%s", strings.Join(differences, "\n"))
	}
}

// readSchemaSnapshot reads every table fingerprint in the temp schema or the
// public schema, skipping goose's own bookkeeping table.
func readSchemaSnapshot(t *testing.T, ctx context.Context, tx pgx.Tx, temp bool) schemaSnapshot {
	t.Helper()
	namespace := "'public'::regnamespace"
	if temp {
		namespace = "pg_my_temp_schema()"
	}
	rows, err := tx.Query(ctx, fmt.Sprintf(`SELECT c.relname FROM pg_class c
WHERE c.relnamespace = %s AND c.relkind = 'r' AND c.relname <> 'goose_db_version'
ORDER BY c.relname`, namespace))
	if err != nil {
		t.Fatal(err)
	}
	tables := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	snapshot := schemaSnapshot{}
	for _, table := range tables {
		indexes := schemaObjectList(t, ctx, tx, namespace, table, schemaIndexesQuery)
		for i := range indexes {
			indexes[i] = normalizeIndexDefinition(indexes[i])
		}
		snapshot[table] = schemaTable{
			columns:     schemaObjectList(t, ctx, tx, namespace, table, schemaColumnsQuery),
			constraints: schemaObjectList(t, ctx, tx, namespace, table, schemaConstraintsQuery),
			indexes:     indexes,
		}
	}
	return snapshot
}

// normalizeIndexDefinition removes the schema qualifier that
// pg_get_indexdef always emits: the same index reads public.* in the
// goose-applied schema and pg_temp.* in the replay.
func normalizeIndexDefinition(definition string) string {
	const on = " ON "
	marker := strings.Index(definition, on)
	if marker < 0 {
		return definition
	}
	prefix := definition[:marker+len(on)]
	relation := definition[marker+len(on):]
	relationEnd := strings.IndexByte(relation, ' ')
	if relationEnd < 0 {
		return definition
	}
	if dot := strings.IndexByte(relation[:relationEnd], '.'); dot >= 0 {
		relation = relation[dot+1:]
	}
	return prefix + relation
}

// schemaObjectList runs one fingerprint query for a table and returns its
// ordered object descriptions.
func schemaObjectList(t *testing.T, ctx context.Context, tx pgx.Tx, namespace, table, query string) []string {
	t.Helper()
	rows, err := tx.Query(ctx, fmt.Sprintf(query, namespace), table)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	objects := []string{}
	for rows.Next() {
		var object string
		if err := rows.Scan(&object); err != nil {
			t.Fatal(err)
		}
		objects = append(objects, object)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return objects
}

// diffSchemaSnapshots reports every table, column, constraint, and index that
// differs between the goose-applied schema and the temp replay.
func diffSchemaSnapshots(applied, replayed schemaSnapshot) []string {
	var differences []string
	for table, appliedTable := range applied {
		replayedTable, ok := replayed[table]
		if !ok {
			differences = append(differences, "table "+table+" is missing from the temp replay")
			continue
		}
		differences = append(differences, diffSchemaList(table+" columns", appliedTable.columns, replayedTable.columns)...)
		differences = append(differences, diffSchemaList(table+" constraints", appliedTable.constraints, replayedTable.constraints)...)
		differences = append(differences, diffSchemaList(table+" indexes", appliedTable.indexes, replayedTable.indexes)...)
	}
	for table := range replayed {
		if _, ok := applied[table]; !ok {
			differences = append(differences, "table "+table+" exists only in the temp replay")
		}
	}
	sort.Strings(differences)
	return differences
}

// diffSchemaList reports objects present on only one side of the comparison.
func diffSchemaList(label string, applied, replayed []string) []string {
	appliedSet := map[string]bool{}
	for _, object := range applied {
		appliedSet[object] = true
	}
	replayedSet := map[string]bool{}
	for _, object := range replayed {
		replayedSet[object] = true
	}
	var differences []string
	for _, object := range applied {
		if !replayedSet[object] {
			differences = append(differences, label+": goose-applied has "+object+", temp replay does not")
		}
	}
	for _, object := range replayed {
		if !appliedSet[object] {
			differences = append(differences, label+": temp replay has "+object+", goose-applied does not")
		}
	}
	return differences
}
