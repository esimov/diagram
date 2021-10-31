module github.com/esimov/diagram

require (
	github.com/fogleman/gg v1.0.1-0.20180308184255-c97f757e6f0e
	github.com/go-gl/gl v0.0.0-20180304232605-eafa86a81d97
	github.com/go-gl/glfw v0.0.0-20170814180746-513e4f2bf85c
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/google/gxui v0.0.0-20151028112939-f85e0a97b3a4
	github.com/gopherjs/gopherjs v0.0.0-20180314020201-768621c88e58
	github.com/goxjs/gl v0.0.0-20171128034433-dc8f4a9a3c9c
	github.com/goxjs/glfw v0.0.0-20171018044755-7dec05603e06
	github.com/jroimartin/gocui v0.3.1-0.20170827195011-4f518eddb04b
	github.com/mattn/go-runewidth v0.0.3-0.20180304235428-a9d6d1e4dc51
	github.com/nsf/termbox-go v0.0.0-20180303152453-e2050e41c884
	golang.org/x/image v0.0.0-20180314180248-f3a9b89b59de
	honnef.co/go/js/dom v0.0.0-20180307180539-662b7b8f3412
)

go 1.13

replace github.com/google/gxui => ./vendor/github.com/google/gxui/
