// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package giv

import (
	"fmt"
	"log"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"goki.dev/enums"
	"goki.dev/gi/v2/gi"
	"goki.dev/girl/paint"
	"goki.dev/girl/states"
	"goki.dev/girl/styles"
	"goki.dev/goosi/events"
	"goki.dev/gti"
	"goki.dev/icons"
	"goki.dev/ki/v2"
	"goki.dev/laser"
	"goki.dev/pi/v2/filecat"
)

// values contains all the Values for basic builtin types

////////////////////////////////////////////////////////////////////////////////////////
//  StructValue

// StructValue presents a button to edit the struct
type StructValue struct {
	ValueBase
}

func (vv *StructValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ButtonType
	return vv.WidgetTyp
}

func (vv *StructValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	bt := vv.Widget.(*gi.Button)
	npv := laser.NonPtrValue(vv.Value)
	if laser.ValueIsZero(vv.Value) || laser.ValueIsZero(npv) {
		bt.SetText("nil")
	} else {
		opv := laser.OnePtrUnderlyingValue(vv.Value)
		if lbler, ok := opv.Interface().(gi.Labeler); ok {
			bt.SetText(lbler.Label())
		} else {
			txt := fmt.Sprintf("%T", npv.Interface())
			if txt == "" {
				fmt.Printf("no label for struct!")
			}
			bt.SetText(txt)
		}
	}
}

func (vv *StructValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	vv.CreateTempIfNotPtr() // we need our value to be a ptr to a struct -- if not make a tmp
	bt := vv.Widget.(*gi.Button)
	bt.SetType(gi.ButtonTonal)
	bt.Icon = icons.Edit
	bt.Tooltip, _ = vv.Desc()
	bt.Config(sc)
	bt.OnClick(func(e events.Event) {
		vv.OpenDialog(bt, nil)
	})
	vv.UpdateWidget()
}

func (vv *StructValue) HasButton() bool {
	return true
}

func (vv *StructValue) OpenDialog(ctx gi.Widget, fun func(dlg *gi.Dialog)) {
	title, newPath, isZero := vv.Label()
	if isZero {
		return
	}
	vpath := vv.ViewPath + "/" + newPath
	opv := laser.OnePtrUnderlyingValue(vv.Value)
	desc, _ := vv.Desc()
	if desc == "list" { // todo: not sure where this comes from but it is uninformative
		desc = ""
	}
	inact := vv.This().(Value).IsInactive()
	StructViewDialog(vv.Widget, DlgOpts{Title: title, Prompt: desc, TmpSave: vv.TmpSave, Inactive: inact, ViewPath: vpath}, opv.Interface(), func(dlg *gi.Dialog) {
		if dlg.Accepted {
			vv.UpdateWidget()
			vv.SendChange()
		}
		if fun != nil {
			fun(dlg)
		}
	}).Run()
}

////////////////////////////////////////////////////////////////////////////////////////
//  StructInlineValue

// StructInlineValue presents a StructViewInline for a struct
type StructInlineValue struct {
	ValueBase
}

func (vv *StructInlineValue) WidgetType() *gti.Type {
	vv.WidgetTyp = StructViewInlineType
	return vv.WidgetTyp
}

func (vv *StructInlineValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	sv := vv.Widget.(*StructViewInline)
	cst := vv.Value.Interface()
	if sv.Struct != cst {
		sv.SetStruct(cst)
	} else {
		sv.UpdateFields()
	}
}

func (vv *StructInlineValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	sv := vv.Widget.(*StructViewInline)
	sv.Sc = sc
	sv.Tooltip, _ = vv.Desc()
	sv.StructValView = vv
	sv.ViewPath = vv.ViewPath
	sv.TmpSave = vv.TmpSave
	vv.CreateTempIfNotPtr() // we need our value to be a ptr to a struct -- if not make a tmp
	sv.SetStruct(vv.Value.Interface())
	sv.OnChange(func(e events.Event) {
		// vv.UpdateWidget() // not needed?
		vv.SendChange()
	})
	vv.UpdateWidget()
}

////////////////////////////////////////////////////////////////////////////////////////
//  SliceValue

// SliceValue presents a button to edit slices
type SliceValue struct {
	ValueBase
	IsArray    bool         // is an array, not a slice
	ElType     reflect.Type // type of element in the slice -- has pointer if slice has pointers
	ElIsStruct bool         // whether non-pointer element type is a struct or not
}

func (vv *SliceValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ButtonType
	return vv.WidgetTyp
}

func (vv *SliceValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	ac := vv.Widget.(*gi.Button)
	npv := laser.NonPtrValue(vv.Value)
	txt := ""
	if !npv.IsValid() {
		txt = "nil"
	} else if npv.Kind() == reflect.Interface {
		txt = fmt.Sprintf("Slice: %T", npv.Interface())
	} else {
		if npv.Kind() == reflect.Array {
			txt = npv.String()
			// txt = fmt.Sprintf("Array [%v]%v", npv.Len(), vv.ElType.String())
		} else if npv.IsNil() {
			txt = "nil"
		} else {
			txt = npv.String()
			// txt = fmt.Sprintf("Slice [%v]%v", npv.Len(), vv.ElType.String())
		}
	}
	ac.SetText(txt)
}

func (vv *SliceValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	slci := vv.Value.Interface()
	vv.IsArray = laser.NonPtrType(reflect.TypeOf(slci)).Kind() == reflect.Array
	if slci != nil && !laser.AnyIsNil(slci) {
		vv.ElType = laser.SliceElType(slci)
		vv.ElIsStruct = (laser.NonPtrType(vv.ElType).Kind() == reflect.Struct)
	}
	bt := vv.Widget.(*gi.Button)
	bt.SetType(gi.ButtonTonal)
	bt.Icon = icons.Edit
	bt.Tooltip, _ = vv.Desc()
	bt.Config(sc)
	bt.OnClick(func(e events.Event) {
		vv.OpenDialog(bt, nil)
	})
	vv.UpdateWidget()
}

func (vv *SliceValue) HasButton() bool {
	return true
}

func (vv *SliceValue) OpenDialog(ctx gi.Widget, fun func(dlg *gi.Dialog)) {
	title, newPath, isZero := vv.Label()
	if isZero {
		return
	}
	vpath := vv.ViewPath + "/" + newPath
	desc, _ := vv.Desc()
	vvp := laser.OnePtrValue(vv.Value)
	if vvp.Kind() != reflect.Ptr {
		slog.Error("giv.SliceValue: Cannot view slices with non-pointer struct elements")
		return
	}
	inact := vv.This().(Value).IsInactive()
	slci := vvp.Interface()
	if !vv.IsArray && vv.ElIsStruct {
		TableViewDialog(vv.Widget, DlgOpts{Title: title, Prompt: desc, TmpSave: vv.TmpSave, Inactive: inact, ViewPath: vpath}, slci, nil, func(dlg *gi.Dialog) {
			if dlg.Accepted {
				vv.UpdateWidget()
				vv.SendChange()
			}
			if fun != nil {
				fun(dlg)
			}

		}).Run()
	} else {
		SliceViewDialog(vv.Widget, DlgOpts{Title: title, Prompt: desc, TmpSave: vv.TmpSave, Inactive: inact, ViewPath: vpath}, slci, nil, func(dlg *gi.Dialog) {
			if dlg.Accepted {
				vv.UpdateWidget()
				vv.SendChange()
			}
			if fun != nil {
				fun(dlg)
			}
		}).Run()
	}
}

////////////////////////////////////////////////////////////////////////////////////////
//  SliceInlineValue

// SliceInlineValue presents a SliceViewInline for a map
type SliceInlineValue struct {
	ValueBase
}

func (vv *SliceInlineValue) WidgetType() *gti.Type {
	vv.WidgetTyp = SliceViewInlineType
	return vv.WidgetTyp
}

func (vv *SliceInlineValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	sv := vv.Widget.(*SliceViewInline)
	csl := vv.Value.Interface()
	if sv.Slice != csl {
		sv.SetSlice(csl)
	} else {
		sv.UpdateValues()
	}
}

func (vv *SliceInlineValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	sv := vv.Widget.(*SliceViewInline)
	sv.Sc = sc
	sv.Tooltip, _ = vv.Desc()
	sv.SliceValView = vv
	sv.ViewPath = vv.ViewPath
	sv.TmpSave = vv.TmpSave
	// npv := vv.Value.Elem()
	sv.SetState(vv.This().(Value).IsInactive(), states.Disabled)
	sv.SetSlice(vv.Value.Interface())
	sv.OnChange(func(e events.Event) {
		vv.SendChange()
	})
}

////////////////////////////////////////////////////////////////////////////////////////
//  MapValue

// MapValue presents a button to edit maps
type MapValue struct {
	ValueBase
}

func (vv *MapValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ButtonType
	return vv.WidgetTyp
}

func (vv *MapValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	bt := vv.Widget.(*gi.Button)
	npv := laser.NonPtrValue(vv.Value)
	mpi := vv.Value.Interface()
	txt := ""
	if !npv.IsValid() || npv.IsNil() {
		txt = "nil"
	} else if npv.Kind() == reflect.Interface {
		txt = fmt.Sprintf("Map: %T", npv.Interface())
	} else {
		txt = fmt.Sprintf("Map: [%v %v]%v", npv.Len(), laser.MapKeyType(mpi).String(), laser.MapValueType(mpi).String())
	}
	bt.SetText(txt)
}

func (vv *MapValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	bt := vv.Widget.(*gi.Button)
	bt.SetType(gi.ButtonTonal)
	bt.Icon = icons.Edit
	bt.Tooltip, _ = vv.Desc()
	bt.Config(sc)
	bt.OnClick(func(e events.Event) {
		vv.OpenDialog(bt, nil)
	})
	vv.UpdateWidget()
}

func (vv *MapValue) HasButton() bool {
	return true
}

func (vv *MapValue) OpenDialog(ctx gi.Widget, fun func(dlg *gi.Dialog)) {
	title, newPath, isZero := vv.Label()
	if isZero {
		return
	}
	vpath := vv.ViewPath + "/" + newPath
	desc, _ := vv.Desc()
	mpi := vv.Value.Interface()
	inact := vv.This().(Value).IsInactive()
	MapViewDialog(vv.Widget, DlgOpts{Title: title, Prompt: desc, TmpSave: vv.TmpSave, Inactive: inact, ViewPath: vpath}, mpi, func(dlg *gi.Dialog) {
		if dlg.Accepted {
			vv.UpdateWidget()
			vv.SendChange()
		}
		if fun != nil {
			fun(dlg)
		}
	}).Run()
}

////////////////////////////////////////////////////////////////////////////////////////
//  MapInlineValue

// MapInlineValue presents a MapViewInline for a map
type MapInlineValue struct {
	ValueBase
}

func (vv *MapInlineValue) WidgetType() *gti.Type {
	vv.WidgetTyp = MapViewInlineType
	return vv.WidgetTyp
}

func (vv *MapInlineValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	sv := vv.Widget.(*MapViewInline)
	cmp := vv.Value.Interface()
	if sv.Map != cmp {
		sv.SetMap(cmp)
	} else {
		sv.UpdateValues()
	}
}

func (vv *MapInlineValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	sv := vv.Widget.(*MapViewInline)
	sv.Sc = sc
	sv.Tooltip, _ = vv.Desc()
	sv.MapValView = vv
	sv.ViewPath = vv.ViewPath
	sv.TmpSave = vv.TmpSave
	// npv := vv.Value.Elem()
	sv.SetState(vv.This().(Value).IsInactive(), states.Disabled)
	sv.SetMap(vv.Value.Interface())
	sv.OnChange(func(e events.Event) {
		vv.SendChange()
	})
}

////////////////////////////////////////////////////////////////////////////////////////
//  KiPtrValue

// KiPtrValue provides a chooser for pointers to Ki objects
type KiPtrValue struct {
	ValueBase
}

func (vv *KiPtrValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ButtonType
	return vv.WidgetTyp
}

// get the Ki struct itself (or nil)
func (vv *KiPtrValue) KiStruct() ki.Ki {
	if !vv.Value.IsValid() {
		return nil
	}
	if vv.Value.IsNil() {
		return nil
	}
	npv := vv.Value
	if vv.Value.Kind() == reflect.Ptr {
		npv = vv.Value.Elem()
	}
	if npv.Kind() == reflect.Struct {
		npv = vv.Value // go back up
	}
	if !npv.IsNil() {
		k, ok := npv.Interface().(ki.Ki)
		if ok && k != nil {
			return k
		}
	}
	return nil
}

func (vv *KiPtrValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	bt := vv.Widget.(*gi.Button)
	path := "nil"
	k := vv.KiStruct()
	if k != nil {
		path = k.Path()
	}
	bt.SetText(path)
}

func (vv *KiPtrValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	bt := vv.Widget.(*gi.Button)
	bt.SetType(gi.ButtonTonal)
	bt.Indicator = icons.KeyboardArrowDown
	bt.Tooltip, _ = vv.Desc()
	bt.Menu = func(m *gi.Scene) {
		gi.NewButton(m, "edit").SetText("Edit").OnClick(func(e events.Event) {
			k := vv.KiStruct()
			if k != nil {
				bt := vv.Widget.(*gi.Button)
				vv.OpenDialog(bt, nil)
			}
		})
		gi.NewButton(m, "gogi-editor").SetText("GoGi editor").OnClick(func(e events.Event) {
			k := vv.KiStruct()
			if k != nil {
				GoGiEditorDialog(k)
			}
		})
	}
	bt.Config(sc)
	vv.UpdateWidget()
}

func (vv *KiPtrValue) HasButton() bool {
	return true
}

func (vv *KiPtrValue) OpenDialog(ctx gi.Widget, fun func(dlg *gi.Dialog)) {
	title, newPath, isZero := vv.Label()
	if isZero {
		return
	}
	k := vv.KiStruct()
	if k == nil {
		return
	}
	vpath := vv.ViewPath + "/" + newPath
	desc, _ := vv.Desc()
	inact := vv.This().(Value).IsInactive()
	StructViewDialog(ctx, DlgOpts{Title: title, Prompt: desc, TmpSave: vv.TmpSave, Inactive: inact, ViewPath: vpath}, k, func(dlg *gi.Dialog) {
		if dlg.Accepted {
			vv.UpdateWidget()
			vv.SendChange()
		}
		if fun != nil {
			fun(dlg)
		}
	})
}

////////////////////////////////////////////////////////////////////////////////////////
//  BoolValue

// BoolValue presents a checkbox for a boolean
type BoolValue struct {
	ValueBase
}

func (vv *BoolValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.SwitchType
	return vv.WidgetTyp
}

func (vv *BoolValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	cb := vv.Widget.(*gi.Switch)
	npv := laser.NonPtrValue(vv.Value)
	bv, _ := laser.ToBool(npv.Interface())
	cb.SetState(bv, states.Checked)
}

func (vv *BoolValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	cb := vv.Widget.(*gi.Switch)
	cb.Tooltip, _ = vv.Desc()
	cb.SetState(vv.This().(Value).IsInactive(), states.Disabled)
	cb.Config(sc)
	cb.OnChange(func(e events.Event) {
		vv.SetValue(cb.StateIs(states.Checked))
	})
	vv.UpdateWidget()
}

////////////////////////////////////////////////////////////////////////////////////////
//  IntValue

// IntValue presents a spinner
type IntValue struct {
	ValueBase
}

func (vv *IntValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.SpinnerType
	return vv.WidgetTyp
}

func (vv *IntValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	sb := vv.Widget.(*gi.Spinner)
	npv := laser.NonPtrValue(vv.Value)
	fv, err := laser.ToFloat32(npv.Interface())
	if err == nil {
		sb.SetValue(fv)
	} else {
		slog.Error("Int Value set", "error:", err)
	}
}

func (vv *IntValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	sb := vv.Widget.(*gi.Spinner)
	sb.Tooltip, _ = vv.Desc()
	sb.SetState(vv.This().(Value).IsInactive(), states.Disabled)
	sb.Step = 1.0
	sb.PageStep = 10.0
	// STYTODO: figure out what to do about this
	// sb.Parts.AddChildStyler("textfield", 0, gi.StylerParent(vv), func(tf *gi.WidgetBase) {
	// 	s.Width.SetCh(5)
	// })
	vk := vv.Value.Kind()
	if vk >= reflect.Uint && vk <= reflect.Uint64 {
		sb.SetMin(0)
	}
	if mintag, ok := vv.Tag("min"); ok {
		minv, err := laser.ToFloat32(mintag)
		if err == nil {
			sb.SetMin(minv)
		} else {
			slog.Error("Int Min Value:", "error:", err)
		}
	}
	if maxtag, ok := vv.Tag("max"); ok {
		maxv, err := laser.ToFloat32(maxtag)
		if err == nil {
			sb.SetMax(maxv)
		} else {
			slog.Error("Int Max Value:", "error:", err)
		}
	}
	if steptag, ok := vv.Tag("step"); ok {
		step, err := laser.ToFloat32(steptag)
		if err == nil {
			sb.Step = step
		} else {
			slog.Error("Int Step Value:", "error:", err)
		}
	}
	if fmttag, ok := vv.Tag("format"); ok {
		sb.Format = fmttag
	}
	sb.Config(sc)
	sb.OnChange(func(e events.Event) {
		vv.SetValue(sb.Value)
	})
	vv.UpdateWidget()
}

////////////////////////////////////////////////////////////////////////////////////////
//  FloatValue

// FloatValue presents a spinner
type FloatValue struct {
	ValueBase
}

func (vv *FloatValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.SpinnerType
	return vv.WidgetTyp
}

func (vv *FloatValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	sb := vv.Widget.(*gi.Spinner)
	npv := laser.NonPtrValue(vv.Value)
	fv, err := laser.ToFloat32(npv.Interface())
	if err == nil {
		sb.SetValue(fv)
	} else {
		slog.Error("Float Value set", "error:", err)
	}
}

func (vv *FloatValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	sb := vv.Widget.(*gi.Spinner)
	sb.Tooltip, _ = vv.Desc()
	sb.SetState(vv.This().(Value).IsInactive(), states.Disabled)
	sb.Step = 1.0
	sb.PageStep = 10.0
	if mintag, ok := vv.Tag("min"); ok {
		minv, err := laser.ToFloat32(mintag)
		if err == nil {
			sb.HasMin = true
			sb.Min = minv
		} else {
			slog.Error("Float Min Value:", "error:", err)
		}
	}
	if maxtag, ok := vv.Tag("max"); ok {
		maxv, err := laser.ToFloat32(maxtag)
		if err == nil {
			sb.HasMax = true
			sb.Max = maxv
		} else {
			slog.Error("Float Max Value:", "error:", err)
		}
	}
	sb.Step = .1 // smaller default
	if steptag, ok := vv.Tag("step"); ok {
		step, err := laser.ToFloat32(steptag)
		if err == nil {
			sb.Step = step
		} else {
			slog.Error("Float Step Value:", "error:", err)
		}
	}
	if fmttag, ok := vv.Tag("format"); ok {
		sb.Format = fmttag
	}
	sb.Config(sc)
	sb.OnChange(func(e events.Event) {
		vv.SetValue(sb.Value)
	})
	vv.UpdateWidget()
}

////////////////////////////////////////////////////////////////////////////////////////
//  EnumValue

// EnumValue presents a chooser for choosing enums
type EnumValue struct {
	ValueBase
}

func (vv *EnumValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ChooserType
	return vv.WidgetTyp
}

func (vv *EnumValue) EnumValue() enums.Enum {
	ev, ok := vv.Value.Interface().(enums.Enum)
	if ok {
		return ev
	}
	slog.Error("giv.EnumValue: type must be enums.Enum")
	return nil
}

// func (vv *EnumValue) SetEnumValueFromInt(ival int64) bool {
// 	// typ := vv.EnumType()
// 	// eval := laser.EnumIfaceFromInt64(ival, typ)
// 	return vv.SetValue(ival)
// }

func (vv *EnumValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	ch := vv.Widget.(*gi.Chooser)
	npv := laser.NonPtrValue(vv.Value)
	ch.SetCurVal(npv.Interface())
	// iv, err := laser.ToInt(npv.Interface())
	// if err == nil {
	// 	ch.SetCurIndex(int(iv)) // todo: currently only working for 0-based values
	// } else {
	// 	slog.Error("Enum Value:", err)
	// }
}

func (vv *EnumValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	ch := vv.Widget.(*gi.Chooser)
	ch.Tooltip, _ = vv.Desc()
	ch.SetState(vv.This().(Value).IsInactive(), states.Disabled)

	ev := vv.EnumValue()
	ch.ItemsFromEnum(ev, false, 50)
	ch.Config(sc)
	ch.OnChange(func(e events.Event) {
		vv.SetValue(ch.CurVal)
		// cval := ch.CurVal.(enums.Enum)
		// vv.SetEnumValueFromInt(cval.Int64()) // todo: using index
	})
	vv.UpdateWidget()
}

////////////////////////////////////////////////////////////////////////////////////////
//  BitFlagView

// BitFlagView presents a ButtonBox for bitflags
type BitFlagView struct {
	ValueBase
	AltType reflect.Type // alternative type, e.g., from EnumType: property
}

func (vv *BitFlagView) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.SwitchesType
	return vv.WidgetTyp
}

func (vv *BitFlagView) EnumValue() enums.BitFlag {
	ev, ok := vv.Value.Interface().(enums.BitFlag)
	if ok {
		return ev
	}
	slog.Error("giv.BitFlagView: type must be enums.BitFlag")
	return nil
}

func (vv *BitFlagView) SetEnumValueFromInt(ival int64) bool {
	// todo: needs to set flags?
	// typ := vv.EnumType()
	// eval := laser.EnumIfaceFromInt64(ival, typ)
	return vv.SetValue(ival)
}

func (vv *BitFlagView) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	bb := vv.Widget.(*gi.Switches)
	_ = bb
	npv := laser.NonPtrValue(vv.Value)
	iv, err := laser.ToInt(npv.Interface())
	_ = iv
	if err == nil {
		// ev := vv.EnumValue() // todo:
		// bb.UpdateFromBitFlags(typ, int64(iv))
	} else {
		slog.Error("BitFlag Value:", "error:", err)
	}
}

func (vv *BitFlagView) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	cb := vv.Widget.(*gi.Switches)
	// vv.StdConfigWidget(cb.Parts)
	// cb.Parts.Lay = gi.LayoutHoriz
	cb.Tooltip, _ = vv.Desc()
	cb.SetState(vv.This().(Value).IsInactive(), states.Disabled)

	// todo!
	ev := vv.EnumValue()
	_ = ev
	// cb.ItemsFromEnum(ev)
	cb.Config(sc)
	// cb.ButtonSig.ConnectOnly(vv.This(), func(recv, send ki.Ki, sig int64, data any) {
	// 	vvv, _ := recv.Embed(TypeBitFlagView).(*BitFlagView)
	// 	cbb := vvv.Widget.(*gi.Switches)
	// 	etyp := vvv.EnumType()
	// 	val := cbb.BitFlagsValue(etyp)
	// 	vvv.SetEnumValueFromInt(val)
	// 	// vvv.UpdateWidget()
	// })
	vv.UpdateWidget()
}

////////////////////////////////////////////////////////////////////////////////////////
//  TypeValue

// TypeValue presents a chooser for choosing types
type TypeValue struct {
	ValueBase
}

func (vv *TypeValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ChooserType
	return vv.WidgetTyp
}

func (vv *TypeValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	sb := vv.Widget.(*gi.Chooser)
	npv := laser.NonPtrValue(vv.Value)
	typ, ok := npv.Interface().(*gti.Type)
	if ok {
		sb.SetCurVal(typ)
	}
}

func (vv *TypeValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	cb := vv.Widget.(*gi.Chooser)
	cb.Tooltip, _ = vv.Desc()
	cb.SetState(vv.This().(Value).IsInactive(), states.Disabled)

	typEmbeds := ki.NodeType
	// if kiv, ok := vv.Owner.(ki.Ki); ok {
	// 	if tep, ok := kiv.PropInherit("type-embeds", ki.Inherit, ki.TypeProps); ok {
	// 		// todo:
	// 		// if te, ok := tep.(reflect.Type); ok {
	// 		// 	typEmbeds = te
	// 		// }
	// 	}
	// }
	if tetag, ok := vv.Tag("type-embeds"); ok {
		typ := gti.TypeByName(tetag)
		if typ != nil {
			typEmbeds = typ
		}
	}

	tl := gti.AllEmbeddersOf(typEmbeds)
	cb.ItemsFromTypes(tl, false, true, 50)
	cb.Config(sc)
	cb.OnChange(func(e events.Event) {
		tval := cb.CurVal.(*gti.Type)
		vv.SetValue(tval)
	})
	vv.UpdateWidget()
}

////////////////////////////////////////////////////////////////////////////////////////
//  ByteSliceValue

// ByteSliceValue presents a textfield of the bytes
type ByteSliceValue struct {
	ValueBase
}

func (vv *ByteSliceValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.TextFieldType
	return vv.WidgetTyp
}

func (vv *ByteSliceValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	tf := vv.Widget.(*gi.TextField)
	npv := laser.NonPtrValue(vv.Value)
	bv, ok := npv.Interface().([]byte)
	if ok {
		tf.SetText(string(bv))
	}
}

func (vv *ByteSliceValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	tf := vv.Widget.(*gi.TextField)
	tf.Tooltip, _ = vv.Desc()
	tf.SetState(vv.This().(Value).IsInactive(), states.Disabled)
	// STYTODO: figure out how how to handle these kinds of styles
	tf.Style(func(s *styles.Style) {
		s.MinWidth.SetCh(16)
		s.MaxWidth.SetDp(-1)
	})
	tf.Config(sc)

	tf.OnChange(func(e events.Event) {
		vv.SetValue(tf.Text())
	})
	vv.UpdateWidget()
}

////////////////////////////////////////////////////////////////////////////////////////
//  RuneSliceValue

// RuneSliceValue presents a textfield of the bytes
type RuneSliceValue struct {
	ValueBase
}

func (vv *RuneSliceValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.TextFieldType
	return vv.WidgetTyp
}

func (vv *RuneSliceValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	tf := vv.Widget.(*gi.TextField)
	npv := laser.NonPtrValue(vv.Value)
	rv, ok := npv.Interface().([]rune)
	if ok {
		tf.SetText(string(rv))
	}
}

func (vv *RuneSliceValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	tf := vv.Widget.(*gi.TextField)
	tf.Tooltip, _ = vv.Desc()
	tf.SetState(vv.This().(Value).IsInactive(), states.Disabled)
	tf.Style(func(s *styles.Style) {
		s.MinWidth.SetCh(16)
		s.MaxWidth.SetDp(-1)
	})
	tf.Config(sc)

	tf.OnChange(func(e events.Event) {
		vv.SetValue(tf.Text())
	})
	vv.UpdateWidget()
}

////////////////////////////////////////////////////////////////////////////////////////
//  NilValue

// NilValue presents a label saying 'nil' -- for any nil or otherwise unrepresentable items
type NilValue struct {
	ValueBase
}

func (vv *NilValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.LabelType
	return vv.WidgetTyp
}

func (vv *NilValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	sb := vv.Widget.(*gi.Label)
	npv := laser.NonPtrValue(vv.Value)
	tstr := ""
	if !laser.ValueIsZero(npv) {
		tstr = npv.String() // npv.Type().String()
	} else if !laser.ValueIsZero(vv.Value) {
		tstr = vv.Value.String() // vv.Value.Type().String()
	}
	sb.SetText("nil " + tstr)
}

func (vv *NilValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	sb := vv.Widget.(*gi.Label)
	sb.Tooltip, _ = vv.Desc()
	sb.Config(sc)
	vv.UpdateWidget()
}

////////////////////////////////////////////////////////////////////////////////////////
//  TimeValue

var DefaultTimeFormat = "2006-01-02 15:04:05 MST"

// TimeValue presents a text field for a time
type TimeValue struct {
	ValueBase
}

func (vv *TimeValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.TextFieldType
	return vv.WidgetTyp
}

// TimeVal decodes Value into a *time.Time value -- also handles FileTime case
func (vv *TimeValue) TimeVal() *time.Time {
	tmi := laser.PtrValue(vv.Value).Interface()
	switch v := tmi.(type) {
	case *time.Time:
		return v
	case *filecat.FileTime:
		return (*time.Time)(v)
	}
	return nil
}

func (vv *TimeValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	tf := vv.Widget.(*gi.TextField)
	tm := vv.TimeVal()
	tf.SetText(tm.Format(DefaultTimeFormat))
}

func (vv *TimeValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	tf := vv.Widget.(*gi.TextField)
	tf.SetStretchMaxWidth()
	tf.Tooltip, _ = vv.Desc()
	tf.SetState(vv.This().(Value).IsInactive(), states.Disabled)
	tf.Style(func(s *styles.Style) {
		tf.Styles.MinWidth.SetCh(float32(len(DefaultTimeFormat) + 2))
	})
	tf.Config(sc)
	tf.OnChange(func(e events.Event) {
		nt, err := time.Parse(DefaultTimeFormat, tf.Text())
		if err != nil {
			log.Println(err)
		} else {
			tm := vv.TimeVal()
			*tm = nt
			// vv.SendChange()
			vv.UpdateWidget()
		}
	})
	vv.UpdateWidget()
}

////////////////////////////////////////////////////////////////////////////////////////
//  IconValue

// IconValue presents an action for displaying an IconName and selecting
// icons from IconChooserDialog
type IconValue struct {
	ValueBase
}

func (vv *IconValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ButtonType
	return vv.WidgetTyp
}

func (vv *IconValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	bt := vv.Widget.(*gi.Button)
	txt := laser.ToString(vv.Value.Interface())
	if icons.Icon(txt).IsNil() {
		bt.SetIcon("blank")
	} else {
		bt.SetIcon(icons.Icon(txt))
	}
	if sntag, ok := vv.Tag("view"); ok {
		if strings.Contains(sntag, "show-name") {
			if txt == "" {
				txt = "none"
			}
			bt.SetText(txt)
		}
	}
}

func (vv *IconValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	bt := vv.Widget.(*gi.Button)
	bt.SetType(gi.ButtonTonal)
	bt.Config(sc)
	bt.OnClick(func(e events.Event) {
		vv.OpenDialog(bt, nil)
	})
	vv.UpdateWidget()
}

func (vv *IconValue) HasDialog() bool {
	return true
}

func (vv *IconValue) OpenDialog(ctx gi.Widget, fun func(dlg *gi.Dialog)) {
	if vv.IsInactive() {
		return
	}
	cur := icons.Icon(laser.ToString(vv.Value.Interface()))
	desc, _ := vv.Desc()
	IconChooserDialog(ctx, DlgOpts{Title: "Select an Icon", Prompt: desc}, cur, func(dlg *gi.Dialog) {
		if dlg.Accepted {
			si := dlg.Data.(int)
			if si >= 0 {
				ic := gi.CurIconList[si]
				vv.SetValue(ic)
				vv.UpdateWidget()
			}
		}
		if fun != nil {
			fun(dlg)
		}
	}).Run()
}

////////////////////////////////////////////////////////////////////////////////////////
//  FontValue

// FontValue presents an action for displaying a FontName and selecting
// fonts from FontChooserDialog
type FontValue struct {
	ValueBase
}

func (vv *FontValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ButtonType
	return vv.WidgetTyp
}

func (vv *FontValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	bt := vv.Widget.(*gi.Button)
	txt := laser.ToString(vv.Value.Interface())
	bt.SetProp("font-family", txt)
	bt.SetText(txt)
}

func (vv *FontValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	bt := vv.Widget.(*gi.Button)
	bt.SetType(gi.ButtonTonal)
	bt.Config(sc)
	bt.OnClick(func(e events.Event) {
		vv.OpenDialog(vv.Widget, nil)
	})
	vv.UpdateWidget()
}

func (vv *FontValue) HasDialog() bool {
	return true
}

func (vv *FontValue) OpenDialog(ctx gi.Widget, fun func(dlg *gi.Dialog)) {
	if vv.IsInactive() {
		return
	}
	// cur := gi.FontName(laser.ToString(vvv.Value.Interface()))
	desc, _ := vv.Desc()
	FontChooserDialog(ctx, DlgOpts{Title: "Select a Font", Prompt: desc}, func(dlg *gi.Dialog) {
		if dlg.Accepted {
			si := dlg.Data.(int)
			if si >= 0 {
				fi := paint.FontLibrary.FontInfo[si]
				vv.SetValue(fi.Name)
				vv.UpdateWidget()
			}
		}
		if fun != nil {
			fun(dlg)
		}
	}).Run()
}

////////////////////////////////////////////////////////////////////////////////////////
//  FileValue

// FileValue presents an action for displaying a FileName and selecting
// icons from FileChooserDialog
type FileValue struct {
	ValueBase
}

func (vv *FileValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ButtonType
	return vv.WidgetTyp
}

func (vv *FileValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	bt := vv.Widget.(*gi.Button)
	txt := laser.ToString(vv.Value.Interface())
	if txt == "" {
		txt = "(click to open file chooser)"
	}
	bt.SetText(txt)
}

func (vv *FileValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	vv.StdConfigWidget(widg)
	bt := vv.Widget.(*gi.Button)
	bt.SetType(gi.ButtonTonal)
	bt.Config(sc)
	bt.OnClick(func(e events.Event) {
		bt := vv.Widget.(*gi.Button)
		vv.OpenDialog(bt, nil)
	})
	vv.UpdateWidget()
}

func (vv *FileValue) HasDialog() bool {
	return true
}

func (vv *FileValue) OpenDialog(ctx gi.Widget, fun func(dlg *gi.Dialog)) {
	if vv.IsInactive() {
		return
	}
	cur := laser.ToString(vv.Value.Interface())
	ext, _ := vv.Tag("ext")
	desc, _ := vv.Desc()
	FileViewDialog(ctx, DlgOpts{Title: vv.Name(), Prompt: desc}, cur, ext, nil, func(dlg *gi.Dialog) {
		if dlg.Accepted {
			fn := dlg.Data.(string)
			vv.SetValue(fn)
			vv.UpdateWidget()
		}
		if fun != nil {
			fun(dlg)
		}
	}).Run()
}

////////////////////////////////////////////////////////////////////////////////////////
//  VersCtrlValue

// VersCtrlSystems is a list of supported Version Control Systems.
// These must match the VCS Types from goki/pi/vci which in turn
// is based on masterminds/vcs
var VersCtrlSystems = []string{"git", "svn", "bzr", "hg"}

// IsVersCtrlSystem returns true if the given string matches one of the
// standard VersCtrlSystems -- uses lowercase version of str.
func IsVersCtrlSystem(str string) bool {
	stl := strings.ToLower(str)
	for _, vcn := range VersCtrlSystems {
		if stl == vcn {
			return true
		}
	}
	return false
}

// VersCtrlName is the name of a version control system
type VersCtrlName string

func VersCtrlNameProper(vc string) VersCtrlName {
	vcl := strings.ToLower(vc)
	for _, vcnp := range VersCtrlSystems {
		vcnpl := strings.ToLower(vcnp)
		if strings.Compare(vcl, vcnpl) == 0 {
			return VersCtrlName(vcnp)
		}
	}
	return ""
}

// Value registers VersCtrlValue as the viewer of VersCtrlName
func (kn VersCtrlName) Value() Value {
	vv := &VersCtrlValue{}
	ki.InitNode(vv)
	return vv
}

// VersCtrlValue presents an action for displaying an VersCtrlName and selecting
// from StringPopup
type VersCtrlValue struct {
	ValueBase
}

func (vv *VersCtrlValue) WidgetType() *gti.Type {
	vv.WidgetTyp = gi.ButtonType
	return vv.WidgetTyp
}

func (vv *VersCtrlValue) UpdateWidget() {
	if vv.Widget == nil {
		return
	}
	bt := vv.Widget.(*gi.Button)
	txt := laser.ToString(vv.Value.Interface())
	if txt == "" {
		txt = "(none)"
	}
	bt.SetText(txt)
}

func (vv *VersCtrlValue) ConfigWidget(widg gi.Widget, sc *gi.Scene) {
	vv.Widget = widg
	bt := vv.Widget.(*gi.Button)
	bt.SetType(gi.ButtonTonal)
	bt.Config(sc)
	bt.OnClick(func(e events.Event) {
		vv.OpenDialog(vv.Widget, nil)
	})
	vv.UpdateWidget()
}

func (vv *VersCtrlValue) HasDialog() bool {
	return true
}

func (vv *VersCtrlValue) OpenDialog(ctx gi.Widget, fun func(dlg *gi.Dialog)) {
	if vv.IsInactive() {
		return
	}
	// TODO(kai/menu): add back StringsChooserPopup here
	// cur := laser.ToString(vv.Value.Interface())
	// gi.StringsChooserPopup(VersCtrlSystems, cur, ctx, func(ac *gi.Button) {
	// 	vv.SetValue(ac.Text)
	// 	vv.UpdateWidget()
	// })
}