package models_test

import (
	"testing"
	"time"

	"github.com/navionguy/quotewall/models"
	"github.com/stretchr/testify/require"
)

func Test_Conversation(t *testing.T) {

	var fields = []struct {
		fn  string
		msg string
	}{
		{"id", "conversation id field not found"},
		{"occurredon", "conversation occurred_on field not found"},
		{"publish", "conversation publish field not found"},
		{"Quotes", "conversation Quotes field not found"},
		{"created_at", "created_at field not found"},
		{"updated_at", "updated_at field not found"},
	}

	a := models.Conversation{}

	js := a.String()

	if len(js) == 0 {
		t.Error("unable to marshal a conversation")
		t.Fail()
	}

	rq := require.New(t)

	for _, fld := range fields {
		rq.Containsf(js, fld.fn, fld.msg)
	}

	var ar models.Conversations
	ar = append(ar, a)

	js = ar.String()

	if len(js) == 0 {
		t.Error("unable to marshal conversations")
		t.Fail()
	}
}

func (ms *modelSuite) Test_CreateConversation() {
	authors, _, _ := loadFixtureData(ms) // re-use from quote_test.go

	q := models.Quote{
		Phrase:   "A shiny new quote.",
		Publish:  true,
		SaidOn:   time.Now(),
		AuthorID: authors[0].ID,
		Sequence: 0,
	}
	conversation := models.Conversation{
		Publish:    true,
		OccurredOn: time.Now(),
	}
	conversation.Quotes = append(conversation.Quotes, q)

	verrs, err := conversation.Create()

	if err != nil {
		ms.Fail("unable to create conversation", err.Error())
	}

	if verrs != nil {
		if verrs.HasAny() {
			ms.Fail("conversation failed to validate", verrs.String())
		}
	}
}

func (ms *modelSuite) Test_CreateConversationInvalidOccurredOn() {
	authors, _, _ := loadFixtureData(ms) // re-use from quote_test.go

	q := models.Quote{
		Phrase:   "A shiny new quote.",
		Publish:  true,
		SaidOn:   time.Now(),
		AuthorID: authors[0].ID,
		Sequence: 0,
	}
	conversation := models.Conversation{
		Publish:    true,
		OccurredOn: time.Now(),
	}
	conversation.Quotes = append(conversation.Quotes, q)
	conversation.OccurredOn = conversation.OccurredOn.AddDate(0, 0, 2)

	verrs, err := conversation.Create()

	if err != nil {
		ms.Fail("unable to create conversation", err.Error())
	}

	if verrs == nil {
		ms.Fail("created conversation with invalid occurredon", "two days in future")
	}

	if verrs != nil {
		if !verrs.HasAny() {
			ms.Fail("conversation validated invalid occurredon", "in the future")
		}
	}
}

func (ms *modelSuite) Test_CreateConversationInvalidSaidOn() {
	authors, _, _ := loadFixtureData(ms) // re-use from quote_test.go

	q := models.Quote{
		Phrase:   "A shiny new quote.",
		Publish:  true,
		SaidOn:   time.Now(),
		AuthorID: authors[0].ID,
		Sequence: 0,
	}
	q.SaidOn = q.SaidOn.AddDate(0, 0, 2)
	conversation := models.Conversation{
		Publish:    true,
		OccurredOn: time.Now(),
	}
	conversation.Quotes = append(conversation.Quotes, q)

	verrs, err := conversation.Create()

	if err != nil {
		ms.Fail("unable to create conversation", err.Error())
	}

	if verrs == nil {
		ms.Fail("created conversation with invalid quote saidon", "two days in future")
	}

	if verrs != nil {
		if !verrs.HasAny() {
			ms.Fail("conversation validated invalid quote saidon", "in the future")
		}
	}
}
