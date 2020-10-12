package models_test

import (
	"testing"
	"time"

	"github.com/gobuffalo/suite/v3"
	"github.com/gobuffalo/uuid"
	"github.com/navionguy/quotewall/models"
	"github.com/stretchr/testify/require"
)

type modelSuite struct {
	*suite.Model
}

func Test_Quote(t *testing.T) {

	var fields = []struct {
		fn  string
		msg string
	}{
		{"id", "quote id field not found"},
		{"created_at", "created_at field not found"},
		{"updated_at", "updated_at field not found"},
		{"said_on", "quote said_on field not found"},
		{"sequence", "quote sequence # field not found"},
		{"publish", "quote publish field not found"},
		{"phrase", "quote phrase field not found"},
		{"Author", "quote Author field not found"},
		{"conversation_id", "quote conversation_id field not found"},
	}

	a := models.Quote{}

	js := a.String()

	if len(js) == 0 {
		t.Error("unable to marshal a quote")
		t.Fail()
	}

	rq := require.New(t)

	for _, fld := range fields {
		rq.Containsf(js, fld.fn, fld.msg)
	}

	var ar models.Quotes
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
func loadFixtureData(ms *modelSuite) ([]models.Author, []models.Annotation, []models.Conversation) {
	ms.LoadFixture("test quotes")

	authors := []models.Author{}
	annotations := []models.Annotation{}
	conversations := []models.Conversation{}

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
func (ms *modelSuite) Test_CreateQuote() {

	authors, _, conversations := loadFixtureData(ms)

	quote := models.Quote{
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
func (ms *modelSuite) Test_CreateQuoteWithAnnotation() {
	authors, annotations, conversations := loadFixtureData(ms)

	quote := models.Quote{
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

func (ms *modelSuite) Test_CreateQuoteWithNewAnnotation() {
	authors, annotations, conversations := loadFixtureData(ms)

	annotations[0].Note = "blahblahblah"
	annotations[0].ID = uuid.Nil

	quote := models.Quote{
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

func (ms *modelSuite) Test_CreateQuoteWithInvalidAnnotation() {
	authors, annotations, conversations := loadFixtureData(ms)

	annotations[0].Note = ""
	annotations[0].ID = uuid.Nil

	quote := models.Quote{
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
