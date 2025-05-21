package ui

import (
	"fmt"
)

type ButtonWidget struct {
	Widget
}

// NewButton creates a new button widget.
func NewButton[T WidgetEmbedder](ui *UI, groupName, viewName string, posX, posY, width int) (*ButtonWidget, error) {
	button, err := NewWidget(
		&ButtonWidget{Widget{groupName: groupName}},
		[]WidgetOption[*ButtonWidget]{
			WithDefaultWidgetOptions[*ButtonWidget](viewName, posX, posY),
			WithWidgetWidth[*ButtonWidget](width),
			WithUIHandler[*ButtonWidget](ui),
		}...,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create button widget: %w", err)
	}

	return *button, nil
}
