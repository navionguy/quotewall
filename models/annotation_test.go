package models

import (
	"fmt"

	"github.com/gobuffalo/uuid"
)

const validAnnotation = "First Timer!"
const invalidAnnotation = "*^*&^*&^*^*^&"

func (ms *ModelSuite) Test_Annotation_FindByNote() {
	tests := []struct {
		test string
		note string
		exp  bool
		jsn  string
		ajsn string
	}{
		{test: "Valid", note: validAnnotation, exp: true, jsn: "{\"note\":\"First Timer!\"}", ajsn: "[{\"note\":\"First Timer!\"}]"},
		{test: "Invalid", note: invalidAnnotation, exp: false},
	}
	ms.LoadFixture("test annotations")

	for _, tt := range tests {
		ano := Annotation{
			Note: tt.note,
		}

		err := ano.FindByNote()

		if err != nil {
			ms.Fail("annotation FindByNote database error", err.Error)
		}

		gotAno := (ano.ID != uuid.Nil)
		ms.EqualValuesf(gotAno, tt.exp, "Annotation_FindByNote(%s) got %t, wanted %t\n", tt.test, gotAno, tt.exp)

		if tt.exp {
			// should have a valid Annotation
			// let's test his string functions
			jsn := ano.String()
			ms.EqualValuesf(jsn, tt.jsn, "Annotation_String(%s) got %t, wanted %t\n", tt.test, jsn, tt.jsn)

			var anoArr Annotations
			anoArr = append(anoArr, ano)
			ajsn := anoArr.String()

			ms.EqualValuesf(ajsn, tt.ajsn, "[]Annotation_String(%s) got %t, wanted %t\n", tt.test, ajsn, tt.ajsn)

		}
	}
}

func (ms *ModelSuite) Test_Annotation_CheckID() {
	valid := validAnnotation
	newAnno := "Something original"
	invAnno := ""

	tests := []struct {
		test string
		note *string
		exp  bool
		verr bool
	}{
		{test: "NoNote", note: nil, exp: false, verr: false},
		{test: "Valid", note: &valid, exp: true, verr: false},
		{test: "New", note: &newAnno, exp: true, verr: false},
		{test: "Invalid", note: &invAnno, exp: false, verr: true},
	}
	ms.LoadFixture("test annotations")

	for _, tt := range tests {
		ano := &Annotation{}

		if tt.note != nil {
			ano.Note = *tt.note
			err := ano.FindByNote()
			fmt.Printf("Finding %s\n", ano.Note)

			if err != nil {
				ms.Fail("annotation FindByNote database error", err.Error)
			}
			fmt.Printf("got uuid %s, note %s\n", ano.ID.String(), ano.Note)
		} else {
			ano = nil
		}

		verrs, err := ano.CheckID(ms.DB)

		ms.EqualValuesf(verrs.HasAny(), tt.verr, "Annotation_CheckID(%s) got %t, wanted %t", tt.test, verrs.HasAny(), tt.verr)

		if err != nil {
			ms.Failf("Annotation_CheckID(%s) got error %s", tt.test, err.Error())
		}
	}
}
