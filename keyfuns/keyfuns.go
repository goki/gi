// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package keyfuns

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"goki.dev/goosi"
	"goki.dev/goosi/events/key"
)

// https://en.wikipedia.org/wiki/Table_of_keyboard_shortcuts
// https://www.cs.colorado.edu/~main/cs1300/lab/emacs.html
// https://help.ubuntu.com/community/KeyboardShortcuts

// KeyFuns are functions that keyboard events can perform in the GUI.
// It seems possible to keep this flat and consistent across different contexts,
// as long as the functions can be appropriately reinterpreted for each context.
type KeyFuns int32 //enums:enum -trim-prefix KeyFun

const (
	Nil KeyFuns = iota
	MoveUp
	MoveDown
	MoveRight
	MoveLeft
	PageUp
	PageDown
	// PageRight
	// PageLeft
	Home          // start-of-line
	KeyFunEnd     // end-of-line
	KeyFunDocHome // start-of-doc -- Control / Alt / Shift +Home
	KeyFunDocEnd  // end-of-doc Control / Alt / Shift +End
	KeyFunWordRight
	KeyFunWordLeft
	KeyFunFocusNext // Tab
	KeyFunFocusPrev // Shift-Tab
	KeyFunEnter     // Enter / return key -- has various special functions
	KeyFunAccept    // Ctrl+Enter = accept any changes and close dialog / move to next
	KeyFunCancelSelect
	KeyFunSelectMode
	KeyFunSelectAll
	KeyFunAbort
	// KeyFunEditItem
	KeyFunCopy
	KeyFunCut
	KeyFunPaste
	KeyFunPasteHist // from history
	KeyFunBackspace
	KeyFunBackspaceWord
	KeyFunDelete
	KeyFunDeleteWord
	KeyFunKill
	KeyFunDuplicate
	KeyFunTranspose
	KeyFunTransposeWord
	KeyFunUndo
	KeyFunRedo
	KeyFunInsert
	KeyFunInsertAfter
	KeyFunZoomOut
	KeyFunZoomIn
	KeyFunPrefs
	KeyFunRefresh
	KeyFunRecenter // Ctrl+L in emacs
	KeyFunComplete
	KeyFunLookup
	KeyFunSearch // Ctrl+S in emacs -- more interactive type of search
	KeyFunFind   // Command+F full-dialog find
	KeyFunReplace
	KeyFunJump // jump to line
	KeyFunHistPrev
	KeyFunHistNext
	KeyFunMenu // put focus on menu
	KeyFunWinFocusNext
	KeyFunWinClose
	KeyFunWinSnapshot
	KeyFunGoGiEditor
	// Below are menu specific functions -- use these as shortcuts for menu buttons
	// allows uniqueness of mapping and easy customization of all key buttons
	KeyFunMenuNew
	KeyFunMenuNewAlt1 // alternative version (e.g., shift)
	KeyFunMenuNewAlt2 // alternative version (e.g., alt)
	KeyFunMenuOpen
	KeyFunMenuOpenAlt1 // alternative version (e.g., shift)
	KeyFunMenuOpenAlt2 // alternative version (e.g., alt)
	KeyFunMenuSave
	KeyFunMenuSaveAs
	KeyFunMenuSaveAlt   // another alt (e.g., alt)
	KeyFunMenuCloseAlt1 // alternative version (e.g., shift)
	KeyFunMenuCloseAlt2 // alternative version (e.g., alt)
)

// KeyMap is a map between a key sequence (chord) and a specific KeyFun
// function.  This mapping must be unique, in that each chord has unique
// KeyFun, but multiple chords can trigger the same function.
type KeyMap map[key.Chord]KeyFuns

// ActiveKeyMap points to the active map -- users can set this to an
// alternative map in Prefs
var ActiveKeyMap *KeyMap

// KeyMapName has an associated Value for selecting from the list of
// available key map names, for use in preferences etc.
type KeyMapName string

func (kn KeyMapName) String() string {
	return string(kn)
}

// ActiveKeyMapName is the name of the active keymap
var ActiveKeyMapName KeyMapName

// SetActiveKeyMap sets the current ActiveKeyMap, calling Update on the map
// prior to setting it to ensure that it is a valid, complete map
func SetActiveKeyMap(km *KeyMap, kmName KeyMapName) {
	km.Update(kmName)
	ActiveKeyMap = km
	ActiveKeyMapName = kmName
}

// SetActiveKeyMapName sets the current ActiveKeyMap by name from those
// defined in AvailKeyMaps, calling Update on the map prior to setting it to
// ensure that it is a valid, complete map
func SetActiveKeyMapName(mapnm KeyMapName) {
	km, _, ok := AvailKeyMaps.MapByName(mapnm)
	if ok {
		SetActiveKeyMap(km, mapnm)
	} else {
		slog.Error("gi.SetActiveKeyMapName: key map named not found, using default", "requested", mapnm, "default", DefaultKeyMap)
		km, _, ok = AvailKeyMaps.MapByName(DefaultKeyMap)
		if ok {
			SetActiveKeyMap(km, DefaultKeyMap)
		} else {
			avail := make([]string, len(AvailKeyMaps))
			for i, km := range AvailKeyMaps {
				avail[i] = km.Name
			}
			slog.Error("gi.SetActiveKeyMapName: DefaultKeyMap not found either; trying first one", "default", DefaultKeyMap, "available", avail)
			if len(AvailKeyMaps) > 0 {
				nkm := AvailKeyMaps[0]
				SetActiveKeyMap(&nkm.Map, KeyMapName(nkm.Name))
			}
		}
	}
}

// KeyFun translates chord into keyboard function -- use oswin key.Chord
// to get chord
func KeyFun(chord key.Chord) KeyFuns {
	kf := Nil
	if chord != "" {
		kf = (*ActiveKeyMap)[chord]
		if KeyEventTrace {
			fmt.Printf("gi.KeyFun chord: %v = %v\n", chord, kf)
		}
	}
	return kf
}

// KeyMapItem records one element of the key map -- used for organizing the map.
type KeyMapItem struct {

	// the key chord that activates a function
	Key key.Chord

	// the function of that key
	Fun KeyFuns
}

// ToSlice copies this keymap to a slice of KeyMapItem's
func (km *KeyMap) ToSlice() []KeyMapItem {
	kms := make([]KeyMapItem, len(*km))
	idx := 0
	for key, fun := range *km {
		kms[idx] = KeyMapItem{key, fun}
		idx++
	}
	return kms
}

// ChordForFun returns first key chord trigger for given KeyFun in map
func (km *KeyMap) ChordForFun(kf KeyFuns) key.Chord {
	for key, fun := range *km {
		if fun == kf {
			return key
		}
	}
	return ""
}

// ShortcutForFun returns OS-specific formatted shortcut for first key chord
// trigger for given KeyFun in map
func (km *KeyMap) ShortcutForFun(kf KeyFuns) key.Chord {
	return km.ChordForFun(kf).OSShortcut()
}

// ShortcutForFun returns OS-specific formatted shortcut for first key chord
// trigger for given KeyFun in the current active map
func ShortcutForFun(kf KeyFuns) key.Chord {
	return ActiveKeyMap.ShortcutForFun(kf)
}

// Update ensures that the given keymap has at least one entry for every
// defined KeyFun, grabbing ones from the default map if not, and also
// eliminates any Nil entries which might reflect out-of-date functions
func (km *KeyMap) Update(kmName KeyMapName) {
	for key, val := range *km {
		if val == Nil {
			slog.Error("gi.KeyMap: key function is nil; probably renamed", "key", key)
			delete(*km, key)
		}
	}
	kms := km.ToSlice()
	addkm := make([]KeyMapItem, 0)

	sort.Slice(kms, func(i, j int) bool {
		return kms[i].Fun < kms[j].Fun
	})

	lfun := Nil
	for _, ki := range kms {
		fun := ki.Fun
		if fun != lfun {
			del := fun - lfun
			if del > 1 {
				for mi := lfun + 1; mi < fun; mi++ {
					slog.Error("gi.KeyMap: key map is missing a key for a key function", "keyMap", kmName, "function", mi)
					s := mi.String()
					s = strings.TrimPrefix(s, "KeyFun")
					s = "- Not Set - " + s
					nski := KeyMapItem{Key: key.Chord(s), Fun: mi}
					addkm = append(addkm, nski)
				}
			}
			lfun = fun
		}
	}

	for _, ai := range addkm {
		(*km)[ai.Key] = ai.Fun
	}
}

/////////////////////////////////////////////////////////////////////////////////
// Shortcuts

// Shortcuts is a map between a key chord and a specific Button that can be
// triggered.  This mapping must be unique, in that each chord has unique
// Button, and generally each Button only has a single chord as well, though
// this is not strictly enforced.  Shortcuts are evaluated *after* the
// standard KeyMap event processing, so any conflicts are resolved in favor of
// the local widget's key event processing, with the shortcut only operating
// when no conflicting widgets are in focus.  Shortcuts are always window-wide
// and are intended for global window / toolbar buttons.  Widget-specific key
// functions should be handled directly within widget key event
// processing.
type Shortcuts map[key.Chord]*Button

/////////////////////////////////////////////////////////////////////////////////
// KeyMaps -- list of KeyMap's

// DefaultKeyMap is the overall default keymap -- reinitialized in gimain init()
// depending on platform
var DefaultKeyMap = KeyMapName("MacEmacs")

// KeyMapsItem is an entry in a KeyMaps list
type KeyMapsItem struct { //gti:add -setters

	// name of keymap
	Name string `width:"20"`

	// description of keymap -- good idea to include source it was derived from
	Desc string

	// to edit key sequence click button and type new key combination; to edit function mapped to key sequence choose from menu
	Map KeyMap
}

// Label satisfies the Labeler interface
func (km KeyMapsItem) Label() string {
	return km.Name
}

// KeyMaps is a list of KeyMap's -- users can edit these in Prefs -- to create
// a custom one, just duplicate an existing map, rename, and customize
type KeyMaps []KeyMapsItem //gti:add

// AvailKeyMaps is the current list of available keymaps for use -- can be
// loaded / saved / edited with preferences.  This is set to StdKeyMaps at
// startup.
var AvailKeyMaps KeyMaps

func init() {
	AvailKeyMaps.CopyFrom(StdKeyMaps)
}

// MapByName returns a keymap and index by name -- returns false and emits a
// message to stdout if not found
func (km *KeyMaps) MapByName(name KeyMapName) (*KeyMap, int, bool) {
	for i, it := range *km {
		if it.Name == string(name) {
			return &it.Map, i, true
		}
	}
	slog.Error("gi.KeyMaps.MapByName: key map not found", "name", name)
	return nil, -1, false
}

// PrefsKeyMapsFileName is the name of the preferences file in GoGi prefs
// directory for saving / loading the default AvailKeyMaps key maps list
var PrefsKeyMapsFileName = "key_maps_prefs.json"

// OpenJSON opens keymaps from a JSON-formatted file.
// You can save and open key maps to / from files to share, experiment, transfer, etc
func (km *KeyMaps) OpenJSON(filename FileName) error { //gti:add
	b, err := os.ReadFile(string(filename))
	if err != nil {
		// Note: keymaps are opened at startup, and this can cause crash if called then
		// PromptDialog(nil, DlgOpts{Title: "File Not Found", Prompt: err.Error()}, true, false, nil, nil)
		log.Println(err)
		return err
	}
	*km = make(KeyMaps, 0, 10) // reset
	return json.Unmarshal(b, km)
}

// SaveJSON saves keymaps to a JSON-formatted file.
// You can save and open key maps to / from files to share, experiment, transfer, etc
func (km *KeyMaps) SaveJSON(filename FileName) error { //gti:add
	b, err := json.MarshalIndent(km, "", "  ")
	if err != nil {
		log.Println(err) // unlikely
		return err
	}
	err = os.WriteFile(string(filename), b, 0644)
	if err != nil {
		// PromptDialog(nil, DlgOpts{Title: "Could not Save to File", Prompt: err.Error()}, true, false, nil, nil)
		log.Println(err)
	}
	return err
}

// OpenPrefs opens KeyMaps from GoGi standard prefs directory, in file key_maps_prefs.json.
// This is called automatically, so calling it manually should not be necessary in most cases.
func (km *KeyMaps) OpenPrefs() error { //gti:add
	pdir := goosi.TheApp.GoGiPrefsDir()
	pnm := filepath.Join(pdir, PrefsKeyMapsFileName)
	AvailKeyMapsChanged = false
	return km.OpenJSON(FileName(pnm))
}

// SavePrefs saves KeyMaps to GoGi standard prefs directory, in file key_maps_prefs.json,
// which will be loaded automatically at startup if prefs SaveKeyMaps is checked
// (should be if you're using custom keymaps)
func (km *KeyMaps) SavePrefs() error { //gti:add
	pdir := goosi.TheApp.GoGiPrefsDir()
	pnm := filepath.Join(pdir, PrefsKeyMapsFileName)
	AvailKeyMapsChanged = false
	return km.SaveJSON(FileName(pnm))
}

// CopyFrom copies keymaps from given other map
func (km *KeyMaps) CopyFrom(cp KeyMaps) {
	*km = make(KeyMaps, 0, len(cp)) // reset
	b, _ := json.Marshal(cp)
	json.Unmarshal(b, km)
}

// RevertToStd reverts the keymaps to using the StdKeyMaps that are compiled into the program
// and have all the lastest key functions defined.  If you have edited your maps, and are finding
// things not working, it is a good idea to save your current maps and try this, or at least do
// ViewStdMaps to see the current standards. Your current map edits will be lost if you proceed!
func (km *KeyMaps) RevertToStd() { //gti:add
	km.CopyFrom(StdKeyMaps)
	AvailKeyMapsChanged = true
}

// ViewStd shows the standard maps that are compiled into the program and have
// all the lastest key functions bound to standard values.  Useful for
// comparing against custom maps.
func (km *KeyMaps) ViewStd() { //gti:add
	TheViewIFace.KeyMapsView(&StdKeyMaps)
}

// AvailKeyMapsChanged is used to update giv.KeyMapsView toolbars via
// following menu, toolbar props update methods -- not accurate if editing any
// other map but works for now..
var AvailKeyMapsChanged = false

// order is: Shift, Control, Alt, Meta
// note: shift and meta modifiers for navigation keys do select + move

// note: where multiple shortcuts exist for a given function, any shortcut
// display of such items in menus will randomly display one of the
// options. This can be considered a feature, not a bug!

// StdKeyMaps is the original compiled-in set of standard keymaps that have
// the lastest key functions bound to standard key chords.
var StdKeyMaps = KeyMaps{
	{"MacStd", "Standard Mac KeyMap", KeyMap{
		"UpArrow":                 MoveUp,
		"Shift+UpArrow":           MoveUp,
		"Meta+UpArrow":            MoveUp,
		"Control+P":               MoveUp,
		"Shift+Control+P":         MoveUp,
		"Meta+Control+P":          MoveUp,
		"DownArrow":               MoveDown,
		"Shift+DownArrow":         MoveDown,
		"Meta+DownArrow":          MoveDown,
		"Control+N":               MoveDown,
		"Shift+Control+N":         MoveDown,
		"Meta+Control+N":          MoveDown,
		"RightArrow":              MoveRight,
		"Shift+RightArrow":        MoveRight,
		"Meta+RightArrow":         KeyFunEnd,
		"Control+F":               MoveRight,
		"Shift+Control+F":         MoveRight,
		"Meta+Control+F":          MoveRight,
		"LeftArrow":               MoveLeft,
		"Shift+LeftArrow":         MoveLeft,
		"Meta+LeftArrow":          Home,
		"Control+B":               MoveLeft,
		"Shift+Control+B":         MoveLeft,
		"Meta+Control+B":          MoveLeft,
		"PageUp":                  PageUp,
		"Shift+PageUp":            PageUp,
		"Control+UpArrow":         PageUp,
		"Control+U":               PageUp,
		"PageDown":                PageDown,
		"Shift+PageDown":          PageDown,
		"Control+DownArrow":       PageDown,
		"Shift+Control+V":         PageDown,
		"Alt+√":                   PageDown,
		"Meta+Home":               KeyFunDocHome,
		"Shift+Home":              KeyFunDocHome,
		"Meta+H":                  KeyFunDocHome,
		"Meta+End":                KeyFunDocEnd,
		"Shift+End":               KeyFunDocEnd,
		"Meta+L":                  KeyFunDocEnd,
		"Control+RightArrow":      KeyFunWordRight,
		"Control+LeftArrow":       KeyFunWordLeft,
		"Alt+RightArrow":          KeyFunWordRight,
		"Shift+Alt+RightArrow":    KeyFunWordRight,
		"Alt+LeftArrow":           KeyFunWordLeft,
		"Shift+Alt+LeftArrow":     KeyFunWordLeft,
		"Home":                    Home,
		"Control+A":               Home,
		"Shift+Control+A":         Home,
		"End":                     KeyFunEnd,
		"Control+E":               KeyFunEnd,
		"Shift+Control+E":         KeyFunEnd,
		"Tab":                     KeyFunFocusNext,
		"Shift+Tab":               KeyFunFocusPrev,
		"ReturnEnter":             KeyFunEnter,
		"KeypadEnter":             KeyFunEnter,
		"Meta+A":                  KeyFunSelectAll,
		"Control+G":               KeyFunCancelSelect,
		"Control+Spacebar":        KeyFunSelectMode,
		"Control+ReturnEnter":     KeyFunAccept,
		"Escape":                  KeyFunAbort,
		"DeleteBackspace":         KeyFunBackspace,
		"Control+DeleteBackspace": KeyFunBackspaceWord,
		"Alt+DeleteBackspace":     KeyFunBackspaceWord,
		"DeleteForward":           KeyFunDelete,
		"Control+DeleteForward":   KeyFunDeleteWord,
		"Alt+DeleteForward":       KeyFunDeleteWord,
		"Control+D":               KeyFunDelete,
		"Control+K":               KeyFunKill,
		"Alt+∑":                   KeyFunCopy,
		"Meta+C":                  KeyFunCopy,
		"Control+W":               KeyFunCut,
		"Meta+X":                  KeyFunCut,
		"Control+Y":               KeyFunPaste,
		"Control+V":               KeyFunPaste,
		"Meta+V":                  KeyFunPaste,
		"Shift+Meta+V":            KeyFunPasteHist,
		"Alt+D":                   KeyFunDuplicate,
		"Control+T":               KeyFunTranspose,
		"Alt+T":                   KeyFunTransposeWord,
		"Control+Z":               KeyFunUndo,
		"Meta+Z":                  KeyFunUndo,
		"Shift+Control+Z":         KeyFunRedo,
		"Shift+Meta+Z":            KeyFunRedo,
		"Control+I":               KeyFunInsert,
		"Control+O":               KeyFunInsertAfter,
		"Shift+Meta+=":            KeyFunZoomIn,
		"Meta+=":                  KeyFunZoomIn,
		"Meta+-":                  KeyFunZoomOut,
		"Control+=":               KeyFunZoomIn,
		"Shift+Control++":         KeyFunZoomIn,
		"Shift+Meta+-":            KeyFunZoomOut,
		"Control+-":               KeyFunZoomOut,
		"Shift+Control+_":         KeyFunZoomOut,
		"Control+Alt+P":           KeyFunPrefs,
		"F5":                      KeyFunRefresh,
		"Control+L":               KeyFunRecenter,
		"Control+.":               KeyFunComplete,
		"Control+,":               KeyFunLookup,
		"Control+S":               KeyFunSearch,
		"Meta+F":                  KeyFunFind,
		"Meta+R":                  KeyFunReplace,
		"Control+J":               KeyFunJump,
		"Control+[":               KeyFunHistPrev,
		"Control+]":               KeyFunHistNext,
		"Meta+[":                  KeyFunHistPrev,
		"Meta+]":                  KeyFunHistNext,
		"F10":                     KeyFunMenu,
		"Meta+`":                  KeyFunWinFocusNext,
		"Meta+W":                  KeyFunWinClose,
		"Control+Alt+G":           KeyFunWinSnapshot,
		"Shift+Control+G":         KeyFunWinSnapshot,
		"Control+Alt+I":           KeyFunGoGiEditor,
		"Shift+Control+I":         KeyFunGoGiEditor,
		"Meta+N":                  KeyFunMenuNew,
		"Shift+Meta+N":            KeyFunMenuNewAlt1,
		"Alt+Meta+N":              KeyFunMenuNewAlt2,
		"Meta+O":                  KeyFunMenuOpen,
		"Shift+Meta+O":            KeyFunMenuOpenAlt1,
		"Alt+Meta+O":              KeyFunMenuOpenAlt2,
		"Meta+S":                  KeyFunMenuSave,
		"Shift+Meta+S":            KeyFunMenuSaveAs,
		"Alt+Meta+S":              KeyFunMenuSaveAlt,
		"Shift+Meta+W":            KeyFunMenuCloseAlt1,
		"Alt+Meta+W":              KeyFunMenuCloseAlt2,
	}},
	{"MacEmacs", "Mac with emacs-style navigation -- emacs wins in conflicts", KeyMap{
		"UpArrow":                 MoveUp,
		"Shift+UpArrow":           MoveUp,
		"Meta+UpArrow":            MoveUp,
		"Control+P":               MoveUp,
		"Shift+Control+P":         MoveUp,
		"Meta+Control+P":          MoveUp,
		"DownArrow":               MoveDown,
		"Shift+DownArrow":         MoveDown,
		"Meta+DownArrow":          MoveDown,
		"Control+N":               MoveDown,
		"Shift+Control+N":         MoveDown,
		"Meta+Control+N":          MoveDown,
		"RightArrow":              MoveRight,
		"Shift+RightArrow":        MoveRight,
		"Meta+RightArrow":         KeyFunEnd,
		"Control+F":               MoveRight,
		"Shift+Control+F":         MoveRight,
		"Meta+Control+F":          MoveRight,
		"LeftArrow":               MoveLeft,
		"Shift+LeftArrow":         MoveLeft,
		"Meta+LeftArrow":          Home,
		"Control+B":               MoveLeft,
		"Shift+Control+B":         MoveLeft,
		"Meta+Control+B":          MoveLeft,
		"PageUp":                  PageUp,
		"Shift+PageUp":            PageUp,
		"Control+UpArrow":         PageUp,
		"Control+U":               PageUp,
		"PageDown":                PageDown,
		"Shift+PageDown":          PageDown,
		"Control+DownArrow":       PageDown,
		"Shift+Control+V":         PageDown,
		"Alt+√":                   PageDown,
		"Control+V":               PageDown,
		"Control+RightArrow":      KeyFunWordRight,
		"Control+LeftArrow":       KeyFunWordLeft,
		"Alt+RightArrow":          KeyFunWordRight,
		"Shift+Alt+RightArrow":    KeyFunWordRight,
		"Alt+LeftArrow":           KeyFunWordLeft,
		"Shift+Alt+LeftArrow":     KeyFunWordLeft,
		"Home":                    Home,
		"Control+A":               Home,
		"Shift+Control+A":         Home,
		"End":                     KeyFunEnd,
		"Control+E":               KeyFunEnd,
		"Shift+Control+E":         KeyFunEnd,
		"Meta+Home":               KeyFunDocHome,
		"Shift+Home":              KeyFunDocHome,
		"Meta+H":                  KeyFunDocHome,
		"Control+H":               KeyFunDocHome,
		"Control+Alt+A":           KeyFunDocHome,
		"Meta+End":                KeyFunDocEnd,
		"Shift+End":               KeyFunDocEnd,
		"Meta+L":                  KeyFunDocEnd,
		"Control+Alt+E":           KeyFunDocEnd,
		"Alt+Ƒ":                   KeyFunWordRight,
		"Alt+∫":                   KeyFunWordLeft,
		"Tab":                     KeyFunFocusNext,
		"Shift+Tab":               KeyFunFocusPrev,
		"ReturnEnter":             KeyFunEnter,
		"KeypadEnter":             KeyFunEnter,
		"Meta+A":                  KeyFunSelectAll,
		"Control+G":               KeyFunCancelSelect,
		"Control+Spacebar":        KeyFunSelectMode,
		"Control+ReturnEnter":     KeyFunAccept,
		"Escape":                  KeyFunAbort,
		"DeleteBackspace":         KeyFunBackspace,
		"Control+DeleteBackspace": KeyFunBackspaceWord,
		"Alt+DeleteBackspace":     KeyFunBackspaceWord,
		"DeleteForward":           KeyFunDelete,
		"Control+DeleteForward":   KeyFunDeleteWord,
		"Alt+DeleteForward":       KeyFunDeleteWord,
		"Control+D":               KeyFunDelete,
		"Control+K":               KeyFunKill,
		"Alt+∑":                   KeyFunCopy,
		"Meta+C":                  KeyFunCopy,
		"Control+W":               KeyFunCut,
		"Meta+X":                  KeyFunCut,
		"Control+Y":               KeyFunPaste,
		"Meta+V":                  KeyFunPaste,
		"Shift+Meta+V":            KeyFunPasteHist,
		"Shift+Control+Y":         KeyFunPasteHist,
		"Alt+∂":                   KeyFunDuplicate,
		"Control+T":               KeyFunTranspose,
		"Alt+T":                   KeyFunTransposeWord,
		"Control+Z":               KeyFunUndo,
		"Meta+Z":                  KeyFunUndo,
		"Control+/":               KeyFunUndo,
		"Shift+Control+Z":         KeyFunRedo,
		"Shift+Meta+Z":            KeyFunRedo,
		"Control+I":               KeyFunInsert,
		"Control+O":               KeyFunInsertAfter,
		"Shift+Meta+=":            KeyFunZoomIn,
		"Meta+=":                  KeyFunZoomIn,
		"Meta+-":                  KeyFunZoomOut,
		"Control+=":               KeyFunZoomIn,
		"Shift+Control++":         KeyFunZoomIn,
		"Shift+Meta+-":            KeyFunZoomOut,
		"Control+-":               KeyFunZoomOut,
		"Shift+Control+_":         KeyFunZoomOut,
		"Control+Alt+P":           KeyFunPrefs,
		"F5":                      KeyFunRefresh,
		"Control+L":               KeyFunRecenter,
		"Control+.":               KeyFunComplete,
		"Control+,":               KeyFunLookup,
		"Control+S":               KeyFunSearch,
		"Meta+F":                  KeyFunFind,
		"Meta+R":                  KeyFunReplace,
		"Control+R":               KeyFunReplace,
		"Control+J":               KeyFunJump,
		"Control+[":               KeyFunHistPrev,
		"Control+]":               KeyFunHistNext,
		"Meta+[":                  KeyFunHistPrev,
		"Meta+]":                  KeyFunHistNext,
		"F10":                     KeyFunMenu,
		"Meta+`":                  KeyFunWinFocusNext,
		"Meta+W":                  KeyFunWinClose,
		"Control+Alt+G":           KeyFunWinSnapshot,
		"Shift+Control+G":         KeyFunWinSnapshot,
		"Control+Alt+I":           KeyFunGoGiEditor,
		"Shift+Control+I":         KeyFunGoGiEditor,
		"Meta+N":                  KeyFunMenuNew,
		"Shift+Meta+N":            KeyFunMenuNewAlt1,
		"Alt+Meta+N":              KeyFunMenuNewAlt2,
		"Meta+O":                  KeyFunMenuOpen,
		"Shift+Meta+O":            KeyFunMenuOpenAlt1,
		"Alt+Meta+O":              KeyFunMenuOpenAlt2,
		"Meta+S":                  KeyFunMenuSave,
		"Shift+Meta+S":            KeyFunMenuSaveAs,
		"Alt+Meta+S":              KeyFunMenuSaveAlt,
		"Shift+Meta+W":            KeyFunMenuCloseAlt1,
		"Alt+Meta+W":              KeyFunMenuCloseAlt2,
	}},
	{"LinuxEmacs", "Linux with emacs-style navigation -- emacs wins in conflicts", KeyMap{
		"UpArrow":                 MoveUp,
		"Shift+UpArrow":           MoveUp,
		"Alt+UpArrow":             MoveUp,
		"Control+P":               MoveUp,
		"Shift+Control+P":         MoveUp,
		"Alt+Control+P":           MoveUp,
		"DownArrow":               MoveDown,
		"Shift+DownArrow":         MoveDown,
		"Alt+DownArrow":           MoveDown,
		"Control+N":               MoveDown,
		"Shift+Control+N":         MoveDown,
		"Alt+Control+N":           MoveDown,
		"RightArrow":              MoveRight,
		"Shift+RightArrow":        MoveRight,
		"Alt+RightArrow":          KeyFunEnd,
		"Control+F":               MoveRight,
		"Shift+Control+F":         MoveRight,
		"Alt+Control+F":           MoveRight,
		"LeftArrow":               MoveLeft,
		"Shift+LeftArrow":         MoveLeft,
		"Alt+LeftArrow":           Home,
		"Control+B":               MoveLeft,
		"Shift+Control+B":         MoveLeft,
		"Alt+Control+B":           MoveLeft,
		"PageUp":                  PageUp,
		"Shift+PageUp":            PageUp,
		"Control+UpArrow":         PageUp,
		"Control+U":               PageUp,
		"Shift+Control+U":         PageUp,
		"Alt+Control+U":           PageUp,
		"PageDown":                PageDown,
		"Shift+PageDown":          PageDown,
		"Control+DownArrow":       PageDown,
		"Control+V":               PageDown,
		"Shift+Control+V":         PageDown,
		"Alt+Control+V":           PageDown,
		"Alt+Home":                KeyFunDocHome,
		"Shift+Home":              KeyFunDocHome,
		"Alt+H":                   KeyFunDocHome,
		"Control+Alt+A":           KeyFunDocHome,
		"Alt+End":                 KeyFunDocEnd,
		"Shift+End":               KeyFunDocEnd,
		"Alt+L":                   KeyFunDocEnd,
		"Control+Alt+E":           KeyFunDocEnd,
		"Control+RightArrow":      KeyFunWordRight,
		"Control+LeftArrow":       KeyFunWordLeft,
		"Home":                    Home,
		"Control+A":               Home,
		"Shift+Control+A":         Home,
		"End":                     KeyFunEnd,
		"Control+E":               KeyFunEnd,
		"Shift+Control+E":         KeyFunEnd,
		"Tab":                     KeyFunFocusNext,
		"Shift+Tab":               KeyFunFocusPrev,
		"ReturnEnter":             KeyFunEnter,
		"KeypadEnter":             KeyFunEnter,
		"Alt+A":                   KeyFunSelectAll,
		"Control+G":               KeyFunCancelSelect,
		"Control+Spacebar":        KeyFunSelectMode,
		"Control+ReturnEnter":     KeyFunAccept,
		"Escape":                  KeyFunAbort,
		"DeleteBackspace":         KeyFunBackspace,
		"Control+DeleteBackspace": KeyFunBackspaceWord,
		"DeleteForward":           KeyFunDelete,
		"Control+D":               KeyFunDelete,
		"Control+DeleteForward":   KeyFunDeleteWord,
		"Alt+DeleteForward":       KeyFunDeleteWord,
		"Control+K":               KeyFunKill,
		"Alt+W":                   KeyFunCopy,
		"Alt+C":                   KeyFunCopy,
		"Control+W":               KeyFunCut,
		"Alt+X":                   KeyFunCut,
		"Control+Y":               KeyFunPaste,
		"Alt+V":                   KeyFunPaste,
		"Shift+Alt+V":             KeyFunPasteHist,
		"Shift+Control+Y":         KeyFunPasteHist,
		"Alt+D":                   KeyFunDuplicate,
		"Control+T":               KeyFunTranspose,
		"Alt+T":                   KeyFunTransposeWord,
		"Control+Z":               KeyFunUndo,
		"Control+/":               KeyFunUndo,
		"Shift+Control+Z":         KeyFunRedo,
		"Control+I":               KeyFunInsert,
		"Control+O":               KeyFunInsertAfter,
		"Control+=":               KeyFunZoomIn,
		"Shift+Control++":         KeyFunZoomIn,
		"Control+-":               KeyFunZoomOut,
		"Shift+Control+_":         KeyFunZoomOut,
		"Control+Alt+P":           KeyFunPrefs,
		"F5":                      KeyFunRefresh,
		"Control+L":               KeyFunRecenter,
		"Control+.":               KeyFunComplete,
		"Control+,":               KeyFunLookup,
		"Control+S":               KeyFunSearch,
		"Alt+F":                   KeyFunFind,
		"Control+R":               KeyFunReplace,
		"Control+J":               KeyFunJump,
		"Control+[":               KeyFunHistPrev,
		"Control+]":               KeyFunHistNext,
		"F10":                     KeyFunMenu,
		"Alt+F6":                  KeyFunWinFocusNext,
		"Shift+Control+W":         KeyFunWinClose,
		"Control+Alt+G":           KeyFunWinSnapshot,
		"Shift+Control+G":         KeyFunWinSnapshot,
		"Control+Alt+I":           KeyFunGoGiEditor,
		"Shift+Control+I":         KeyFunGoGiEditor,
		"Alt+N":                   KeyFunMenuNew, // ctrl keys conflict..
		"Shift+Alt+N":             KeyFunMenuNewAlt1,
		"Control+Alt+N":           KeyFunMenuNewAlt2,
		"Alt+O":                   KeyFunMenuOpen,
		"Shift+Alt+O":             KeyFunMenuOpenAlt1,
		"Control+Alt+O":           KeyFunMenuOpenAlt2,
		"Alt+S":                   KeyFunMenuSave,
		"Shift+Alt+S":             KeyFunMenuSaveAs,
		"Control+Alt+S":           KeyFunMenuSaveAlt,
		"Shift+Alt+W":             KeyFunMenuCloseAlt1,
		"Control+Alt+W":           KeyFunMenuCloseAlt2,
	}},
	{"LinuxStd", "Standard Linux KeyMap", KeyMap{
		"UpArrow":                 MoveUp,
		"Shift+UpArrow":           MoveUp,
		"DownArrow":               MoveDown,
		"Shift+DownArrow":         MoveDown,
		"RightArrow":              MoveRight,
		"Shift+RightArrow":        MoveRight,
		"LeftArrow":               MoveLeft,
		"Shift+LeftArrow":         MoveLeft,
		"PageUp":                  PageUp,
		"Shift+PageUp":            PageUp,
		"Control+UpArrow":         PageUp,
		"PageDown":                PageDown,
		"Shift+PageDown":          PageDown,
		"Control+DownArrow":       PageDown,
		"Home":                    Home,
		"Alt+LeftArrow":           Home,
		"End":                     KeyFunEnd,
		"Alt+Home":                KeyFunDocHome,
		"Shift+Home":              KeyFunDocHome,
		"Alt+End":                 KeyFunDocEnd,
		"Shift+End":               KeyFunDocEnd,
		"Control+RightArrow":      KeyFunWordRight,
		"Control+LeftArrow":       KeyFunWordLeft,
		"Alt+RightArrow":          KeyFunEnd,
		"Tab":                     KeyFunFocusNext,
		"Shift+Tab":               KeyFunFocusPrev,
		"ReturnEnter":             KeyFunEnter,
		"KeypadEnter":             KeyFunEnter,
		"Control+A":               KeyFunSelectAll,
		"Shift+Control+A":         KeyFunCancelSelect,
		"Control+G":               KeyFunCancelSelect,
		"Control+Spacebar":        KeyFunSelectMode, // change input method / keyboard
		"Control+ReturnEnter":     KeyFunAccept,
		"Escape":                  KeyFunAbort,
		"DeleteBackspace":         KeyFunBackspace,
		"Control+DeleteBackspace": KeyFunBackspaceWord,
		"DeleteForward":           KeyFunDelete,
		"Control+DeleteForward":   KeyFunDeleteWord,
		"Alt+DeleteForward":       KeyFunDeleteWord,
		"Control+K":               KeyFunKill,
		"Control+C":               KeyFunCopy,
		"Control+X":               KeyFunCut,
		"Control+V":               KeyFunPaste,
		"Shift+Control+V":         KeyFunPasteHist,
		"Alt+D":                   KeyFunDuplicate,
		"Control+T":               KeyFunTranspose,
		"Alt+T":                   KeyFunTransposeWord,
		"Control+Z":               KeyFunUndo,
		"Control+Y":               KeyFunRedo,
		"Shift+Control+Z":         KeyFunRedo,
		"Control+Alt+I":           KeyFunInsert,
		"Control+Alt+O":           KeyFunInsertAfter,
		"Control+=":               KeyFunZoomIn,
		"Shift+Control++":         KeyFunZoomIn,
		"Control+-":               KeyFunZoomOut,
		"Shift+Control+_":         KeyFunZoomOut,
		"Shift+Control+P":         KeyFunPrefs,
		"Control+Alt+P":           KeyFunPrefs,
		"F5":                      KeyFunRefresh,
		"Control+L":               KeyFunRecenter,
		"Control+.":               KeyFunComplete,
		"Control+,":               KeyFunLookup,
		"Alt+S":                   KeyFunSearch,
		"Control+F":               KeyFunFind,
		"Control+H":               KeyFunReplace,
		"Control+R":               KeyFunReplace,
		"Control+J":               KeyFunJump,
		"Control+[":               KeyFunHistPrev,
		"Control+]":               KeyFunHistNext,
		"Control+N":               KeyFunMenuNew,
		"F10":                     KeyFunMenu,
		"Alt+F6":                  KeyFunWinFocusNext,
		"Control+W":               KeyFunWinClose,
		"Control+Alt+G":           KeyFunWinSnapshot,
		"Shift+Control+G":         KeyFunWinSnapshot,
		"Shift+Control+I":         KeyFunGoGiEditor,
		"Shift+Control+N":         KeyFunMenuNewAlt1,
		"Control+Alt+N":           KeyFunMenuNewAlt2,
		"Control+O":               KeyFunMenuOpen,
		"Shift+Control+O":         KeyFunMenuOpenAlt1,
		"Shift+Alt+O":             KeyFunMenuOpenAlt2,
		"Control+S":               KeyFunMenuSave,
		"Shift+Control+S":         KeyFunMenuSaveAs,
		"Control+Alt+S":           KeyFunMenuSaveAlt,
		"Shift+Control+W":         KeyFunMenuCloseAlt1,
		"Control+Alt+W":           KeyFunMenuCloseAlt2,
	}},
	{"WindowsStd", "Standard Windows KeyMap", KeyMap{
		"UpArrow":                 MoveUp,
		"Shift+UpArrow":           MoveUp,
		"DownArrow":               MoveDown,
		"Shift+DownArrow":         MoveDown,
		"RightArrow":              MoveRight,
		"Shift+RightArrow":        MoveRight,
		"LeftArrow":               MoveLeft,
		"Shift+LeftArrow":         MoveLeft,
		"PageUp":                  PageUp,
		"Shift+PageUp":            PageUp,
		"Control+UpArrow":         PageUp,
		"PageDown":                PageDown,
		"Shift+PageDown":          PageDown,
		"Control+DownArrow":       PageDown,
		"Home":                    Home,
		"Alt+LeftArrow":           Home,
		"End":                     KeyFunEnd,
		"Alt+RightArrow":          KeyFunEnd,
		"Alt+Home":                KeyFunDocHome,
		"Shift+Home":              KeyFunDocHome,
		"Alt+End":                 KeyFunDocEnd,
		"Shift+End":               KeyFunDocEnd,
		"Control+RightArrow":      KeyFunWordRight,
		"Control+LeftArrow":       KeyFunWordLeft,
		"Tab":                     KeyFunFocusNext,
		"Shift+Tab":               KeyFunFocusPrev,
		"ReturnEnter":             KeyFunEnter,
		"KeypadEnter":             KeyFunEnter,
		"Control+A":               KeyFunSelectAll,
		"Shift+Control+A":         KeyFunCancelSelect,
		"Control+G":               KeyFunCancelSelect,
		"Control+Spacebar":        KeyFunSelectMode, // change input method / keyboard
		"Control+ReturnEnter":     KeyFunAccept,
		"Escape":                  KeyFunAbort,
		"DeleteBackspace":         KeyFunBackspace,
		"Control+DeleteBackspace": KeyFunBackspaceWord,
		"DeleteForward":           KeyFunDelete,
		"Control+DeleteForward":   KeyFunDeleteWord,
		"Alt+DeleteForward":       KeyFunDeleteWord,
		"Control+K":               KeyFunKill,
		"Control+C":               KeyFunCopy,
		"Control+X":               KeyFunCut,
		"Control+V":               KeyFunPaste,
		"Shift+Control+V":         KeyFunPasteHist,
		"Alt+D":                   KeyFunDuplicate,
		"Control+T":               KeyFunTranspose,
		"Alt+T":                   KeyFunTransposeWord,
		"Control+Z":               KeyFunUndo,
		"Control+Y":               KeyFunRedo,
		"Shift+Control+Z":         KeyFunRedo,
		"Control+Alt+I":           KeyFunInsert,
		"Control+Alt+O":           KeyFunInsertAfter,
		"Control+=":               KeyFunZoomIn,
		"Shift+Control++":         KeyFunZoomIn,
		"Control+-":               KeyFunZoomOut,
		"Shift+Control+_":         KeyFunZoomOut,
		"Shift+Control+P":         KeyFunPrefs,
		"Control+Alt+P":           KeyFunPrefs,
		"F5":                      KeyFunRefresh,
		"Control+L":               KeyFunRecenter,
		"Control+.":               KeyFunComplete,
		"Control+,":               KeyFunLookup,
		"Alt+S":                   KeyFunSearch,
		"Control+F":               KeyFunFind,
		"Control+H":               KeyFunReplace,
		"Control+R":               KeyFunReplace,
		"Control+J":               KeyFunJump,
		"Control+[":               KeyFunHistPrev,
		"Control+]":               KeyFunHistNext,
		"F10":                     KeyFunMenu,
		"Alt+F6":                  KeyFunWinFocusNext,
		"Control+W":               KeyFunWinClose,
		"Control+Alt+G":           KeyFunWinSnapshot,
		"Shift+Control+G":         KeyFunWinSnapshot,
		"Shift+Control+I":         KeyFunGoGiEditor,
		"Control+N":               KeyFunMenuNew,
		"Shift+Control+N":         KeyFunMenuNewAlt1,
		"Control+Alt+N":           KeyFunMenuNewAlt2,
		"Control+O":               KeyFunMenuOpen,
		"Shift+Control+O":         KeyFunMenuOpenAlt1,
		"Shift+Alt+O":             KeyFunMenuOpenAlt2,
		"Control+S":               KeyFunMenuSave,
		"Shift+Control+S":         KeyFunMenuSaveAs,
		"Control+Alt+S":           KeyFunMenuSaveAlt,
		"Shift+Control+W":         KeyFunMenuCloseAlt1,
		"Control+Alt+W":           KeyFunMenuCloseAlt2,
	}},
	{"ChromeStd", "Standard chrome-browser and linux-under-chrome bindings", KeyMap{
		"UpArrow":                 MoveUp,
		"Shift+UpArrow":           MoveUp,
		"DownArrow":               MoveDown,
		"Shift+DownArrow":         MoveDown,
		"RightArrow":              MoveRight,
		"Shift+RightArrow":        MoveRight,
		"LeftArrow":               MoveLeft,
		"Shift+LeftArrow":         MoveLeft,
		"PageUp":                  PageUp,
		"Shift+PageUp":            PageUp,
		"Control+UpArrow":         PageUp,
		"PageDown":                PageDown,
		"Shift+PageDown":          PageDown,
		"Control+DownArrow":       PageDown,
		"Home":                    Home,
		"Alt+LeftArrow":           Home,
		"End":                     KeyFunEnd,
		"Alt+Home":                KeyFunDocHome,
		"Shift+Home":              KeyFunDocHome,
		"Alt+End":                 KeyFunDocEnd,
		"Shift+End":               KeyFunDocEnd,
		"Control+RightArrow":      KeyFunWordRight,
		"Control+LeftArrow":       KeyFunWordLeft,
		"Alt+RightArrow":          KeyFunEnd,
		"Tab":                     KeyFunFocusNext,
		"Shift+Tab":               KeyFunFocusPrev,
		"ReturnEnter":             KeyFunEnter,
		"KeypadEnter":             KeyFunEnter,
		"Control+A":               KeyFunSelectAll,
		"Shift+Control+A":         KeyFunCancelSelect,
		"Control+G":               KeyFunCancelSelect,
		"Control+Spacebar":        KeyFunSelectMode, // change input method / keyboard
		"Control+ReturnEnter":     KeyFunAccept,
		"Escape":                  KeyFunAbort,
		"DeleteBackspace":         KeyFunBackspace,
		"Control+DeleteBackspace": KeyFunBackspaceWord,
		"DeleteForward":           KeyFunDelete,
		"Control+DeleteForward":   KeyFunDeleteWord,
		"Alt+DeleteForward":       KeyFunDeleteWord,
		"Control+K":               KeyFunKill,
		"Control+C":               KeyFunCopy,
		"Control+X":               KeyFunCut,
		"Control+V":               KeyFunPaste,
		"Shift+Control+V":         KeyFunPasteHist,
		"Alt+D":                   KeyFunDuplicate,
		"Control+T":               KeyFunTranspose,
		"Alt+T":                   KeyFunTransposeWord,
		"Control+Z":               KeyFunUndo,
		"Control+Y":               KeyFunRedo,
		"Shift+Control+Z":         KeyFunRedo,
		"Control+Alt+I":           KeyFunInsert,
		"Control+Alt+O":           KeyFunInsertAfter,
		"Control+=":               KeyFunZoomIn,
		"Shift+Control++":         KeyFunZoomIn,
		"Control+-":               KeyFunZoomOut,
		"Shift+Control+_":         KeyFunZoomOut,
		"Shift+Control+P":         KeyFunPrefs,
		"Control+Alt+P":           KeyFunPrefs,
		"F5":                      KeyFunRefresh,
		"Control+L":               KeyFunRecenter,
		"Control+.":               KeyFunComplete,
		"Control+,":               KeyFunLookup,
		"Alt+S":                   KeyFunSearch,
		"Control+F":               KeyFunFind,
		"Control+H":               KeyFunReplace,
		"Control+R":               KeyFunReplace,
		"Control+J":               KeyFunJump,
		"Control+[":               KeyFunHistPrev,
		"Control+]":               KeyFunHistNext,
		"F10":                     KeyFunMenu,
		"Alt+F6":                  KeyFunWinFocusNext,
		"Control+W":               KeyFunWinClose,
		"Control+Alt+G":           KeyFunWinSnapshot,
		"Shift+Control+G":         KeyFunWinSnapshot,
		"Shift+Control+I":         KeyFunGoGiEditor,
		"Control+N":               KeyFunMenuNew,
		"Shift+Control+N":         KeyFunMenuNewAlt1,
		"Control+Alt+N":           KeyFunMenuNewAlt2,
		"Control+O":               KeyFunMenuOpen,
		"Shift+Control+O":         KeyFunMenuOpenAlt1,
		"Shift+Alt+O":             KeyFunMenuOpenAlt2,
		"Control+S":               KeyFunMenuSave,
		"Shift+Control+S":         KeyFunMenuSaveAs,
		"Control+Alt+S":           KeyFunMenuSaveAlt,
		"Shift+Control+W":         KeyFunMenuCloseAlt1,
		"Control+Alt+W":           KeyFunMenuCloseAlt2,
	}},
}