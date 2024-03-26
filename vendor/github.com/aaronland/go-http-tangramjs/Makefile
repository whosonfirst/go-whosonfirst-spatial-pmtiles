GOMOD=readonly

example:
	go run -mod $(GOMOD) cmd/example/main.go -api-key $(APIKEY) -javascript-at-eof -rollup-assets

tangram: 
	curl -s -o static/javascript/tangram.debug.js https://raw.githubusercontent.com/tangrams/tangram/master/dist/tangram.debug.js
	curl -s -o static/javascript/tangram.min.js https://raw.githubusercontent.com/tangrams/tangram/master/dist/tangram.min.js

styles: refill walkabout

refill:
	curl -s -o static/tangram/refill-style.zip https://www.nextzen.org/carto/refill-style/refill-style.zip
	curl -s -o static/tangram/refill-style-themes-label.zip https://www.nextzen.org/carto/refill-style/themes/label-10.zip

walkabout:
	curl -s -o static/tangram/walkabout-style.zip https://www.nextzen.org/carto/refill-style/walkabout-style.zip
