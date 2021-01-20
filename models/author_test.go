package models

import (
	"github.com/gofrs/uuid"
)

const invalidUUID = "563cd207-ab16-4a46-b44e-7317b96c6ba9"
const validAuthor = "b39300f0-6760-4feb-bc32-4b8682b0175d" // matches entry in author.toml
const validName = "George P. Burdell"

func (ms *ModelSuite) Test_Author_FindByID() {
	tests := []struct {
		test   string
		id     string
		expErr bool
	}{
		{test: "Good ID", id: validAuthor, expErr: false},
		{test: "Bad ID", id: invalidUUID, expErr: true},
	}

	ms.LoadFixture(("test authors"))

	for _, tt := range tests {
		id, err := uuid.FromString(tt.id)

		if err != nil {
			ms.Fail("uuid.FromString failed", err.Error())
		}

		auth := Author{ID: id}
		err = auth.FindByID()

		goterr := (err != nil)
		ms.EqualValuesf(goterr, tt.expErr, "Author_FindByID(%s) got %t, wanted %t\n", tt.test, goterr, tt.expErr)
	}
}

func (ms *ModelSuite) Test_Author_Create() {
	tests := []struct {
		test    string
		name    string
		expErr  bool
		expJSON string
		arrJSON string
	}{
		{test: "Good Name", name: "Shari Freeman", expErr: false, expJSON: "{\"name\":\"Shari Freeman\"}", arrJSON: "[{\"name\":\"Shari Freeman\"}]"},
		{test: "Blank Name", expErr: true},
	}

	ms.LoadFixture(("test authors"))

	for _, tt := range tests {
		auth := &Author{Name: tt.name}
		verrs, _ := auth.Create()

		ms.EqualValuesf(verrs.HasAny(), tt.expErr, "Author_Create(%s) got %t, wanted %t\n", tt.test, verrs.HasAny(), tt.expErr)

		if !tt.expErr {
			// got a valid author, let's test his string functions
			jsn := auth.String()

			ms.EqualValuesf(tt.expJSON, jsn, "Author.String(%s) got %s, wanted %s\n", tt.test, jsn, tt.expJSON)

			var authAry Authors
			authAry = append(authAry, *auth)
			ajsn := authAry.String()

			ms.EqualValuesf(tt.arrJSON, ajsn, "Author[].String(%s) got %s, wanted %s\n", tt.test, ajsn, tt.arrJSON)

			// and his select functions

			ms.EqualValuesf(auth.SelectValue(), auth.ID.String(), "Author.SelectValue(%s) got %s, wanted %s\n", tt.test, auth.SelectValue(), auth.ID.String())
			ms.EqualValuesf(auth.SelectLabel(), auth.Name, "Author.SelectLabel(%s) got %s, wanted %s\n", tt.test, auth.SelectLabel(), auth.Name)
		}
	}
}

func (ms *ModelSuite) Test_Author_FindByName() {
	tests := []struct {
		test   string
		name   string
		expErr bool
	}{
		{test: "Good Name", name: "Shari Freeman", expErr: false},
		{test: "Bad Name", name: "Alfred E. Neuman", expErr: true},
		{test: "Blank Name", name: "", expErr: true},
	}

	ms.LoadFixture(("test authors"))

	for _, tt := range tests {
		auth := Author{Name: tt.name}
		err := auth.FindByName()

		goterr := (err != nil)
		ms.EqualValuesf(goterr, tt.expErr, "Author_FindByName(%s) got %t, wanted %t\n", tt.test, goterr, tt.expErr)
	}
}

func (ms *ModelSuite) Test_Author_Update() {
	tests := []struct {
		test    string
		id      string
		newName string
		expErr  bool
	}{
		{test: "Good Name", id: validAuthor, newName: "Alfred E. Neuman", expErr: false},
		{test: "Bad Name", id: validAuthor, expErr: true},
	}

	ms.LoadFixture(("test authors"))

	for _, tt := range tests {
		id, err := uuid.FromString(tt.id)

		if err != nil {
			ms.Fail("uuid.FromString failed", err.Error())
		}

		auth := Author{ID: id}

		if len(tt.newName) > 0 {
			auth.Name = tt.newName
		}

		verrs, _ := auth.Update()
		ms.EqualValuesf(verrs.HasAny(), tt.expErr, "Author_Update(%s) got %t, wanted %t\n", tt.test, verrs.HasAny(), tt.expErr)
	}
}
