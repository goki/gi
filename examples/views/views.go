// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate goki generate

import (
	"fmt"

	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/gimain"
	"goki.dev/gi/v2/giv"
	"goki.dev/goosi/events"
	"goki.dev/icons"
	"goki.dev/mat32/v2"
)

func main() { gimain.Run(app) }

// TableStruct is a testing struct for table view
type TableStruct struct { //gti:add

	// an icon
	Icon icons.Icon

	// an integer field
	IntField int

	// a float field
	FloatField float32

	// a string field
	StrField string

	// a file
	File gi.FileName
}

// ILStruct is an inline-viewed struct
type ILStruct struct { //gti:add

	// click to show next
	On bool

	// can u see me?
	ShowMe string `viewif:"On"`

	// a conditional
	Cond int `viewif:"On"`

	// On and Cond=0 -- note that slbool as bool cannot be used directly..
	Cond1 string `viewif:"On&&Cond==0"`

	// if Cond=0
	Cond2 TableStruct `viewif:"On&&Cond<=1"`

	// a value
	Val float32
}

// Struct is a testing struct for struct view
type Struct struct { //gti:add

	// an enum
	Stripes gi.Stripes

	// a string
	Name string `viewif:"!(Stripes==[RowStripes,ColStripes])"`

	// click to show next
	ShowNext bool

	// can u see me?
	ShowMe string `viewif:"ShowNext"`

	// how about that
	Inline ILStruct `view:"inline"`

	// a conditional
	Cond int

	// if Cond=0
	Cond1 string `viewif:"Cond==0"`

	// if Cond=0
	Cond2 TableStruct `viewif:"Cond>=0"`

	// a value
	Val float32

	Vec mat32.Vec2

	Things []*TableStruct

	Stuff []float32
}

func app() {
	tstslice := make([]string, 20)

	for i := 0; i < len(tstslice); i++ {
		tstslice[i] = fmt.Sprintf("el: %v", i)
	}

	tstmap := make(map[string]string)

	tstmap["mapkey1"] = "whatever"
	tstmap["mapkey2"] = "testing"
	tstmap["mapkey3"] = "boring"

	tsttable := make([]*TableStruct, 100)

	for i := range tsttable {
		ts := &TableStruct{IntField: i, FloatField: float32(i) / 10.0}
		tsttable[i] = ts
	}

	var stru Struct
	stru.Name = "happy"
	stru.Cond = 2
	stru.Val = 3.1415
	stru.Vec.Set(5, 7)
	stru.Inline.Val = 3
	stru.Cond2.IntField = 22
	stru.Cond2.FloatField = 44.4
	stru.Cond2.StrField = "fi"
	// stru.Cond2.File = gi.FileName("views.go")
	stru.Things = make([]*TableStruct, 2)
	stru.Stuff = make([]float32, 3)

	// turn this on to see a trace of the rendering
	// gi.WinEventTrace = true
	// gi.RenderTrace = true
	// gi.LayoutTrace = true
	// gi.WinRenderTrace = true
	// gi.UpdateTrace = true
	// gi.KeyEventTrace = true

	b := gi.NewAppBody("views").SetTitle("GoGi Views Test")

	b.App().About = `This is a demo of the MapView and SliceView views in the <b>GoGi</b> graphical interface system, within the <b>GoKi</b> tree framework.  See <a href="https://github.com/goki">GoKi on GitHub</a>`

	b.AddAppBar(func(tb *gi.Toolbar) {
		gi.NewButton(tb, "slice-test").SetText("SliceDialog").
			SetTooltip("open a SliceViewDialog slice view with a lot of elments, for performance testing").
			OnClick(func(e events.Event) {
				sl := make([]float32, 2880)
				d := gi.NewBody().AddTitle("SliceView Test").AddText("It should open quickly.")
				giv.NewSliceView(d).SetSlice(&sl)
				d.NewFullDialog(tb).Run()
			})
		gi.NewButton(tb, "table-test").SetText("TableDialog").
			SetTooltip("open a TableViewDialog view").
			OnClick(func(e events.Event) {
				d := gi.NewBody().AddTitle("TableView Test").AddText("how does it resize.")
				giv.NewTableView(d).SetSlice(&tsttable)
				d.NewFullDialog(tb).Run()
			})
	})

	// split := gi.NewSplits(b, "split")
	// split.Dim = mat32.X
	// split.SetSplits(.3, .2, .2, .3)
	// split.SetSplits(.5, .5)

	ts := gi.NewTabs(b)
	tst := ts.NewTab("StructView")
	tmv := ts.NewTab("MapView")
	tsl := ts.NewTab("SliceView")
	ttv := ts.NewTab("TabView")

	strv := giv.NewStructView(tst, "strv")
	strv.SetStruct(&stru)

	mv := giv.NewMapView(tmv, "mv")
	mv.SetMap(&tstmap)

	sv := giv.NewSliceView(tsl, "sv")
	// sv.SetState(true, states.ReadOnly)
	sv.SetSlice(&tstslice)

	tv := giv.NewTableView(ttv, "tv")
	// tv.SetState(true, states.ReadOnly)
	tv.SetSlice(&tsttable)

	b.NewWindow().Run().Wait()
}
