build:
	gomobile bind -v -target=android -o mobile.aar -tags notray .

android: build
	cd android ; ./gradlew bundle

clean:
	rm mobile-sources.jar mobile.aar

tools:
	go install golang.org/x/mobile/cmd/gomobile@latest
	gomobile init
