package containers

import (
	"database/sql"
	"errors"
)

/*
	So, the original reason for me to use pointers is to check
	weither the value was provided with the invokation or not.

	But... I learned the following
	1. sql ids start by default at 1 not 0 and so as go uninitialied int.
	2. I thought of pointers in this scenario act like cpp references, 
	   but I was wrong and it is a bit deeper than I initially thought.
	   So, this would be a new exploration for me. but for not it works.
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
