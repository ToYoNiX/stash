package containers

import (
	"database/sql"
	"errors"
)

/*
	Two reason I made the whole Container structure with pointers
		1. if the value is nil it would insert NULL
		2. to make the code a tad bit faster since we don't
			have to copy as much elements
*/

type Container struct {
	Id          *int
	Name        *string
	Description *string
	ImgPath     *string
	Parent      *int
}

func DoesExist(db *sql.DB, id int) (bool, error) {
	query := `
	SELECT c_id FROM containers WHERE c_id = ?;
	`
	// Scan returns sql.ErrNoRows if no record exists. Any other error
	// means the query itself failed, so we hand it to the caller
	// instead of pretending the container is missing.
	var temp int
	err := db.QueryRow(query, id).Scan(&temp)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// Insert stores the container and returns its new id.
func Insert(db *sql.DB, c *Container) (int, error) {
	// Validate if name, and desc are missing
	if c.Name == nil {
		return 0, errors.New("Name not provided")
	}

	if c.Description == nil {
		return 0, errors.New("Description not provided")
	}

	// Check if parent exits
	if c.Parent != nil {
		exists, err := DoesExist(db, *c.Parent)
		if err != nil {
			return 0, err
		}
		if !exists {
			return 0, errors.New("Provided parent not found")
		}
	}

	sqlStmt := `
	INSERT
	INTO containers(name, description, img_path, c_parent_id)
	VALUES(?, ?, ?, ?);
	`

	res, err := db.Exec(sqlStmt, c.Name, c.Description, c.ImgPath, c.Parent)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func Remove(db *sql.DB, c *Container) error {
	// Validate if there is an id provided
	if c.Id == nil {
		return errors.New("Id not provided")
	}

	sqlStmt := `
	DELETE 
	FROM containers
	WHERE c_id = ?;
	`

	_, err := db.Exec(sqlStmt, c.Id)
	if err != nil {
		return err
	}

	return nil
}
