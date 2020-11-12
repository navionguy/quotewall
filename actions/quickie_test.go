package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/google/uuid"
)

type testContext struct {
	buffalo.Context
}

type testCookies struct {
	buffalo.Cookies
}

//*github.com/gobuffalo/buffalo.Cookies
func (tc testContext) Cookies() *buffalo.Cookies {
	var tck buffalo.Cookies

	return &tck
}

func (as *ActionSuite) Test_QuickieBeforeDate() {
	res := as.HTML("/quickie?before=03/15/1997").Get()

	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "see us ever needing to change the product")

}

func (as *ActionSuite) Test_QuickieAfterDate() {
	res := as.HTML("/quickie?after=06/24/2019").Get()

	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Chris is wearing pants!")
}

func (as *ActionSuite) Test_BadQuickieAfterDate() {
	res := as.HTML("/quickie?after=FRED").Get()

	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Quote Wall Quickie")
}

func (as *ActionSuite) Test_QuickieSpeaker() {
	res := as.HTML("/quickie?speaker=Freeman").Get()

	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Shari Freeman")
}

func (as *ActionSuite) Test_QuickieAnnotation() {
	res := as.HTML("/quickie?after=03/20/2019&before=03/22/2019").Get()

	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "First timer!")
}

func Test_setDefaultQuote(t *testing.T) {
	p := setDefaultConversation()

	if strings.Compare(p.Conversation[0].Quote, "Life isn't about quotes about life.") != 0 {
		t.Fatal("setDefaultQuote didn't!")
	}
}

func Test_EncryptDecrypt(t *testing.T) {
	key, err := uuid.New().MarshalBinary()

	if err != nil {
		t.Fatal("stdlib failed!!!")
	}

	text := "Now is the time for all good men to come to the aid of their country."

	mash, err := encrypt(key, []byte(text))

	if err != nil {
		t.Fatalf("encrypt failed with %s\n", err.Error())
	}

	out, err := decrypt(key, mash)

	if strings.Compare(out, text) != 0 {
		t.Fatalf("expected %s, got %s\n", text, out)
	}
}
