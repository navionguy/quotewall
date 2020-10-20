package grifts

import (
	"errors"
	"fmt"
	"strings"

	"github.com/markbates/grift/grift"
	"github.com/navionguy/quotewall/models"
)

// some string constants I use

const nameSpace = "user"
const addCmd = "add"
const emailParam = "email"
const pwdParam = "pwd"

const rmvCmd = "rmv"

var _ = grift.Namespace(nameSpace, func() {
	// "add" creates a new user in the database
	grift.Desc(addCmd, "Adds a user account for working with quotes, example: buffalo task user:add email:emailaddr pwd:initialpassword")
	grift.Add(addCmd, func(c *grift.Context) error {
		// Accpets two options
		// email:emailaddr the users email address
		// pwd:initalpassword password to set on the account, users can change this

		if len(c.Args) == 0 {
			return errors.New("no valid arguements to user:add")
		}

		u := models.User{}
		u.Email = ""
		u.Password = ""

		// look for my arguements

		for _, arg := range c.Args {
			fmt.Printf("arg = %s\n", arg)
			parts := strings.Split(arg, ":")

			if len(parts) == 2 && strings.Compare(parts[0], emailParam) == 0 {
				u.Email = parts[1]
			}

			if len(parts) == 2 && strings.Compare(parts[0], pwdParam) == 0 {
				u.Password = parts[1]
			}
		}

		if len(u.Email) == 0 || len(u.Password) == 0 {
			return errors.New("required parameter not supplied")
		}

		// okay, try to create the user
		// make sure that both password fields match

		u.PasswordConfirmation = u.Password
		verrs, err := u.Create(models.DB)

		if verrs.HasAny() {
			return errors.New("user parameters failed validation")
		}

		return err
	})

	grift.Desc(rmvCmd, "Removes a user account from the quotes archive, example: buffalo task user:rmv email:emailaddr")
	grift.Add(rmvCmd, func(c *grift.Context) error {
		// Accpets one options
		// email:emailaddr the users email address

		if len(c.Args) == 0 {
			return errors.New("no valid arguements to user:rmv")
		}

		u := &models.User{}
		u.Email = ""

		// look for my arguements

		for _, arg := range c.Args {
			fmt.Printf("arg = %s\n", arg)
			parts := strings.Split(arg, ":")

			if len(parts) == 2 && strings.Compare(parts[0], emailParam) == 0 {
				u.Email = parts[1]
			}
		}

		if len(u.Email) == 0 {
			return errors.New("required parameter not supplied")
		}

		//okay, go see if I can find that user

		err := models.DB.Where("Email = ?", u.Email).First(u)

		if err != nil {
			return err
		}

		return models.DB.Destroy(u)
	})
})
