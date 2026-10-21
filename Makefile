BUILD_DIR := "./bin"
DB_FILE := "./db.sqlite3"

build:
	mkdir -p ${BUILD_DIR}
	go build -tags "linux" -o ${BUILD_DIR} ./...

clean:
	rm -rf ${DB_FILE}
	rm -rf ${BUILD_DIR}

