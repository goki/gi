// Code generated by "goki generate"; DO NOT EDIT.

package main

import (
	"goki.dev/gti"
	"goki.dev/ordmap"
)

var _ = gti.AddType(&gti.Type{
	Name:      "main.TableStruct",
	ShortName: "main.TableStruct",
	IDName:    "table-struct",
	Doc:       "TableStruct is a testing struct for table view",
	Directives: gti.Directives{
		&gti.Directive{Tool: "gti", Directive: "add", Args: []string{}},
	},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Icon", &gti.Field{Name: "Icon", Type: "icons.Icon", Doc: "an icon", Directives: gti.Directives{}}},
		{"IntField", &gti.Field{Name: "IntField", Type: "int", Doc: "an integer field", Directives: gti.Directives{}}},
		{"FloatField", &gti.Field{Name: "FloatField", Type: "float32", Doc: "a float field", Directives: gti.Directives{}}},
		{"StrField", &gti.Field{Name: "StrField", Type: "string", Doc: "a string field", Directives: gti.Directives{}}},
	}),
	Embeds:  ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:      "main.ILStruct",
	ShortName: "main.ILStruct",
	IDName:    "il-struct",
	Doc:       "ILStruct is an inline-viewed struct",
	Directives: gti.Directives{
		&gti.Directive{Tool: "gti", Directive: "add", Args: []string{}},
	},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"On", &gti.Field{Name: "On", Type: "bool", Doc: "click to show next", Directives: gti.Directives{}}},
		{"ShowMe", &gti.Field{Name: "ShowMe", Type: "string", Doc: "can u see me?", Directives: gti.Directives{}}},
		{"Cond", &gti.Field{Name: "Cond", Type: "int", Doc: "a conditional", Directives: gti.Directives{}}},
		{"Cond1", &gti.Field{Name: "Cond1", Type: "string", Doc: "On and Cond=0 -- note that slbool as bool cannot be used directly..", Directives: gti.Directives{}}},
		{"Cond2", &gti.Field{Name: "Cond2", Type: "TableStruct", Doc: "if Cond=0", Directives: gti.Directives{}}},
		{"Val", &gti.Field{Name: "Val", Type: "float32", Doc: "a value", Directives: gti.Directives{}}},
	}),
	Embeds:  ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:      "main.Struct",
	ShortName: "main.Struct",
	IDName:    "struct",
	Doc:       "Struct is a testing struct for struct view",
	Directives: gti.Directives{
		&gti.Directive{Tool: "gti", Directive: "add", Args: []string{}},
	},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Stripes", &gti.Field{Name: "Stripes", Type: "gi.Stripes", Doc: "an enum", Directives: gti.Directives{}}},
		{"Name", &gti.Field{Name: "Name", Type: "string", Doc: "a string", Directives: gti.Directives{}}},
		{"ShowNext", &gti.Field{Name: "ShowNext", Type: "bool", Doc: "click to show next", Directives: gti.Directives{}}},
		{"ShowMe", &gti.Field{Name: "ShowMe", Type: "string", Doc: "can u see me?", Directives: gti.Directives{}}},
		{"Inline", &gti.Field{Name: "Inline", Type: "ILStruct", Doc: "how about that", Directives: gti.Directives{}}},
		{"Cond", &gti.Field{Name: "Cond", Type: "int", Doc: "a conditional", Directives: gti.Directives{}}},
		{"Cond1", &gti.Field{Name: "Cond1", Type: "string", Doc: "if Cond=0", Directives: gti.Directives{}}},
		{"Cond2", &gti.Field{Name: "Cond2", Type: "TableStruct", Doc: "if Cond=0", Directives: gti.Directives{}}},
		{"Val", &gti.Field{Name: "Val", Type: "float32", Doc: "a value", Directives: gti.Directives{}}},
		{"Things", &gti.Field{Name: "Things", Type: "[]*TableStruct", Doc: "", Directives: gti.Directives{}}},
		{"Stuff", &gti.Field{Name: "Stuff", Type: "[]float32", Doc: "", Directives: gti.Directives{}}},
	}),
	Embeds:  ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})