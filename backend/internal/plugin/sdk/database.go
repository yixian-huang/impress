package sdk

import (
	"context"
	"fmt"
	"strings"

	"blotting-consultancy/internal/plugin"

	"gorm.io/gorm"
)

// tablePrefix returns the table name prefix for a given plugin ID.
// All plugin-owned tables are prefixed with "plg_<pluginID>_" to avoid
// collisions with core CMS tables and between plugins.
func tablePrefix(pluginID string) string {
	return "plg_" + pluginID + "_"
}

// PrefixedTable returns the fully-qualified table name for a plugin-scoped table.
func PrefixedTable(pluginID, table string) string {
	return tablePrefix(pluginID) + table
}

// PluginDB wraps a GORM DB instance and enforces per-plugin table name prefixing.
// All CRUD helpers automatically scope queries to the plugin's table namespace.
type PluginDB struct {
	pluginID string
	db       *gorm.DB
	sandbox  *plugin.Sandbox
}

// newPluginDB creates a new PluginDB. The sandbox is used to gate write operations.
func newPluginDB(pluginID string, db *gorm.DB, sandbox *plugin.Sandbox) *PluginDB {
	return &PluginDB{pluginID: pluginID, db: db, sandbox: sandbox}
}

// Table returns the prefixed table name for the given base name.
func (d *PluginDB) Table(name string) string {
	return PrefixedTable(d.pluginID, name)
}

// Raw returns a GORM scoped to the prefixed table for custom queries.
// Only call this for read operations — write operations should use Create/Update/Delete.
func (d *PluginDB) Raw(table string) *gorm.DB {
	return d.db.Table(d.Table(table))
}

// AutoMigrate creates or updates the table schema for the given model pointer.
// The model's TableName() method (if present) is ignored; the table is always
// the prefixed name derived from baseName.
func (d *PluginDB) AutoMigrate(baseName string, model interface{}) error {
	if err := d.sandbox.Check(plugin.PermDatabaseWrite); err != nil {
		return fmt.Errorf("AutoMigrate: %w", err)
	}
	return d.db.Table(d.Table(baseName)).AutoMigrate(model)
}

// Create inserts a new record into the plugin-scoped table.
func (d *PluginDB) Create(table string, value interface{}) error {
	if err := d.sandbox.Check(plugin.PermDatabaseWrite); err != nil {
		return fmt.Errorf("Create: %w", err)
	}
	return d.db.Table(d.Table(table)).Create(value).Error
}

// Save saves (upsert by primary key) a record into the plugin-scoped table.
func (d *PluginDB) Save(table string, value interface{}) error {
	if err := d.sandbox.Check(plugin.PermDatabaseWrite); err != nil {
		return fmt.Errorf("Save: %w", err)
	}
	return d.db.Table(d.Table(table)).Save(value).Error
}

// First finds the first record matching the conditions into dest.
func (d *PluginDB) First(table string, dest interface{}, conds ...interface{}) error {
	if err := d.sandbox.Check(plugin.PermDatabaseRead); err != nil {
		return fmt.Errorf("First: %w", err)
	}
	return d.db.Table(d.Table(table)).First(dest, conds...).Error
}

// Find retrieves all records matching the conditions into dest.
func (d *PluginDB) Find(table string, dest interface{}, conds ...interface{}) error {
	if err := d.sandbox.Check(plugin.PermDatabaseRead); err != nil {
		return fmt.Errorf("Find: %w", err)
	}
	return d.db.Table(d.Table(table)).Find(dest, conds...).Error
}

// Where returns a scoped GORM DB for chained query building on the prefixed table.
// The caller is responsible for checking read permissions before calling this.
func (d *PluginDB) Where(table string, query interface{}, args ...interface{}) (*gorm.DB, error) {
	if err := d.sandbox.Check(plugin.PermDatabaseRead); err != nil {
		return nil, fmt.Errorf("Where: %w", err)
	}
	return d.db.Table(d.Table(table)).Where(query, args...), nil
}

// Update modifies columns on records matching the condition.
func (d *PluginDB) Update(table string, condition map[string]interface{}, updates map[string]interface{}) error {
	if err := d.sandbox.Check(plugin.PermDatabaseWrite); err != nil {
		return fmt.Errorf("Update: %w", err)
	}
	return d.db.Table(d.Table(table)).Where(condition).Updates(updates).Error
}

// Delete removes records from the plugin-scoped table matching the condition.
func (d *PluginDB) Delete(table string, model interface{}, conds ...interface{}) error {
	if err := d.sandbox.Check(plugin.PermDatabaseWrite); err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	return d.db.Table(d.Table(table)).Delete(model, conds...).Error
}

// Count returns the number of records in the plugin-scoped table matching the condition.
func (d *PluginDB) Count(table string, model interface{}, conds ...interface{}) (int64, error) {
	if err := d.sandbox.Check(plugin.PermDatabaseRead); err != nil {
		return 0, fmt.Errorf("Count: %w", err)
	}
	var count int64
	err := d.db.Table(d.Table(table)).Model(model).Find(nil, conds...).Count(&count).Error
	return count, err
}

// DropTable removes a plugin-scoped table (for uninstall cleanup).
func (d *PluginDB) DropTable(table string) error {
	if err := d.sandbox.Check(plugin.PermDatabaseWrite); err != nil {
		return fmt.Errorf("DropTable: %w", err)
	}
	// Quote the table name with backticks for SQLite / double-quotes for other dialects.
	// GORM's Migrator.DropTable handles this correctly for all supported dialects.
	return d.db.Migrator().DropTable(d.Table(table))
}

// ListTables returns all table names owned by this plugin (prefixed with the plugin's prefix).
// This is useful during uninstall to clean up all plugin data.
func (d *PluginDB) ListTables(ctx context.Context) ([]string, error) {
	if err := d.sandbox.Check(plugin.PermDatabaseRead); err != nil {
		return nil, fmt.Errorf("ListTables: %w", err)
	}

	prefix := tablePrefix(d.pluginID)
	var tables []string

	// GORM Migrator can list tables; fall back to a raw query if needed.
	migrator := d.db.WithContext(ctx).Migrator()
	all, err := migrator.GetTables()
	if err != nil {
		return nil, fmt.Errorf("ListTables: failed to query tables: %w", err)
	}

	for _, t := range all {
		if strings.HasPrefix(t, prefix) {
			tables = append(tables, t)
		}
	}
	return tables, nil
}
