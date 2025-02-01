package main

import (
	"net/http"

	"github.com/Althaf66/go-backend-social-app/internal/store"
)

// getUserFeedHandler godoc
//
//	@Summary		Fetches the user feed
//	@Description	Fetches the user feed
//	@Tags			feed
//	@Accept			json
//	@Produce		json
//	@Param			since	query		string	false	"Since"
//	@Param			until	query		string	false	"Until"
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			tags	query		string	false	"Tags"
//	@Param			search	query		string	false	"Search"
//
// @Success		200		{object}	[]store.PostMetadata
//
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/feed [get]
func (app *application) getUserFeed(w http.ResponseWriter, r *http.Request) {
	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
		Tags:   []string{},
		Search: "",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	err = Validate.Struct(fq)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	feed, err := app.store.Posts.GetUserFeed(r.Context(), int64(177), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	err = JsonResponse(w, http.StatusOK, feed)
	if err != nil {
		app.internalServerError(w, r, err)
	}
}
