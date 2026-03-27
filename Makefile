.PHONY: build clean

build:
	@cp VERSION src/VERSION
	@cd src && go build -o ../bin/makedo .
	@rm src/VERSION

clean:
	@rm -f bin/makedo
	@rm -f src/VERSION
