package models_test

import (
	"testing"

	"github.com/gobuffalo/uuid"
	"github.com/navionguy/quotewall/models"
	"github.com/stretchr/testify/require"
)

func Test_Annotation(t *testing.T) {

	var fields = []struct {
		fn  string
		msg string
	}{
		{"id", "annotation id field not found"},
		{"note", "annotation note field not found"},
		{"created_at", "created_at field not found"},
		{"updated_at", "updated_at field not found"},
	}

	a := models.Annotation{
		Note: "Snarky comment",
	}

	js := a.String()

	if len(js) == 0 {
		t.Error("unable to marshal an annotation")
		t.Fail()
	}

	rq := require.New(t)

	for _, fld := range fields {
		rq.Containsf(js, fld.fn, fld.msg)
	}

	var ar models.Annotations
	ar = append(ar, a)

	js = ar.String()

	if len(js) == 0 {
		t.Error("unable to marshal array of annotations")
		t.Fail()
	}
}

const validAnnotation = "First Timer!"
const invalidAnnotation = "*^*&^*&^*^*^&"

func (ms *modelSuite) Test_Annotation_FindByNote() {
	ms.LoadFixture("test annotations")

	ano := models.Annotation{
		Note: validAnnotation,
	}

	err := ano.FindByNote()

	if err != nil {
		ms.Fail("annotation FindByNote database error", err.Error)
	}

	if ano.ID == uuid.Nil {
		ms.Fail("annotation FindByNote didn't find", ano.Note)
	}
}

func (ms *modelSuite) Test_Annotation_FindByNote_NoFind() {
	ms.LoadFixture("test annotations")

	ano := models.Annotation{
		Note: invalidAnnotation,
	}

	err := ano.FindByNote()

	if err != nil {
		ms.Fail("annotation FindByNote database error", err.Error)
	}

	if ano.ID != uuid.Nil {
		ms.Fail("annotation FindByNote found invalid", ano.Note)
	}
}
