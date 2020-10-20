package grifts

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/markbates/grift/grift"
)

// some string constants I use
const vParam = "v"
const srcParam = "src"
const destParam = "dest"
const seedCmd = "seed"
const exportCmd = "export"

var _ = grift.Namespace("db", func() {

	// "seed" is used to load the contents of a json file.  Look at the source in
	// loader.go for the format of the json.
	grift.Desc(seedCmd, "Seeds the QuoteArchive database from a json file, example: buffalo task db:seed src:filename v:[0-4]")
	grift.Add(seedCmd, func(c *grift.Context) error {
		// Add DB seeding stuff here

		// Accpets two options
		// src:filename (reqd) example: 'buffalo task db:seed src:file.json'
		// v:level (optional) example: 'buffalo task db:seed v:4 src:file.json'
		// v default is zero, max is 4

		if len(c.Args) == 0 {
			return errors.New("no valid arguement to seed")
		}

		// look for my aruguement

		for _, arg := range c.Args {
			fmt.Printf("arg = %s\n", arg)
			parts := strings.Split(arg, ":")

			if len(parts) == 2 && strings.Compare(parts[0], vParam) == 0 {
				nv, err := strconv.Atoi(parts[1])
				if err != nil {
					return err
				}

				setVerbosity(nv)
				tracemsg(fmt.Sprintf("verbosity set to %d", verbosity), 1)
			}

			if len(parts) == 2 && strings.Compare(parts[0], srcParam) == 0 {
				tracemsg(fmt.Sprintf("seeding from file %s", parts[1]), 1)
				defer seedQuoteDB(parts[1])
			}
		}
		return nil
	})

	grift.Desc(exportCmd, "Exports the QuoteArchive to a json file, example: buffalo task db:export dest:filename")

	grift.Add(exportCmd, func(c *grift.Context) error {
		// Drop the archive into a json for the online quotewall

		for _, arg := range c.Args {
			fmt.Printf("arg = %s\n", arg)
			parts := strings.Split(arg, ":")

			if len(parts) == 2 && strings.Compare(parts[0], destParam) == 0 {
				tracemsg(fmt.Sprintf("exporting to file %s", parts[1]), 1)
				exportArchive(parts[1])
			}
		}

		return nil
	})

})
