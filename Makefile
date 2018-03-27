all: 
	@./build.sh
clean:
	@rm -f diagram
install: all
	@cp diagram /usr/local/bin
uninstall: 
	@rm -f /usr/local/bin/diagram
package:
	@NOCOPY=1 ./build.sh package