build:
	gomobile bind -v -target=android -o mobile.aar -tags notray .

clean:
	rm mobile-sources.jar mobile.aar

tools:
	go install golang.org/x/mobile/cmd/gomobile@latest
	gomobile init
