package data

import (
	"testing"

	"github.com/tmgasek/calendar-app/internal/assert"
)

func TestGroupModelInsert(t *testing.T) {
	db := newTestDB(t)
	m := GroupModel{DB: db}

	userID := 1
	name := "Test Group"
	description := "This is a test group"

	groupID, err := m.Insert(userID, name, description)
	assert.NilError(t, err)
	assert.Greater(t, groupID, 0)

	// Check if the group and user_group records are inserted correctly
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM groups WHERE id = $1", groupID).Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, count, 1)

	err = db.QueryRow("SELECT COUNT(*) FROM user_groups WHERE user_id = $1 AND group_id = $2", userID, groupID).Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, count, 1)
}

func TestGroupModelGet(t *testing.T) {
	db := newTestDB(t)
	m := GroupModel{DB: db}

	groupID := 1

	group, err := m.Get(groupID)
	assert.NilError(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, group.ID, groupID)
	assert.Equal(t, group.Name, "Group 1")
	assert.Equal(t, group.Description, "Description for Group 1")
	assert.Equal(t, len(group.Members), 2)
	assert.Equal(t, group.Members[0].ID, 1)
	assert.Equal(t, group.Members[0].Name, "Alice Jones")
	assert.Equal(t, group.Members[0].Email, "alice@example.com")
	assert.Equal(t, group.Members[1].ID, 2)
	assert.Equal(t, group.Members[1].Name, "Bob")
	assert.Equal(t, group.Members[1].Email, "bob@example.com")
}

func TestGroupModelGetAllForUser(t *testing.T) {
	db := newTestDB(t)
	m := GroupModel{DB: db}

	userID := 1

	groups, err := m.GetAllForUser(userID)
	assert.NilError(t, err)
	assert.Equal(t, len(groups), 2)
	assert.Equal(t, groups[0].ID, 1)
	assert.Equal(t, groups[0].Name, "Group 1")
	assert.Equal(t, groups[0].Description, "Description for Group 1")
	assert.Equal(t, groups[1].ID, 2)
	assert.Equal(t, groups[1].Name, "Group 2")
	assert.Equal(t, groups[1].Description, "Description for Group 2")
}

func TestGroupModelAddMember(t *testing.T) {
	db := newTestDB(t)
	m := GroupModel{DB: db}

	groupID := 1
	userID := 3

	err := m.AddMember(groupID, userID)
	assert.NilError(t, err)

	// Check if the user_group record is inserted correctly
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM user_groups WHERE user_id = $1 AND group_id = $2", userID, groupID).Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, count, 1)
}
