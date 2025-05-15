package main

import (
    stdctx "context"
  "fmt"
  "lenslockd/models"
    "lenslockd/context"
)

func main() {
	ctx := stdctx.Background()

	user := models.User{
		Email: "jon@calhoun.io",
	}

    ctx = context.WithUser(ctx, &user)
    retrievedUser := context.User(ctx)
fmt.Println(retrievedUser.Email)
}