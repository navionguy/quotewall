package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
)

// Quote holds what one person said
type Quote struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	SaidOn    time.Time `json:"said_on" db:"saidon" form:"SaidOn"`
	Sequence  int       `json:"sequence" db:"sequence" form:"sequence"`
	Phrase    string    `json:"phrase" db:"phrase" form:"Phrase"`
	Publish   bool      `json:"publish" db:"publish" form:"MakePublic"`

	// Relationships
	Conversation Conversation `json:"-" belongs_to:"conversation" db:"-"`
	Author       Author       `belongs_to:"author" db:"-"`
	Annotation   *Annotation  `belongs_to:"annotation" db:"-"`

	// Foreign keys
	ConversationID uuid.UUID  `json:"conversation_id" db:"conversation_id"`
	AuthorID       uuid.UUID  `json:"author_id" db:"author_id"`
	AnnotationID   *uuid.UUID `json:"annotation_id" db:"annotation_id"`
}

// String is not required by pop and may be deleted
func (q Quote) String() string {
	jq, _ := json.Marshal(q)
	return string(jq)
}

// Quotes is not required by pop and may be deleted
type Quotes []Quote

// String is not required by pop and may be deleted
func (q Quotes) String() string {
	jq, _ := json.Marshal(q)
	return string(jq)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (q *Quote) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.TimeIsPresent{Field: q.SaidOn, Name: "SaidOn"},
		&validators.TimeIsBeforeTime{FirstTime: q.SaidOn, SecondTime: time.Now().AddDate(0, 0, 1), FirstName: "Said On", SecondName: "Tomorrow"},

		&validators.IntIsGreaterThan{Field: q.Sequence, Name: "sequence", Compared: -1, Message: "sequence must be >= 0"},

		&validators.StringIsPresent{Field: q.Phrase, Name: "Phrase"},
		&validators.StringLengthInRange{Field: q.Phrase, Name: "Phrase", Min: 1, Max: 255, Message: "length must be <255"},

		&validators.FuncValidator{
			Field:   q.AuthorID.String(),
			Name:    "AuthorID",
			Message: "quote.AuthorID %s is NIL",
			Fn: func() bool {
				return !(q.AuthorID == uuid.Nil)
			},
		},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (q *Quote) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (q *Quote) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Create saves a quote pointing at the conversation
func (q *Quote) Create(db *pop.Connection, id uuid.UUID) (*validate.Errors, error) {

	verrs, err := q.Annotation.CheckID(db)

	if err != nil || verrs.HasAny() {
		return verrs, err
	}

	// save the ConversationID into the quote
	q.ConversationID = id

	// add the quote
	verrs, err = db.ValidateAndCreate(q)

	return verrs, err
}

// Update saves a quote pointing at the conversation
func (q *Quote) Update(db *pop.Connection, id uuid.UUID) (*validate.Errors, error) {

	verrs, err := q.Annotation.CheckID(db)

	if err != nil || verrs.HasAny() {
		return verrs, err
	}

	// save the ConversationID into the quote
	q.ConversationID = id

	// add the quote
	verrs, err = db.ValidateAndUpdate(q)

	return verrs, err
}
