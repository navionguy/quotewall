package actions

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strconv"
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

const ageParam = "max-age" // only pull quotes that happened in the last n days
const endRange = "before"  // only pull quotes that happened before specified date
const startRange = "after" // only pull quotes that happened after specified date
const speaker = "speaker"  // only pull quotes that involved the specified speaker

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

type filterSet map[string]string

var quotes archivetype
var skips int // counter of how many times I hit quotes I shouldn't display

// these variables are used to provide a better selecting algorithm for picking
// the next quote to display

const nextQuote = "NextQuote"
const filtercookie = "FilteredList"
const nextquoteoffset = 16

type nextQuoteBlob struct {
	FrontJunk    string
	NextQuote    int
	FilteredList []int
	ParamHash    []byte
	BackJunk     string
}

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

type quickieRequest struct {
	c          buffalo.Context
	rqParams   map[string]string
	paramsHash []byte
	paramsChgd bool
	quoteID    *uuid.UUID
}

var curShuffle *ShuffleData
var filterKey []byte

// QuickieQuote displays a random quote
func (v ConversationsResource) QuickieQuote(c buffalo.Context) error {
	rq := newRequest(c)

	err := rq.getShuffleData()

	if err != nil {
		return c.Error(404, err)
	}

	err = rq.pickQuote()

	if err != nil {
		return c.Error(404, err)
	}

	conv := models.Conversation{}
	err = models.DB.Eager("Quotes.Conversation").Eager("Quotes").Eager("Quotes.Author").Eager("Quotes.Annotation").Find(&conv, rq.quoteID)

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

// initialize a quickieRequest object
func newRequest(c buffalo.Context) *quickieRequest {
	var rq quickieRequest
	rq.c = c
	rq.rqParams = make(map[string]string)

	// ToDo retrieve the filterkey from the session

	if len(filterKey) == 0 {
		// set encryption key for cookies
		var err error
		filterKey, err = uuid.New().MarshalBinary()

		if err != nil {
			filterKey = []byte("reallyweakkey")
		}
	}

	return &rq
}

func prepareConv(conv models.Conversation, c buffalo.Context) pageParams {
	var p pageParams

	p.Datestr = time.Now().Format("Mon Jan _2 15:04:05 2006")
	p.Title = "Quote Wall Quickie"

	for _, qt := range conv.Quotes {
		var utt utterancestype
		utt.Date = qt.SaidOn.Format("Jan 2, 2006")
		utt.Name = qt.Author.Name
		utt.Quote = qt.Phrase
		p.Conversation = append(p.Conversation, utt)
	}

	p.QuoteShare = 80 / len(p.Conversation)
	fmt.Printf("QuoteShare = %d\n", p.QuoteShare)

	return p
}

func (rq *quickieRequest) pickQuote() error {
	// see if there is a "nextQuote" on the request
	index := rq.nextQuoteCookie()

	// just give him a random start point
	if index == 0 {
		index = rand.Intn(curShuffle.Size)
		var blob nextQuoteBlob
		blob.NextQuote = index + 1
		rq.saveNextQuoteCookie(&blob)
	}

	var sc ShuffledConversations
	err := models.DB.RawQuery("SELECT * FROM shuffled_conversations WHERE sequence = ?", index).First(&sc)

	if err != nil {
		return err
	}

	rq.quoteID = &sc.ID

	return nil
}

func (rq *quickieRequest) chkParams(blob *nextQuoteBlob) error {
	// go check for any parameters on the request
	err := rq.checkForFilters()

	if err != nil {
		return err
	}

	// if no parameters, nothing more for me to do
	if (len(rq.rqParams) == 0) || !rq.paramsChgd {
		return nil
	}

	// let's try to apply the parameters and build a query

	qry := "SELECT s.sequence FROM authors a JOIN quotes q ON a.id = q.author_id JOIN shuffled_conversations s ON q.conversation_id = s.id"

	first := true
	for _, cond := range rq.rqParams {
		if first {
			qry = qry + fmt.Sprintf(" WHERE %s", cond)
		} else {
			qry = qry + fmt.Sprintf(" AND %s", cond)
		}
		first = false
	}
	qry = qry + ";"
	fmt.Printf("Query = %s\n", qry)

	//var blob nextQuoteBlob

	var filteredConvs []ShuffledConversations

	err = models.DB.RawQuery(qry).All(&filteredConvs)

	if err != nil {
		return err
	}

	blob.FilteredList = blob.FilteredList[:0]
	for _, fil := range filteredConvs {
		blob.FilteredList = append(blob.FilteredList, fil.Sequence)
	}

	blob.NextQuote = rand.Intn(len(blob.FilteredList))

	return nil
}

// I support letting the users apply the following filters:
// 		max-age=n  			: only pull quotes that happened in the last n days (saved as 'startRange')
//		before=date	: only pull quotes that happened before specified date
//		after=date		: only pull quotes that happened after specified date
//		speaker=name		: only pull quotes that involved the specified speaker
//								speaker name is in a 'LIKE' clause so partials will match
//								name of Sha would return both "Shari Freeman" quotes and "Mitesh Shah" quotes

func (rq *quickieRequest) checkForFilters() error {
	// clear my param map in case it has changed
	for _, p := range rq.rqParams {
		rq.rqParams[p] = ""
	}

	val, ok := rq.checkNumericFilter(ageParam)

	if ok {
		d, _ := time.ParseDuration(fmt.Sprintf("%dh", val*24))
		dt := time.Now().Add(-d)
		rq.rqParams[startRange] = fmt.Sprintf("q.saidon > '%s'", dt.Format("01/02/2006"))
	}

	hash := sha256.Sum256([]byte(rq.rqParams[startRange] + rq.rqParams[endRange] + rq.rqParams[speaker]))

	if !bytes.Equal(rq.paramsHash, hash[:]) {
		rq.paramsChgd = true
		rq.paramsHash = hash[:]
	}

	return nil
}

func (rq *quickieRequest) checkNumericFilter(name string) (int, bool) {
	strVal, ok := rq.c.Request().URL.Query()[name]

	if !ok {
		return 0, false
	}

	val, err := strconv.Atoi(strVal[0])

	if err != nil {
		return 0, false
	}

	return val, true
}

var mutex sync.Mutex

// If necessary, create the shuffle data and store it
func (rq *quickieRequest) getShuffleData() error {

	var err error

	// if the conversation list has never been shuffled, go shuffle it
	if curShuffle == nil {
		// go create the shuffled table
		err = rq.shuffleConversations()
	}

	if err != nil {
		return err
	}

	//make sure the shuffle happened today
	mutex.Lock()
	defer mutex.Unlock()

	if curShuffle.ShuffledDay.Day() != time.Now().Day() {
		err = rq.shuffleConversations()

		if err != nil {
			return err
		}
	}

	return err
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

func (rq *quickieRequest) shuffleConversations() error {
	err := models.DB.RawQuery("SELECT * FROM shuffle_deck();").Exec()

	if err != nil {
		return err
	}

	count, err := models.DB.Count(ShuffledConversations{})

	if err != nil {
		return err
	}

	var date ShuffleDate
	err = models.DB.RawQuery("SELECT OBJ_DESCRIPTION('public.shuffled_conversations'::regclass);").First(&date)

	if err != nil {
		return err
	}

	date.ObjDescription = strings.Trim(date.ObjDescription, "\"")
	var t time.Time
	t, err = time.Parse("2006-01-02", date.ObjDescription)

	if err != nil {
		return err
	}

	var corr *time.Duration
	corr, err = getDBTimeDiff()

	if err != nil {
		return err
	}

	curShuffle = &ShuffleData{
		Size:        count,
		ShuffledDay: t,
		ServCor:     *corr,
	}

	return nil
}

// check that the shuffle happened today
func (rq *quickieRequest) shuffleCurrent() bool {
	tTime := time.Now().UTC().Add(curShuffle.ServCor)

	// if DOYs match, we're done
	if tTime.YearDay() == curShuffle.ShuffledDay.YearDay() {
		return true
	}

	// did my day change in the last five minutes?
	fiveMins, _ := time.ParseDuration("-5m")

	if tTime.Add(fiveMins).YearDay() == curShuffle.ShuffledDay.YearDay() {
		return true
	}

	// time to re-shuffle!
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
// if there is one, return the index he holds
// and save a new cookie for the next time
//
func (rq *quickieRequest) nextQuoteCookie() int {
	cookieBytes, err := rq.c.Cookies().Get(nextQuote)

	var cookie nextQuoteBlob
	if err == nil {
		cookieJSON, err := decrypt(filterKey, cookieBytes)

		if err != nil {
			return 0
		}

		err = json.Unmarshal([]byte(cookieJSON), &cookie)

		if err != nil {
			return 0
		}
	}

	err = rq.chkParams(&cookie)

	if len(cookie.FilteredList) == 0 {
		return cookie.nextShuffledQuote(rq)
	}

	return cookie.nextFilteredQuote(rq)
}

// iterate over the shuffled table and return the index to display
func (ck *nextQuoteBlob) nextShuffledQuote(rq *quickieRequest) int {
	ind := ck.NextQuote
	ck.NextQuote = ck.NextQuote + 1

	if ck.NextQuote >= curShuffle.Size {
		ck.NextQuote = 0
	}
	rq.saveNextQuoteCookie(ck)

	return ind
}

// iterate over the users filtered list of quotes and return the index to display
func (ck *nextQuoteBlob) nextFilteredQuote(rq *quickieRequest) int {
	ind := ck.FilteredList[ck.NextQuote]
	ck.NextQuote = ck.NextQuote + 1

	if ck.NextQuote >= len(ck.FilteredList) {
		ck.NextQuote = 0
	}
	rq.saveNextQuoteCookie(ck)

	return ind
}

// saveNextQuoteCookie()
//
// Builds a blob of fairly random data and hides the next index into the
// middle of it.  This then gets signed and encrypted so it can be safely
// sent down to the browser.
//
func (rq *quickieRequest) saveNextQuoteCookie(tblob *nextQuoteBlob) {
	tblob.FrontJunk = uuid.New().String()
	tblob.BackJunk = uuid.New().String()

	rq.saveCookie(nextQuote, tblob)
}

func (rq *quickieRequest) saveCookie(name string, jdata interface{}) {
	jblob, _ := json.Marshal(jdata)

	value, _ := encrypt(filterKey, jblob)
	life, _ := time.ParseDuration("3h")
	rq.c.Cookies().Set(name, value, life)
}

// encrypt string to base64 crypto using AES
func encrypt(key []byte, plaintext []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext)+crypto.SHA256.Size())
	// slice off the start of ciphertext for the initial value
	iv := ciphertext[:aes.BlockSize]
	rand.Read(iv)
	if _, err := io.ReadFull(bytes.NewReader(iv), iv); err != nil {
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
