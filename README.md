## Diagram

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
A shell script is included to watch the output folder changes and automatically open the generated png files, however `inotifywait` is required for Linux distribution. To install it under Linux please run:

```bash
sudo apt install inotify-tools
```

Then you can use the provided shell script by typing `$ ./watch`.

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
<kbd>Home</kbd>                         | Jump to the start line
<kbd>End</kbd>                          | Jump to the end line
<kbd>Ctrl+c</kbd>                       | Quit

### Example
| Input | Output |
|:--:|:--:|
| <img src="https://user-images.githubusercontent.com/883386/29396424-9200a978-8320-11e7-9c60-17d2be989136.png" height="300"> | <img src="https://user-images.githubusercontent.com/883386/29396385-529a23a4-8320-11e7-9d70-bf9b33d769cc.png" height="300"> |

## Issues

The app was tested on **Ubuntu** and **MacOS**, but on Mac the panels are not selectables with clicks.

### Acknowledgements
The ascii -> png conversion was ported from [shaky.dart](https://github.com/mraleph/moe-js/blob/master/talks/jsconfeu2012/tools/shaky/web/shaky.dart).

## License

This project is under the MIT License. See the [LICENSE](https://github.com/esimov/diagram/blob/master/LICENSE) file for the full license text.
