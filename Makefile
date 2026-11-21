BUILD_DIR := "./bin"
DB_DIR := "./data/"

build:
	mkdir -p ${BUILD_DIR}
	go build -tags "linux" -o ${BUILD_DIR} ./...

clean:
	rm -rf ${DB_DIR}
	rm -rf ${BUILD_DIR}

