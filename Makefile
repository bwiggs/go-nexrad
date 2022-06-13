test:
	go test -v ./...

nexrad-tui:
	cd cmd/nexrad-tui && go install

nexrad-render:
	cd cmd/nexrad-render && go install