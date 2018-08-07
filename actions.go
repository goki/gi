// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	"fmt"
	"log"

	"github.com/goki/gi/oswin/key"
	"github.com/goki/gi/units"
	"github.com/goki/ki"
	"github.com/goki/ki/kit"
)

////////////////////////////////////////////////////////////////////////////////////////
// Action -- for menu items and tool bars

// Action is a button widget that can display a text label and / or an icon
// and / or a keyboard shortcut -- this is what is put in menus and toolbars.
type Action struct {
	ButtonBase
	Data      interface{} `json:"-" xml:"-" desc:"optional data that is sent with the ActionSig when it is emitted"`
	ActionSig ki.Signal   `json:"-" xml:"-" desc:"signal for action -- does not have a signal type, as there is only one type: Action triggered -- data is Data of this action"`
	Shortcut  string      `desc:"optional shortcut keyboard chord to trigger this action -- always window-wide in scope, and should generally not conflict other shortcuts (a log message will be emitted if so).  Shortcuts are processed after all other processing of keyboard input."`
}

var KiT_Action = kit.Types.AddType(&Action{}, ActionProps)

var ActionProps = ki.Props{
	"border-width":     units.NewValue(0, units.Px), // todo: should be default
	"border-radius":    units.NewValue(0, units.Px),
	"border-color":     &Prefs.BorderColor,
	"border-style":     BorderSolid,
	"padding":          units.NewValue(2, units.Px),
	"margin":           units.NewValue(0, units.Px),
	"box-shadow.color": &Prefs.ShadowColor,
	"text-align":       AlignCenter,
	"background-color": &Prefs.ControlColor,
	"color":            &Prefs.FontColor,
	"#icon": ki.Props{
		"width":   units.NewValue(1, units.Em),
		"height":  units.NewValue(1, units.Em),
		"margin":  units.NewValue(0, units.Px),
		"padding": units.NewValue(0, units.Px),
		"fill":    &Prefs.IconColor,
		"stroke":  &Prefs.FontColor,
	},
	"#label": ki.Props{
		"margin":  units.NewValue(0, units.Px),
		"padding": units.NewValue(0, units.Px),
	},
	"#indicator": ki.Props{
		"width":          units.NewValue(1.5, units.Ex),
		"height":         units.NewValue(1.5, units.Ex),
		"margin":         units.NewValue(0, units.Px),
		"padding":        units.NewValue(0, units.Px),
		"vertical-align": AlignBottom,
		"fill":           &Prefs.IconColor,
		"stroke":         &Prefs.FontColor,
	},
	"#shortcut": ki.Props{
		"vertical-align": AlignCenter,
	},
	"#sc-stretch": ki.Props{
		"min-width": units.NewValue(2, units.Em),
	},
	ButtonSelectors[ButtonActive]: ki.Props{
		"background-color": "lighter-0",
	},
	ButtonSelectors[ButtonInactive]: ki.Props{
		"border-color": "highlight-50",
		"color":        "highlight-50",
	},
	ButtonSelectors[ButtonHover]: ki.Props{
		"background-color": "highlight-10",
	},
	ButtonSelectors[ButtonFocus]: ki.Props{
		"border-width":     units.NewValue(2, units.Px),
		"background-color": "samelight-50",
	},
	ButtonSelectors[ButtonDown]: ki.Props{
		"color":            "highlight-90",
		"background-color": "highlight-30",
	},
	ButtonSelectors[ButtonSelected]: ki.Props{
		"background-color": &Prefs.SelectColor,
	},
}

// ButtonWidget interface

func (g *Action) ButtonAsBase() *ButtonBase {
	return &(g.ButtonBase)
}

// Trigger triggers the action signal -- for external activation of action --
// only works if action is not inactive
func (g *Action) Trigger() {
	if g.IsInactive() {
		return
	}
	g.ActionSig.Emit(g.This, 0, g.Data)
}

// trigger action signal
func (g *Action) ButtonRelease() {
	if g.IsInactive() {
		return
	}
	wasPressed := (g.State == ButtonDown)
	updt := g.UpdateStart()
	g.SetButtonState(ButtonActive)
	g.ButtonSig.Emit(g.This, int64(ButtonReleased), nil)
	menOpen := false
	if wasPressed {
		g.ActionSig.Emit(g.This, 0, g.Data)
		g.ButtonSig.Emit(g.This, int64(ButtonClicked), g.Data)
		menOpen = g.OpenMenu()
	}
	if !menOpen && g.IsMenu() && g.Viewport != nil {
		win := g.Viewport.Win
		if win != nil {
			win.ClosePopup(g.Viewport) // in case we are a menu popup -- no harm if not
		}
	}
	g.UpdateEnd(updt)
}

func (g *Action) Init2D() {
	g.Init2DWidget()
	g.ConfigParts()
}

// ConfigPartsAddShortcut adds a menu shortcut, with a stretch space -- only called when needed
func (g *Action) ConfigPartsAddShortcut(config *kit.TypeAndNameList) int {
	config.Add(KiT_Stretch, "sc-stretch")
	scIdx := len(*config)
	config.Add(KiT_Label, "shortcut")
	return scIdx
}

func (g *Action) ConfigPartsShortcut(scIdx int) {
	if scIdx < 0 {
		return
	}
	sc := g.Parts.KnownChild(scIdx).(*Label)
	sclbl := key.ChordShortcut(g.Shortcut)
	if sc.Text != sclbl {
		sc.Text = sclbl
		g.StylePart(Node2D(sc))
		g.StylePart(g.Parts.KnownChild(scIdx - 1).(Node2D)) // also get the stretch
	}
}

func (g *Action) ConfigPartsButton() {
	config, icIdx, lbIdx := g.ConfigPartsIconLabel(string(g.Icon), g.Text)
	indIdx := g.ConfigPartsAddIndicator(&config, false) // default off
	mods, updt := g.Parts.ConfigChildren(config, false) // not unique names
	g.ConfigPartsSetIconLabel(string(g.Icon), g.Text, icIdx, lbIdx)
	g.ConfigPartsIndicator(indIdx)
	if g.Tooltip == "" {
		if g.Shortcut != "" {
			g.Tooltip = fmt.Sprintf("Shortcut: %v", g.Shortcut)
		}
	}
	if mods {
		g.UpdateEnd(updt)
	}
}

func (g *Action) ConfigPartsMenu() {
	config, icIdx, lbIdx := g.ConfigPartsIconLabel(string(g.Icon), g.Text)
	indIdx := g.ConfigPartsAddIndicator(&config, false) // default off
	scIdx := -1
	if indIdx < 0 && g.Shortcut != "" {
		scIdx = g.ConfigPartsAddShortcut(&config)
	} else if g.Shortcut != "" {
		log.Printf("gi.Action shortcut cannot be used on a sub-menu for action: %v\n", g.Text)
	}
	mods, updt := g.Parts.ConfigChildren(config, false) // not unique names
	if mods {
		g.SetProp("max-width", -1) // note: this is needed to get menus to fill out
		// well, but doesn't work in menubars
		g.SetProp("indicator", "widget-wedge-right")
	}
	g.ConfigPartsSetIconLabel(string(g.Icon), g.Text, icIdx, lbIdx)
	g.ConfigPartsIndicator(indIdx)
	g.ConfigPartsShortcut(scIdx)
	if mods {
		g.UpdateEnd(updt)
	}
}

func (g *Action) ConfigParts() {
	ismbar := false
	if g.Par != nil {
		_, ismbar = g.Par.(*MenuBar)
		if ismbar {
			g.Indicator = "none"
			g.SetProp("background-color", "linear-gradient(pref(ControlColor), highlight-10)")
		}
	}
	if g.IsMenu() && !ismbar {
		g.ConfigPartsMenu()
	} else {
		g.ConfigPartsButton()
	}
}
