package actions

import (
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

func (as *ActionSuite) Test_SaveNextCookieQuote() {
	as.HTML("/conversations/quickie?max-age=550").Get()

}

func Test_setDefaultQuote(t *testing.T) {
	var p pageParams
	setDefaultConversation(p)

	if strings.Compare(quotes.Quotearchive.Conversations[0].Conversation[0].Quote, "Life isn't about quotes about life.") != 0 {
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
