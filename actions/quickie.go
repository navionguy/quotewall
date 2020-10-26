package actions

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/google/uuid"
	"github.com/navionguy/quotewall/models"
	"github.com/pkg/errors"

	"crypto"
	rand "math/rand"
	"strconv"
)

const quotewallhtml = `<!DOCTYPE html><html lang="en">
<head><meta charset="utf-8">
<meta name="dcterms.created" content="{{.Datestr}}">
<meta name="description" content="">
<meta name="keywords" content="">
<meta http-equiv="refresh" content="10" />
<title>{{.Title}}</title>
<h1><div align="center">{{.Title}}</div></h1>
</head>
<body>
<div align="center"><table><col width=100%>
<tr height="80%"><td><table width="100%">
{{$height := .QuoteShare}}
{{range $element := .Conversation}}
<tr><td ALIGN="CENTER"><p style="font-size:64px; height:{{$height}}%">{{.Quote}}</p></td></tr><tr><td/>
</tr><tr><td ALIGN="RIGHT"><font color="blue">{{.Name}}</font></td></tr>
<tr><td ALIGN="RIGHT"><font color="blue">{{.Date}}</font></td></tr>
{{end}}
</table></td></tr>
</div></body></html>`

const errorhtml = `<!DOCTYPE html><html lang="en">
<head><meta charset="utf-8">
<meta name="dcterms.created" content="{{.Datestr}}">
<meta name="description" content="">
<meta name="keywords" content="">
<meta http-equiv="refresh" content="10" />
<title>{{.Title}}</title>
<h1><div align="center">{{.ErrorMsg}}</div></h1>
</head>
<body>
</body></html>`

const ctLayout = "2006-01-02" // the date format used in my json file

var sdft time.Time // date time stamp of quote file I'm using

//Utterancestype --
type utterancestype struct {
	Name    string
	Quote   string
	Date    string
	Publish string
}

type conversationtype struct {
	Conversation []utterancestype
}

type archivedatatype struct {
	Conversations []conversationtype
}

type archivetype struct {
	Quotearchive archivedatatype
}

type pageParams struct {
	Datestr      string
	Title        string
	QuoteShare   int
	Conversation []utterancestype
	ErrorMsg     string
}

var quotes archivetype
var skips int // counter of how many times I hit quotes I shouldn't display

// these variables are used to provide a better selecting algorithm for picking
// the next quote to display

const nextQuote = "NextQuote"
const filtercookie = "FilteredList"
const nextquoteoffset = 16

type nextQuoteBlob struct {
	FrontJunk string
	NextQuote int
	BackJunk  string
}

var filterkey []byte

type filteredlist struct {
	Iterator int
	DaysOld  int
	List     []int
}

var shuffled []int
var shuffledday time.Time // once I day I should re-shuffle

// ShuffledConversations holds id of a random conversation
type ShuffledConversations struct {
	ID       uuid.UUID `json:"-" db:"id"`
	Sequence int       `json:"-" db:"sequence"`
}

// ShuffleData holds information about the shuffled conversation table
type ShuffleData struct {
	Size        int           // count of conversations in shuffled list
	ShuffledDay time.Time     // server date of last shuffle
	ServCor     time.Duration // correction to get time closer to server
}

// QuickieQuote displays a random quote
func (v ConversationsResource) QuickieQuote(c buffalo.Context) error {
	// Get the DB connection from the context
	/*	tx, ok := c.Value("tx").(*pop.Connection)
		if !ok {
			return errors.WithStack(errors.New("no transaction found"))
		}*/

	if len(filterkey) == 0 {
		// set encryption key for cookies
		var err error
		filterkey, err = uuid.New().MarshalBinary()

		if err != nil {
			filterkey = []byte("reallyweakkey")
		}
	}

	shuffle, err := getShuffleData(c)

	if err != nil {
		return c.Error(404, err)
	}

	id, err := pickQuote(c, shuffle)

	if err != nil {
		return c.Error(404, err)
	}

	conv := models.Conversation{}
	err = models.DB.Eager("Quotes.Conversation").Eager("Quotes").Eager("Quotes.Author").Eager("Quotes.Annotation").Find(&conv, id)

	if err != nil {
		return c.Error(404, err)
	}

	// prepare conversation for display
	page := prepareConv(conv, c)
	templ := template.New("quote wall")
	templ = template.Must((templ.Parse((quotewallhtml))))

	return c.Render(200, render.Func("html", func(w io.Writer, d render.Data) error {
		return templ.Execute(w, page)
	}))
}

func prepareConv(conv models.Conversation, c buffalo.Context) pageParams {
	var p pageParams

	p.Datestr = time.Now().Format("Mon Jan _2 15:04:05 2006")
	p.Title = "Quote Wall Quickie"

	p.Conversation = make([]utterancestype, len(conv.Quotes))
	for _, qt := range conv.Quotes {
		var utt utterancestype
		utt.Date = qt.SaidOn.Format("Jan 2, 2006")
		utt.Name = qt.Author.Name
		utt.Quote = qt.Phrase
		p.Conversation = append(p.Conversation, utt)
	}

	p.QuoteShare = 80 / len(p.Conversation)

	return p
}

func pickQuote(c buffalo.Context, shuffle *ShuffleData) (*uuid.UUID, error) {

	index := nextQuoteCookie(c)

	if index == 0 {
		index = rand.Intn(shuffle.Size)
	}

	var sc ShuffledConversations
	err := models.DB.RawQuery("SELECT * FROM shuffled_conversations WHERE sequence = ?", index).First(&sc)

	if err != nil {
		return nil, err
	}

	index++
	if index >= shuffle.Size {
		index = 0
	}
	saveNextQuoteCookie(index, c)

	return &sc.ID, err
}

// Get the shuffle data from the Session
// If necessary, create the shuffle data and store it
func getShuffleData(c buffalo.Context) (*ShuffleData, error) {

	shuffle, ok := c.Session().Get(&ShuffleData{}).(*ShuffleData)

	var err error
	// if the conversation list has never been shuffled, go shuffle it
	if !ok {
		// go create the shuffled table
		shuffle, err = shuffleConversations()

		if err != nil {
			return nil, err
		}
		// ShuffleData likely need to be registered since I couldn't find him
		gob.Register(&ShuffleData{})

		c.Session().Set(&ShuffleData{}, shuffle)
	}

	//make sure the shuffle happened today

	if !shuffleCurrent(*shuffle) {
		mutex.Lock()
		defer mutex.Unlock()
		shuffle, err = shuffleConversations()

		if err != nil {
			return nil, err
		}
		c.Session().Set(&ShuffleData{}, shuffle)
	}

	return shuffle, nil
}

// Shuffling the conversations is actuall done in a stored procedure.
// I call him here, then query for the number of conversations and
// the date of the last shuffle.
// WARNING!  If the database is running on a server whose clock is wrong
// or disagrees with the local clock about utc, you can go into a
// shuffle loop until they agree.

// ShuffleDate is only used to pull back the Shuffled Date comment on the table
type ShuffleDate struct {
	ObjDescription string `db:"obj_description"`
}

func shuffleConversations() (*ShuffleData, error) {
	err := models.DB.RawQuery("SELECT * FROM shuffle_deck();").Exec()

	if err != nil {
		return nil, err
	}

	count, err := models.DB.Count(models.ShuffledConversation{})

	if err != nil {
		return nil, err
	}

	var date ShuffleDate
	err = models.DB.RawQuery("SELECT OBJ_DESCRIPTION('public.shuffled_conversations'::regclass);").First(&date)

	if err != nil {
		return nil, err
	}

	date.ObjDescription = strings.Trim(date.ObjDescription, "\"")
	var t time.Time
	t, err = time.Parse("2006-01-02", date.ObjDescription)

	if err != nil {
		return nil, err
	}

	var corr *time.Duration
	corr, err = getDBTimeDiff()

	if err != nil {
		return nil, err
	}

	shuffle := &ShuffleData{
		Size:        count,
		ShuffledDay: t,
		ServCor:     *corr,
	}

	return shuffle, nil
}

// check that the shuffle happened today
func shuffleCurrent(shuffle ShuffleData) bool {
	tTime := time.Now().UTC().Add(shuffle.ServCor)

	// if DOYs match, we're done
	if tTime.YearDay() == shuffle.ShuffledDay.YearDay() {
		return true
	}

	// did my day change in the last five minutes?
	fiveMins, _ := time.ParseDuration("-5m")

	if tTime.Add(fiveMins).YearDay() == shuffle.ShuffledDay.YearDay() {
		return true
	}

	// time to re-shuffle!
	fmt.Println("Calling for a shuffle!")
	return false
}

// ServerTime is used to get the SQL servers current time in UTC
type ServerTime struct {
	UTCTime string `db:"timezone"`
}

// calculate clock difference between me and SQL server
func getDBTimeDiff() (*time.Duration, error) {
	var sTime ServerTime
	err := models.DB.RawQuery("SELECT CURRENT_TIMESTAMP AT TIME ZONE 'UTC';").First(&sTime)

	if err != nil {
		return nil, err
	}

	// drop the fraction of a second
	sTime.UTCTime = strings.Split(sTime.UTCTime, ".")[0]
	sTimeVal, err := time.Parse("2006-01-02T15:04:05", sTime.UTCTime)

	if err != nil {
		return nil, err
	}

	sCorr := sTimeVal.Sub(time.Now().UTC())

	return &sCorr, nil
}

// nextQuoteCookie()
//
// See if there is a cookie telling me where I am in the shuffled list
// of quotes.  Returning zero means there is not.
//
func nextQuoteCookie(c buffalo.Context) int {
	cookieBytes, err := c.Cookies().Get(nextQuote)

	if err != nil {
		return 0
	}

	var cookie nextQuoteBlob
	err = json.Unmarshal([]byte(cookieBytes), &cookie)

	if err != nil {
		return 0
	}

	return cookie.NextQuote
}

// saveNextQuoteCookie()
//
// Builds a blob of fairly random data and hides the next index into the
// middle of it.  This then gets signed and encrypted so it can be safely
// sent down to the browser.
//
func saveNextQuoteCookie(nextIndex int, c buffalo.Context) {
	fmt.Println("saveNextQuoteCookie!")
	var tblob nextQuoteBlob
	tblob.FrontJunk = uuid.New().String()
	tblob.BackJunk = uuid.New().String()
	tblob.NextQuote = nextIndex

	saveCookie(nextQuote, tblob, c)
}

func saveCookie(name string, jdata interface{}, c buffalo.Context) {
	jblob, _ := json.Marshal(jdata)

	value, _ := encrypt(filterkey, jblob)
	fmt.Printf("c_Value: %s\n", value)
	life, _ := time.ParseDuration("3h")
	c.Cookies().Set(name, value, life)
}

// encrypt string to base64 crypto using AES
func encrypt(key []byte, plaintext []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Printf("NewCipher failed! %s\n", err.Error())
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext)+crypto.SHA256.Size())
	// slice off the start of ciphertext for the initial value
	iv := ciphertext[:aes.BlockSize]
	rand.Read(iv)
	if _, err := io.ReadFull(bytes.NewReader(iv), iv); err != nil {
		fmt.Println("encrypt reader failed!")
		return "", err
	}

	sha := sha256.Sum256(plaintext)

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	stream.XORKeyStream(ciphertext[aes.BlockSize+len(plaintext):], sha[:crypto.SHA256.Size()])

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// decrypt from base64 to decrypted string
func decrypt(key []byte, cryptoText string) (string, error) {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return "", err
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	hv := ciphertext[:len(ciphertext)-crypto.SHA256.Size()]
	sha := sha256.Sum256(hv)

	if 0 != bytes.Compare(sha[:crypto.SHA256.Size()], ciphertext[len(ciphertext)-crypto.SHA256.Size():]) {
		err := errors.New("Signature failed verification,")
		return "", err
	}

	return fmt.Sprintf("%s", ciphertext[:len(ciphertext)-crypto.SHA256.Size()]), nil
}

// SetDefaultQuote()
//
// If the file won't load, I still want to have something to show
//
func setDefaultQuote(p pageParams) {
	quotes.Quotearchive.Conversations = make([]conversationtype, 1)
	quotes.Quotearchive.Conversations[0].Conversation = make([]utterancestype, 1)
	quotes.Quotearchive.Conversations[0].Conversation[0].Date = ""
	quotes.Quotearchive.Conversations[0].Conversation[0].Name = "Unknown"
	quotes.Quotearchive.Conversations[0].Conversation[0].Publish = "True"
	quotes.Quotearchive.Conversations[0].Conversation[0].Quote = "Life isn't about quotes about life."
	p.Title = "Çá´ÊÉá´nQ llÉM ÇÊonQ"
}

/*********************************************************************/
// below here is old stuff
// acutal
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/
/*********************************************************************/

// QuoteWallQuickie Put up a random quote, sounds deceptively easy.  But just you wait.
func QuoteWallQuickie(w http.ResponseWriter, r *http.Request) {
	var p pageParams

	p.Datestr = time.Now().Format("Mon Jan _2 15:04:05 2006")
	p.Title = "Quote Wall Quickie"

	err := loadquotedata()

	if err != nil {
		// couldn't read quote file.  Put up a fill-in quote
		setDefaultQuote(p)
	}

	checkShuffle()

	quoteindex, gotQuote := pickquote(w, r, p)

	if !gotQuote {
		return
	}

	p.Conversation = make([]utterancestype, len(quotes.Quotearchive.Conversations[quoteindex].Conversation))
	copy(p.Conversation, quotes.Quotearchive.Conversations[quoteindex].Conversation)
	p.QuoteShare = 80 / len(p.Conversation)

	templ := template.New("quote wall")
	templ = template.Must(templ.Parse(quotewallhtml))
	templ.Execute(w, p)
}

// pickquote()
//
// First, check for a filter, if that results in a list, pick from the list.
// Second, check for a next quote cookie.  If there is one, return that quote.
//
func pickquote(w http.ResponseWriter, r *http.Request, page pageParams) (index int, ValidQuote bool) {
	// Check for a current max quote age filter
	/*
		filteredList, gotFilter := maxAgeFilter(r)

		// there is a filter, go pick from it
		if gotFilter {
			index, ValidQuote = pickFromFilterList(filteredList, w, page)
			return
		}

		// if there isn't a filter in place, check for a next quote cookie

		//cookievalue, gotNextQuote := nextQuoteCookie(r)
		size := len(shuffled)

		if !gotNextQuote {
			cookievalue = rand.Intn(size)
		}

		// put down some boundry checking

		if cookievalue < 0 || cookievalue >= size {
			cookievalue = 0 // cookie had junk in it, start over
		}

		// go pick a quote, skipping the ones that are to remain "unpublished"

		for ; strings.Compare(quotes.Quotearchive.Conversations[shuffled[cookievalue]].Conversation[0].Publish, "False") == 0; cookievalue++ {
			skips = skips + 1

			// put down some boundry checking

			if cookievalue >= size {
				cookievalue = 0 // reached the end of the array, start over
			}
		}

		saveNextQuoteCookie(cookievalue+1, w)

		return shuffled[cookievalue], true*/
	return 0, false
}

// pickFromFilterList()
//
// If I have a filtered list, pick the next entry in the list to display.
//
func pickFromFilterList(curList filteredlist, w http.ResponseWriter, page pageParams) (index int, validQuote bool) {
	// if he put on a filter that results in no quotes
	// that is a special case.  I need to display something to the user
	// that communicates that the filter criteria eliminates all quotes.
	if len(curList.List) == 0 {
		page.ErrorMsg = "No quotes meet criteria."
		templ := template.New("quote error")
		templ = template.Must(templ.Parse(errorhtml))
		templ.Execute(w, page)
		http.Error(w, "No data returned. - 204", http.StatusNoContent)

		return 0, false
	}

	// At this point I have been asked for a filtered list and have created one
	// return the next quote, bump the iterator and save the filter cookie

	index = curList.List[curList.Iterator]
	curList.Iterator = curList.Iterator + 1

	if curList.Iterator >= len(curList.List) {
		curList.Iterator = 0
	}
	saveFilter(curList, w)

	return index, true
}

// maxAgeFilter()
//
// Retrieves the passed values for the max-age parameter and the
// currently active filter.  If there is a mis-match, the filtered
// list *may* have to be recalculated.  If all goes happily, the
// final list of filtered quotes, with the iterator pointing at
// the next one to display.
//
func maxAgeFilter(r *http.Request) (newlist filteredlist, valid bool) {
	//newlist.Iterator = 0
	valid = false

	// go see if user specified a "maximum age" for quotes

	maxdays, ok := maxAgeParam(r)

	if !ok {
		return newlist, false
	}

	// okay, we now have a filter turned on, was it already in place?

	newlist, valid = filterCookie(r)

	if !valid {
		// no list stored in a cookie, need to create it
		return applyAgeFilter(shuffled, maxdays), true
	}

	// make sure the current parameter matches the saved cookie

	if newlist.DaysOld != maxdays {
		// they don't, build a new list
		return applyAgeFilter(shuffled, maxdays), true
	}

	// list is good, go with it

	return newlist, true
}

// maxAgeParam()
//
// See if the request for a quote had max-age parameter
//
func maxAgeParam(r *http.Request) (maxdays int, valid bool) {
	maxage, found := r.URL.Query()["max-age"]

	if !found {
		return 0, false
	}

	var err error
	maxdays, err = strconv.Atoi(maxage[0])

	if err != nil {
		return 0, false
	}

	if maxdays < 0 {
		return 0, true
	}

	return maxdays, true
}

// filterCookie()
//
// Check the incoming http request and see if it has a filter in place already
// if it does, I create a filteredlist structure and return it.
//
func filterCookie(r *http.Request) (activefilter filteredlist, valid bool) {
	tintf, err := getCookie(r, filtercookie)

	// I'm cheating a little here.  Elsewhere I create and save an empty
	// cookie with an expired lifetime so the browser will delete my old
	// cookies.  It is possible for one of them to come back, but I don't
	// check for that case.  Instead, I rely on the fact that such a cookie
	// will not decrypt correctly, no hash at the end, so it will be caught
	// in getCookie() function.  BTW, if you ever change this code to encrypt the
	// empty cookie, the decrpt will pass, but the json unmarshal will fail.
	// Unless you convert the cookie to contain empty json, in which case,
	// you're being wierd.

	if err != nil {
		return activefilter, false
	}

	err = json.Unmarshal([]byte(tintf), &activefilter)

	if err != nil {
		return activefilter, false
	}

	return activefilter, true
}

func getCookie(r *http.Request, name string) (newdata string, err error) {
	tdata, err := r.Cookie(name)

	if err != nil {
		return newdata, err
	}

	newdata, err = decrypt(filterkey, tdata.Value)

	if err != nil {
		return newdata, err
	}

	return newdata, err
}

// applyAgeFilter()
//
// Given the input list of quotes, role through and pull out all that meet
// the max days old value.
//
func applyAgeFilter(initList []int, maxdays int) (newlist filteredlist) {
	cutoff := time.Now().AddDate(0, 0, 0-maxdays)
	newlist.Iterator = 0
	newlist.DaysOld = maxdays
	for i := 0; i < len(initList); i++ {
		date, err := time.Parse("1/2/2006", quotes.Quotearchive.Conversations[initList[i]].Conversation[0].Date)
		if err == nil {
			if date.Sub(cutoff) >= 0 && strings.Compare(quotes.Quotearchive.Conversations[initList[i]].Conversation[0].Publish, "True") == 0 {
				newlist.List = append(newlist.List, initList[i])
			}
		}
	}

	// shuffle the filtered list

	j := len(newlist.List)
	for i := 0; i < j; i++ {
		a := rand.Intn(j)
		b := rand.Intn(j)

		newlist.List[a], newlist.List[b] = newlist.List[b], newlist.List[a]
	}

	return
}

// saveFilter()
//
// Converts the filteredlist into json, then bakes it into a cookie to send
// down to the client.
//
func saveFilter(newlist filteredlist, w http.ResponseWriter) {
	sf := new(http.Cookie)
	sf.Name = filtercookie
	jvalB, _ := json.Marshal(newlist)
	sf.Value, _ = encrypt(filterkey, jvalB)
	http.SetCookie(w, sf)
}

// clearFilterCookie()
//
// I create an empty cookie and send it down to the client in the hope
// that the receiving browser will only the MaxAge value and delete
// it from the cookie store.  Even if it doesn't work, I can sleep
// well knowing I tried.
//
func clearFilterCookie(w http.ResponseWriter) {
	sf := new(http.Cookie)
	sf.Name = filtercookie
	sf.Value = "empty"
	sf.MaxAge = -1
	http.SetCookie(w, sf)
}

// loadquotedata()
//
// Checks to see if the datafile on the file system has a new date/time stamp
// than it did when I last read it.  If it does, then I *attempt* to read it
// into memory and convert it into an array of quotes.
//
func loadquotedata() error {

	return nil
}

var mutex sync.Mutex

// checkShuffle()
//
// I re-shuffle the quotes once a day just to make it interesting.
func checkShuffle() {
	if shuffledday.Day() != time.Now().Day() {
		mutex.Lock()
		defer mutex.Unlock()

		shuffle(len(quotes.Quotearchive.Conversations))
		shuffledday = time.Now()

		var err error
		filterkey, err = uuid.New().MarshalBinary()

		if err != nil {
			filterkey = []byte("reallyweakkey")
		}
	}
}

// shuffle()
//
// Rather than use rand to generate a random index into the array of quotes,
// I'm going to treat the quote array like a deck of cards.  By creating and
// array of indexs, that starts out sequientially, and then randomly swapping
// elements in the index array.  I can then just step through the array item
// by item display the quote whose index is in the shuffled index. For the
// computer science fans in the room, this is called a Fisher-Yates Shuffle.
//
// Doing so means that if you sit on the web page long enough, you will eventually
// cycle through all of the quotes, but in a random order.  And that no quote
// will display twice, until you have seen all the quotes.
//
// Also, I re-shuffle the quotes everytime the quote file is re-loaded.
// Generally, re-loading the quotes file happens when quotes are added.
//
func shuffle(size int) {
	shuffled = make([]int, size)

	for i := 0; i < size; i++ {
		shuffled[i] = i
	}

	rand.Seed(int64(time.Now().Nanosecond()))

	for i := 0; i < 10000; i++ {
		a := rand.Intn(size)
		b := rand.Intn(size)

		shuffled[a], shuffled[b] = shuffled[b], shuffled[a]
	}
}
