# Diagram
[![Build Status](https://travis-ci.org/esimov/diagram.svg?branch=master)](https://travis-ci.org/esimov/diagram)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/esimov/diagram)

Diagram is a CLI tool to generate hand drawn diagrams from ASCII arts. 

It's a full featured CLI application which converts the ASCII text into hand drawn diagrams. The CLI part is based on [gocui](https://github.com/jroimartin/gocui) and the ascii to png conversion is done using the [gg](https://github.com/fogleman/gg) library.

![screencast](images/screencast.gif)

## Installation and usage

```bash
$ go get github.com/esimov/diagram
$ go install

# Start the application
$ diagram
```
A shell script is included to watch the output folder and automatically open the generated image files, however `inotifywait` is required for the Linux distribution. Use the following command to install it on Ubuntu:

```bash
sudo apt install inotify-tools
```
Then you can use the provided shell script to activate it by running the following command `$ ./watch`.

**Update:**
*The included shell script is not needed anymore, because an internal image viewer is bundled into the application.*

### Command Line support

The application supports the generation of hand drawn diagrams directly via command line. Typing `$ diagram --help` will show the supported commands for generating the diagrams without to enter the CLI tool:

```bash
Usage of diagram:
  -font string
    	path to font file (default "${GOPATH}/src/github.com/esimov/diagram/font/gloriahallelujah.ttf")
  -in string
    	Source
  -out string
    	Destination
  -preview
    	Show the preview window (default true)
```

#### CLI Examples

Read input from `sample.txt` and write image to `sample.png` showing a preview window with the hand drawn diagram:

```bash
diagram -in sample.txt -out sample.png
```

Read input from `sample.txt` and write image to `sample.png`, and exit immediately without showing a preview window:

```bash
diagram -in sample.txt -out sample.png -preview=false
```

Generate diagram as above but use a font at a different location:

```bash
diagram -in sample.txt -out sample.png -preview=false -font /path/to/my/font/MyHandwriting.ttf
```



### Key bindings
Key                                     | Description
----------------------------------------|---------------------------------------
<kbd>Tab</kbd>                          | Next Panel
<kbd>Shift+Tab</kbd>                    | Previous Panel
<kbd>Ctrl+s</kbd>                       | Open Save Diagram Modal
<kbd>Ctrl+s</kbd>                       | Save Diagram
<kbd>Ctrl+d</kbd>                       | Convert Ascii to PNG
<kbd>Ctrl+x</kbd>                       | Clear the editor content
<kbd>Ctrl+z</kbd>                       | Restore the editor content
<kbd>PageUp</kbd>                       | Jump to the top
<kbd>PageDown</kbd>                     | Jump to the bottom
<kbd>Home</kbd>                         | Jump to the line start
<kbd>End</kbd>                          | Jump to the line end
<kbd>Delete/Backspace</kbd>            | Delete diagram
<kbd>Ctrl+c</kbd>                       | Quit

### Example
| Input | Output |
|:--:|:--:|
| <img src="https://user-images.githubusercontent.com/883386/29396424-9200a978-8320-11e7-9c60-17d2be989136.png" height="300"> | <img src="https://user-images.githubusercontent.com/883386/29396385-529a23a4-8320-11e7-9d70-bf9b33d769cc.png" height="300"> |

## Known issues

The app was tested on **Ubuntu** and **MacOS**, but on Mac the panels are not selectables with clicks.

### Acknowledgements
The ascii to png conversion was ported from [shaky.dart](https://github.com/mraleph/moe-js/blob/master/talks/jsconfeu2012/tools/shaky/web/shaky.dart).

## License

This project is under the MIT License. See the [LICENSE](https://github.com/esimov/diagram/blob/master/LICENSE) file for the full license text.
