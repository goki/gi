// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filetree

import (
	"fmt"
	"strings"

	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/gi/v2/keyfun"
	"goki.dev/girl/states"
	"goki.dev/girl/styles"
	"goki.dev/goosi/events"
	"goki.dev/icons"
	"goki.dev/vci/v2"
)

func (fn *Node) HandleFileNodeEvents() {
	// note: put all mouse events in parts!
	// note: OnClick is grabbed by the parts first -- we don't see it
	/*
		fn.On(events.KeyChord, func(e events.Event) {
			kt := e.(*events.Key)
			fn.KeyInput(kt)
		})
			ftvwe.AddFunc(goosi.DNDEvent, gi.RegPri, func(recv, send ki.Ki, sig int64, d any) {
				de := d.(events.Event)
				tvvi := recv.Embed(TypeNode)
				if tvvi == nil {
					return
				}
				tvv := tvvi.(*Node)
				switch de.Action {
				case events.Start:
					tvv.DragNDropStart()
				case events.DropOnTarget:
					tvv.DragNDropTarget(de)
				case events.DropFmSource:
					tvv.This().(gi.DragNDropper).Dragged(de)
				case events.External:
					tvv.DragNDropExternal(de)
				}
			})
			ftvwe.AddFunc(goosi.DNDFocusEvent, gi.RegPri, func(recv, send ki.Ki, sig int64, d any) {
				de := d.(*events.FocusEvent)
				tvvi := recv.Embed(TypeNode)
				if tvvi == nil {
					return
				}
				tvv := tvvi.(*Node)
				switch de.Action {
				case events.Enter:
					tvv.ParentRenderWin().DNDSetCursor(de.Mod)
				case events.Exit:
					tvv.ParentRenderWin().DNDNotCursor()
				case events.Hover:
					tvv.Open()
				}
			})
	*/
	fn.HandleTreeViewEvents()
}

func (fn *Node) KeyInput(kt events.Event) {
	if gi.KeyEventTrace {
		fmt.Printf("TreeView KeyInput: %v\n", fn.Path())
	}
	kf := keyfun.Of(kt.KeyChord())
	selMode := events.SelectModeBits(kt.Modifiers())

	if selMode == events.SelectOne {
		if fn.SelectMode() {
			selMode = events.ExtendContinuous
		}
	}

	// first all the keys that work for ReadOnly and active
	if !fn.IsReadOnly() && !kt.IsHandled() {
		switch kf {
		case keyfun.Delete:
			fn.DeleteFiles()
			kt.SetHandled()
			// todo: remove when gi issue 237 is resolved
		case keyfun.Backspace:
			fn.DeleteFiles()
			kt.SetHandled()
		case keyfun.Duplicate:
			fn.DuplicateFiles()
			kt.SetHandled()
		case keyfun.Insert: // New File
			giv.CallFunc(fn, fn.NewFile)
			kt.SetHandled()
		case keyfun.InsertAfter: // New Folder
			giv.CallFunc(fn, fn.NewFolder)
			kt.SetHandled()
		}
	}
	if !kt.IsHandled() {
		fn.HandleTreeViewKeyChord(kt)
	}
}

// VCSLabelFunc gets the appropriate label for removing from version control
func VCSLabelFunc(fn *Node, label string) string {
	repo, _ := fn.Repo()
	if repo != nil {
		label = strings.Replace(label, "VCS", string(repo.Vcs()), 1)
	}
	return label
}

func (fn *Node) FileNodeContextMenu(m *gi.Scene) {
	giv.NewFuncButton(m, fn.ShowFileInfo).SetIcon(icons.Info).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection(), states.Disabled)
		})
	giv.NewFuncButton(m, fn.OpenFilesDefault).SetText("Open (w/default app)").SetIcon(icons.Open).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection(), states.Disabled)
		})
	gi.NewSeparator(m)

	giv.NewFuncButton(m, fn.DuplicateFiles).SetText("Duplicate").SetIcon(icons.Copy).
		SetKey(keyfun.Duplicate).Style(func(s *styles.Style) {
		s.State.SetFlag(!fn.HasSelection(), states.Disabled)
	})
	giv.NewFuncButton(m, fn.DeleteFiles).SetText("Delete").SetIcon(icons.Delete).
		SetKey(keyfun.Delete).Style(func(s *styles.Style) {
		s.State.SetFlag(!fn.HasSelection(), states.Disabled)
	})
	giv.NewFuncButton(m, fn.RenameFiles).SetText("Rename").SetIcon(icons.NewLabel).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection(), states.Disabled)
		})
	gi.NewSeparator(m)

	giv.NewFuncButton(m, fn.OpenAll).SetIcon(icons.KeyboardArrowDown).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection() || !fn.IsDir(), states.Disabled)
		})
	giv.NewFuncButton(m, fn.CloseAll).SetIcon(icons.KeyboardArrowRight).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection() || !fn.IsDir(), states.Disabled)
		})
	giv.NewFuncButton(m, fn.SortBys).SetText("Sort by").SetIcon(icons.Sort).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection() || !fn.IsDir(), states.Disabled)
		})
	gi.NewSeparator(m)

	giv.NewFuncButton(m, fn.NewFiles).SetText("New file").SetIcon(icons.OpenInNew).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection(), states.Disabled)
		})
	giv.NewFuncButton(m, fn.NewFolders).SetText("New folder").SetIcon(icons.CreateNewFolder).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection(), states.Disabled)
		})
	gi.NewSeparator(m)

	giv.NewFuncButton(m, fn.AddToVcsSel).SetText(VCSLabelFunc(fn, "Add to VCS")).SetIcon(icons.Add).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection() || fn.Info.Vcs != vci.Untracked, states.Disabled)
		})
	giv.NewFuncButton(m, fn.DeleteFromVcsSel).SetText(VCSLabelFunc(fn, "Delete from VCS")).SetIcon(icons.Delete).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection() || fn.Info.Vcs == vci.Untracked, states.Disabled)
		})
	giv.NewFuncButton(m, fn.CommitToVcsSel).SetText(VCSLabelFunc(fn, "Commit to VCS")).SetIcon(icons.Star).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection() || fn.Info.Vcs == vci.Untracked, states.Disabled)
		})
	giv.NewFuncButton(m, fn.RevertVcsSel).SetText(VCSLabelFunc(fn, "Revert from VCS")).SetIcon(icons.Undo).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection() || fn.Info.Vcs == vci.Untracked, states.Disabled)
		})
	gi.NewSeparator(m)

	giv.NewFuncButton(m, fn.DiffVcsSel).SetText(VCSLabelFunc(fn, "Diff VCS")).SetIcon(icons.Add).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection() || fn.Info.Vcs == vci.Untracked, states.Disabled)
		})
	giv.NewFuncButton(m, fn.LogVcsSel).SetText(VCSLabelFunc(fn, "Log VCS")).SetIcon(icons.List).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection() || fn.Info.Vcs == vci.Untracked, states.Disabled)
		})
	giv.NewFuncButton(m, fn.BlameVcsSel).SetText(VCSLabelFunc(fn, "Blame VCS")).SetIcon(icons.CreditScore).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection() || fn.Info.Vcs == vci.Untracked, states.Disabled)
		})
	gi.NewSeparator(m)

	giv.NewFuncButton(m, fn.RemoveFromExterns).SetIcon(icons.Delete).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection(), states.Disabled)
		})

	gi.NewSeparator(m)
	gi.NewButton(m).SetText("Copy").SetIcon(icons.ContentCopy).SetKey(keyfun.Copy).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection(), states.Disabled)
		}).
		OnClick(func(e events.Event) {
			fn.This().(gi.Clipper).Copy(true)
		})
	gi.NewButton(m).SetText("Cut").SetIcon(icons.ContentCut).SetKey(keyfun.Cut).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection(), states.Disabled)
		}).
		OnClick(func(e events.Event) {
			fn.This().(gi.Clipper).Cut()
		})
	pbt := gi.NewButton(m).SetText("Paste").SetIcon(icons.ContentPaste).SetKey(keyfun.Paste).
		Style(func(s *styles.Style) {
			s.State.SetFlag(!fn.HasSelection(), states.Disabled)
		}).
		OnClick(func(e events.Event) {
			fn.This().(gi.Clipper).Paste()
		})
	cb := fn.Sc.EventMgr.ClipBoard()
	if cb != nil {
		pbt.SetState(cb.IsEmpty(), states.Disabled)
	}
}

func (fn *Node) ContextMenu(m *gi.Scene) {
	// derived types put native menu code here
	if fn.CustomContextMenu != nil {
		fn.CustomContextMenu(m)
	}
	// TODO(kai/menu): need a replacement for this:
	// if tv.SyncNode != nil && CtxtMenuView(tv.SyncNode, tv.RootIsReadOnly(), tv.Scene, m) { // our viewed obj's menu
	// 	if tv.ShowViewCtxtMenu {
	// 		m.AddSeparator("sep-tvmenu")
	// 		CtxtMenuView(tv.This(), tv.RootIsReadOnly(), tv.Scene, m)
	// 	}
	// } else {
	if fn.IsReadOnly() {
		fn.TreeViewContextMenuReadOnly(m)
	} else {
		fn.FileNodeContextMenu(m)
	}
}
