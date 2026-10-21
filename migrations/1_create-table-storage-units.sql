CREATE TABLE IF NOT EXISTS containers (
	c_id INTEGER NOT NULL PRIMARY KEY,
	c_parent_id INT,
	name TEXT,
	description TEXT,
	img_path TEXT,
	FOREIGN KEY (c_parent_id) REFERENCES containers(c_id)
);

