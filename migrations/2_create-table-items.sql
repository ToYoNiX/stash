CREATE TABLE IF NOT EXISTS items (
	item_id INTEGER NOT NULL PRIMARY KEY,
	c_id INT NOT NULL,
	name TEXT,
	description TEXT,
	amount INT,
	img_path TEXT,
	FOREIGN KEY (c_id) REFERENCES containers(c_id)
);

