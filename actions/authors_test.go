package actions

import "github.com/navionguy/quotewall/models"

func (as *ActionSuite) Test_AuthorList() {
	as.LoadFixture("test authors")

	res := as.HTML("/authors").Get()

	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, "Test Author")
	as.Contains(body, "George P. Burdell")
	as.Contains(body, "Shari Freeman")
}

func (as *ActionSuite) Test_AuthorNew() {
	as.LoadFixture("test authors")

	res := as.HTML("/authors/new").Get()

	as.Equal(200, res.Code)
	body := res.Body.String()
	as.Contains(body, "Create a new Speaker")
}

func (as *ActionSuite) Test_Author_Create() {
	authname := "Senthil Krisnipali"
	auth := &models.Author{Name: authname}
	res := as.HTML("/authors").Post(auth)

	// if he succeeds he tries to send us back where we came
	as.Equal(201, res.Code)

	// see if I can get my Author back
	err := as.DB.First(auth)
	as.NoErrorf(err, "Test_Author_Create() db failed.")

	as.Equal(authname, auth.Name)
	as.Equal("/authors/authors", res.Location())
}
