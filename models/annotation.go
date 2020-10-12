package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Annotation holds
type Annotation struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Note      string    `json:"note" db:"note" form:"Annotation"`
}

// String is not required by pop and may be deleted
func (a Annotation) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// Annotations is not required by pop and may be deleted
type Annotations []Annotation

// String is not required by pop and may be deleted
func (a Annotations) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (a *Annotation) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: a.Note, Name: "Note"},
		&validators.StringLengthInRange{Field: a.Note, Name: "Note", Min: 0, Max: 255, Message: "length must be <255"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (a *Annotation) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (a *Annotation) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// FindByNote tries to find the note already saved in the
// annotations table.
// If it does, it returns the full annotation record.
// If it does not, it returns an annotation object that
// holds the note, but has the ID set to uuid.NULL
func (a *Annotation) FindByNote() error {

	annoRecs := []Annotation{}
	query := DB.Where(fmt.Sprintf("note = '%s'", a.Note))
	err := query.All(&annoRecs)

	if err != nil {
		return err
	}

	if len(annoRecs) > 0 {
		a.ID = annoRecs[0].ID
		return nil // found him!
	}

	a.ID = uuid.Nil

	return nil
}

// CheckID makes sure that if there is an annotation, the db record for it exists
func (a *Annotation) CheckID(db *pop.Connection) (*validate.Errors, error) {

	var ve validate.Errors

	if a == nil {
		return &ve, nil
	}

	if a.ID != uuid.Nil {
		return &ve, nil
	}

	verrs, err := db.ValidateAndCreate(a)

	if err != nil || verrs.HasAny() {
		return verrs, err
	}

	// he created
	return &ve, nil
}
