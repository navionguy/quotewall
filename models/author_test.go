package models

import (
	"github.com/gofrs/uuid"
)

const invalidUUID = "563cd207-ab16-4a46-b44e-7317b96c6ba9"
const validAuthor = "b39300f0-6760-4feb-bc32-4b8682b0175d" // matches entry in author.toml
const validName = "George P. Burdell"

func (ms *ModelSuite) Test_AuthorGet() {
	ms.LoadFixture("test authors")

	auths := []Author{}

	err := ms.DB.All(&auths)

	if err != nil {
		ms.Fail(err.Error())
	}

	ms.Equal(3, len(auths))

	var authors Authors
	for _, auth := range auths {
		authors = append(authors, auth)

		ms.Equal(auth.Name, auth.SelectLabel())
		ms.Equal(auth.ID.String(), auth.SelectValue())
	}

	authsJS := authors.String()

	for _, auth := range auths {
		js := auth.String()

		ms.Contains(authsJS, js)
	}
}

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
		ms.EqualValuesf(goterr, tt.expErr, "Author_FindByID(%s) got %b, wanted %b\n", tt.test, goterr, tt.expErr)
	}
}

func (ms *ModelSuite) Test_Author_Create() {
	tests := []struct {
		test   string
		name   string
		expErr bool
	}{
		{test: "Good Name", name: "Shari Freeman", expErr: false},
		{test: "Blank Name", expErr: true},
	}

	ms.LoadFixture(("test authors"))

	for _, tt := range tests {
		auth := &Author{Name: tt.name}
		verrs, _ := auth.Create()

		ms.EqualValuesf(verrs.HasAny(), tt.expErr, "Author_Create(%s) got %b, wanted %b\n", tt.test, verrs.HasAny(), tt.expErr)
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
		ms.EqualValuesf(goterr, tt.expErr, "Author_FindByName(%s) got %b, wanted %b\n", tt.test, goterr, tt.expErr)
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
		ms.EqualValuesf(verrs.HasAny(), tt.expErr, "Author_Update(%s) got %b, wanted %b\n", tt.test, verrs.HasAny(), tt.expErr)
	}
}
