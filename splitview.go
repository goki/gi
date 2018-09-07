// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	"image"

	"github.com/goki/gi/units"
	"github.com/goki/ki"
	"github.com/goki/ki/bitflag"
	"github.com/goki/ki/kit"
)

////////////////////////////////////////////////////////////////////////////////////////
//    SplitView

// SplitView allocates a fixed proportion of space to each child, along given
// dimension, always using only the available space given to it by its parent
// (i.e., it will force its children, which should be layouts (typically
// Frame's), to have their own scroll bars as necesssary).  It should
// generally be used as a main outer-level structure within a window,
// providing a framework for inner elements -- it allows individual child
// elements to update indpendently and thus is important for speeding update
// performance.  It uses the Widget Parts to hold the splitter widgets
// separately from the children that contain the rest of the scenegraph to be
// displayed within each region.
type SplitView struct {
	PartsWidgetBase
	HandleSize  units.Value `xml:"handle-size" desc:"size of the handle region in the middle of each split region, where the splitter can be dragged -- other-dimension size is 2x of this"`
	Splits      []float32   `desc:"proportion (0-1 normalized, enforced) of space allocated to each element -- can enter 0 to collapse a given element"`
	SavedSplits []float32   `desc:"A saved version of the splits which can be restored -- for dynamic collapse / expand operations"`
	Dim         Dims2D      `desc:"dimension along which to split the space"`
}

var KiT_SplitView = kit.Types.AddType(&SplitView{}, SplitViewProps)

// auto-max-stretch
var SplitViewProps = ki.Props{
	"handle-size": units.NewValue(10, units.Px),
	"max-width":   -1.0,
	"max-height":  -1.0,
	"margin":      0,
	"padding":     0,
}

// UpdateSplits updates the splits to be same length as number of children,
// and normalized
func (g *SplitView) UpdateSplits() {
	sz := len(g.Kids)
	if sz == 0 {
		return
	}
	if g.Splits == nil || len(g.Splits) != sz {
		g.Splits = make([]float32, sz)
	}
	sum := float32(0.0)
	for _, sp := range g.Splits {
		sum += sp
	}
	if sum == 0 { // set default even splits
		even := 1.0 / float32(sz)
		for i := range g.Splits {
			g.Splits[i] = even
		}
		sum = 1.0
	} else {
		norm := 1.0 / sum
		for i := range g.Splits {
			g.Splits[i] *= norm
		}
	}
}

// SetSplits sets the split proportions -- can use 0 to hide / collapse a
// child entirely -- does an Update
func (g *SplitView) SetSplits(splits ...float32) {
	updt := g.UpdateStart()
	g.UpdateSplits()
	sz := len(g.Kids)
	mx := kit.MinInt(sz, len(splits))
	for i := 0; i < mx; i++ {
		g.Splits[i] = splits[i]
	}
	g.UpdateSplits()
	g.UpdateEnd(updt)
}

// SaveSplits saves the current set of splits in SavedSplits, for a later RestoreSplits
func (g *SplitView) SaveSplits() {
	sz := len(g.Splits)
	if sz == 0 {
		return
	}
	if g.SavedSplits == nil || len(g.SavedSplits) != sz {
		g.SavedSplits = make([]float32, sz)
	}
	for i, sp := range g.Splits {
		g.SavedSplits[i] = sp
	}
}

// RestoreSplits restores a previously-saved set of splits (if it exists), does an update
func (g *SplitView) RestoreSplits() {
	if g.SavedSplits == nil {
		return
	}
	g.SetSplits(g.SavedSplits...)
}

// CollapseChild collapses given child(ren) (sets split proportion to 0),
// optionally saving the prior splits for later Restore function -- does an
// Update -- triggered by double-click of splitter
func (g *SplitView) CollapseChild(save bool, idxs ...int) {
	updt := g.UpdateStart()
	if save {
		g.SaveSplits()
	}
	sz := len(g.Kids)
	for _, idx := range idxs {
		if idx >= 0 && idx < sz {
			g.Splits[idx] = 0
		}
	}
	g.UpdateSplits()
	g.UpdateEnd(updt)
}

// SetSplitsAction sets the new splitter value, for given splitter -- new
// value is 0..1 value of position of that splitter -- it is a sum of all the
// positions up to that point.  Splitters are updated to ensure that selected
// position is achieved, while dividing remainder appropriately.
func (g *SplitView) SetSplitsAction(idx int, nwval float32) {
	updt := g.UpdateStart()
	g.SetFullReRender()
	sz := len(g.Splits)
	oldsum := float32(0)
	for i := 0; i <= idx; i++ {
		oldsum += g.Splits[i]
	}
	delta := nwval - oldsum
	oldval := g.Splits[idx]
	uval := oldval + delta
	if uval < 0 {
		uval = 0
		delta = -oldval
		nwval = oldsum + delta
	}
	rmdr := 1 - nwval
	if idx < sz-1 {
		oldrmdr := 1 - oldsum
		if oldrmdr <= 0 {
			if rmdr > 0 {
				dper := rmdr / float32((sz-1)-idx)
				for i := idx + 1; i < sz; i++ {
					g.Splits[i] = dper
				}
			}
		} else {
			for i := idx + 1; i < sz; i++ {
				curval := g.Splits[i]
				g.Splits[i] = rmdr * (curval / oldrmdr) // proportional
			}
		}
	}
	g.Splits[idx] = uval
	// fmt.Printf("splits: %v value: %v  splts: %v\n", idx, nwval, g.Splits)
	g.UpdateSplits()
	// fmt.Printf("splits: %v\n", g.Splits)
	g.UpdateEnd(updt)
}

func (g *SplitView) Init2D() {
	g.Parts.Lay = LayoutNil
	g.Init2DWidget()
	g.UpdateSplits()
	g.ConfigSplitters()
}

func (g *SplitView) ConfigSplitters() {
	sz := len(g.Kids)
	mods, updt := g.Parts.SetNChildren(sz-1, KiT_Splitter, "Splitter")
	odim := OtherDim(g.Dim)
	spc := g.Sty.BoxSpace()
	size := g.LayData.AllocSize.Dim(g.Dim) - 2*spc
	handsz := g.HandleSize.Dots
	mid := 0.5 * (g.LayData.AllocSize.Dim(odim) - 2*spc)
	spicon := IconName("")
	if g.Dim == X {
		spicon = IconName("widget-handle-circles-vert")
	} else {
		spicon = IconName("widget-handle-circles-horiz")
	}
	for i, spk := range *g.Parts.Children() {
		sp := spk.(*Splitter)
		sp.Defaults()
		sp.SplitterNo = i
		sp.Icon = spicon
		sp.Dim = g.Dim
		sp.LayData.AllocSize.SetDim(g.Dim, size)
		sp.LayData.AllocSize.SetDim(odim, handsz*2)
		sp.LayData.AllocSizeOrig = sp.LayData.AllocSize
		sp.LayData.AllocPosRel.SetDim(g.Dim, 0)
		sp.LayData.AllocPosRel.SetDim(odim, mid-handsz)
		sp.LayData.AllocPosOrig = sp.LayData.AllocPosRel
		sp.Min = 0.0
		sp.Max = 1.0
		sp.Snap = false
		sp.SetProp("thumb-size", g.HandleSize)
		sp.ThumbSize = g.HandleSize
		if mods {
			sp.SliderSig.ConnectOnly(g.This, func(recv, send ki.Ki, sig int64, data interface{}) {
				if sig == int64(SliderReleased) {
					spr, _ := recv.Embed(KiT_SplitView).(*SplitView)
					spl := send.(*Splitter)
					spr.SetSplitsAction(spl.SplitterNo, spl.Value)
				}
			})
		}
	}
	if mods {
		g.Parts.UpdateEnd(updt)
	}
}

func (g *SplitView) Style2D() {
	g.Style2DWidget()
	g.HandleSize.SetFmInheritProp("handle-size", g.This, false, true) // no inherit, yes type defaults
	g.HandleSize.ToDots(&g.Sty.UnContext)
	g.UpdateSplits()
	g.ConfigSplitters()
}

func (g *SplitView) Layout2D(parBBox image.Rectangle, iter int) bool {
	g.ConfigSplitters()
	g.Layout2DBase(parBBox, true, iter) // init style
	g.Layout2DParts(parBBox, iter)
	g.UpdateSplits()

	handsz := g.HandleSize.Dots
	// fmt.Printf("handsz: %v\n", handsz)
	sz := len(g.Kids)
	odim := OtherDim(g.Dim)
	spc := g.Sty.BoxSpace()
	size := g.LayData.AllocSize.Dim(g.Dim) - 2*spc
	avail := size - handsz*float32(sz-1)
	// fmt.Printf("avail: %v\n", avail)
	osz := g.LayData.AllocSize.Dim(odim) - 2*spc
	pos := float32(0.0)

	spsum := float32(0)
	for i, sp := range g.Splits {
		gis := g.Kids[i].(Node2D).AsWidget()
		if gis == nil {
			continue
		}
		if gis.TypeEmbeds(KiT_Frame) {
			gis.SetReRenderAnchor()
		}
		isz := sp * avail
		gis.LayData.AllocSize.SetDim(g.Dim, isz)
		gis.LayData.AllocSize.SetDim(odim, osz)
		gis.LayData.AllocSizeOrig = gis.LayData.AllocSize
		gis.LayData.AllocPosRel.SetDim(g.Dim, pos)
		gis.LayData.AllocPosRel.SetDim(odim, spc)
		gis.LayData.AllocPosOrig = gis.LayData.AllocPos

		// fmt.Printf("spl: %v sp: %v size: %v alloc: %v  pos: %v\n", i, sp, isz, gis.LayData.AllocSizeOrig, gis.LayData.AllocPosOrig)

		pos += isz + handsz

		spsum += sp
		if i < sz-1 {
			spl := g.Parts.KnownChild(i).(*Splitter)
			spl.Value = spsum
			spl.UpdatePosFromValue()
		}
	}

	return g.Layout2DChildren(iter)
}

func (g *SplitView) Render2D() {
	if g.FullReRenderIfNeeded() {
		return
	}
	if g.PushBounds() {
		for i, kid := range g.Kids {
			nii, ni := KiToNode2D(kid)
			if nii != nil {
				sp := g.Splits[i]
				if sp <= 0 {
					ni.SetInactive()
					continue
				}
				ni.ClearInactive()
				nii.Render2D()
			}
		}
		g.Parts.Render2DTree()
		g.PopBounds()
	}
}

////////////////////////////////////////////////////////////////////////////////////////
//    Splitter

// Splitter provides the splitter handle and line separating two elements in a
// SplitView, with draggable resizing of the splitter -- parent is Parts
// layout of the SplitView -- based on SliderBase
type Splitter struct {
	SliderBase
	SplitterNo int `desc:"splitter number this one is"`
}

var KiT_Splitter = kit.Types.AddType(&Splitter{}, SplitterProps)

var SplitterProps = ki.Props{
	"padding":          units.NewValue(6, units.Px),
	"margin":           units.NewValue(0, units.Px),
	"background-color": &Prefs.Colors.Background,
	"color":            &Prefs.Colors.Font,
	"#icon": ki.Props{
		"max-width":      units.NewValue(1, units.Em),
		"max-height":     units.NewValue(5, units.Em),
		"min-width":      units.NewValue(1, units.Em),
		"min-height":     units.NewValue(5, units.Em),
		"margin":         units.NewValue(0, units.Px),
		"padding":        units.NewValue(0, units.Px),
		"vertical-align": AlignMiddle,
		"fill":           &Prefs.Colors.Icon,
		"stroke":         &Prefs.Colors.Font,
	},
	SliderSelectors[SliderActive]: ki.Props{},
	SliderSelectors[SliderInactive]: ki.Props{
		"border-color": "highlight-50",
		"color":        "highlight-50",
	},
	SliderSelectors[SliderHover]: ki.Props{
		"background-color": "highlight-10",
	},
	SliderSelectors[SliderFocus]: ki.Props{
		"border-width":     units.NewValue(2, units.Px),
		"background-color": "samelight-50",
	},
	SliderSelectors[SliderDown]: ki.Props{},
	SliderSelectors[SliderValue]: ki.Props{
		"border-color":     &Prefs.Colors.Icon,
		"background-color": &Prefs.Colors.Icon,
	},
	SliderSelectors[SliderBox]: ki.Props{
		"border-color":     &Prefs.Colors.Background,
		"background-color": &Prefs.Colors.Background,
	},
}

func (g *Splitter) Defaults() {
	g.ValThumb = false
	g.ThumbSize = units.NewValue(10, units.Px) // will be replaced by parent HandleSize
	g.Step = 0.01
	g.PageStep = 0.1
	g.Max = 1.0
	g.Snap = false
	g.Prec = 4
	bitflag.Set(&g.Flag, int(InstaDrag))
}

func (g *Splitter) Init2D() {
	g.Init2DSlider()
	g.Defaults()
	g.ConfigParts()
}

func (g *Splitter) ConfigPartsIfNeeded(render bool) {
	if g.PartsNeedUpdateIconLabel(string(g.Icon), "") {
		g.ConfigParts()
	}
	if !g.Icon.IsValid() || !g.Parts.HasChildren() {
		return
	}
	ick, ok := g.Parts.Children().ElemByType(KiT_Icon, true, 0)
	if !ok {
		return
	}
	ic := ick.(*Icon)
	handsz := g.ThumbSize.Dots
	spc := g.Sty.BoxSpace()
	odim := OtherDim(g.Dim)
	g.LayData.AllocSize.SetDim(odim, 2*(handsz+2*spc))
	g.LayData.AllocSizeOrig = g.LayData.AllocSize

	ic.LayData.AllocSize.SetDim(odim, 2*handsz)
	ic.LayData.AllocSize.SetDim(g.Dim, handsz)
	ic.LayData.AllocPosRel.SetDim(g.Dim, g.Pos-(0.5*(handsz+spc)))
	ic.LayData.AllocPosRel.SetDim(odim, 0)
	if render {
		ic.Layout2DTree()
	}
}

func (g *Splitter) Style2D() {
	bitflag.Clear(&g.Flag, int(CanFocus))
	g.Style2DWidget()
	pst := &(g.Par.(Node2D).AsWidget().Sty)
	for i := 0; i < int(SliderStatesN); i++ {
		g.StateStyles[i].CopyFrom(&g.Sty)
		g.StateStyles[i].SetStyleProps(pst, g.StyleProps(SliderSelectors[i]))
		g.StateStyles[i].CopyUnitContext(&g.Sty.UnContext)
	}
	SliderFields.Style(g, nil, g.Props)
	SliderFields.ToDots(g, &g.Sty.UnContext)
	g.ThSize = g.ThumbSize.Dots
	g.ConfigParts()
}

func (g *Splitter) Size2D(iter int) {
	g.InitLayout2D()
	if g.ThSize == 0.0 {
		g.Defaults()
	}
}

func (g *Splitter) Layout2D(parBBox image.Rectangle, iter int) bool {
	g.ConfigPartsIfNeeded(false)
	g.Layout2DBase(parBBox, true, iter) // init style
	g.Layout2DParts(parBBox, iter)
	// g.SizeFromAlloc()
	g.Size = g.LayData.AllocSize.Dim(g.Dim)
	g.UpdatePosFromValue()
	g.DragPos = g.Pos
	g.OrigWinBBox = g.WinBBox
	return g.Layout2DChildren(iter)
}

func (g *Splitter) UpdateSplitterPos() {
	spc := g.Sty.BoxSpace()
	ispc := int(spc)
	handsz := g.ThumbSize.Dots
	off := 0
	if g.Dim == X {
		off = g.OrigWinBBox.Min.X
	} else {
		off = g.OrigWinBBox.Min.Y
	}
	sz := handsz
	if !g.IsDragging() {
		sz += 2 * spc
	}
	pos := off + int(g.Pos-0.5*sz)
	mxpos := off + int(g.Pos+0.5*sz)
	if g.Dim == X {
		g.VpBBox = image.Rect(pos, g.ObjBBox.Min.Y+ispc, mxpos, g.ObjBBox.Max.Y+ispc)
		g.WinBBox = image.Rect(pos, g.ObjBBox.Min.Y+ispc, mxpos, g.ObjBBox.Max.Y+ispc)
	} else {
		g.VpBBox = image.Rect(g.ObjBBox.Min.X+ispc, pos, g.ObjBBox.Max.X+ispc, mxpos)
		g.WinBBox = image.Rect(g.ObjBBox.Min.X+ispc, pos, g.ObjBBox.Max.X+ispc, mxpos)
	}
}

func (g *Splitter) Render2D() {
	vp := g.Viewport
	win := vp.Win
	g.SliderEvents()
	if g.IsDragging() {
		// spc := g.Sty.BoxSpace()
		// odim := OtherDim(g.Dim)
		ick, ok := g.Parts.Children().ElemByType(KiT_Icon, true, 0)
		if !ok {
			return
		}
		ic := ick.(*Icon)
		icvp, ok := ic.Children().ElemByType(KiT_Viewport2D, true, 0)
		if !ok {
			return
		}
		ovk, ok := win.OverlayVp.ChildByName(g.UniqueName(), 0)
		var ovb *Bitmap
		if !ok {
			ovb = &Bitmap{}
			ovb.SetName(g.UniqueName())
			win.OverlayVp.AddChild(ovb)
			ovk = ovb.This
		}
		ovb = ovk.(*Bitmap)
		ovb.GrabRenderFrom(icvp.(Node2D))
		ovb.LayData = ic.LayData // copy
		// ovb.LayData.AllocPos.SetDim(odim, ovb.LayData.AllocPos.Dim(odim)+spc)
		g.UpdateSplitterPos()
		ovb.LayData.AllocPos.SetPoint(g.VpBBox.Min)
		win.RenderOverlays()
	} else {
		ovidx, ok := win.OverlayVp.Children().IndexByName(g.UniqueName(), 0)
		if ok {
			win.OverlayVp.DeleteChildAtIndex(ovidx, true)
			win.RenderOverlays()
		}
		// todo: still not rendering properly
		if g.FullReRenderIfNeeded() {
			return
		}
		if g.PushBounds() {
			g.Render2DDefaultStyle()
			g.Render2DChildren()
			g.PopBounds()
		}
	}
}

// render using a default style if not otherwise styled
func (g *Splitter) Render2DDefaultStyle() {
	st := &g.Sty
	rs := &g.Viewport.Render
	pc := &rs.Paint

	g.UpdateSplitterPos()
	g.ConfigPartsIfNeeded(true)

	if g.Icon.IsValid() && g.Parts.HasChildren() {
		g.Parts.Render2DTree()
	} else {
		pc.StrokeStyle.SetColor(nil)
		pc.FillStyle.SetColorSpec(&st.Font.BgColor)

		pos := NewVec2DFmPoint(g.VpBBox.Min)
		pos.SetSubDim(OtherDim(g.Dim), 10.0)
		sz := NewVec2DFmPoint(g.VpBBox.Size())
		g.RenderBoxImpl(pos, sz, 0)
	}
}

func (g *Splitter) FocusChanged2D(change FocusChanges) {
	switch change {
	case FocusLost:
		g.SetSliderState(SliderActive) // lose any hover state but whatever..
		g.UpdateSig()
	case FocusGot:
		g.SetSliderState(SliderFocus)
		g.EmitFocusedSignal()
		g.UpdateSig()
	case FocusInactive: // don't care..
	case FocusActive:
	}
}
