package version

import "github.com/esimov/diagram/color"

// Name of application
const Name = "diagram"

// Description of application
const Description = "Transform ASCII arts into hand drawn diagrams"

// Version number
const Version = "v.1.0.1"

//Colorize logo
func DrawLogo() string {
	var logo string

	logo += "\n\n"
	logo += color.StringRandom("  ██████╗ ██╗ █████╗  ██████╗ ██████╗  █████╗ ███╗   ███╗\n")
	logo += color.StringRandom("  ██╔══██╗██║██╔══██╗██╔════╝ ██╔══██╗██╔══██╗████╗ ████║\n")
	logo += color.StringRandom("  ██║  ██║██║███████║██║  ███╗██████╔╝███████║██╔████╔██║\n")
	logo += color.StringRandom("  ██║  ██║██║██╔══██║██║   ██║██╔══██╗██╔══██║██║╚██╔╝██║\n")
	logo += color.StringRandom("  ██████╔╝██║██║  ██║╚██████╔╝██║  ██║██║  ██║██║ ╚═╝ ██║\n")
	logo += color.StringRandom("  ╚═════╝ ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝ " + Version)

	return logo
}