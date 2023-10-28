all: install

clean:
	rm -rf gohookr

build:
	go mod tidy
	go build -o gohookr

install: build
	cp gohookr /usr/local/bin/
	cp gohookr.service /usr/lib/systemd/system/
	cp -n config.json /etc/gohookr.json
	systemctl daemon-reload
	systemctl enable --now gohookr

uninstall:
	systemctl disable --now gohookr
	rm -rf /usr/local/bin/gohookr /usr/lib/systemd/system/gohookr.service
	systemctl daemon-reload
