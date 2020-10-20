package grifts

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/navionguy/quotewall/models"
)

/*
{ "quotearchive" : {
	"conversations" : [
		{ "conversation" : [
			{ "name" : "Bob McGowan", "Quote" : "I don\u0027t see us ever needing to change the product name again.", "date" : "3/14/1997", "publish" : "True" }
			]
		},
		{ "conversation" : [
			{ "name" : "Beth Smith", "Quote" : "Anytime they say, \u0027All you have to do...\u0027, you\u0027re screwed", "date" : "10/1/1997", "publish" : "True" }
			]
		}
    ]}
}
*/

const ctLayout = "1/2/2006" // the date format used in my json file

// CustomTime -- allows me to handle my date format in the JSON File
type CustomTime struct {
	time.Time
}

// UnmarshalJSON  -- extracts out my format
func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}

	s := string(b)
	fmt.Printf("unmarshaling %s\n", s)

	parts := strings.Split(s, "/")

	if len(parts) != 3 {
		err = errors.New("invalid date found")
		return
	}
	ct.Time, err = time.Parse(ctLayout, string(b))

	if err != nil {
		fmt.Printf("unmarshal got %s\n", err)
	}
	return
}

// MarshalJSON CustomTime to my JSON format
func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	s := ct.Format(ctLayout)

	return json.Marshal(s)
}

var sdft time.Time // date time stamp of quote file I'm using

//Utterancestype --
type utterancestype struct {
	Name       string
	Quote      string
	Date       CustomTime
	Publish    string
	Annotation string
}

type conversationtype struct {
	Conversation []utterancestype
}

type archivedatatype struct {
	Conversations []conversationtype
}

type archivetype struct {
	Quotearchive archivedatatype
}

var quotes archivetype

var authorCache authormap

type authormap map[string]uuid.UUID

// seedQuoteDB()
//
// 1. Load the json file of quotes
// 2. for each conversation in the file
//	a. Add each utterance to the database
//	b. Create the conversation with
//

func seedQuoteDB(seedfile string) error {
	err := loadquotedata(seedfile)

	// check if he had an issue

	if err != nil {
		return err
	}

	// but did he unmarshal any useable data

	if len(quotes.Quotearchive.Conversations) == 0 {
		return errors.New("no quotes found in seed file")
	}

	// re-create the authorCache
	authorCache = make(authormap)

	for _, cv := range quotes.Quotearchive.Conversations {
		err = createConversation(cv)
	}

	return err
}

// createConversation()
//
// Builds the database entries for one conversation
//
func createConversation(cv conversationtype) error {
	// create the conversation and give it a unique ID

	conv := &models.Conversation{}
	conv.OccurredOn = cv.Conversation[0].Date.Time
	conv.Publish = (strings.Compare("true", strings.ToLower(cv.Conversation[0].Publish)) == 0)
	tracemsg(fmt.Sprintf("creating conversation, publish = %v, src = %s", conv.Publish, cv.Conversation[0].Publish), 4)

	if err := models.DB.Create(conv); err != nil {
		fmt.Printf("CreateConversation got error %s\n", err)
		return err
	}

	for i, quote := range cv.Conversation {
		err := createQuote(conv.ID, i, quote)

		if err != nil {
			return err
		}
	}

	return nil
}

// createQuote()
//
// Creates a quote record for a quote in a conversation.
//
// He gets told the uuid of the conversation and the sequence number of this quote
// Everything else, he pulls out of the "utterance"
//
func createQuote(cv uuid.UUID, sequence int, qt utterancestype) error {
	tracemsg(fmt.Sprintf("creating quote %s", qt.Quote), 4)

	// find or create the ID for the author
	authID, err := findOrCreateAuthor(qt.Name)

	// if that didn't go well, get out

	if err != nil {
		return err
	}

	// create the quote with as much stuff as we know

	aQuote := &models.Quote{
		SaidOn:         qt.Date.Time,
		Sequence:       sequence,
		Phrase:         qt.Quote,
		AuthorID:       authID,
		Publish:        strings.Compare("true", strings.ToLower(qt.Publish)) == 0,
		ConversationID: cv,
		Annotation:     nil,
		AnnotationID:   nil,
	}

	// if quote is annotated, find or create an ID for that

	if len(qt.Annotation) > 0 {
		aQuote.AnnotationID, err = findOrCreateAnnotation(qt.Annotation)

		if err != nil {
			return err
		}
	}

	fmt.Println("about to create")

	err = models.DB.Create(aQuote)

	if err != nil {
		fmt.Printf("createQuote failed with error %s\n", err)
		return err
	}

	return nil
}

// findOrCreateAuthor()
//
// Try to find the ID for the Author passed.  If you can't, create one.
// First, look in the authorCache.
// Second, look in the database
// If all else fails, create a record in the database
//
func findOrCreateAuthor(author string) (uuid.UUID, error) {
	id := authorCache[author]

	// got him from the cache!
	if id != uuid.Nil {
		tracemsg(fmt.Sprintf("found author %s in the cache", author), 4)
		return id, nil
	}

	// try the database

	authRecs := []models.Author{}
	query := models.DB.Where(fmt.Sprintf("name = '%s'", author))
	err := query.All(&authRecs)

	if err != nil {
		fmt.Printf("Author query returns an error: %s.", err)
		return id, err
	}

	if len(authRecs) > 0 {
		tracemsg(fmt.Sprintf("found author %s in database at %s", authRecs[0].Name, authRecs[0].ID), 4)

		// and add to the cache

		authorCache[authRecs[0].Name] = authRecs[0].ID
		tracemsg(fmt.Sprintf("added author %s to cache", authRecs[0].Name), 3)

		return authRecs[0].ID, nil
	}

	id, err = createAuthor(author)
	if err != nil {
		return uuid.Nil, err
	}

	// and add to the cache

	authorCache[author] = id
	tracemsg(fmt.Sprintf("added new author %s to cache", author), 3)

	return id, nil
}

// findOrCreateAnnotation()
//
// Try to find the ID for the passed Annotation.  If you can't create one.
//

func findOrCreateAnnotation(annotation string) (*uuid.UUID, error) {
	annotateRecs := []models.Annotation{}
	query := models.DB.Where(fmt.Sprintf("note = '%s'", annotation))
	err := query.All(&annotateRecs)

	if err != nil {
		fmt.Printf("annotation query returned an error: %s\n", err)
		return nil, err
	}

	// if I found him, just return his id

	if len(annotateRecs) > 0 {
		id := annotateRecs[0].ID

		return &id, nil
	}

	// create a new annotaton

	id, err := createAnnotation(annotation)

	if err != nil {
		fmt.Printf("createAnnotation() returned an error: %s\n", err)
		return nil, err
	}

	tracemsg(fmt.Sprintf("added annotation %s to database", annotation), 4)

	return &id, nil
}

// createAuthor() - Create his record in the database
//
// Creates a ID value for the author and then writes a record into the database
//
func createAuthor(author string) (uuid.UUID, error) {
	// create a database record

	rec := models.Author{
		Name: author,
	}

	err := models.DB.Create(&rec)
	if err != nil {
		fmt.Printf("author create failed with %s\n", err)
		return uuid.Nil, err
	}

	tracemsg(fmt.Sprintf("adding author %s with ID %s", rec.Name, rec.ID), 4)

	return rec.ID, nil
}

// createAnnotation() -- Create an annotation record in the database
//
// Stores the annotation string into the database and returns its ID.
//
func createAnnotation(annotation string) (uuid.UUID, error) {

	// create new object with the annotation text
	rec := models.Annotation{
		Note: annotation,
	}

	err := models.DB.Create(&rec)
	if err != nil {
		fmt.Printf("annotation create failed with %s\n", err)
		return uuid.Nil, err
	}

	tracemsg(fmt.Sprintf("adding annotation %s with ID %s", rec.Note, rec.ID), 4)

	return rec.ID, nil
}

// loadquotedata()
//
// Accepts a filename, or full path, and *attempts* to read it
// into memory and convert it into an array of quotes.
//
func loadquotedata(filename string) error {

	file, e := ioutil.ReadFile(filename)
	if e != nil {
		return e
	}
	json.Unmarshal(file, &quotes)
	tracemsg(fmt.Sprintf("found %d quotes", len(quotes.Quotearchive.Conversations)), 1)

	return nil
}

var verbosity int

func init() {
	verbosity = 0
}

// tracemsg()
//
// Decides to show, or not show, a trace message based on the message trace level
// versus the specified verbosity
//
func tracemsg(msg string, lvl int) {
	if lvl <= verbosity {
		fmt.Println(msg)
	}
}

// setVerbosity()
//
// Sets the verboisty level for tracing,
// should be based on command line
//
func setVerbosity(v int) {
	verbosity = v
}
