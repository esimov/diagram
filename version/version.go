package version

import "github.com/esimov/diagram/color"

// Name of application.
const Name = "diagram"

// Description of application.
const Description = " ...transforms your ASCII arts into hand drawn diagrams!"

// Version number.
const Version = "v1.0.4"

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
	logo += color.String(c, "  ╚═════╝ ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝ "+Version)
	logo += "\n\n\n\n"
	logo += color.String(c, Description)

	return logo
}
