package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/gobuffalo/uuid"
	"github.com/stretchr/testify/require"
)

func Test_Quote(t *testing.T) {

	var fields = []struct {
		fn  string
		msg string
	}{
		{"said_on", "quote said_on field not found"},
		{"publish", "quote publish field not found"},
		{"phrase", "quote phrase field not found"},
		{"Author", "quote Author field not found"},
	}

	a := Quote{}

	js := a.String()

	if len(js) == 0 {
		t.Error("unable to marshal a quote")
		t.Fail()
	}

	rq := require.New(t)

	for _, fld := range fields {
		rq.Containsf(js, fld.fn, fld.msg)
	}

	var ar Quotes
	ar = append(ar, a)

	js = ar.String()

	if len(js) == 0 {
		t.Error("unable to marshal quotes")
		t.Fail()
	}
}

// loadFixtureData loads three of the four tables created in the "test quotes" scenario
//
// Not much code needed to load the data, but lots of steps to check for any failures to
// get the data I expect.  Not all the tests use the loaded data.
func loadFixtureData(ms *ModelSuite) ([]Author, []Annotation, []Conversation) {
	ms.LoadFixture("test authors")
	ms.LoadFixture("test annotations")
	ms.LoadFixture("test conversations")

	authors := []Author{}
	annotations := []Annotation{}
	conversations := []Conversation{}

	err := ms.DB.All(&authors)

	if err != nil {
		ms.FailNow("error getting authors", err.Error())
	}

	if len(authors) == 0 {
		ms.FailNow("no authors found", "no test authors")
	}

	err = ms.DB.All(&annotations)

	if err != nil {
		ms.FailNow("error getting annotations", err.Error())
	}

	if len(annotations) == 0 {
		ms.FailNow("no annotations found", "no test annotations")
	}

	err = ms.DB.All(&conversations)

	if err != nil {
		ms.FailNow("error getting conversations", err.Error())
	}

	if len(conversations) == 0 {
		ms.FailNow("no conversations found", "no test conversations")
	}

	return authors, annotations, conversations
}

// CreateQuote tries to create a simple quote
func (ms *ModelSuite) Test_CreateQuote() {

	authors, _, conversations := loadFixtureData(ms)

	quote := Quote{
		SaidOn:   time.Now(),
		Sequence: 0,
		Phrase:   "A test quote.",
		Publish:  true,
		AuthorID: authors[0].ID,
	}

	verrs, err := quote.Create(ms.DB, conversations[0].ID)

	if err != nil {
		ms.Fail("CreateQuote failed", err.Error())
	}

	if verrs.HasAny() {
		ms.Fail("CreateQuote validation fail", verrs.String())
	}
}

// CreateQuoteWithAnnnotation adds a known annotation to the quote
func (ms *ModelSuite) Test_CreateQuoteWithAnnotation() {
	authors, annotations, conversations := loadFixtureData(ms)

	quote := Quote{
		SaidOn:     time.Now(),
		Sequence:   0,
		Phrase:     "A test quote.",
		Publish:    true,
		Annotation: &annotations[0],
		AuthorID:   authors[0].ID,
	}

	verrs, err := quote.Create(ms.DB, conversations[0].ID)

	if err != nil {
		ms.Fail("CreateQuote failed", err.Error())
	}

	if verrs.HasAny() {
		ms.Fail("CreateQuote validation fail", verrs.String())
	}
}

func (ms *ModelSuite) Test_CreateQuoteWithNewAnnotation() {
	authors, _, conversations := loadFixtureData(ms)

	annotation := &Annotation{Note: "blahblahblah"}

	quote := Quote{
		SaidOn:     time.Now(),
		Sequence:   0,
		Phrase:     "A test quote.",
		Publish:    true,
		Annotation: annotation,
		AuthorID:   authors[0].ID,
	}

	verrs, err := quote.Create(ms.DB, conversations[0].ID)

	if err != nil {
		fmt.Printf("err = %s\n", err.Error())
		ms.Fail("CreateQuote failed", err.Error())
	}

	if verrs.HasAny() {
		ms.Fail("CreateQuote validation fail", verrs.String())
	}
}

func (ms *ModelSuite) Test_CreateQuoteWithInvalidAnnotation() {
	authors, annotations, conversations := loadFixtureData(ms)

	annotations[0].Note = ""
	annotations[0].ID = uuid.Nil

	quote := Quote{
		SaidOn:     time.Now(),
		Sequence:   0,
		Phrase:     "A test quote.",
		Publish:    true,
		Annotation: &annotations[0],
		AuthorID:   authors[0].ID,
	}

	verrs, err := quote.Create(ms.DB, conversations[0].ID)

	if err != nil {
		ms.Fail("CreateQuote failed", err.Error())
	}

	if !verrs.HasAny() {
		ms.Fail("CreateQuote failed", "validation failed to catch empty note")
	}
}

func (ms *ModelSuite) Test_UpdateQuoteWithNewAnnotation() {
	authors, _, conversations := loadFixtureData(ms)
	ms.LoadFixture(("test quotes"))

	quotes := []Quote{}

	err := ms.DB.All(&quotes)

	if err != nil {
		ms.Fail("Update fetch all quotes failed", err.Error())
	}

	quote := &quotes[0]
	quote.AuthorID = authors[0].ID

	verrs, err := quote.Update(ms.DB, conversations[0].ID)

	if err != nil {
		fmt.Printf("err = %s\n", err.Error())
		ms.Fail("UpdateQuote failed", err.Error())
	}

	if verrs.HasAny() {
		ms.Fail("UpdateQuote validation fail", verrs.String())
	}
}
