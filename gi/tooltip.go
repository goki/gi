// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	"goki.dev/colors"
	"goki.dev/girl/styles"
	"goki.dev/girl/units"
	"goki.dev/goosi/events"
)

// TooltipConfigStyles configures the default styles
// for the given tooltip frame with the given parent.
// It should be called on tooltips when they are created.
func TooltipConfigStyles(tooltip *Frame) {
	tooltip.AddStyler(func(w *WidgetBase, s *styles.Style) {
		s.Border.Style.Set(styles.BorderNone)
		s.Border.Radius = styles.BorderRadiusExtraSmall
		s.Padding.Set(units.Px(8 * Prefs.DensityMul()))
		s.BackgroundColor.SetSolid(colors.Scheme.InverseSurface)
		s.Color = colors.Scheme.InverseOnSurface
		s.BoxShadow = BoxShadow1 // STYTODO: not sure whether we should have this
	})
}

// PopupTooltip pops up a scene displaying the tooltip text
func PopupTooltip(tooltip string, x, y int, parSc *Scene, name string) *Scene {
	/*
		win := parSc.Win
		mainSc := win.Scene
		psc := &Scene{}
		psc.Name = name + "Tooltip"
		psc.Win = win
		psc.Type = ScTooltip

		psc.Frame.AddStyler(func(w *WidgetBase, s *styles.Style) {
			// TOOD: get border radius actually working
			// without having parent background color workaround
			s.Border.Radius = styles.BorderRadiusExtraSmall
			s.BackgroundColor = psc.Frame.ParentBackgroundColor()
		})

		psc.Geom.Pos = image.Point{x, y}
		psc.SetFlag(true, ScPopupDestroyAll) // nuke it all

		frame := &psc.Frame
		lbl := NewLabel(frame, "ttlbl")
		lbl.Text = tooltip
		lbl.Type = LabelBodyMedium

		TooltipConfigStyles(frame)

		lbl.AddStyler(func(w *WidgetBase, s *styles.Style) {
			mwdots := parSc.Frame.Style.UnContext.ToDots(40, units.UnitEm)
			mwdots = mat32.Min(mwdots, float32(mainSc.Geom.Size.X-20))

			s.MaxWidth.SetDot(mwdots)
		})

		frame.ConfigTree(psc)
		frame.ApplyStyleTree(psc) // sufficient to get sizes
		mainSz := mat32.NewVec2FmPoint(mainSc.Geom.Size)
		frame.LayState.Alloc.Size = mainSz // give it the whole vp initially
		frame.GetSizeTree(psc, 0)          // collect sizes
		psc.Win = nil
		vpsz := frame.LayState.Size.Pref.Min(mainSz).ToPoint()

		x = min(x, mainSc.Geom.Size.X-vpsz.X) // fit
		y = min(y, mainSc.Geom.Size.Y-vpsz.Y) // fit
		psc.Resize(vpsz)
		psc.Geom.Pos = image.Point{x, y}

		// win.PushPopup(psc)
		return psc
	*/
	return nil
}

// HoverTooltipEvent connects to HoverEvent and pops up a tooltip -- most
// widgets should call this as part of their event connection method
func (wb *WidgetBase) HoverTooltipEvent(we *events.Handlers) {
	/*
		we.AddFunc(events.LongHoverStart, RegPri, func(recv, send ki.Ki, sig int64, d any) {
			me := d.(*mouse.Event)
			wbb := AsWidgetBase(recv)
			if wbb.Tooltip != "" {
				me.SetHandled()
				pos := wbb.WinBBox.Max
				pos.X -= 20
				mvp := wbb.Sc
				PopupTooltip(wbb.Tooltip, pos.X, pos.Y, mvp, wbb.Nm)
			}
		})
	*/
}
