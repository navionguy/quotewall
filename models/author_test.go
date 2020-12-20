package models_test

import (
	"fmt"

	"github.com/navionguy/quotewall/models"
)

/*type modelSuite struct {
	*suite.Model
}*/
/*
func Test_Author(t *testing.T) {

	var fields = []struct {
		fn  string
		msg string
	}{
		{"id", "author id field not found"},
		{"name", "author name field not found"},
		{"created_at", "created_at field not found"},
		{"updated_at", "updated_at field not found"},
	}

	a := models.Author{
		Name: "George B. Burdell",
	}

	// convert author to json

	js := a.String()

	// if you get nothing, that's a problem

	if len(js) == 0 {
		t.Error("unable to marshal author")
		t.FailNow()
	}

	// make sure expected fields are there

	rq := require.New(t)

	for _, fld := range fields {
		rq.Containsf(js, fld.fn, fld.msg)
	}

	var ar models.Authors

	ar = append(ar, a)

	js = ar.String()

	if len(js) == 0 {
		t.Error("Unable to marshal array of authors")
		t.Fail()
	}
}

const invalidUUID = "563cd207-ab16-4a46-b44e-7317b96c6ba9"
const validUUID = "b39300f0-6760-4feb-bc32-4b8682b0175d" // matches entry in testauthors.toml
const validName = "George P. Burdell"

// Test for finding an existing author
func (ms *modelSuite) Test_Author_FindByID() {
	//	ms.LoadFixture("test authors")

	id, err := uuid.FromString(validUUID)

	if err != nil {
		ms.Fail("uuid.FromString failed", err.Error())
	}

	auth := models.Author{
		ID: id,
	}

	err = auth.FindByID()

	if err != nil {
		ms.Fail("FindByID failed", err.Error())
	}

	if len(auth.Name) == 0 {
		ms.Fail("FindByID failed", "validUUID was not found in database")
	}

	if strings.Compare(auth.Name, validName) != 0 {
		ms.Fail("FindByID didn't find expected author", auth.Name)
	}

	// as long as I have a valid author, check some other functions

	if strings.Compare(auth.SelectLabel(), validName) != 0 {
		ms.Fail("unexpected SelectLabel", auth.SelectLabel())
	}

	v := auth.SelectValue()
	s, ok := v.(string)

	if ok {
		if strings.Compare(s, validUUID) != 0 {
			ms.Fail("unexpected SelectValue", s)
		}
	} else {
		ms.Fail("unexpected SelectValue", "not a string")
	}
}

// test that FindByID correctly handles NOT finding the author
func (ms *modelSuite) Test_Author_FindByID_BadID() {
	ms.LoadFixture("test authors")

	id, err := uuid.FromString(invalidUUID)

	if err != nil {
		ms.Fail("uuid.FromString failed", err.Error())
	}

	auth := models.Author{
		ID: id,
	}

	err = auth.FindByID()

	if err != nil {
		ms.Fail("FindByID failed", err.Error())
	}

	if len(auth.Name) != 0 {
		ms.Fail("FindByID succeeded with an invalid UUID", auth.Name)
	}
}
*/

func (ms *modelSuite) Test_Author_Create() {
	ms.LoadFixture("test authors")

	fmt.Println("Testing author creation.")

	auth := models.Author{
		Name: "Brand New Author",
	}

	verrs, err := ms.DB.ValidateAndCreate(&auth)

	if err != nil {
		ms.Fail("failed to create author", err.Error())
	}

	if verrs.HasAny() {
		ms.Fail("unable to validate new author", verrs.String())
	}
}

/*
func (ms *modelSuite) Test_Author_CreateInvalid() {
	ms.LoadFixture("test authors")

	auth := models.Author{}

	verrs, err := ms.DB.ValidateAndCreate(&auth)

	if err != nil {
		ms.Fail("failed to create author", err.Error())
	}

	if !verrs.HasAny() {
		ms.Fail("invalid author validated", "no name supplied")
	}
}
*/
