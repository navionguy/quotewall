package models

import (
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func Test_Conversation(t *testing.T) {

	var fields = []struct {
		fn  string
		msg string
	}{
		{"occurredon", "conversation occurred_on field not found"},
		{"publish", "conversation publish field not found"},
		{"Quotes", "conversation Quotes field not found"},
	}

	a := Conversation{}

	js := a.String()

	if len(js) == 0 {
		t.Error("unable to marshal a conversation")
		t.Fail()
	}

	rq := require.New(t)

	for _, fld := range fields {
		rq.Containsf(js, fld.fn, fld.msg)
	}

	var ar Conversations
	ar = append(ar, a)

	js = ar.String()

	if len(js) == 0 {
		t.Error("unable to marshal conversations")
		t.Fail()
	}
}

func (ms *ModelSuite) Test_CreateConversation() {
	tests := []struct {
		saidOnAdd     int // #days to add
		OccurredOnAdd int // #days to add
		getVErrors    bool
	}{
		{0, 0, false},
		{3, 0, true},
		{0, 3, true},
	}
	authors, _, _ := loadFixtureData(ms) // re-use from quote_test.go

	for _, tt := range tests {

		q := Quote{
			Phrase:   "A shiny new quote.",
			Publish:  true,
			SaidOn:   time.Now(),
			AuthorID: authors[0].ID,
			Sequence: 0,
		}
		conversation := Conversation{
			Publish:    true,
			OccurredOn: time.Now(),
		}

		q.SaidOn = q.SaidOn.AddDate(0, 0, tt.saidOnAdd)
		conversation.OccurredOn = conversation.OccurredOn.AddDate(0, 0, tt.OccurredOnAdd)

		conversation.Quotes = append(conversation.Quotes, q)

		verrs, err := conversation.Create()

		if err != nil {
			ms.Fail("unable to create conversation", err.Error())
		}

		if verrs != nil {
			gv := verrs.HasAny()
			if gv != tt.getVErrors {
				ms.Failf("conversation validate result ", "was %t, wanted %t", gv, tt.getVErrors)
			}
		}

	}
}

func (ms *ModelSuite) Test_UpdateConversation() {
	tests := []struct {
		saidOnAdd     int // #days to add
		OccurredOnAdd int // #days to add
		getVErrors    bool
	}{
		{0, 0, false},
		{3, 0, true},
		{0, 3, true},
	}

	authors, _, conversations := loadFixtureData(ms)
	ms.LoadFixture("test quotes")

	quotes := []Quote{}

	err := ms.DB.All(&quotes)

	if err != nil {
		ms.Fail("Test_UpdateConversation failed to load quotes ", err.Error())
	}

	conversations[0].Publish = false
	conversations[0].Quotes = append(conversations[0].Quotes, quotes[0])
	conversations[0].Quotes[0].AuthorID = authors[0].ID

	for _, tt := range tests {
		conversations[0].OccurredOn = conversations[0].OccurredOn.AddDate(0, 0, tt.OccurredOnAdd)
		conversations[0].Quotes[0].SaidOn = conversations[0].Quotes[0].SaidOn.AddDate(0, 0, tt.saidOnAdd)

		verrs, err := conversations[0].Update()

		if err != nil {
			ms.Fail("Test_UpdateConversation failed on Update ", err.Error())
		}

		if verrs != nil {
			gv := verrs.HasAny()
			if gv != tt.getVErrors {
				ms.Failf("conversation validate result ", "was %t, wanted %t", gv, tt.getVErrors)
			}
		}

		// backout changes for this test
		conversations[0].OccurredOn = conversations[0].OccurredOn.AddDate(0, 0, -tt.OccurredOnAdd)
		conversations[0].Quotes[0].SaidOn = conversations[0].Quotes[0].SaidOn.AddDate(0, 0, -tt.saidOnAdd)
	}
}

func (ms *ModelSuite) Test_UpdateConversationAddingQuote() {
	authors, _, conversations := loadFixtureData(ms)
	ms.LoadFixture("test quotes")

	quotes := []Quote{}

	quote := Quote{
		SaidOn:   time.Now(),
		Sequence: 1,
		Phrase:   "A test quote.",
		Publish:  true,
		AuthorID: authors[1].ID,
		Author:   authors[1],
	}

	err := ms.DB.All(&quotes)

	if err != nil {
		ms.Fail("Test_UpdateConversation failed to load quotes ", err.Error())
	}

	conversations[0].Publish = false
	conversations[0].Quotes = append(conversations[0].Quotes, quotes[0])
	conversations[0].Quotes = append(conversations[0].Quotes, quote)
	conversations[0].Quotes[0].AuthorID = authors[0].ID

	verrs, err := conversations[0].Update()

	if err != nil {
		ms.Fail("Test_UpdateConversation failed on Update ", err.Error())
	}

	if verrs.HasAny() {
		ms.Fail("Test_UpdateConversation failed verification", verrs.String())
	}
}

func Test_Marshal(t *testing.T) {
	auth := Author{
		Name: "George P.Burdell",
		ID:   uuid.FromStringOrNil("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
	}

	note := Annotation{
		Note: "A snide comment",
	}

	q := Quote{
		Phrase:     "A shiny new quote.",
		Publish:    true,
		SaidOn:     time.Now(),
		AuthorID:   auth.ID,
		Sequence:   0,
		Annotation: &note,
	}
	conversation := Conversation{
		Publish:    true,
		OccurredOn: time.Now(),
	}
	conversation.Quotes = append(conversation.Quotes, q)

	res, err := conversation.MarshalConversation()

	if err != nil {
		t.Fatalf("marshal failed with error %s", err.Error())
	}

	var cv2 Conversation
	err = cv2.ExtractConversationFromJSON(res)

	if err != nil {
		t.Fatalf("ExtractConversationFromJSON failed with error %s", err.Error())
	}

	if !cv2.OccurredOn.Equal(conversation.OccurredOn) {
		t.Fatal("OccurredOn time changed!")
	}

	if cv2.Publish != conversation.Publish {
		t.Fatal("Publish flag changed!")
	}

	if len(cv2.Quotes) != len(conversation.Quotes) {
		t.Fatal("Quote count changed!")
	}

	if strings.Compare(cv2.Quotes[0].Phrase, conversation.Quotes[0].Phrase) != 0 {
		t.Fatal("Quote text changed!")
	}

	if !cv2.Quotes[0].SaidOn.Equal(conversation.Quotes[0].SaidOn) {
		t.Fatal("Quote SaidOn changed!")
	}

	if strings.Compare(cv2.Quotes[0].Author.Name, conversation.Quotes[0].Author.Name) != 0 {
		t.Fatal("Quote Author changed!")
	}

	if strings.Compare(cv2.Quotes[0].Annotation.Note, conversation.Quotes[0].Annotation.Note) != 0 {
		t.Fatal("Quote Annotation changed!")
	}

	/*
			%7B%22id%22:%2200000000-0000-0000-0000-000000000000%22%2C%22created_at%22:%220001-01-01T00:00:00Z%22%2C
			%22updated_at%22:%220001-01-01T00:00:00Z%22%2C%22occurredon%22:%222020-10-15T11:31:21.320123-04:00%22%2C%22publish%22:true%2C
			%22Quotes%22:%5B%7B%22id%22:%2200000000-0000-0000-0000-000000000000%22%2C%22created_at%22:%220001-01-01T00:00:00Z%22%2C%22updated_at%22:%220001-01-01T00:00:00Z%22%2C
			%22said_on%22:%222020-10-15T11:31:21.320123-04:00%22%2C%22sequence%22:0%2C%22phrase%22:%22A%20shiny%20new%20quote.%22%2C%22publish%22:true%2C
			%22Author%22:%7B%22id%22:%2200000000-0000-0000-0000-000000000000%22%2C%22created_at%22:%220001-01-01T00:00:00Z%22%2C%22updated_at%22:%220001-01-01T00:00:00Z%22%2C
			%22name%22:%22%22%7D%2C%22Annotation%22:%7B%22id%22:%2200000000-0000-0000-0000-000000000000%22%2C%22created_at%22:%220001-01-01T00:00:00Z%22%2C
			%22updated_at%22:%220001-01-01T00:00:00Z%22%2C%22note%22:%22A%20snide%20comment%22%7D%2C%22conversation_id%22:%2200000000-0000-0000-0000-000000000000%22%2C
			%22author_id%22:%226ba7b810-9dad-11d1-80b4-00c04fd430c8%22%2C%22annotation_id%22:null%7D%5D%7D

			{"id":"00000000-0000-0000-0000-000000000000",
			"created_at":"0001-01-01T00:00:00Z",
			"updated_at":"0001-01-01T00:00:00Z",
			"occurredon":"2020-10-15T11:07:40.236165-04:00",
			"publish":true,
			"Quotes":[{
				"id":"00000000-0000-0000-0000-000000000000",
				"created_at":"0001-01-01T00:00:00Z",
				"updated_at":"0001-01-01T00:00:00Z",
				"said_on":"2020-10-15T11:07:40.236164-04:00",
				"sequence":0,
				"phrase":"A shiny new quote.",
				"publish":true,
				"Author":{
					"id":"00000000-0000-0000-0000-000000000000",
					"created_at":"0001-01-01T00:00:00Z",
					"updated_at":"0001-01-01T00:00:00Z",
					"name":""
				},
				"Annotation":{
					"id":"00000000-0000-0000-0000-000000000000",
					"created_at":"0001-01-01T00:00:00Z",
					"updated_at":"0001-01-01T00:00:00Z",
					"note":"A snide comment"
				},
				"conversation_id":"00000000-0000-0000-0000-000000000000",
				"author_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				"annotation_id":null
			}]
		}
	*/
}
