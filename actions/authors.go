package actions

import (
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"

	"github.com/navionguy/quotewall/models"
	"github.com/pkg/errors"
)

// AuthorsResource is the resource for the Conversation model
type AuthorsResource struct {
	buffalo.Resource
}

// List all the known authors
func (v AuthorsResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".

	// Get all the authors names and their quote count
	cq := tx.PaginateFromParams(c.Params()).RawQuery("SELECT authors.id, authors.name, COUNT(DISTINCT quotes.id) FROM authors LEFT JOIN quotes ON quotes.author_id = authors.id GROUP BY authors.id ORDER BY authors.name")

	authorCredits := &models.AuthorCredits{}

	if err := cq.All(authorCredits); err != nil {
		return errors.WithStack(err)
	}

	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", cq.Paginator)

	return c.Render(200, r.Auto(c, authorCredits))
}

// New author about to be entered
func (v AuthorsResource) New(c buffalo.Context) error {
	spkr := &models.Author{}

	c.Set("author", spkr)
	c.Set("cvj", "")

	return c.Render(200, r.HTML("authors/new.html"))
}

// Create default implementation.
func (v AuthorsResource) Create(c buffalo.Context) error {

	speaker := &models.Author{}

	// Bind quote to the html form elements
	if err := c.Bind(speaker); err != nil {
		return errors.WithStack(err)
	}

	fmt.Printf("new speaker %s, %s\n", speaker.Name, speaker.ID.String())

	tx, ok := c.Value("tx").(*pop.Connection)

	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	if nil != speaker.FindByName() {
		verrs, err := tx.ValidateAndCreate(speaker)

		if err != nil {
			return err
		}

		if verrs.HasAny() {
			c.Set("author", speaker)
			c.Set("gotoPage", "new")

			// set the verification errors into the context and send back the author
			c.Set("errors", verrs)

			return c.Render(422, r.Auto(c, speaker))
		}
		c.Flash().Add("success", "Speaker created successfully!")
	}

	cvjson := c.Request().Form.Get("cvjson")

	fmt.Printf("json length %d, %s\n", len(cvjson), cvjson)

	if len(cvjson) == 0 {
		return c.Redirect(201, "authors")
	}

	// put the conversation back into the form
	conv := models.Conversation{}
	err := conv.UnmarshalConversation(cvjson)

	if err != nil {
		return errors.WithStack(err)
	}

	authors := []models.Author{}

	// Retrieve all Authors from the DB
	if err := tx.Order("name").All(&authors); err != nil {
		return errors.WithStack(err)
	}

	c.Set("conversation", conv)
	c.Set("authors", authors)
	c.Set("cvj", cvjson)

	return c.Render(200, r.HTML("conversations/new"))
}

// Edit renders a edit form for an Author. This function is
// mapped to the path GET /authors/{author_id}/edit
func (v AuthorsResource) Edit(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	spkr := models.Author{}

	if err := tx.Find(&spkr, c.Param("author_id")); err != nil {
		return c.Error(404, err)
	}

	c.Set("author", spkr)
	c.Set("cvj", "")

	return c.Render(200, r.Auto(c, spkr))
}

/*
func (v AuthorsResource) unMarshalConversation(c buffalo.Context) (*models.Conversation, bool) {
	cvv := c.Request().Form.Get("cvjson")
	conv := &models.Conversation{}
	err := conv.UnmarshalConversation(cvv)

	if err != nil {
		return nil, false
	}

	return conv, true
}*/

// Update changes an Author in the DB. This function is mapped to
// the path PUT /authors/{author_id}
func (v AuthorsResource) Update(c buffalo.Context) error {
	fmt.Println("authors PUT")
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	speaker := &models.Author{}

	// Bind quote to the html form elements
	if err := c.Bind(speaker); err != nil {
		return errors.WithStack(err)
	}

	fmt.Printf("modified speaker %s, %s\n", speaker.Name, c.Param("author_id"))
	verrs, err := tx.ValidateAndUpdate(speaker)

	if err != nil {
		return err
	}

	if verrs.HasAny() {
		c.Set("author", speaker)
		c.Set("gotoPage", "edit")

		// set the verification errors into the context and send back the author
		c.Set("errors", verrs)

		return c.Render(422, r.Auto(c, speaker))
	}
	c.Flash().Add("success", "Speaker updated successfully!")

	fmt.Println("Moving back to author list.")

	return c.Redirect(302, fmt.Sprintf("/author//%%7B%s%%7D/", speaker.ID.String()))

	//return c.Render(201, r.Auto(c, speaker))

	//rc := v.List(c)

	//fmt.Println("About to return")
	//return rc
}
