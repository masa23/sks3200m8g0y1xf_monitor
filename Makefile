all: build install

build:
	CGO_ENABLED=0 go build -trimpath -tags netgo

install:
	install sks3200m8g0y1xf_monitor /usr/local/bin
	install config.yaml /usr/local/etc/sks3200m8g0y1xf_monitor.yaml
	install misc/sks3200m8g0y1xf_monitor.service /etc/systemd/system
	systemctl daemon-reload

uninstall:
	rm -f /usr/local/bin/sks3200m8g0y1xf_monitor
	rm -f /usr/local/etc/sks3200m8g0y1xf_monitor.yaml
	rm -f /etc/systemd/system/sks3200m8g0y1xf_monitor.service
	systemctl daemon-reload

clean:
	go clean
