package items

import (
	"database/sql"
	"errors"
)

type Item struct {
	Id          *int
	ContainerId *int
	Name        *string
	Description *string
	Amount      *int
	ImgPath     *string
}

// Insert stores the item and returns its new id.
func Insert(db *sql.DB, i *Item) (int, error) {
	// Validate if name, desc, amount, and containerId are missing
	if i.Name == nil {
		return 0, errors.New("Name not provided")
	}

	if i.Description == nil {
		return 0, errors.New("Description not provided")
	}

	if i.ContainerId == nil {
		return 0, errors.New("Parent container not provided")
	}

	if i.Amount == nil {
		return 0, errors.New("Amount not provided")
	}

	// Check if amount is negative number
	if *i.Amount < 0 {
		return 0, errors.New("Amount can't be negative")
	}

	sqlStmt := `
	INSERT
	INTO items(name, description, amount, img_path, c_id)
	VALUES(?, ?, ?, ?, ?);
	`

	res, err := db.Exec(sqlStmt, i.Name, i.Description, i.Amount, i.ImgPath, i.ContainerId)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func Remove(db *sql.DB, i *Item) error {
	// Valudate if id is provided
	if i.Id == nil {
		return errors.New("Id not provided")
	}

	sqlStmt := `
	DELETE
	FROM items
	WHERE item_id = ?;
	`
	_, err := db.Exec(sqlStmt, i.Id)
	if err != nil {
		return err
	}

	return nil
}
