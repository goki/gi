// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package girl

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"image"
	"image/color"
	"io"
	"math"
	"strings"
	"sync"

	"unicode"
	"unicode/utf8"

	"github.com/chewxy/math32"
	"github.com/goki/gi/gist"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/units"
	"github.com/goki/ki/bitflag"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/ki"
	"github.com/goki/mat32"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/f64"
	"golang.org/x/net/html/charset"
)

// text.go contains all the core text rendering and formatting code -- see
// font.go for basic font-level style and management
//
// Styling, Formatting / Layout, and Rendering are each handled separately as
// three different levels in the stack -- simplifies many things to separate
// in this way, and makes the final render pass maximally efficient and
// high-performance, at the potential cost of some memory redundancy.

////////////////////////////////////////////////////////////////////////////////////////
// RuneRender

// RuneRender contains fully explicit data needed for rendering a single rune
// -- Face and Color can be nil after first element, in which case the last
// non-nil is used -- likely slightly more efficient to avoid setting all
// those pointers -- float32 values used to support better accuracy when
// transforming points
type RuneRender struct {
	Face    font.Face            `json:"-" xml:"-" desc:"fully-specified font rendering info, includes fully computed font size -- this is exactly what will be drawn -- no further transforms"`
	Color   color.Color          `json:"-" xml:"-" desc:"color to draw characters in"`
	BgColor color.Color          `json:"-" xml:"-" desc:"background color to fill background of color -- for highlighting, <mark> tag, etc -- unlike Face, Color, this must be non-nil for every case that uses it, as nil is also used for default transparent background"`
	Deco    gist.TextDecorations `desc:"additional decoration to apply -- underline, strike-through, etc -- also used for encoding a few special layout hints to pass info from styling tags to separate layout algorithms (e.g., &lt;P&gt; vs &lt;BR&gt;)"`
	RelPos  mat32.Vec2           `desc:"relative position from start of TextRender for the lower-left baseline rendering position of the font character"`
	Size    mat32.Vec2           `desc:"size of the rune itself, exclusive of spacing that might surround it"`
	RotRad  float32              `desc:"rotation in radians for this character, relative to its lower-left baseline rendering position"`
	ScaleX  float32              `desc:"scaling of the X dimension, in case of non-uniform scaling, 0 = no separate scaling"`
}

// HasNil returns error if any of the key info (face, color) is nil -- only
// the first element must be non-nil
func (rr *RuneRender) HasNil() error {
	if rr.Face == nil {
		return errors.New("gi.RuneRender: Face is nil")
	}
	if rr.Color == nil {
		return errors.New("gi.RuneRender: Color is nil")
	}
	// note: BgColor can be nil -- transparent
	return nil
}

// CurFace is convenience for updating current font face if non-nil
func (rr *RuneRender) CurFace(curFace font.Face) font.Face {
	if rr.Face != nil {
		return rr.Face
	}
	return curFace
}

// CurColor is convenience for updating current color if non-nil
func (rr *RuneRender) CurColor(curColor color.Color) color.Color {
	if rr.Color != nil {
		return rr.Color
	}
	return curColor
}

// RelPosAfterLR returns the relative position after given rune for LR order: RelPos.X + Size.X
func (rr *RuneRender) RelPosAfterLR() float32 {
	return rr.RelPos.X + rr.Size.X
}

// RelPosAfterRL returns the relative position after given rune for RL order: RelPos.X - Size.X
func (rr *RuneRender) RelPosAfterRL() float32 {
	return rr.RelPos.X - rr.Size.X
}

// RelPosAfterTB returns the relative position after given rune for TB order: RelPos.Y + Size.Y
func (rr *RuneRender) RelPosAfterTB() float32 {
	return rr.RelPos.Y + rr.Size.Y
}

//////////////////////////////////////////////////////////////////////////////////
//  SpanRender

// SpanRender contains fully explicit data needed for rendering a span of text
// as a slice of runes, with rune and RuneRender elements in one-to-one
// correspondence (but any nil values will use prior non-nil value -- first
// rune must have all non-nil). Text can be oriented in any direction -- the
// only constraint is that it starts from a single starting position.
// Typically only text within a span will obey kerning.  In standard
// TextRender context, each span is one line of text -- should not have new
// lines within the span itself.  In SVG special cases (e.g., TextPath), it
// can be anything.  It is NOT synonymous with the HTML <span> tag, as many
// styling applications of that tag can be accommodated within a larger
// span-as-line.  The first RuneRender RelPos for LR text should be at X=0
// (LastPos = 0 for RL) -- i.e., relpos positions are minimal for given span.
type SpanRender struct {
	Text    []rune               `desc:"text as runes"`
	Render  []RuneRender         `desc:"render info for each rune in one-to-one correspondence"`
	RelPos  mat32.Vec2           `desc:"position for start of text relative to an absolute coordinate that is provided at the time of rendering -- this typically includes the baseline offset to align all rune rendering there -- individual rune RelPos are added to this plus the render-time offset to get the final position"`
	LastPos mat32.Vec2           `desc:"rune position for further edge of last rune -- for standard flat strings this is the overall length of the string -- used for size / layout computations -- you do not add RelPos to this -- it is in same TextRender relative coordinates"`
	Dir     gist.TextDirections  `desc:"where relevant, this is the (default, dominant) text direction for the span"`
	HasDeco gist.TextDecorations `desc:"mask of decorations that have been set on this span -- optimizes rendering passes"`
}

// Init initializes a new span with given capacity
func (sr *SpanRender) Init(capsz int) {
	sr.Text = make([]rune, 0, capsz)
	sr.Render = make([]RuneRender, 0, capsz)
	sr.HasDeco = 0
}

// IsValid ensures that at least some text is represented and the sizes of
// Text and Render slices are the same, and that the first render info is non-nil
func (sr *SpanRender) IsValid() error {
	if len(sr.Text) == 0 {
		return errors.New("gi.TextRender: Text is empty")
	}
	if len(sr.Text) != len(sr.Render) {
		return fmt.Errorf("gi.TextRender: Render length %v != Text length %v for text: %v", len(sr.Render), len(sr.Text), string(sr.Text))
	}
	return sr.Render[0].HasNil()
}

// SizeHV computes the size of the text span from the first char to the last
// position, which is valid for purely horizontal or vertical text lines --
// either X or Y will be zero depending on orientation
func (sr *SpanRender) SizeHV() mat32.Vec2 {
	if sr.IsValid() != nil {
		return mat32.Vec2{}
	}
	sz := sr.Render[0].RelPos.Sub(sr.LastPos)
	if sz.X < 0 {
		sz.X = -sz.X
	}
	if sz.Y < 0 {
		sz.Y = -sz.Y
	}
	return sz
}

// RuneRelPos returns the relative (starting) position of the given rune index
// (adds Span RelPos and rune RelPos) -- this is typically the baseline
// position where rendering will start, not the upper left corner. if index >
// length, then uses LastPos
func (sr *SpanRender) RuneRelPos(idx int) mat32.Vec2 {
	if idx >= len(sr.Render) {
		return sr.LastPos
	}
	return sr.RelPos.Add(sr.Render[idx].RelPos)
}

// RuneEndPos returns the relative ending position of the given rune index
// (adds Span RelPos and rune RelPos + rune Size.X for LR writing). If index >
// length, then uses LastPos
func (sr *SpanRender) RuneEndPos(idx int) mat32.Vec2 {
	if idx >= len(sr.Render) {
		return sr.LastPos
	}
	spos := sr.RelPos.Add(sr.Render[idx].RelPos)
	spos.X += sr.Render[idx].Size.X
	return spos
}

// AppendRune adds one rune and associated formatting info
func (sr *SpanRender) HasDecoUpdate(bg color.Color, deco gist.TextDecorations) {
	sr.HasDeco |= deco
	if bg != nil {
		bitflag.Set32((*int32)(&sr.HasDeco), int(gist.DecoBgColor))
	}
}

// IsNewPara returns true if this span starts a new paragraph
func (sr *SpanRender) IsNewPara() bool {
	if len(sr.Render) == 0 {
		return false
	}
	return bitflag.Has32(int32(sr.Render[0].Deco), int(gist.DecoParaStart))
}

// SetNewPara sets this as starting a new paragraph
func (sr *SpanRender) SetNewPara() {
	if len(sr.Render) > 0 {
		bitflag.Set32((*int32)(&sr.Render[0].Deco), int(gist.DecoParaStart))
	}
}

// AppendRune adds one rune and associated formatting info
func (sr *SpanRender) AppendRune(r rune, face font.Face, clr, bg color.Color, deco gist.TextDecorations) {
	sr.Text = append(sr.Text, r)
	rr := RuneRender{Face: face, Color: clr, BgColor: bg, Deco: deco}
	sr.Render = append(sr.Render, rr)
	sr.HasDecoUpdate(bg, deco)
}

// AppendString adds string and associated formatting info, optimized with
// only first rune having non-nil face and color settings
func (sr *SpanRender) AppendString(str string, face font.Face, clr, bg color.Color, deco gist.TextDecorations, sty *gist.Font, ctxt *units.Context) {
	if len(str) == 0 {
		return
	}
	ucfont := &gist.Font{}
	if oswin.TheApp.Platform() == oswin.MacOS {
		ucfont.Family = "Arial Unicode"
	} else {
		ucfont.Family = "Arial"
	}
	ucfont.Size = sty.Size
	OpenFont(ucfont, ctxt)

	nwr := []rune(str)
	sz := len(nwr)
	sr.Text = append(sr.Text, nwr...)
	rr := RuneRender{Face: face, Color: clr, BgColor: bg, Deco: deco}
	r := nwr[0]
	lastUc := false
	if r > 0xFF && unicode.IsSymbol(r) {
		rr.Face = ucfont.Face.Face
		lastUc = true
	}
	sr.HasDecoUpdate(bg, deco)
	sr.Render = append(sr.Render, rr)
	for i := 1; i < sz; i++ { // optimize by setting rest to nil for same
		rp := RuneRender{Deco: deco, BgColor: bg}
		r := nwr[i]
		if oswin.TheApp.Platform() == oswin.MacOS {
			if r > 0xFF && unicode.IsSymbol(r) {
				if !lastUc {
					rp.Face = ucfont.Face.Face
					lastUc = true
				}
			} else {
				if lastUc {
					rp.Face = face
					lastUc = false
				}
			}
		}
		sr.Render = append(sr.Render, rp)
	}
}

// SetRenders sets rendering parameters based on style
func (sr *SpanRender) SetRenders(sty *gist.Font, ctxt *units.Context, noBG bool, rot, scalex float32) {
	sz := len(sr.Text)
	if sz == 0 {
		return
	}

	bgc := (color.Color)(&sty.BgColor.Color)
	if noBG {
		bgc = nil
	}

	ucfont := &gist.Font{}
	ucfont.Family = "Arial Unicode"
	ucfont.Size = sty.Size
	OpenFont(ucfont, ctxt)

	sr.HasDecoUpdate(bgc, sty.Deco)
	sr.Render = make([]RuneRender, sz)
	if sty.Face == nil {
		sr.Render[0].Face = ucfont.Face.Face
	} else {
		sr.Render[0].Face = sty.Face.Face
	}
	sr.Render[0].Color = sty.Color
	sr.Render[0].BgColor = bgc
	sr.Render[0].RotRad = rot
	sr.Render[0].ScaleX = scalex
	if bgc != nil {
		for i := range sr.Text {
			sr.Render[i].BgColor = bgc
		}
	}
	if rot != 0 || scalex != 0 {
		for i := range sr.Text {
			sr.Render[i].RotRad = rot
			sr.Render[i].ScaleX = scalex
		}
	}
	if sty.Deco != gist.DecoNone {
		for i := range sr.Text {
			sr.Render[i].Deco = sty.Deco
		}
	}
	// use unicode font for all non-ascii symbols
	lastUc := false
	for i, r := range sr.Text {
		if r > 0xFF && unicode.IsSymbol(r) {
			if !lastUc {
				sr.Render[i].Face = ucfont.Face.Face
				lastUc = true
			}
		} else {
			if lastUc {
				sr.Render[i].Face = sty.Face.Face
				lastUc = false
			}
		}
	}
}

// SetString initializes to given plain text string, with given default style
// parameters that are set for the first render element -- constructs Render
// slice of same size as Text
func (sr *SpanRender) SetString(str string, sty *gist.Font, ctxt *units.Context, noBG bool, rot, scalex float32) {
	sr.Text = []rune(str)
	sr.SetRenders(sty, ctxt, noBG, rot, scalex)
}

// SetRunes initializes to given plain rune string, with given default style
// parameters that are set for the first render element -- constructs Render
// slice of same size as Text
func (sr *SpanRender) SetRunes(str []rune, sty *gist.Font, ctxt *units.Context, noBG bool, rot, scalex float32) {
	sr.Text = str
	sr.SetRenders(sty, ctxt, noBG, rot, scalex)
}

// TextFontRenderMu mutex is required because multiple different goroutines
// associated with different windows can (and often will be) call font stuff
// at the same time (curFace.GlyphAdvance, rendering font) at the same time, on
// the same font face -- and that turns out not to work!
var TextFontRenderMu sync.Mutex

// SetRunePosLR sets relative positions of each rune using a flat
// left-to-right text layout, based on font size info and additional extra
// letter and word spacing parameters (which can be negative)
func (sr *SpanRender) SetRunePosLR(letterSpace, wordSpace, chsz float32, tabSize int) {
	if err := sr.IsValid(); err != nil {
		// log.Println(err)
		return
	}
	sr.Dir = gist.LRTB
	sz := len(sr.Text)
	prevR := rune(-1)
	lspc := letterSpace
	wspc := wordSpace
	if tabSize == 0 {
		tabSize = 4
	}
	var fpos float32
	curFace := sr.Render[0].Face
	TextFontRenderMu.Lock()
	defer TextFontRenderMu.Unlock()
	for i, r := range sr.Text {
		rr := &(sr.Render[i])
		curFace = rr.CurFace(curFace)

		fht := mat32.FromFixed(curFace.Metrics().Height)
		if prevR >= 0 {
			fpos += mat32.FromFixed(curFace.Kern(prevR, r))
		}
		rr.RelPos.X = fpos
		rr.RelPos.Y = 0

		if bitflag.Has32(int32(rr.Deco), int(gist.DecoSuper)) {
			rr.RelPos.Y = -0.45 * mat32.FromFixed(curFace.Metrics().Ascent)
		}
		if bitflag.Has32(int32(rr.Deco), int(gist.DecoSub)) {
			rr.RelPos.Y = 0.15 * mat32.FromFixed(curFace.Metrics().Ascent)
		}

		// todo: could check for various types of special unicode space chars here
		a, _ := curFace.GlyphAdvance(r)
		a32 := mat32.FromFixed(a)
		if a32 == 0 {
			a32 = .1 * fht // something..
		}
		rr.Size = mat32.Vec2{a32, fht}

		if r == '\t' {
			col := int(math32.Ceil(fpos / chsz))
			curtab := col / tabSize
			curtab++
			col = curtab * tabSize
			cpos := chsz * float32(col)
			if cpos > fpos {
				fpos = cpos
			}
		} else {
			fpos += a32
			if i < sz-1 {
				fpos += lspc
				if unicode.IsSpace(r) {
					fpos += wspc
				}
			}
		}
		prevR = r
	}
	sr.LastPos.X = fpos
	sr.LastPos.Y = 0
}

// SetRunePosTB sets relative positions of each rune using a flat
// top-to-bottom text layout -- i.e., letters are in their normal
// upright orientation, but arranged vertically.
func (sr *SpanRender) SetRunePosTB(letterSpace, wordSpace, chsz float32, tabSize int) {
	if err := sr.IsValid(); err != nil {
		// log.Println(err)
		return
	}
	sr.Dir = gist.TB
	sz := len(sr.Text)
	lspc := letterSpace
	wspc := wordSpace
	if tabSize == 0 {
		tabSize = 4
	}
	var fpos float32
	curFace := sr.Render[0].Face
	TextFontRenderMu.Lock()
	defer TextFontRenderMu.Unlock()
	col := 0 // current column position -- todo: does NOT deal with indent
	for i, r := range sr.Text {
		rr := &(sr.Render[i])
		curFace = rr.CurFace(curFace)

		fht := mat32.FromFixed(curFace.Metrics().Height)
		rr.RelPos.X = 0
		rr.RelPos.Y = fpos

		if bitflag.Has32(int32(rr.Deco), int(gist.DecoSuper)) {
			rr.RelPos.Y = -0.45 * mat32.FromFixed(curFace.Metrics().Ascent)
		}
		if bitflag.Has32(int32(rr.Deco), int(gist.DecoSub)) {
			rr.RelPos.Y = 0.15 * mat32.FromFixed(curFace.Metrics().Ascent)
		}

		// todo: could check for various types of special unicode space chars here
		a, _ := curFace.GlyphAdvance(r)
		a32 := mat32.FromFixed(a)
		if a32 == 0 {
			a32 = .1 * fht // something..
		}
		rr.Size = mat32.Vec2{a32, fht}

		if r == '\t' {
			curtab := col / tabSize
			curtab++
			col = curtab * tabSize
			cpos := chsz * float32(col)
			if cpos > fpos {
				fpos = cpos
			}
		} else {
			fpos += fht
			col++
			if i < sz-1 {
				fpos += lspc
				if unicode.IsSpace(r) {
					fpos += wspc
				}
			}
		}
	}
	sr.LastPos.Y = fpos
	sr.LastPos.X = 0
}

// SetRunePosTBRot sets relative positions of each rune using a flat
// top-to-bottom text layout, with characters rotated 90 degress
// based on font size info and additional extra letter and word spacing
// parameters (which can be negative)
func (sr *SpanRender) SetRunePosTBRot(letterSpace, wordSpace, chsz float32, tabSize int) {
	if err := sr.IsValid(); err != nil {
		// log.Println(err)
		return
	}
	sr.Dir = gist.TB
	sz := len(sr.Text)
	prevR := rune(-1)
	lspc := letterSpace
	wspc := wordSpace
	if tabSize == 0 {
		tabSize = 4
	}
	var fpos float32
	curFace := sr.Render[0].Face
	TextFontRenderMu.Lock()
	defer TextFontRenderMu.Unlock()
	col := 0 // current column position -- todo: does NOT deal with indent
	for i, r := range sr.Text {
		rr := &(sr.Render[i])
		rr.RotRad = math32.Pi / 2
		curFace = rr.CurFace(curFace)

		fht := mat32.FromFixed(curFace.Metrics().Height)
		if prevR >= 0 {
			fpos += mat32.FromFixed(curFace.Kern(prevR, r))
		}
		rr.RelPos.Y = fpos
		rr.RelPos.X = 0

		if bitflag.Has32(int32(rr.Deco), int(gist.DecoSuper)) {
			rr.RelPos.X = -0.45 * mat32.FromFixed(curFace.Metrics().Ascent)
		}
		if bitflag.Has32(int32(rr.Deco), int(gist.DecoSub)) {
			rr.RelPos.X = 0.15 * mat32.FromFixed(curFace.Metrics().Ascent)
		}

		// todo: could check for various types of special unicode space chars here
		a, _ := curFace.GlyphAdvance(r)
		a32 := mat32.FromFixed(a)
		if a32 == 0 {
			a32 = .1 * fht // something..
		}
		rr.Size = mat32.Vec2{fht, a32}

		if r == '\t' {
			curtab := col / tabSize
			curtab++
			col = curtab * tabSize
			cpos := chsz * float32(col)
			if cpos > fpos {
				fpos = cpos
			}
		} else {
			fpos += a32
			col++
			if i < sz-1 {
				fpos += lspc
				if unicode.IsSpace(r) {
					fpos += wspc
				}
			}
		}
		prevR = r
	}
	sr.LastPos.Y = fpos
	sr.LastPos.X = 0
}

// FindWrapPosLR finds a position to do word wrapping to fit within trgSize --
// RelPos positions must have already been set (e.g., SetRunePosLR)
func (sr *SpanRender) FindWrapPosLR(trgSize, curSize float32) int {
	sz := len(sr.Text)
	if sz == 0 {
		return -1
	}
	idx := int(float32(sz) * (trgSize / curSize))
	if idx >= sz {
		idx = sz - 1
	}
	// find starting index that is just within size
	csz := sr.RelPos.X + sr.Render[idx].RelPosAfterLR()
	if csz > trgSize {
		for idx > 0 {
			csz = sr.RelPos.X + sr.Render[idx].RelPosAfterLR()
			if csz <= trgSize {
				break
			}
			idx--
		}
	} else {
		for idx < sz-1 {
			nsz := sr.RelPos.X + sr.Render[idx+1].RelPosAfterLR()
			if nsz > trgSize {
				break
			}
			csz = nsz
			idx++
		}
	}
	if unicode.IsSpace(sr.Text[idx]) {
		idx++
		for idx < sz && unicode.IsSpace(sr.Text[idx]) { // break at END of whitespace
			idx++
		}
		return idx
	}
	// find earlier space
	for idx > 0 && !unicode.IsSpace(sr.Text[idx-1]) {
		idx--
	}
	if idx > 0 {
		return idx
	}
	// no spaces within size -- find next break going up
	for idx < sz && !unicode.IsSpace(sr.Text[idx]) {
		idx++
	}
	if idx == sz-1 { // unbreakable
		return -1
	}
	idx++
	for idx < sz && unicode.IsSpace(sr.Text[idx]) { // break at END of whitespace
		idx++
	}
	return idx
}

// ZeroPos ensures that the positions start at 0, for LR direction
func (sr *SpanRender) ZeroPosLR() {
	sz := len(sr.Text)
	if sz == 0 {
		return
	}
	sx := sr.Render[0].RelPos.X
	if sx == 0 {
		return
	}
	for i := range sr.Render {
		sr.Render[i].RelPos.X -= sx
	}
	sr.LastPos.X -= sx
}

// TrimSpaceLeft trims leading space elements from span, and updates the
// relative positions accordingly, for LR direction
func (sr *SpanRender) TrimSpaceLeftLR() {
	srr0 := sr.Render[0]
	for range sr.Text {
		if unicode.IsSpace(sr.Text[0]) {
			sr.Text = sr.Text[1:]
			sr.Render = sr.Render[1:]
			if len(sr.Render) > 0 {
				if sr.Render[0].Face == nil {
					sr.Render[0].Face = srr0.Face
				}
				if sr.Render[0].Color == nil {
					sr.Render[0].Color = srr0.Color
				}
			}
		} else {
			break
		}
	}
	sr.ZeroPosLR()
}

// TrimSpaceRight trims trailing space elements from span, and updates the
// relative positions accordingly, for LR direction
func (sr *SpanRender) TrimSpaceRightLR() {
	for range sr.Text {
		lidx := len(sr.Text) - 1
		if unicode.IsSpace(sr.Text[lidx]) {
			sr.Text = sr.Text[:lidx]
			sr.Render = sr.Render[:lidx]
			lidx--
			if lidx >= 0 {
				sr.LastPos.X = sr.Render[lidx].RelPosAfterLR()
			} else {
				sr.LastPos.X = sr.Render[0].Size.X
			}
		} else {
			break
		}
	}
}

// TrimSpace trims leading and trailing space elements from span, and updates
// the relative positions accordingly, for LR direction
func (sr *SpanRender) TrimSpaceLR() {
	sr.TrimSpaceLeftLR()
	sr.TrimSpaceRightLR()
}

// SplitAt splits current span at given index, returning a new span with
// remainder after index -- space is trimmed from both spans and relative
// positions updated, for LR direction
func (sr *SpanRender) SplitAtLR(idx int) *SpanRender {
	if idx <= 0 || idx >= len(sr.Text)-1 { // shouldn't happen
		return nil
	}
	nsr := SpanRender{Text: sr.Text[idx:], Render: sr.Render[idx:], Dir: sr.Dir, HasDeco: sr.HasDeco}
	sr.Text = sr.Text[:idx]
	sr.Render = sr.Render[:idx]
	sr.LastPos.X = sr.Render[idx-1].RelPosAfterLR()
	// sr.TrimSpaceLR()
	// nsr.TrimSpaceLeftLR() // don't trim right!
	// go back and find latest face and color -- each sr must start with valid one
	if len(nsr.Render) > 0 {
		nrr0 := &(nsr.Render[0])
		face, color := sr.LastFont()
		if nrr0.Face == nil {
			nrr0.Face = face
		}
		if nrr0.Color == nil {
			nrr0.Color = color
		}
	}
	return &nsr
}

// LastFont finds the last font and color from given span
func (sr *SpanRender) LastFont() (face font.Face, color color.Color) {
	for i := len(sr.Render) - 1; i >= 0; i-- {
		srr := sr.Render[i]
		if face == nil && srr.Face != nil {
			face = srr.Face
			if face != nil && color != nil {
				break
			}
		}
		if color == nil && srr.Color != nil {
			color = srr.Color
			if face != nil && color != nil {
				break
			}
		}
	}
	return
}

// todo: TB, RL cases -- layout is complicated.. with unicode-bidi, direction,
// writing-mode styles all interacting: https://www.w3.org/TR/SVG11/text.html#TextLayout

//////////////////////////////////////////////////////////////////////////////////
//  TextLink

// TextLink represents a hyperlink within rendered text
type TextLink struct {
	Label     string   `desc:"text label for the link"`
	URL       string   `desc:"full URL for the link"`
	Props     ki.Props `desc:"properties defined for the link"`
	StartSpan int      `desc:"span index where link starts"`
	StartIdx  int      `desc:"index in StartSpan where link starts"`
	EndSpan   int      `desc:"span index where link ends (can be same as EndSpan)"`
	EndIdx    int      `desc:"index in EndSpan where link ends (index of last rune in label)"`
	Widget    ki.Ki    `desc:"the widget that owns this text link -- only set prior to passing off to handler function"`
}

// Bounds returns the bounds of the link
func (tl *TextLink) Bounds(tr *TextRender, pos mat32.Vec2) image.Rectangle {
	stsp := &tr.Spans[tl.StartSpan]
	tpos := pos.Add(stsp.RelPos)
	sr := &(stsp.Render[tl.StartIdx])
	sp := tpos.Add(sr.RelPos)
	sp.Y -= sr.Size.Y
	ep := sp
	if tl.EndSpan == tl.StartSpan {
		er := &(stsp.Render[tl.EndIdx])
		ep = tpos.Add(er.RelPos)
		ep.X += er.Size.X
	} else {
		er := &(stsp.Render[len(stsp.Render)-1])
		ep = tpos.Add(er.RelPos)
		ep.X += er.Size.X
	}
	return image.Rectangle{Min: sp.ToPointFloor(), Max: ep.ToPointCeil()}
}

// TextLinkHandlerFunc is a function that handles TextLink links -- returns
// true if the link was handled, false if not (in which case it might be
// passed along to someone else)
type TextLinkHandlerFunc func(tl TextLink) bool

// TextLinkHandler is used to handle TextLink links, if non-nil -- set this to
// your own handler to get first crack at all the text link clicks -- if this
// function returns false (or is nil) then the URL is sent to URLHandler (the
// default one just calls oswin.TheApp.OpenURL)
var TextLinkHandler TextLinkHandlerFunc

// URLHandlerFunc is a function that handles URL links -- returns
// true if the link was handled, false if not (in which case it might be
// passed along to someone else).
type URLHandlerFunc func(url string) bool

// URLHandler is used to handle URL links, if non-nil -- set this to your own
// handler to process URL's, depending on TextLinkHandler -- the default
// version of this function just calls oswin.TheApp.OpenURL -- setting this to
// nil will prevent any links from being open that way, and your own function
// will have full responsibility for links if set (i.e., the return value is ignored)
var URLHandler = func(url string) bool {
	oswin.TheApp.OpenURL(url)
	return true
}

//////////////////////////////////////////////////////////////////////////////////
//  TextRender

// TextRender contains one or more SpanRender elements, typically with each
// representing a separate line of text (but they can be anything).
type TextRender struct {
	Spans []SpanRender
	Size  mat32.Vec2          `desc:"last size of overall rendered text"`
	Dir   gist.TextDirections `desc:"where relevant, this is the (default, dominant) text direction for the span"`
	Links []TextLink          `desc:"hyperlinks within rendered text"`
}

// InsertSpan inserts a new span at given index
func (tr *TextRender) InsertSpan(at int, ns *SpanRender) {
	sz := len(tr.Spans)
	tr.Spans = append(tr.Spans, SpanRender{})
	if at > sz-1 {
		tr.Spans[sz] = *ns
		return
	}
	copy(tr.Spans[at+1:], tr.Spans[at:])
	tr.Spans[at] = *ns
}

// Render does text rendering into given image, within given bounds, at given
// absolute position offset (specifying position of text baseline) -- any
// applicable transforms (aside from the char-specific rotation in Render)
// must be applied in advance in computing the relative positions of the
// runes, and the overall font size, etc.  todo: does not currently support
// stroking, only filling of text -- probably need to grab path from font and
// use paint rendering for stroking
func (tr *TextRender) Render(rs *RenderState, pos mat32.Vec2) {
	// pr := prof.Start("RenderText")
	// defer pr.End()

	rs.BackupPaint()
	defer rs.RestorePaint()

	rs.PushXForm(mat32.Identity2D()) // needed for SVG
	defer rs.PopXForm()
	rs.XForm = mat32.Identity2D()

	TextFontRenderMu.Lock()
	defer TextFontRenderMu.Unlock()

	for _, sr := range tr.Spans {
		if sr.IsValid() != nil {
			continue
		}
		curFace := sr.Render[0].Face
		curColor := sr.Render[0].Color
		tpos := pos.Add(sr.RelPos)

		d := &font.Drawer{
			Dst:  rs.Image,
			Src:  image.NewUniform(curColor),
			Face: curFace,
		}

		// todo: cache flags if these are actually needed
		if bitflag.Has32(int32(sr.HasDeco), int(gist.DecoBgColor)) {
			sr.RenderBg(rs, tpos)
		}
		if bitflag.HasAny32(int32(sr.HasDeco), int(gist.DecoUnderline), int(gist.DecoDottedUnderline)) {
			sr.RenderUnderline(rs, tpos)
		}
		if bitflag.Has32(int32(sr.HasDeco), int(gist.DecoOverline)) {
			sr.RenderLine(rs, tpos, gist.DecoOverline, 1.1)
		}

		for i, r := range sr.Text {
			rr := &(sr.Render[i])
			if rr.Color != nil {
				curColor = rr.Color
				d.Src = image.NewUniform(curColor)
			}
			curFace = rr.CurFace(curFace)
			if !unicode.IsPrint(r) {
				continue
			}
			dsc32 := mat32.FromFixed(curFace.Metrics().Descent)
			rp := tpos.Add(rr.RelPos)
			scx := float32(1)
			if rr.ScaleX != 0 {
				scx = rr.ScaleX
			}
			tx := mat32.Scale2D(scx, 1).Rotate(rr.RotRad)
			ll := rp.Add(tx.MulVec2AsVec(mat32.Vec2{0, dsc32}))
			ur := ll.Add(tx.MulVec2AsVec(mat32.Vec2{rr.Size.X, -rr.Size.Y}))
			if int(math32.Floor(ll.X)) > rs.Bounds.Max.X || int(math32.Floor(ur.Y)) > rs.Bounds.Max.Y ||
				int(math32.Ceil(ur.X)) < rs.Bounds.Min.X || int(math32.Ceil(ll.Y)) < rs.Bounds.Min.Y {
				continue
			}
			d.Face = curFace
			d.Dot = rp.Fixed()
			dr, mask, maskp, _, ok := d.Face.Glyph(d.Dot, r)
			if !ok {
				// fmt.Printf("not ok rendering rune: %v\n", string(r))
				continue
			}
			if rr.RotRad == 0 && (rr.ScaleX == 0 || rr.ScaleX == 1) {
				idr := dr.Intersect(rs.Bounds)
				soff := image.ZP
				if dr.Min.X < rs.Bounds.Min.X {
					soff.X = rs.Bounds.Min.X - dr.Min.X
					maskp.X += rs.Bounds.Min.X - dr.Min.X
				}
				if dr.Min.Y < rs.Bounds.Min.Y {
					soff.Y = rs.Bounds.Min.Y - dr.Min.Y
					maskp.Y += rs.Bounds.Min.Y - dr.Min.Y
				}
				draw.DrawMask(d.Dst, idr, d.Src, soff, mask, maskp, draw.Over)
			} else {
				srect := dr.Sub(dr.Min)
				dbase := mat32.Vec2{rp.X - float32(dr.Min.X), rp.Y - float32(dr.Min.Y)}

				transformer := draw.BiLinear
				fx, fy := float32(dr.Min.X), float32(dr.Min.Y)
				m := mat32.Translate2D(fx+dbase.X, fy+dbase.Y).Scale(scx, 1).Rotate(rr.RotRad).Translate(-dbase.X, -dbase.Y)
				s2d := f64.Aff3{float64(m.XX), float64(m.XY), float64(m.X0), float64(m.YX), float64(m.YY), float64(m.Y0)}
				transformer.Transform(d.Dst, s2d, d.Src, srect, draw.Over, &draw.Options{
					SrcMask:  mask,
					SrcMaskP: maskp,
				})
			}
		}
		if bitflag.Has32(int32(sr.HasDeco), int(gist.DecoLineThrough)) {
			sr.RenderLine(rs, tpos, gist.DecoLineThrough, 0.25)
		}
	}
}

// RenderBg renders the background behind chars
func (sr *SpanRender) RenderBg(rs *RenderState, tpos mat32.Vec2) {
	curFace := sr.Render[0].Face
	didLast := false
	// first := true
	pc := &rs.Paint

	for i := range sr.Text {
		rr := &(sr.Render[i])
		if rr.BgColor == nil {
			if didLast {
				pc.Fill(rs)
			}
			didLast = false
			continue
		}
		curFace = rr.CurFace(curFace)
		dsc32 := mat32.FromFixed(curFace.Metrics().Descent)
		rp := tpos.Add(rr.RelPos)
		scx := float32(1)
		if rr.ScaleX != 0 {
			scx = rr.ScaleX
		}
		tx := mat32.Scale2D(scx, 1).Rotate(rr.RotRad)
		ll := rp.Add(tx.MulVec2AsVec(mat32.Vec2{0, dsc32}))
		ur := ll.Add(tx.MulVec2AsVec(mat32.Vec2{rr.Size.X, -rr.Size.Y}))
		if int(math32.Floor(ll.X)) > rs.Bounds.Max.X || int(math32.Floor(ur.Y)) > rs.Bounds.Max.Y ||
			int(math32.Ceil(ur.X)) < rs.Bounds.Min.X || int(math32.Ceil(ll.Y)) < rs.Bounds.Min.Y {
			if didLast {
				pc.Fill(rs)
			}
			didLast = false
			continue
		}
		pc.FillStyle.Color.SetColor(rr.BgColor)
		szt := mat32.Vec2{rr.Size.X, -rr.Size.Y}
		sp := rp.Add(tx.MulVec2AsVec(mat32.Vec2{0, dsc32}))
		ul := sp.Add(tx.MulVec2AsVec(mat32.Vec2{0, szt.Y}))
		lr := sp.Add(tx.MulVec2AsVec(mat32.Vec2{szt.X, 0}))
		pc.DrawPolygon(rs, []mat32.Vec2{sp, ul, ur, lr})
		didLast = true
	}
	if didLast {
		pc.Fill(rs)
	}
}

// RenderUnderline renders the underline for span -- ensures continuity to do it all at once
func (sr *SpanRender) RenderUnderline(rs *RenderState, tpos mat32.Vec2) {
	curFace := sr.Render[0].Face
	curColor := sr.Render[0].Color
	didLast := false
	pc := &rs.Paint

	for i, r := range sr.Text {
		if !unicode.IsPrint(r) {
			continue
		}
		rr := &(sr.Render[i])
		if !bitflag.HasAny32(int32(rr.Deco), int(gist.DecoUnderline), int(gist.DecoDottedUnderline)) {
			if didLast {
				pc.Stroke(rs)
			}
			didLast = false
			continue
		}
		curFace = rr.CurFace(curFace)
		if rr.Color != nil {
			curColor = rr.Color
		}
		dsc32 := mat32.FromFixed(curFace.Metrics().Descent)
		rp := tpos.Add(rr.RelPos)
		scx := float32(1)
		if rr.ScaleX != 0 {
			scx = rr.ScaleX
		}
		tx := mat32.Scale2D(scx, 1).Rotate(rr.RotRad)
		ll := rp.Add(tx.MulVec2AsVec(mat32.Vec2{0, dsc32}))
		ur := ll.Add(tx.MulVec2AsVec(mat32.Vec2{rr.Size.X, -rr.Size.Y}))
		if int(math32.Floor(ll.X)) > rs.Bounds.Max.X || int(math32.Floor(ur.Y)) > rs.Bounds.Max.Y ||
			int(math32.Ceil(ur.X)) < rs.Bounds.Min.X || int(math32.Ceil(ll.Y)) < rs.Bounds.Min.Y {
			if didLast {
				pc.Stroke(rs)
			}
			continue
		}
		dw := .05 * rr.Size.Y
		if !didLast {
			pc.StrokeStyle.Width.Dots = dw
			pc.StrokeStyle.Color.SetColor(curColor)
		}
		if bitflag.Has32(int32(rr.Deco), int(gist.DecoDottedUnderline)) {
			pc.StrokeStyle.Dashes = []float64{2, 2}
		}
		sp := rp.Add(tx.MulVec2AsVec(mat32.Vec2{0, 2 * dw}))
		ep := rp.Add(tx.MulVec2AsVec(mat32.Vec2{rr.Size.X, 2 * dw}))

		if didLast {
			pc.LineTo(rs, sp.X, sp.Y)
		} else {
			pc.NewSubPath(rs)
			pc.MoveTo(rs, sp.X, sp.Y)
		}
		pc.LineTo(rs, ep.X, ep.Y)
		didLast = true
	}
	if didLast {
		pc.Stroke(rs)
	}
	pc.StrokeStyle.Dashes = nil
}

// RenderLine renders overline or line-through -- anything that is a function of ascent
func (sr *SpanRender) RenderLine(rs *RenderState, tpos mat32.Vec2, deco gist.TextDecorations, ascPct float32) {
	curFace := sr.Render[0].Face
	curColor := sr.Render[0].Color
	didLast := false
	pc := &rs.Paint

	for i, r := range sr.Text {
		if !unicode.IsPrint(r) {
			continue
		}
		rr := &(sr.Render[i])
		if !bitflag.Has32(int32(rr.Deco), int(deco)) {
			if didLast {
				pc.Stroke(rs)
			}
			didLast = false
			continue
		}
		curFace = rr.CurFace(curFace)
		dsc32 := mat32.FromFixed(curFace.Metrics().Descent)
		asc32 := mat32.FromFixed(curFace.Metrics().Ascent)
		rp := tpos.Add(rr.RelPos)
		scx := float32(1)
		if rr.ScaleX != 0 {
			scx = rr.ScaleX
		}
		tx := mat32.Scale2D(scx, 1).Rotate(rr.RotRad)
		ll := rp.Add(tx.MulVec2AsVec(mat32.Vec2{0, dsc32}))
		ur := ll.Add(tx.MulVec2AsVec(mat32.Vec2{rr.Size.X, -rr.Size.Y}))
		if int(math32.Floor(ll.X)) > rs.Bounds.Max.X || int(math32.Floor(ur.Y)) > rs.Bounds.Max.Y ||
			int(math32.Ceil(ur.X)) < rs.Bounds.Min.X || int(math32.Ceil(ll.Y)) < rs.Bounds.Min.Y {
			if didLast {
				pc.Stroke(rs)
			}
			continue
		}
		if rr.Color != nil {
			curColor = rr.Color
		}
		dw := 0.05 * rr.Size.Y
		if !didLast {
			pc.StrokeStyle.Width.Dots = dw
			pc.StrokeStyle.Color.SetColor(curColor)
		}
		yo := ascPct * asc32
		sp := rp.Add(tx.MulVec2AsVec(mat32.Vec2{0, -yo}))
		ep := rp.Add(tx.MulVec2AsVec(mat32.Vec2{rr.Size.X, -yo}))

		if didLast {
			pc.LineTo(rs, sp.X, sp.Y)
		} else {
			pc.NewSubPath(rs)
			pc.MoveTo(rs, sp.X, sp.Y)
		}
		pc.LineTo(rs, ep.X, ep.Y)
		didLast = true
	}
	if didLast {
		pc.Stroke(rs)
	}
}

// RenderTopPos renders at given top position -- uses first font info to
// compute baseline offset and calls overall Render -- convenience for simple
// widget rendering without layouts
func (tr *TextRender) RenderTopPos(rs *RenderState, tpos mat32.Vec2) {
	if len(tr.Spans) == 0 {
		return
	}
	sr := &(tr.Spans[0])
	if sr.IsValid() != nil {
		return
	}
	curFace := sr.Render[0].Face
	pos := tpos
	pos.Y += mat32.FromFixed(curFace.Metrics().Ascent)
	tr.Render(rs, pos)
}

// SetString is for basic text rendering with a single style of text (see
// SetHTML for tag-formatted text) -- configures a single SpanRender with the
// entire string, and does standard layout (LR currently).  rot and scalex are
// general rotation and x-scaling to apply to all chars -- alternatively can
// apply these per character after.  Be sure that OpenFont has been run so a
// valid Face is available.  noBG ignores any BgColor in font style, and never
// renders background color
func (tr *TextRender) SetString(str string, fontSty *gist.Font, ctxt *units.Context, txtSty *gist.Text, noBG bool, rot, scalex float32) {
	if len(tr.Spans) != 1 {
		tr.Spans = make([]SpanRender, 1)
	}
	tr.Links = nil
	sr := &(tr.Spans[0])
	sr.SetString(str, fontSty, ctxt, noBG, rot, scalex)
	sr.SetRunePosLR(txtSty.LetterSpacing.Dots, txtSty.WordSpacing.Dots, fontSty.Face.Metrics.Ch, txtSty.TabSize)
	ssz := sr.SizeHV()
	vht := fontSty.Face.Face.Metrics().Height
	tr.Size = mat32.Vec2{ssz.X, mat32.FromFixed(vht)}

}

// SetStringRot90 is for basic text rendering with a single style of text (see
// SetHTML for tag-formatted text) -- configures a single SpanRender with the
// entire string, and does TB rotated layout (-90 deg).
// Be sure that OpenFont has been run so a valid Face is available.
// noBG ignores any BgColor in font style, and never renders background color
func (tr *TextRender) SetStringRot90(str string, fontSty *gist.Font, ctxt *units.Context, txtSty *gist.Text, noBG bool, scalex float32) {
	if len(tr.Spans) != 1 {
		tr.Spans = make([]SpanRender, 1)
	}
	tr.Links = nil
	sr := &(tr.Spans[0])
	rot := math32.Pi / 2
	sr.SetString(str, fontSty, ctxt, noBG, rot, scalex)
	sr.SetRunePosTBRot(txtSty.LetterSpacing.Dots, txtSty.WordSpacing.Dots, fontSty.Face.Metrics.Ch, txtSty.TabSize)
	ssz := sr.SizeHV()
	vht := fontSty.Face.Face.Metrics().Height
	tr.Size = mat32.Vec2{mat32.FromFixed(vht), ssz.Y}
}

// SetRunes is for basic text rendering with a single style of text (see
// SetHTML for tag-formatted text) -- configures a single SpanRender with the
// entire string, and does standard layout (LR currently).  rot and scalex are
// general rotation and x-scaling to apply to all chars -- alternatively can
// apply these per character after Be sure that OpenFont has been run so a
// valid Face is available.  noBG ignores any BgColor in font style, and never
// renders background color
func (tr *TextRender) SetRunes(str []rune, fontSty *gist.Font, ctxt *units.Context, txtSty *gist.Text, noBG bool, rot, scalex float32) {
	if len(tr.Spans) != 1 {
		tr.Spans = make([]SpanRender, 1)
	}
	tr.Links = nil
	sr := &(tr.Spans[0])
	sr.SetRunes(str, fontSty, ctxt, noBG, rot, scalex)
	sr.SetRunePosLR(txtSty.LetterSpacing.Dots, txtSty.WordSpacing.Dots, fontSty.Face.Metrics.Ch, txtSty.TabSize)
	ssz := sr.SizeHV()
	vht := fontSty.Face.Face.Metrics().Height
	tr.Size = mat32.Vec2{ssz.X, mat32.FromFixed(vht)}
}

// SetHTMLSimpleTag sets the styling parameters for simple html style tags
// that only require updating the given font spec values -- returns true if handled
// https://www.w3schools.com/cssref/css_default_values.asp
func SetHTMLSimpleTag(tag string, fs *gist.Font, ctxt *units.Context, cssAgg ki.Props) bool {
	did := false
	switch tag {
	case "b", "strong":
		fs.Weight = gist.WeightBold
		OpenFont(fs, ctxt)
		did = true
	case "i", "em", "var", "cite":
		fs.Style = gist.FontItalic
		OpenFont(fs, ctxt)
		did = true
	case "ins":
		fallthrough
	case "u":
		fs.SetDeco(gist.DecoUnderline)
		did = true
	case "s", "del", "strike":
		fs.SetDeco(gist.DecoLineThrough)
		did = true
	case "sup":
		fs.SetDeco(gist.DecoSuper)
		curpts := math.Round(float64(fs.Size.Convert(units.Pt, ctxt).Val))
		curpts -= 2
		fs.Size = units.NewPt(float32(curpts))
		fs.Size.ToDots(ctxt)
		OpenFont(fs, ctxt)
		did = true
	case "sub":
		fs.SetDeco(gist.DecoSub)
		fallthrough
	case "small":
		curpts := math.Round(float64(fs.Size.Convert(units.Pt, ctxt).Val))
		curpts -= 2
		fs.Size = units.NewPt(float32(curpts))
		fs.Size.ToDots(ctxt)
		OpenFont(fs, ctxt)
		did = true
	case "big":
		curpts := math.Round(float64(fs.Size.Convert(units.Pt, ctxt).Val))
		curpts += 2
		fs.Size = units.NewPt(float32(curpts))
		fs.Size.ToDots(ctxt)
		OpenFont(fs, ctxt)
		did = true
	case "xx-small", "x-small", "smallf", "medium", "large", "x-large", "xx-large":
		fs.Size = units.NewPt(gist.FontSizePoints[tag])
		fs.Size.ToDots(ctxt)
		OpenFont(fs, ctxt)
		did = true
	case "mark":
		fs.BgColor.SetColor(gist.ThePrefs.PrefColor("highlight"))
		did = true
	case "abbr", "acronym":
		fs.SetDeco(gist.DecoDottedUnderline)
		did = true
	case "tt", "kbd", "samp", "code":
		fs.Family = "monospace"
		OpenFont(fs, ctxt)
		did = true
	}
	return did
}

// SetHTML sets text by decoding all standard inline HTML text style
// formatting tags in the string and sets the per-character font information
// appropriately, using given font style info.  <P> and <BR> tags create new
// spans, with <P> marking start of subsequent span with DecoParaStart.
// Critically, it does NOT deal at all with layout (positioning) except in
// breaking lines into different spans, but not with word wrapping -- only
// sets font, color, and decoration info, and strips out the tags it processes
// -- result can then be processed by different layout algorithms as needed.
// cssAgg, if non-nil, should contain CSSAgg properties -- will be tested for
// special css styling of each element.
func (tr *TextRender) SetHTML(str string, font *gist.Font, txtSty *gist.Text, ctxt *units.Context, cssAgg ki.Props) {
	if txtSty.HasPre() {
		tr.SetHTMLPre([]byte(str), font, txtSty, ctxt, cssAgg)
	} else {
		tr.SetHTMLNoPre([]byte(str), font, txtSty, ctxt, cssAgg)
	}
}

// SetHTMLBytes does SetHTML with bytes as input -- more efficient -- use this
// if already in bytes
func (tr *TextRender) SetHTMLBytes(str []byte, font *gist.Font, txtSty *gist.Text, ctxt *units.Context, cssAgg ki.Props) {
	if txtSty.HasPre() {
		tr.SetHTMLPre(str, font, txtSty, ctxt, cssAgg)
	} else {
		tr.SetHTMLNoPre(str, font, txtSty, ctxt, cssAgg)
	}
}

// This is the No-Pre parser that uses the golang XML decoder system, which
// strips all whitespace and is thus unsuitable for any Pre case
func (tr *TextRender) SetHTMLNoPre(str []byte, font *gist.Font, txtSty *gist.Text, ctxt *units.Context, cssAgg ki.Props) {
	//	errstr := "gi.TextRender SetHTML"
	sz := len(str)
	if sz == 0 {
		return
	}
	tr.Spans = make([]SpanRender, 1)
	tr.Links = nil
	curSp := &(tr.Spans[0])
	initsz := ints.MinInt(sz, 1020)
	curSp.Init(initsz)

	spcstr := bytes.Join(bytes.Fields(str), []byte(" "))

	reader := bytes.NewReader(spcstr)
	decoder := xml.NewDecoder(reader)
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose
	decoder.Entity = xml.HTMLEntity
	decoder.CharsetReader = charset.NewReaderLabel

	OpenFont(font, ctxt)

	// set when a </p> is encountered
	nextIsParaStart := false
	curLinkIdx := -1 // if currently processing an <a> link element

	fstack := make([]*gist.Font, 1, 10)
	fstack[0] = font
	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			// log.Printf("%v parsing error: %v for string\n%v\n", errstr, err, string(str))
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			curf := fstack[len(fstack)-1]
			fs := *curf
			nm := strings.ToLower(se.Name.Local)
			curLinkIdx = -1
			if !SetHTMLSimpleTag(nm, &fs, ctxt, cssAgg) {
				switch nm {
				case "a":
					fs.Color.SetColor(gist.ThePrefs.PrefColor("link"))
					fs.SetDeco(gist.DecoUnderline)
					curLinkIdx = len(tr.Links)
					tl := &TextLink{StartSpan: len(tr.Spans) - 1, StartIdx: len(curSp.Text)}
					sprop := make(ki.Props, len(se.Attr))
					tl.Props = sprop
					for _, attr := range se.Attr {
						if attr.Name.Local == "href" {
							tl.URL = attr.Value
						}
						sprop[attr.Name.Local] = attr.Value
					}
					tr.Links = append(tr.Links, *tl)
				case "span":
					// just uses props
				case "q":
					curf := fstack[len(fstack)-1]
					atStart := len(curSp.Text) == 0
					curSp.AppendRune('“', curf.Face.Face, curf.Color, curf.BgColor.ColorOrNil(), curf.Deco)
					if nextIsParaStart && atStart {
						curSp.SetNewPara()
					}
					nextIsParaStart = false
				case "dfn":
					// no default styling
				case "bdo":
					// bidirectional override..
				case "p":
					if len(curSp.Text) > 0 {
						// fmt.Printf("para start: '%v'\n", string(curSp.Text))
						tr.Spans = append(tr.Spans, SpanRender{})
						curSp = &(tr.Spans[len(tr.Spans)-1])
					}
					nextIsParaStart = true
				case "br":
				default:
					// log.Printf("%v tag not recognized: %v for string\n%v\n", errstr, nm, string(str))
				}
			}
			if len(se.Attr) > 0 {
				sprop := make(ki.Props, len(se.Attr))
				for _, attr := range se.Attr {
					switch attr.Name.Local {
					case "style":
						gist.SetStylePropsXML(attr.Value, &sprop)
					case "class":
						if cssAgg != nil {
							clnm := "." + attr.Value
							if aggp, ok := ki.SubProps(cssAgg, clnm); ok {
								fs.SetStyleProps(nil, aggp, nil)
								OpenFont(&fs, ctxt)
							}
						}
					default:
						sprop[attr.Name.Local] = attr.Value
					}
				}
				fs.SetStyleProps(nil, sprop, nil)
				OpenFont(&fs, ctxt)
			}
			if cssAgg != nil {
				FontStyleCSS(&fs, nm, cssAgg, ctxt, nil)
			}
			fstack = append(fstack, &fs)
		case xml.EndElement:
			switch se.Name.Local {
			case "p":
				tr.Spans = append(tr.Spans, SpanRender{})
				curSp = &(tr.Spans[len(tr.Spans)-1])
				nextIsParaStart = true
			case "br":
				tr.Spans = append(tr.Spans, SpanRender{})
				curSp = &(tr.Spans[len(tr.Spans)-1])
			case "q":
				curf := fstack[len(fstack)-1]
				curSp.AppendRune('”', curf.Face.Face, curf.Color, curf.BgColor.ColorOrNil(), curf.Deco)
			case "a":
				if curLinkIdx >= 0 {
					tl := &tr.Links[curLinkIdx]
					tl.EndSpan = len(tr.Spans) - 1
					tl.EndIdx = len(curSp.Text)
					curLinkIdx = -1
				}
			}
			if len(fstack) > 1 {
				fstack = fstack[:len(fstack)-1]
			}
		case xml.CharData:
			curf := fstack[len(fstack)-1]
			atStart := len(curSp.Text) == 0
			sstr := html.UnescapeString(string(se))
			if nextIsParaStart && atStart {
				sstr = strings.TrimLeftFunc(sstr, func(r rune) bool {
					return unicode.IsSpace(r)
				})
			}
			curSp.AppendString(sstr, curf.Face.Face, curf.Color, curf.BgColor.ColorOrNil(), curf.Deco, font, ctxt)
			if nextIsParaStart && atStart {
				curSp.SetNewPara()
			}
			nextIsParaStart = false
			if curLinkIdx >= 0 {
				tl := &tr.Links[curLinkIdx]
				tl.Label = sstr
			}
		}
	}
}

// note: adding print / log statements to following when inside gide will cause
// an infinite loop because the console redirection uses this very same code!

// SetHTMLPre sets preformatted HTML-styled text by decoding all standard
// inline HTML text style formatting tags in the string and sets the
// per-character font information appropriately, using given font style info.
// Only basic styling tags, including <span> elements with style parameters
// (including class names) are decoded.  Whitespace is decoded as-is,
// including LF \n etc, except in WhiteSpacePreLine case which only preserves LF's
func (tr *TextRender) SetHTMLPre(str []byte, font *gist.Font, txtSty *gist.Text, ctxt *units.Context, cssAgg ki.Props) {
	// errstr := "gi.TextRender SetHTMLPre"

	sz := len(str)
	tr.Spans = make([]SpanRender, 1)
	tr.Links = nil
	if sz == 0 {
		return
	}
	curSp := &(tr.Spans[0])
	initsz := ints.MinInt(sz, 1020)
	curSp.Init(initsz)

	OpenFont(font, ctxt)

	nextIsParaStart := false
	curLinkIdx := -1 // if currently processing an <a> link element

	fstack := make([]*gist.Font, 1, 10)
	fstack[0] = font

	tagstack := make([]string, 0, 10)

	tmpbuf := make([]byte, 0, 1020)

	bidx := 0
	curTag := ""
	for bidx < sz {
		cb := str[bidx]
		ftag := ""
		if cb == '<' && sz > bidx+1 {
			eidx := bytes.Index(str[bidx+1:], []byte(">"))
			if eidx > 0 {
				ftag = string(str[bidx+1 : bidx+1+eidx])
				bidx += eidx + 2
			} else { // get past <
				curf := fstack[len(fstack)-1]
				curSp.AppendString(string(str[bidx:bidx+1]), curf.Face.Face, curf.Color, curf.BgColor.ColorOrNil(), curf.Deco, font, ctxt)
				bidx++
			}
		}
		if ftag != "" {
			if ftag[0] == '/' {
				etag := strings.ToLower(ftag[1:])
				// fmt.Printf("%v  etag: %v\n", bidx, etag)
				if etag == "pre" {
					continue // ignore
				}
				if etag != curTag {
					// log.Printf("%v end tag: %v doesn't match current tag: %v for string\n%v\n", errstr, etag, curTag, string(str))
				}
				switch etag {
				// case "p":
				// 	tr.Spans = append(tr.Spans, SpanRender{})
				// 	curSp = &(tr.Spans[len(tr.Spans)-1])
				// 	nextIsParaStart = true
				// case "br":
				// 	tr.Spans = append(tr.Spans, SpanRender{})
				// 	curSp = &(tr.Spans[len(tr.Spans)-1])
				case "q":
					curf := fstack[len(fstack)-1]
					curSp.AppendRune('”', curf.Face.Face, curf.Color, curf.BgColor.ColorOrNil(), curf.Deco)
				case "a":
					if curLinkIdx >= 0 {
						tl := &tr.Links[curLinkIdx]
						tl.EndSpan = len(tr.Spans) - 1
						tl.EndIdx = len(curSp.Text)
						curLinkIdx = -1
					}
				}
				if len(fstack) > 1 { // pop at end
					fstack = fstack[:len(fstack)-1]
				}
				tslen := len(tagstack)
				if tslen > 1 {
					curTag = tagstack[tslen-2]
					tagstack = tagstack[:tslen-1]
				} else if tslen == 1 {
					curTag = ""
					tagstack = tagstack[:0]
				}
			} else { // start tag
				parts := strings.Split(ftag, " ")
				stag := strings.ToLower(strings.TrimSpace(parts[0]))
				// fmt.Printf("%v  stag: %v\n", bidx, stag)
				attrs := parts[1:]
				attr := strings.Split(strings.Join(attrs, " "), "=")
				nattr := len(attr) / 2
				curf := fstack[len(fstack)-1]
				fs := *curf
				curLinkIdx = -1
				if !SetHTMLSimpleTag(stag, &fs, ctxt, cssAgg) {
					switch stag {
					case "a":
						fs.Color.SetColor(gist.ThePrefs.PrefColor("link"))
						fs.SetDeco(gist.DecoUnderline)
						curLinkIdx = len(tr.Links)
						tl := &TextLink{StartSpan: len(tr.Spans) - 1, StartIdx: len(curSp.Text)}
						if nattr > 0 {
							sprop := make(ki.Props, len(parts)-1)
							tl.Props = sprop
							for ai := 0; ai < nattr; ai++ {
								nm := strings.TrimSpace(attr[ai*2])
								vl := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(attr[ai*2+1]), `"`), `"`)
								if nm == "href" {
									tl.URL = vl
								}
								sprop[nm] = vl
							}
						}
						tr.Links = append(tr.Links, *tl)
					case "span":
						// just uses props
					case "q":
						curf := fstack[len(fstack)-1]
						atStart := len(curSp.Text) == 0
						curSp.AppendRune('“', curf.Face.Face, curf.Color, curf.BgColor.ColorOrNil(), curf.Deco)
						if nextIsParaStart && atStart {
							curSp.SetNewPara()
						}
						nextIsParaStart = false
					case "dfn":
						// no default styling
					case "bdo":
						// bidirectional override..
					// case "p":
					// 	if len(curSp.Text) > 0 {
					// 		// fmt.Printf("para start: '%v'\n", string(curSp.Text))
					// 		tr.Spans = append(tr.Spans, SpanRender{})
					// 		curSp = &(tr.Spans[len(tr.Spans)-1])
					// 	}
					// 	nextIsParaStart = true
					// case "br":
					case "pre":
						continue // ignore
					default:
						// log.Printf("%v tag not recognized: %v for string\n%v\n", errstr, stag, string(str))
						// just ignore it and format as is, for pre case!
						// todo: need to include
					}
				}
				if nattr > 0 { // attr
					sprop := make(ki.Props, nattr)
					for ai := 0; ai < nattr; ai++ {
						nm := strings.TrimSpace(attr[ai*2])
						vl := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(attr[ai*2+1]), `"`), `"`)
						// fmt.Printf("nm: %v  val: %v\n", nm, vl)
						switch nm {
						case "style":
							gist.SetStylePropsXML(vl, &sprop)
						case "class":
							if cssAgg != nil {
								clnm := "." + vl
								if aggp, ok := ki.SubProps(cssAgg, clnm); ok {
									fs.SetStyleProps(nil, aggp, nil)
									OpenFont(&fs, ctxt)
								}
							}
						default:
							sprop[nm] = vl
						}
					}
					fs.SetStyleProps(nil, sprop, nil)
					OpenFont(&fs, ctxt)
				}
				if cssAgg != nil {
					FontStyleCSS(&fs, stag, cssAgg, ctxt, nil)
				}
				fstack = append(fstack, &fs)
				curTag = stag
				tagstack = append(tagstack, curTag)
			}
		} else { // raw chars
			// todo: deal with WhiteSpacePreLine -- trim out non-LF ws
			curf := fstack[len(fstack)-1]
			// atStart := len(curSp.Text) == 0
			tmpbuf := tmpbuf[0:0]
			didNl := false
		aggloop:
			for ; bidx < sz; bidx++ {
				nb := str[bidx] // re-gets cb so it can be processed here..
				switch nb {
				case '<':
					if (bidx > 0 && str[bidx-1] == '<') || sz == bidx+1 {
						tmpbuf = append(tmpbuf, nb)
						didNl = false
					} else {
						didNl = false
						break aggloop
					}
				case '\n': // todo absorb other line endings
					unestr := html.UnescapeString(string(tmpbuf))
					curSp.AppendString(unestr, curf.Face.Face, curf.Color, curf.BgColor.ColorOrNil(), curf.Deco, font, ctxt)
					tmpbuf = tmpbuf[0:0]
					tr.Spans = append(tr.Spans, SpanRender{})
					curSp = &(tr.Spans[len(tr.Spans)-1])
					didNl = true
				default:
					didNl = false
					tmpbuf = append(tmpbuf, nb)
				}
			}
			if !didNl {
				unestr := html.UnescapeString(string(tmpbuf))
				// fmt.Printf("%v added: %v\n", bidx, unestr)
				curSp.AppendString(unestr, curf.Face.Face, curf.Color, curf.BgColor.ColorOrNil(), curf.Deco, font, ctxt)
				if curLinkIdx >= 0 {
					tl := &tr.Links[curLinkIdx]
					tl.Label = unestr
				}
			}
		}
	}
}

// RuneSpanPos returns the position (span, rune index within span) within a
// sequence of spans of a given absolute rune index, starting in the first
// span -- returns false if index is out of range (and returns the last position).
func (tx *TextRender) RuneSpanPos(idx int) (si, ri int, ok bool) {
	if idx < 0 || len(tx.Spans) == 0 {
		return 0, 0, false
	}
	ri = idx
	for si = range tx.Spans {
		if ri < 0 {
			ri = 0
		}
		sr := &tx.Spans[si]
		if ri >= len(sr.Render) {
			ri -= len(sr.Render)
			continue
		}
		return si, ri, true
	}
	si = len(tx.Spans) - 1
	ri = len(tx.Spans[si].Render) - 1
	return si, ri, false
}

// SpanPosToRuneIdx returns the absolute rune index for a given span, rune
// index position -- i.e., the inverse of RuneSpanPos.  Returns false if given
// input position is out of range, and returns last valid index in that case.
func (tx *TextRender) SpanPosToRuneIdx(si, ri int) (idx int, ok bool) {
	idx = 0
	for i := range tx.Spans {
		sr := &tx.Spans[i]
		if si > i {
			idx += len(sr.Render)
			continue
		}
		if ri < len(sr.Render) {
			return idx + ri, true
		}
		return idx + (len(sr.Render) - 1), false
	}
	return 0, false
}

// RuneRelPos returns the relative (starting) position of the given rune
// index, counting progressively through all spans present (adds Span RelPos
// and rune RelPos) -- this is typically the baseline position where rendering
// will start, not the upper left corner. If index > length, then uses
// LastPos.  Returns also the index of the span that holds that char (-1 = no
// spans at all) and the rune index within that span, and false if index is
// out of range.
func (tx *TextRender) RuneRelPos(idx int) (pos mat32.Vec2, si, ri int, ok bool) {
	si, ri, ok = tx.RuneSpanPos(idx)
	if ok {
		sr := &tx.Spans[si]
		return sr.RelPos.Add(sr.Render[ri].RelPos), si, ri, true
	}
	nsp := len(tx.Spans)
	if nsp > 0 {
		sr := &tx.Spans[nsp-1]
		return sr.LastPos, nsp - 1, len(sr.Render), false
	}
	return mat32.Vec2Zero, -1, -1, false
}

// RuneEndPos returns the relative ending position of the given rune index,
// counting progressively through all spans present(adds Span RelPos and rune
// RelPos + rune Size.X for LR writing). If index > length, then uses LastPos.
// Returns also the index of the span that holds that char (-1 = no spans at
// all) and the rune index within that span, and false if index is out of
// range.
func (tx *TextRender) RuneEndPos(idx int) (pos mat32.Vec2, si, ri int, ok bool) {
	si, ri, ok = tx.RuneSpanPos(idx)
	if ok {
		sr := &tx.Spans[si]
		spos := sr.RelPos.Add(sr.Render[ri].RelPos)
		spos.X += sr.Render[ri].Size.X
		return spos, si, ri, true
	}
	nsp := len(tx.Spans)
	if nsp > 0 {
		sr := &tx.Spans[nsp-1]
		return sr.LastPos, nsp - 1, len(sr.Render), false
	}
	return mat32.Vec2Zero, -1, -1, false
}

//////////////////////////////////////////////////////////////////////////////////
//  TextStyle-based Layout Routines

// LayoutStdLR does basic standard layout of text in LR direction, assigning
// relative positions to spans and runes according to given styles, and given
// size overall box (nonzero values used to constrain). Returns total
// resulting size box for text.  Font face in gist.Font is used for
// determining line spacing here -- other versions can do more expensive
// calculations of variable line spacing as needed.
func (tr *TextRender) LayoutStdLR(txtSty *gist.Text, fontSty *gist.Font, ctxt *units.Context, size mat32.Vec2) mat32.Vec2 {
	if len(tr.Spans) == 0 {
		return mat32.Vec2Zero
	}

	// pr := prof.Start("TextRenderLayout")
	// defer pr.End()
	//
	tr.Dir = gist.LRTB
	OpenFont(fontSty, ctxt)
	fht := fontSty.Face.Metrics.Height
	dsc := mat32.FromFixed(fontSty.Face.Face.Metrics().Descent)
	lspc := fht * txtSty.EffLineHeight()
	lpad := (lspc - fht) / 2 // padding above / below text box for centering in line

	maxw := float32(0)

	// first pass gets rune positions and wraps text as needed, and gets max width
	si := 0
	for si < len(tr.Spans) {
		sr := &(tr.Spans[si])
		if err := sr.IsValid(); err != nil {
			// log.Print(err)
			si++
			continue
		}
		if sr.LastPos.X == 0 { // don't re-do unless necessary
			sr.SetRunePosLR(txtSty.LetterSpacing.Dots, txtSty.WordSpacing.Dots, fontSty.Face.Metrics.Ch, txtSty.TabSize)
		}
		if sr.IsNewPara() {
			sr.RelPos.X = txtSty.Indent.Dots
		} else {
			sr.RelPos.X = 0
		}
		ssz := sr.SizeHV()
		ssz.X += sr.RelPos.X
		if size.X > 0 && ssz.X > size.X && txtSty.HasWordWrap() {
			for {
				wp := sr.FindWrapPosLR(size.X, ssz.X)
				if wp > 0 && wp < len(sr.Text)-1 {
					nsr := sr.SplitAtLR(wp)
					tr.InsertSpan(si+1, nsr)
					ssz = sr.SizeHV()
					ssz.X += sr.RelPos.X
					if ssz.X > maxw {
						maxw = ssz.X
					}
					si++
					sr = &(tr.Spans[si]) // keep going with nsr
					sr.SetRunePosLR(txtSty.LetterSpacing.Dots, txtSty.WordSpacing.Dots, fontSty.Face.Metrics.Ch, txtSty.TabSize)
					ssz = sr.SizeHV()

					// fixup links
					for li := range tr.Links {
						tl := &tr.Links[li]
						if tl.StartSpan == si-1 {
							if tl.StartIdx >= wp {
								tl.StartIdx -= wp
								tl.StartSpan++
							}
						} else if tl.StartSpan > si-1 {
							tl.StartSpan++
						}
						if tl.EndSpan == si-1 {
							if tl.EndIdx >= wp {
								tl.EndIdx -= wp
								tl.EndSpan++
							}
						} else if tl.EndSpan > si-1 {
							tl.EndSpan++
						}
					}

					if ssz.X <= size.X {
						if ssz.X > maxw {
							maxw = ssz.X
						}
						break
					}
				} else {
					if ssz.X > maxw {
						maxw = ssz.X
					}
					break
				}
			}
		} else {
			if ssz.X > maxw {
				maxw = ssz.X
			}
		}
		si++
	}
	// have maxw, can do alignment cases..

	// make sure links are still in range
	for li := range tr.Links {
		tl := &tr.Links[li]
		stsp := tr.Spans[tl.StartSpan]
		if tl.StartIdx >= len(stsp.Text) {
			tl.StartIdx = len(stsp.Text) - 1
		}
		edsp := tr.Spans[tl.EndSpan]
		if tl.EndIdx >= len(edsp.Text) {
			tl.EndIdx = len(edsp.Text) - 1
		}
	}

	if maxw > size.X {
		size.X = maxw
	}

	// vertical alignment
	nsp := len(tr.Spans)
	npara := 0
	for si := 1; si < nsp; si++ {
		sr := &(tr.Spans[si])
		if sr.IsNewPara() {
			npara++
		}
	}

	vht := lspc*float32(nsp) + float32(npara)*txtSty.ParaSpacing.Dots
	if vht > size.Y {
		size.Y = vht
	}

	tr.Size = mat32.Vec2{maxw, vht}

	vpad := float32(0) // padding at top to achieve vertical alignment
	vextra := size.Y - vht
	if vextra > 0 {
		switch {
		case gist.IsAlignMiddle(txtSty.AlignV):
			vpad = vextra / 2
		case gist.IsAlignEnd(txtSty.AlignV):
			vpad = vextra
		}
	}

	vbaseoff := lspc - lpad - dsc // offset of baseline within overall line
	vpos := vpad + vbaseoff

	for si := range tr.Spans {
		sr := &(tr.Spans[si])
		if si > 0 && sr.IsNewPara() {
			vpos += txtSty.ParaSpacing.Dots
		}
		sr.RelPos.Y = vpos
		sr.LastPos.Y = vpos
		ssz := sr.SizeHV()
		ssz.X += sr.RelPos.X
		hextra := size.X - ssz.X
		if hextra > 0 {
			switch {
			case gist.IsAlignMiddle(txtSty.Align):
				sr.RelPos.X += hextra / 2
			case gist.IsAlignEnd(txtSty.Align):
				sr.RelPos.X += hextra
			}
		}
		vpos += lspc
	}
	return size
}

//////////////////////////////////////////////////////////////////////////////////
//  Utilities

// NextRuneAt returns the next rune starting from given index -- could be at
// that index or some point thereafter -- returns utf8.RuneError if no valid
// rune could be found -- this should be a standard function!
func NextRuneAt(str string, idx int) rune {
	r, _ := utf8.DecodeRuneInString(str[idx:])
	return r
}
