// Code generated by "goki generate"; DO NOT EDIT.

package texteditor

import (
	"errors"
	"strconv"
	"strings"
	"sync/atomic"

	"goki.dev/enums"
)

var _BufSignalsValues = []BufSignals{0, 1, 2, 3, 4, 5, 6}

// BufSignalsN is the highest valid value
// for type BufSignals, plus one.
const BufSignalsN BufSignals = 7

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the enumgen command to generate them again.
func _BufSignalsNoOp() {
	var x [1]struct{}
	_ = x[BufDone-(0)]
	_ = x[BufNew-(1)]
	_ = x[BufMods-(2)]
	_ = x[BufInsert-(3)]
	_ = x[BufDelete-(4)]
	_ = x[BufMarkUpdt-(5)]
	_ = x[BufClosed-(6)]
}

var _BufSignalsNameToValueMap = map[string]BufSignals{
	`BufDone`:     0,
	`bufdone`:     0,
	`BufNew`:      1,
	`bufnew`:      1,
	`BufMods`:     2,
	`bufmods`:     2,
	`BufInsert`:   3,
	`bufinsert`:   3,
	`BufDelete`:   4,
	`bufdelete`:   4,
	`BufMarkUpdt`: 5,
	`bufmarkupdt`: 5,
	`BufClosed`:   6,
	`bufclosed`:   6,
}

var _BufSignalsDescMap = map[BufSignals]string{
	0: `BufDone means that editing was completed and applied to Txt field -- data is Txt bytes`,
	1: `BufNew signals that entirely new text is present. All views should do full layout update.`,
	2: `BufMods signals that potentially diffuse modifications have been made. Views should do a Layout and Render.`,
	3: `BufInsert signals that some text was inserted. data is textbuf.Edit describing change. The Buf always reflects the current state *after* the edit.`,
	4: `BufDelete signals that some text was deleted. data is textbuf.Edit describing change. The Buf always reflects the current state *after* the edit.`,
	5: `BufMarkUpdt signals that the Markup text has been updated This signal is typically sent from a separate goroutine, so should be used with a mutex`,
	6: `BufClosed signals that the textbuf was closed.`,
}

var _BufSignalsMap = map[BufSignals]string{
	0: `BufDone`,
	1: `BufNew`,
	2: `BufMods`,
	3: `BufInsert`,
	4: `BufDelete`,
	5: `BufMarkUpdt`,
	6: `BufClosed`,
}

// String returns the string representation
// of this BufSignals value.
func (i BufSignals) String() string {
	if str, ok := _BufSignalsMap[i]; ok {
		return str
	}
	return strconv.FormatInt(int64(i), 10)
}

// SetString sets the BufSignals value from its
// string representation, and returns an
// error if the string is invalid.
func (i *BufSignals) SetString(s string) error {
	if val, ok := _BufSignalsNameToValueMap[s]; ok {
		*i = val
		return nil
	}
	if val, ok := _BufSignalsNameToValueMap[strings.ToLower(s)]; ok {
		*i = val
		return nil
	}
	return errors.New(s + " is not a valid value for type BufSignals")
}

// Int64 returns the BufSignals value as an int64.
func (i BufSignals) Int64() int64 {
	return int64(i)
}

// SetInt64 sets the BufSignals value from an int64.
func (i *BufSignals) SetInt64(in int64) {
	*i = BufSignals(in)
}

// Desc returns the description of the BufSignals value.
func (i BufSignals) Desc() string {
	if str, ok := _BufSignalsDescMap[i]; ok {
		return str
	}
	return i.String()
}

// BufSignalsValues returns all possible values
// for the type BufSignals.
func BufSignalsValues() []BufSignals {
	return _BufSignalsValues
}

// Values returns all possible values
// for the type BufSignals.
func (i BufSignals) Values() []enums.Enum {
	res := make([]enums.Enum, len(_BufSignalsValues))
	for i, d := range _BufSignalsValues {
		res[i] = d
	}
	return res
}

// IsValid returns whether the value is a
// valid option for type BufSignals.
func (i BufSignals) IsValid() bool {
	_, ok := _BufSignalsMap[i]
	return ok
}

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i BufSignals) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *BufSignals) UnmarshalText(text []byte) error {
	return i.SetString(string(text))
}

var _BufFlagsValues = []BufFlags{10, 11, 12, 13}

// BufFlagsN is the highest valid value
// for type BufFlags, plus one.
const BufFlagsN BufFlags = 14

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the enumgen command to generate them again.
func _BufFlagsNoOp() {
	var x [1]struct{}
	_ = x[BufAutoSaving-(10)]
	_ = x[BufMarkingUp-(11)]
	_ = x[BufChanged-(12)]
	_ = x[BufFileModOk-(13)]
}

var _BufFlagsNameToValueMap = map[string]BufFlags{
	`BufAutoSaving`: 10,
	`bufautosaving`: 10,
	`BufMarkingUp`:  11,
	`bufmarkingup`:  11,
	`BufChanged`:    12,
	`bufchanged`:    12,
	`BufFileModOk`:  13,
	`buffilemodok`:  13,
}

var _BufFlagsDescMap = map[BufFlags]string{
	10: `BufAutoSaving is used in atomically safe way to protect autosaving`,
	11: `BufMarkingUp indicates current markup operation in progress -- don&#39;t redo`,
	12: `BufChanged indicates if the text has been changed (edited) relative to the original, since last save`,
	13: `BufFileModOk have already asked about fact that file has changed since being opened, user is ok`,
}

var _BufFlagsMap = map[BufFlags]string{
	10: `BufAutoSaving`,
	11: `BufMarkingUp`,
	12: `BufChanged`,
	13: `BufFileModOk`,
}

// String returns the string representation
// of this BufFlags value.
func (i BufFlags) String() string {
	str := ""
	for _, ie := range _BufFlagsValues {
		if i.HasFlag(ie) {
			ies := ie.BitIndexString()
			if str == "" {
				str = ies
			} else {
				str += "|" + ies
			}
		}
	}
	return str
}

// BitIndexString returns the string
// representation of this BufFlags value
// if it is a bit index value
// (typically an enum constant), and
// not an actual bit flag value.
func (i BufFlags) BitIndexString() string {
	if str, ok := _BufFlagsMap[i]; ok {
		return str
	}
	return strconv.FormatInt(int64(i), 10)
}

// SetString sets the BufFlags value from its
// string representation, and returns an
// error if the string is invalid.
func (i *BufFlags) SetString(s string) error {
	*i = 0
	return i.SetStringOr(s)
}

// SetStringOr sets the BufFlags value from its
// string representation while preserving any
// bit flags already set, and returns an
// error if the string is invalid.
func (i *BufFlags) SetStringOr(s string) error {
	flgs := strings.Split(s, "|")
	for _, flg := range flgs {
		if val, ok := _BufFlagsNameToValueMap[flg]; ok {
			i.SetFlag(true, &val)
		} else if val, ok := _BufFlagsNameToValueMap[strings.ToLower(flg)]; ok {
			i.SetFlag(true, &val)
		} else {
			return errors.New(flg + " is not a valid value for type BufFlags")
		}
	}
	return nil
}

// Int64 returns the BufFlags value as an int64.
func (i BufFlags) Int64() int64 {
	return int64(i)
}

// SetInt64 sets the BufFlags value from an int64.
func (i *BufFlags) SetInt64(in int64) {
	*i = BufFlags(in)
}

// Desc returns the description of the BufFlags value.
func (i BufFlags) Desc() string {
	if str, ok := _BufFlagsDescMap[i]; ok {
		return str
	}
	return i.String()
}

// BufFlagsValues returns all possible values
// for the type BufFlags.
func BufFlagsValues() []BufFlags {
	return _BufFlagsValues
}

// Values returns all possible values
// for the type BufFlags.
func (i BufFlags) Values() []enums.Enum {
	res := make([]enums.Enum, len(_BufFlagsValues))
	for i, d := range _BufFlagsValues {
		res[i] = d
	}
	return res
}

// IsValid returns whether the value is a
// valid option for type BufFlags.
func (i BufFlags) IsValid() bool {
	_, ok := _BufFlagsMap[i]
	return ok
}

// HasFlag returns whether these
// bit flags have the given bit flag set.
func (i BufFlags) HasFlag(f enums.BitFlag) bool {
	return atomic.LoadInt64((*int64)(&i))&(1<<uint32(f.Int64())) != 0
}

// SetFlag sets the value of the given
// flags in these flags to the given value.
func (i *BufFlags) SetFlag(on bool, f ...enums.BitFlag) {
	var mask int64
	for _, v := range f {
		mask |= 1 << v.Int64()
	}
	in := int64(*i)
	if on {
		in |= mask
		atomic.StoreInt64((*int64)(i), in)
	} else {
		in &^= mask
		atomic.StoreInt64((*int64)(i), in)
	}
}

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i BufFlags) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *BufFlags) UnmarshalText(text []byte) error {
	return i.SetString(string(text))
}

var _ViewFlagsValues = []ViewFlags{10, 11, 12}

// ViewFlagsN is the highest valid value
// for type ViewFlags, plus one.
const ViewFlagsN ViewFlags = 13

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the enumgen command to generate them again.
func _ViewFlagsNoOp() {
	var x [1]struct{}
	_ = x[ViewHasLineNos-(10)]
	_ = x[ViewLastWasTabAI-(11)]
	_ = x[ViewLastWasUndo-(12)]
}

var _ViewFlagsNameToValueMap = map[string]ViewFlags{
	`ViewHasLineNos`:   10,
	`viewhaslinenos`:   10,
	`ViewLastWasTabAI`: 11,
	`viewlastwastabai`: 11,
	`ViewLastWasUndo`:  12,
	`viewlastwasundo`:  12,
}

var _ViewFlagsDescMap = map[ViewFlags]string{
	10: `ViewHasLineNos indicates that this view has line numbers (per Buf option)`,
	11: `ViewLastWasTabAI indicates that last key was a Tab auto-indent`,
	12: `ViewLastWasUndo indicates that last key was an undo`,
}

var _ViewFlagsMap = map[ViewFlags]string{
	10: `ViewHasLineNos`,
	11: `ViewLastWasTabAI`,
	12: `ViewLastWasUndo`,
}

// String returns the string representation
// of this ViewFlags value.
func (i ViewFlags) String() string {
	str := ""
	for _, ie := range _ViewFlagsValues {
		if i.HasFlag(ie) {
			ies := ie.BitIndexString()
			if str == "" {
				str = ies
			} else {
				str += "|" + ies
			}
		}
	}
	return str
}

// BitIndexString returns the string
// representation of this ViewFlags value
// if it is a bit index value
// (typically an enum constant), and
// not an actual bit flag value.
func (i ViewFlags) BitIndexString() string {
	if str, ok := _ViewFlagsMap[i]; ok {
		return str
	}
	return strconv.FormatInt(int64(i), 10)
}

// SetString sets the ViewFlags value from its
// string representation, and returns an
// error if the string is invalid.
func (i *ViewFlags) SetString(s string) error {
	*i = 0
	return i.SetStringOr(s)
}

// SetStringOr sets the ViewFlags value from its
// string representation while preserving any
// bit flags already set, and returns an
// error if the string is invalid.
func (i *ViewFlags) SetStringOr(s string) error {
	flgs := strings.Split(s, "|")
	for _, flg := range flgs {
		if val, ok := _ViewFlagsNameToValueMap[flg]; ok {
			i.SetFlag(true, &val)
		} else if val, ok := _ViewFlagsNameToValueMap[strings.ToLower(flg)]; ok {
			i.SetFlag(true, &val)
		} else {
			return errors.New(flg + " is not a valid value for type ViewFlags")
		}
	}
	return nil
}

// Int64 returns the ViewFlags value as an int64.
func (i ViewFlags) Int64() int64 {
	return int64(i)
}

// SetInt64 sets the ViewFlags value from an int64.
func (i *ViewFlags) SetInt64(in int64) {
	*i = ViewFlags(in)
}

// Desc returns the description of the ViewFlags value.
func (i ViewFlags) Desc() string {
	if str, ok := _ViewFlagsDescMap[i]; ok {
		return str
	}
	return i.String()
}

// ViewFlagsValues returns all possible values
// for the type ViewFlags.
func ViewFlagsValues() []ViewFlags {
	return _ViewFlagsValues
}

// Values returns all possible values
// for the type ViewFlags.
func (i ViewFlags) Values() []enums.Enum {
	res := make([]enums.Enum, len(_ViewFlagsValues))
	for i, d := range _ViewFlagsValues {
		res[i] = d
	}
	return res
}

// IsValid returns whether the value is a
// valid option for type ViewFlags.
func (i ViewFlags) IsValid() bool {
	_, ok := _ViewFlagsMap[i]
	return ok
}

// HasFlag returns whether these
// bit flags have the given bit flag set.
func (i ViewFlags) HasFlag(f enums.BitFlag) bool {
	return atomic.LoadInt64((*int64)(&i))&(1<<uint32(f.Int64())) != 0
}

// SetFlag sets the value of the given
// flags in these flags to the given value.
func (i *ViewFlags) SetFlag(on bool, f ...enums.BitFlag) {
	var mask int64
	for _, v := range f {
		mask |= 1 << v.Int64()
	}
	in := int64(*i)
	if on {
		in |= mask
		atomic.StoreInt64((*int64)(i), in)
	} else {
		in &^= mask
		atomic.StoreInt64((*int64)(i), in)
	}
}

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i ViewFlags) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *ViewFlags) UnmarshalText(text []byte) error {
	return i.SetString(string(text))
}