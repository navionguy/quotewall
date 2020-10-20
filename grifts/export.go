package grifts

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/navionguy/quotewall/models"
)

// exportArchive loads the entire database of quotes and then converts them into
// the old style array of conversations, then marshales that into json

func exportArchive(dest string) error {
	f, err := os.Create(dest)

	if err != nil {
		fmt.Printf("unable to create json file %s, error = %s\n", dest, err.Error())
		return err
	}

	conversations := []models.Conversation{}

	// Load the whole database

	err = models.DB.Eager("Quotes.Conversation").Eager("Quotes").Eager("Quotes.Author").Eager("Quotes.Annotation").All(&conversations)

	if err != nil {
		fmt.Printf("query db failed, %s\n", err.Error())
		return err
	}

	defer f.Close()

	// allocate the new array

	var arc archivetype

	for _, cv := range conversations {
		var nc conversationtype

		for _, qt := range cv.Quotes {
			note := ""

			if qt.Annotation != nil {
				note = qt.Annotation.Note
			}

			nq := utterancestype{
				Name:       qt.Author.Name,
				Quote:      qt.Phrase,
				Date:       CustomTime{Time: qt.SaidOn},
				Publish:    strconv.FormatBool(qt.Publish),
				Annotation: note,
			}

			nc.Conversation = append(nc.Conversation, nq)
		}

		arc.Quotearchive.Conversations = append(arc.Quotearchive.Conversations, nc)
	}

	raw, err := json.MarshalIndent(arc, " ", "   ")

	if err != nil {
		fmt.Printf("marshal failed with %s\n", err.Error())
		return err
	}

	_, err = f.Write(raw)

	if err != nil {
		fmt.Printf("write failed with %s\n", err.Error())
	}

	return err
}
