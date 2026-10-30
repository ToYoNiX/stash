package app

/*
	The idea behind this package is I didn't want the items
	package to import something from containers package.
	I am learning more about code design and patterns so I
	wanted to experiment with this and see how far I will 
	get using it and what I will face.
*/

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ToYoNiX/stash/internal/containers"
	"github.com/ToYoNiX/stash/internal/items"
)

/*
	When adding an item you need the container it will live in.
	Since we can't import anything from container to item to 
	check if the container exists this lives here.
*/
func AddItem(db *sql.DB, i *items.Item) (int, error) {
	if i.ContainerId == nil {
		return 0, errors.New("Parent container not provided")
	}

	exists, err := containers.DoesExist(db, *i.ContainerId)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("Container %d not found", *i.ContainerId)
	}

	return items.Insert(db, i)
}
