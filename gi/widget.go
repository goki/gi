// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

//go:generate goki generate

import (
	"fmt"
	"image"
	"log"
	"reflect"
	"sync"

	"goki.dev/enums"
	"goki.dev/girl/styles"
	"goki.dev/girl/units"
	"goki.dev/goosi/events"
	"goki.dev/ki/v2"
	"goki.dev/laser"
)

// Widget is the interface for all GoGi Widget Nodes
type Widget interface {
	ki.Ki

	// todo: rename .Style to .Styles and AddStyles to Style()

	// AddStyles sets the styling of the widget by adding a Styler function
	AddStyles(s Styler) Widget

	// SetTooltip sets the Tooltip message when hovering over the widget
	SetTooltip(tt string) Widget

	// AsWidget returns the WidgetBase embedded field for any Widget node.
	// The Widget interface defines only methods that can be overridden
	// or need to be called on other nodes.  Everything else that is common
	// to all Widgets is in the WidgetBase.
	AsWidget() *WidgetBase

	// Config configures the widget, primarily configuring its Parts.
	// it does _not_ call Config on children, just self.
	// ApplyStyle must generally be called after Config - it is called
	// automatically when Scene is first shown, but must be called
	// manually thereafter as needed after configuration changes.
	// See ReConfig for a convenience function that does both.
	// ConfigScene on Scene handles full tree configuration.
	// This config calls UpdateStart / End, and SetNeedsLayout,
	// and calls ConfigWidget to do the actual configuration,
	// so it does not need to manage this housekeeping.
	// Thus, this Config call is typically never changed, and
	// all custom configuration should happen in ConfigWidget.
	Config(sc *Scene)

	// ConfigWidget does the actual configuration of the widget,
	// primarily configuring its Parts.
	// All configuration should be robust to multiple calls
	// (i.e., use Parts.ConfigChildren with Config).
	// Outer Config call handles all the other infrastructure,
	// so this call just does the core configuration.
	ConfigWidget(sc *Scene)

	// ReConfig calls Config and ApplyStyle on this widget.
	// This should be called if any config options are changed,
	// while the Scene is being viewed.
	ReConfig()

	// StateIs returns true if given Style.State flag is set
	StateIs(flag enums.BitFlag) bool

	// AbilityIs returns true if given Style.Abilities flag is set
	AbilityIs(flag enums.BitFlag) bool

	// SetState sets given Style.State flags
	SetState(on bool, state ...enums.BitFlag)

	// ApplyStyle applies style functions to the widget based on current state.
	// It is typically not overridden -- set style funcs to apply custom styling.
	ApplyStyle(sc *Scene)

	// GetSize: MeLast downward pass, each node first calls
	// g.Layout.Reset(), then sets their LayoutSize according to their own
	// intrinsic size parameters, and/or those of its children if it is a
	// Layout.
	GetSize(sc *Scene, iter int)

	// DoLayout: MeFirst downward pass (each node calls on its children at
	// appropriate point) with relevant parent BBox that the children are
	// constrained to render within -- they then intersect this BBox with
	// their own BBox (from BBoxes) -- typically just call DoLayoutBase for
	// default behavior -- and add parent position to AllocPos, and then
	// return call to DoLayoutChildren. Layout does all its sizing and
	// positioning of children in this pass, based on the GetSize data gathered
	// bottom-up and constraints applied top-down from higher levels.
	// Typically only a single iteration is required (iter = 0) but multiple
	// are supported (needed for word-wrapped text or flow layouts) -- return
	// = true indicates another iteration required (pass this up the chain).
	DoLayout(sc *Scene, parBBox image.Rectangle, iter int) bool

	// LayoutScroll: optional MeFirst downward pass to move all elements by given
	// delta -- used for scrolling -- the layout pass assigns canonical
	// positions, saved in AllocPosOrig and BBox, and this adds the given
	// delta to that AllocPosOrig -- each node must call ComputeBBoxes to
	// update its bounding box information given the new position.
	LayoutScroll(sc *Scene, delta image.Point, parBBox image.Rectangle)

	// BBoxes: compute the raw bounding box of this node relative to its
	// parent scene -- called during DoLayout to set node BBox field, which
	// is then used in setting ScBBox.
	BBoxes() image.Rectangle

	// Compute ScBBox and WinBBox from BBox, given parent ScBBox -- most nodes
	// call ComputeBBoxesBase but scenes require special code -- called
	// during Layout and Move.
	ComputeBBoxes(sc *Scene, parBBox image.Rectangle, delta image.Point)

	// ChildrenBBoxes: compute the bbox available to my children (content),
	// adjusting for margins, border, padding (BoxSpace) taken up by me --
	// operates on the existing ScBBox for this node -- this is what is passed
	// down as parBBox do the children's DoLayout.
	ChildrenBBoxes(sc *Scene) image.Rectangle

	// Render: Actual rendering pass, each node is fully responsible for
	// calling Render on its own children, to provide maximum flexibility
	// (see RenderChildren for default impl) -- bracket the render calls in
	// PushBounds / PopBounds and a false from PushBounds indicates that
	// ScBBox is empty and no rendering should occur.
	Render(sc *Scene)

	// On adds an event listener function for the given event type
	On(etype events.Types, fun func(e events.Event)) Widget

	// HandleEvent calls registered event Listener functions for given event
	HandleEvent(ev events.Event)

	// Send sends an event of given type to this widget,
	// optionally starting from values in the given original event
	// (recommended to include where possible).
	Send(ev events.Types, orig events.Event)

	// MakeContextMenu creates the context menu items (typically Action
	// elements, but it can be anything) for a given widget, typically
	// activated by the right mouse button or equivalent.  Widget has a
	// function parameter that can be set to add context items (e.g., by Views
	// or other complex widgets) to extend functionality.
	MakeContextMenu(menu *MenuActions)

	// ContextMenuPos returns the default position for popup menus --
	// by default in the middle its Bounding Box, but can be adapted as
	// appropriate for different widgets.
	ContextMenuPos() image.Point

	// ContextMenu displays the context menu of various actions to perform on
	// a node -- returns immediately, and actions are all executed directly
	// (later) via the action signals.  Calls MakeContextMenu and
	// ContextMenuPos.
	ContextMenu()

	// IsVisible provides the definitive answer as to whether a given node
	// is currently visible.  It is only entirely valid after a render pass
	// for widgets in a visible window, but it checks the window and scene
	// for their visibility status as well, which is available always.
	// This does *not* check for ScBBox level visibility, which is a further check.
	// Non-visible nodes are automatically not rendered and do not get
	// window events.  The Invisible flag is one key element of the IsVisible
	// calculus -- it is set by e.g., TabView for invisible tabs, and is also
	// set if a widget is entirely out of render range.  But again, use
	// IsVisible as the main end-user method.
	// For robustness, it recursively calls the parent -- this is typically
	// a short path -- propagating the Invisible flag properly can be
	// very challenging without mistakenly overwriting invisibility at various
	// levels.
	IsVisible() bool

	// SetMinPrefWidth sets minimum and preferred width;
	// will get at least this amount; max unspecified.
	// This adds a styler that calls [styles.Style.SetMinPrefWidth].
	SetMinPrefWidth(val units.Value) Widget

	// SetMinPrefHeight sets minimum and preferred height;
	// will get at least this amount; max unspecified.
	// This adds a styler that calls [styles.Style.SetMinPrefHeight].
	SetMinPrefHeight(val units.Value) Widget

	// SetStretchMaxWidth sets stretchy max width (-1);
	// can grow to take up avail room.
	// This adds a styler that calls [styles.Style.SetStretchMaxWidth].
	SetStretchMaxWidth() Widget

	// SetStretchMaxHeight sets stretchy max height (-1);
	// can grow to take up avail room.
	// This adds a styler that calls [styles.Style.SetStretchMaxHeight].
	SetStretchMaxHeight() Widget

	// SetStretchMax sets stretchy max width and height (-1);
	// can grow to take up avail room.
	// This adds a styler that calls [styles.Style.SetStretchMax].
	SetStretchMax() Widget

	// SetFixedWidth sets all width style options
	// (Width, MinWidth, and MaxWidth) to
	// the given fixed width unit value.
	// This adds a styler that calls [styles.Style.SetFixedWidth].
	SetFixedWidth(val units.Value) Widget

	// SetFixedHeight sets all height style options
	// (Height, MinHeight, and MaxHeight) to
	// the given fixed height unit value.
	// This adds a styler that calls [styles.Style.SetFixedHeight].
	SetFixedHeight(val units.Value) Widget

	// todo: revisit this -- in general anything with a largish image (including svg,
	// SubScene, but not Icon) should get put on a list so the RenderWin Drawer just
	// directly uploads its image.

	// IsDirectWinUpload returns true if this is a node that does a direct window upload
	// e.g., for gi3d.Scene which renders directly to the window texture for maximum efficiency
	IsDirectWinUpload() bool

	// DirectWinUpload does a direct upload of contents to a window
	// Drawer compositing image, which will then be used for drawing
	// the window during a Publish() event (triggered by the window Update
	// event).  This is called by the scene in its Update signal processing
	// routine on nodes that respond true to IsDirectWinUpload().
	// The node is also free to update itself of its own accord at any point.
	DirectWinUpload()
}

// WidgetBase is the base type for all Widget Widget elements, which are
// managed by a containing Layout, and use all 5 rendering passes.  All
// elemental widgets must support the Inactive and Selected states in a
// reasonable way (Selected only essential when also Inactive), so they can
// function appropriately in a chooser (e.g., SliceView or TableView) -- this
// includes toggling selection on left mouse press.
type WidgetBase struct {
	ki.Node

	// todo: remove CSS stuff from here??

	// user-defined class name(s) used primarily for attaching CSS styles to different display elements -- multiple class names can be used to combine properties: use spaces to separate per css standard
	Class string `desc:"user-defined class name(s) used primarily for attaching CSS styles to different display elements -- multiple class names can be used to combine properties: use spaces to separate per css standard"`

	// cascading style sheet at this level -- these styles apply here and to everything below, until superceded -- use .class and #name Props elements to apply entire styles to given elements, and type for element type
	CSS ki.Props `xml:"css" desc:"cascading style sheet at this level -- these styles apply here and to everything below, until superceded -- use .class and #name Props elements to apply entire styles to given elements, and type for element type"`

	// [view: no-inline] aggregated css properties from all higher nodes down to me
	CSSAgg ki.Props `copy:"-" json:"-" xml:"-" view:"no-inline" desc:"aggregated css properties from all higher nodes down to me"`

	// todo: need to fully revisit scrolling logic!

	// raw original bounding box for the widget within its parent Scene -- used for computing ScBBox.  This is not updated by LayoutScroll, whereas ScBBox is
	BBox image.Rectangle `copy:"-" json:"-" xml:"-" desc:"raw original bounding box for the widget within its parent Scene -- used for computing ScBBox.  This is not updated by LayoutScroll, whereas ScBBox is"`

	// full object bbox -- this is BBox + LayoutScroll delta, but NOT intersected with parent's parBBox -- used for computing color gradients or other object-specific geometry computations
	ObjBBox image.Rectangle `copy:"-" json:"-" xml:"-" desc:"full object bbox -- this is BBox + LayoutScroll delta, but NOT intersected with parent's parBBox -- used for computing color gradients or other object-specific geometry computations"`

	// 2D bounding box for region occupied within immediate parent Scene object that we render onto -- these are the pixels we draw into, filtered through parent bounding boxes -- used for render Bounds clipping
	ScBBox image.Rectangle `copy:"-" json:"-" xml:"-" desc:"2D bounding box for region occupied within immediate parent Scene object that we render onto -- these are the pixels we draw into, filtered through parent bounding boxes -- used for render Bounds clipping"`

	// text for tooltip for this widget -- can use HTML formatting
	Tooltip string `desc:"text for tooltip for this widget -- can use HTML formatting"`

	// a slice of stylers that are called in sequential descending order (so the first added styler is called last and thus overrides all other functions) to style the element; these should be set using AddStyles, which can be called by end-user and internal code
	Stylers []Styler `json:"-" xml:"-" copy:"-" desc:"a slice of stylers that are called in sequential descending order (so the first added styler is called last and thus overrides all other functions) to style the element; these should be set using AddStyles, which can be called by end-user and internal code"`

	// override the computed styles and allow directly editing Style
	OverrideStyle bool `json:"-" xml:"-" desc:"override the computed styles and allow directly editing Style"`

	// styling settings for this widget -- set in SetApplyStyle during an initialization step, and when the structure changes; they are determined by, in increasing priority order, the default values, the ki node properties, and the StyleFunc (the recommended way to set styles is through the StyleFunc -- setting this field directly outside of that will have no effect unless OverrideStyle is on)
	Style styles.Style `json:"-" xml:"-" desc:"styling settings for this widget -- set in SetApplyStyle during an initialization step, and when the structure changes; they are determined by, in increasing priority order, the default values, the ki node properties, and the StyleFunc (the recommended way to set styles is through the StyleFunc -- setting this field directly outside of that will have no effect unless OverrideStyle is on)"`

	// Listeners are event listener functions for processing events on this widget.
	// type specific Listeners are added in OnInit when the widget is initialized.
	Listeners events.Listeners

	// a separate tree of sub-widgets that implement discrete parts of a widget -- positions are always relative to the parent widget -- fully managed by the widget and not saved
	Parts *Layout `json:"-" xml:"-" view-closed:"true" desc:"a separate tree of sub-widgets that implement discrete parts of a widget -- positions are always relative to the parent widget -- fully managed by the widget and not saved"`

	// all the layout state information for this widget
	LayState LayoutState `copy:"-" json:"-" xml:"-" desc:"all the layout state information for this widget"`

	// [view: -] optional context menu function called by MakeContextMenu AFTER any native items are added -- this function can decide where to insert new elements -- typically add a separator to disambiguate
	CtxtMenuFunc CtxtMenuFunc `copy:"-" view:"-" json:"-" xml:"-" desc:"optional context menu function called by MakeContextMenu AFTER any native items are added -- this function can decide where to insert new elements -- typically add a separator to disambiguate"`

	// parent scene.  Only for use as a last resort when arg is not available -- otherwise always use the arg.  Set during Config.
	Sc *Scene `copy:"-" json:"-" xml:"-" desc:"parent scene.  Only for use as a last resort when arg is not available -- otherwise always use the arg.  Set during Config."`

	// [view: -] mutex protecting the Style field
	StyMu sync.RWMutex `copy:"-" view:"-" json:"-" xml:"-" desc:"mutex protecting the Style field"`

	// [view: -] mutex protecting the BBox fields
	BBoxMu sync.RWMutex `copy:"-" view:"-" json:"-" xml:"-" desc:"mutex protecting the BBox fields"`
}

func (wb *WidgetBase) OnInit() {
}

// AsWidget returns the given Ki object
// as a Widget interface and a WidgetBase.
func AsWidget(k ki.Ki) (Widget, *WidgetBase) {
	if k == nil || k.This() == nil {
		return nil, nil
	}
	if w, ok := k.This().(Widget); ok {
		return w, w.AsWidget()
	}
	return nil, nil
}

func (wb *WidgetBase) AsWidget() *WidgetBase {
	return wb
}

// AsWidgetBase returns the given Ki object as a WidgetBase, or nil.
// for direct use of the return value in cases where that is needed.
func AsWidgetBase(k ki.Ki) *WidgetBase {
	_, wb := AsWidget(k)
	return wb
}

func (wb *WidgetBase) CopyFieldsFrom(frm any) {
	fr, ok := frm.(*WidgetBase)
	if !ok {
		log.Printf("GoGi node of type: %v needs a CopyFieldsFrom method defined\n", wb.KiType().Name)
		return
	}
	wb.Class = fr.Class
	wb.CSS.CopyFrom(fr.CSS, true)
	wb.Tooltip = fr.Tooltip
	wb.Style.CopyFrom(&fr.Style)
}

func (wb *WidgetBase) BaseIface() reflect.Type {
	return laser.TypeFor[Widget]()
}

func (wb *WidgetBase) StateIs(flag enums.BitFlag) bool {
	return wb.Style.State.HasFlag(flag)
}

func (wb *WidgetBase) AbilityIs(flag enums.BitFlag) bool {
	return wb.Style.Abilities.HasFlag(flag)
}

// SetState sets the Style.State flags
func (wb *WidgetBase) SetState(on bool, state ...enums.BitFlag) {
	wb.Style.State.SetFlag(on, state...)
}

func (wb *WidgetBase) SetTooltip(tt string) Widget {
	wb.Tooltip = tt
	return wb.This().(Widget)
}

// NewParts makes the Parts layout if not already there,
// with given layout orientation
func (wb *WidgetBase) NewParts(lay Layouts) *Layout {
	if wb.Parts != nil {
		return wb.Parts
	}
	parts := &Layout{}
	parts.InitName(parts, "parts")
	parts.Lay = lay
	ki.SetParent(parts, wb.This())
	parts.SetFlag(true, ki.Field)
	wb.Parts = parts
	return parts
}

// ParentWidget returns the parent as a (Widget, *WidgetBase)
// or nil if this is the root and has no parent.
func (wb *WidgetBase) ParentWidget() (Widget, *WidgetBase) {
	if wb.Par == nil {
		return nil, nil
	}
	wi := wb.Par.(Widget)
	return wi, wi.AsWidget()
}

// ParentWidgetIf returns the nearest widget parent
// of the widget for which the given function returns true.
// It returns nil if no such parent is found;
// see [ParentWidgetIfTry] for a version with an error.
func (wb *WidgetBase) ParentWidgetIf(fun func(p *WidgetBase) bool) (Widget, *WidgetBase) {
	pwi, pwb, _ := wb.ParentWidgetIfTry(fun)
	return pwi, pwb
}

// ParentWidgetIfTry returns the nearest widget parent
// of the widget for which the given function returns true.
// It returns an error if no such parent is found; see
// [ParentWidgetIf] for a version without an error.
func (wb *WidgetBase) ParentWidgetIfTry(fun func(p *WidgetBase) bool) (Widget, *WidgetBase, error) {
	cur := wb
	for {
		par := cur.Par
		if par == nil {
			return nil, nil, fmt.Errorf("(gi.WidgetBase).ParentWidgetIfTry: got to root: %v without finding", cur)
		}
		pwi, ok := par.(Widget)
		if !ok {
			return nil, nil, fmt.Errorf("(gi.WidgetBase).ParentWidgetIfTry: parent is not a widget: %v", par)
		}
		pwb := pwi.AsWidget()
		if fun(pwb) {
			return pwi, pwb, nil
		}
		cur = pwb
	}
	return nil, nil, fmt.Errorf("(gi.WidgetBase).ParentWidgetIfTry: shouldn't get here: %v", wb)
}

func (wb *WidgetBase) IsVisible() bool {
	if wb == nil || wb.This() == nil || wb.Is(Invisible) {
		return false
	}
	if wb.Par == nil || wb.Par.This() == nil {
		return true
	}
	return wb.Par.This().(Widget).IsVisible()
}

func (wb *WidgetBase) IsDirectWinUpload() bool {
	return false
}

func (wb *WidgetBase) DirectWinUpload() {
}
