package data

import (
	"database/sql"
)

// CREATE TABLE groups (
//     id SERIAL PRIMARY KEY,
//     name VARCHAR(255) NOT NULL,
//     description TEXT,
//     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
//     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
// );
// CREATE TABLE user_groups (
//     id SERIAL PRIMARY KEY,
//     user_id INT NOT NULL,
//     group_id INT NOT NULL,
//     CONSTRAINT fk_user_groups_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
//     CONSTRAINT fk_user_groups_group_id FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE,
//     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
//     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
//     UNIQUE (user_id, group_id)
// );

type GroupMember struct {
	ID    int
	Name  string
	Email string
}

type Group struct {
	ID          int
	Name        string
	Description string
	CreatedAt   string
	UpdatedAt   string
	Members     []*GroupMember
}

type GroupModel struct {
	DB *sql.DB
}

// Insert a new group. Also auto insert the creator as a member.
func (m *GroupModel) Insert(userID int, name, description string) (int, error) {
	query := `
		INSERT INTO groups (name, description)
		VALUES ($1, $2)
		RETURNING id
	`

	var newGroupID int
	err := m.DB.QueryRow(query, name, description).Scan(&newGroupID)
	if err != nil {
		return 0, err
	}

	query = `
		INSERT INTO user_groups (user_id, group_id)
		VALUES ($1, $2)
	`

	_, err = m.DB.Exec(query, userID, newGroupID)
	if err != nil {
		return 0, err
	}

	return newGroupID, nil
}

func (m *GroupModel) Get(id int) (*Group, error) {
	query := `
        SELECT g.id, g.name, g.description, g.created_at, g.updated_at, u.id, u.name, u.email
        FROM groups g
        INNER JOIN user_groups ug ON g.id = ug.group_id
        INNER JOIN users u ON ug.user_id = u.id
        WHERE g.id = $1
    `

	rows, err := m.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	group := &Group{}
	members := []*GroupMember{}

	for rows.Next() {
		member := &GroupMember{}
		err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.CreatedAt, &group.UpdatedAt, &member.ID, &member.Name, &member.Email)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	group.Members = members

	return group, nil
}

// Get groups user is a member of without the members.
func (m *GroupModel) GetAllForUser(userID int) ([]*Group, error) {
	query := `
		SELECT g.id, g.name, g.description, g.created_at, g.updated_at
		FROM groups g
		INNER JOIN user_groups ug ON g.id = ug.group_id
		WHERE ug.user_id = $1
	`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []*Group{}

	for rows.Next() {
		group := &Group{}
		err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.CreatedAt, &group.UpdatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

// Add a new member to a group.
func (m *GroupModel) AddMember(groupID, userID int) error {
	query := `
		INSERT INTO user_groups (user_id, group_id)
		VALUES ($1, $2)
	`

	_, err := m.DB.Exec(query, userID, groupID)
	if err != nil {
		return err
	}

	return nil
}
