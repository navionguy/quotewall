package actions

import "net/http"

func (as *ActionSuite) Test_QuotesResource_List() {
	res := as.HTML("/conversations/").Get()

	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Welcome to the Quote Archive")
}

func (as *ActionSuite) Test_QuotesResource_Show() {
	//as.Fail("Not Implemented!")
}

func (as *ActionSuite) Test_QuotesResource_New() {
	//as.Fail("Not Implemented!")
}

func (as *ActionSuite) Test_QuotesResource_Create() {
	//as.Fail("Not Implemented!")
}

func (as *ActionSuite) Test_QuotesResource_Edit() {
	//as.Fail("Not Implemented!")
}

func (as *ActionSuite) Test_QuotesResource_Update() {
	//as.Fail("Not Implemented!")
}

func (as *ActionSuite) Test_QuotesResource_Destroy() {
	//as.Fail("Not Implemented!")
}
