package sdk

import (
	"context"
	"testing"

	"blotting-consultancy/internal/plugin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	return db
}

func newTestPluginDB(t *testing.T, pluginID string, perms ...plugin.Permission) (*PluginDB, *gorm.DB) {
	t.Helper()
	db := newTestDB(t)
	sandbox := plugin.NewSandbox(pluginID, perms)
	return &PluginDB{pluginID: pluginID, db: db, sandbox: sandbox}, db
}

func TestTablePrefix(t *testing.T) {
	assert.Equal(t, "plg_my-plugin_", tablePrefix("my-plugin"))
}

func TestPrefixedTable(t *testing.T) {
	assert.Equal(t, "plg_my-plugin_items", PrefixedTable("my-plugin", "items"))
}

func TestPluginDB_Table(t *testing.T) {
	pdb, _ := newTestPluginDB(t,"abc-plugin", plugin.PermDatabaseRead)
	assert.Equal(t, "plg_abc-plugin_notes", pdb.Table("notes"))
}

func TestPluginDB_AutoMigrate_Denied(t *testing.T) {
	pdb, _ := newTestPluginDB(t,"test-plugin", plugin.PermDatabaseRead) // no write
	type Row struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}
	err := pdb.AutoMigrate("items", &Row{})
	assert.Error(t, err)
}

func TestPluginDB_CRUD(t *testing.T) {
	pdb, _ := newTestPluginDB(t,"test-plugin", plugin.PermDatabaseRead, plugin.PermDatabaseWrite)

	type Row struct {
		ID    uint   `gorm:"primaryKey;autoIncrement"`
		Label string `gorm:"size:200"`
	}

	// AutoMigrate
	err := pdb.AutoMigrate("rows", &Row{})
	require.NoError(t, err)

	// Create
	row := &Row{Label: "hello"}
	err = pdb.Create("rows", row)
	require.NoError(t, err)
	assert.Greater(t, row.ID, uint(0))

	// First
	var found Row
	err = pdb.First("rows", &found, row.ID)
	require.NoError(t, err)
	assert.Equal(t, "hello", found.Label)

	// Find
	var all []Row
	err = pdb.Find("rows", &all)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	// Update
	err = pdb.Update("rows", map[string]interface{}{"id": row.ID}, map[string]interface{}{"label": "world"})
	require.NoError(t, err)

	var updated Row
	err = pdb.First("rows", &updated, row.ID)
	require.NoError(t, err)
	assert.Equal(t, "world", updated.Label)

	// Delete
	err = pdb.Delete("rows", &Row{}, row.ID)
	require.NoError(t, err)

	var afterDel []Row
	err = pdb.Find("rows", &afterDel)
	require.NoError(t, err)
	assert.Empty(t, afterDel)
}

func TestPluginDB_Find_NoPerm(t *testing.T) {
	pdb, _ := newTestPluginDB(t,"test-plugin") // no perms
	var rows []map[string]any
	err := pdb.Find("rows", &rows)
	assert.Error(t, err)
}

func TestPluginDB_ListTables(t *testing.T) {
	pdb, _ := newTestPluginDB(t,"myplugin", plugin.PermDatabaseRead, plugin.PermDatabaseWrite)

	type Row struct {
		ID uint `gorm:"primaryKey;autoIncrement"`
	}
	require.NoError(t, pdb.AutoMigrate("things", &Row{}))
	require.NoError(t, pdb.AutoMigrate("stuff", &Row{}))

	tables, err := pdb.ListTables(context.Background())
	require.NoError(t, err)

	assert.Contains(t, tables, "plg_myplugin_things")
	assert.Contains(t, tables, "plg_myplugin_stuff")
}

func TestPluginDB_DropTable(t *testing.T) {
	pdb, _ := newTestPluginDB(t,"drop-test", plugin.PermDatabaseRead, plugin.PermDatabaseWrite)

	type Row struct {
		ID uint `gorm:"primaryKey;autoIncrement"`
	}
	require.NoError(t, pdb.AutoMigrate("temp", &Row{}))

	err := pdb.DropTable("temp")
	require.NoError(t, err)

	// table should no longer exist
	tables, err := pdb.ListTables(context.Background())
	require.NoError(t, err)
	assert.NotContains(t, tables, "plg_drop-test_temp")
}

func TestPluginDB_Save(t *testing.T) {
	pdb, _ := newTestPluginDB(t,"save-test", plugin.PermDatabaseRead, plugin.PermDatabaseWrite)

	type Row struct {
		ID    uint   `gorm:"primaryKey;autoIncrement"`
		Label string `gorm:"size:200"`
	}
	require.NoError(t, pdb.AutoMigrate("saverows", &Row{}))

	row := &Row{Label: "initial"}
	require.NoError(t, pdb.Create("saverows", row))

	row.Label = "updated"
	err := pdb.Save("saverows", row)
	require.NoError(t, err)

	var found Row
	require.NoError(t, pdb.First("saverows", &found, row.ID))
	assert.Equal(t, "updated", found.Label)
}
