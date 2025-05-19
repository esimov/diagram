package version

import "github.com/esimov/diagram/color"

// Name of application.
const Name = "diagram"

// description of application.
const description = " Transform your ASCII arts into hand drawn diagrams."
const author = " Developed by: Endre Simo (https://github.com/esimov)"

// version number.
const version = "v1.1.0"

// DrawLogo draws diagram logo.
func DrawLogo() string {
	var logo string

	c := color.Random(180, 231)

	logo += "\n\n"
	logo += color.String(c, "  ██████╗ ██╗ █████╗  ██████╗ ██████╗  █████╗ ███╗   ███╗\n")
	logo += color.String(c, "  ██╔══██╗██║██╔══██╗██╔════╝ ██╔══██╗██╔══██╗████╗ ████║\n")
	logo += color.String(c, "  ██║  ██║██║███████║██║  ███╗██████╔╝███████║██╔████╔██║\n")
	logo += color.String(c, "  ██║  ██║██║██╔══██║██║   ██║██╔══██╗██╔══██║██║╚██╔╝██║\n")
	logo += color.String(c, "  ██████╔╝██║██║  ██║╚██████╔╝██║  ██║██║  ██║██║ ╚═╝ ██║\n")
	logo += color.String(c, "  ╚═════╝ ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝ "+version)
	logo += "\n\n\n"
	logo += color.String(c, description+"\n")
	logo += color.String(c, author)

	return logo
}
