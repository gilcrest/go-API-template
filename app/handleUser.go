package app

import (
	"encoding/json"
	"net/http"

	"github.com/gilcrest/go-API-template/datastore"
	"github.com/gilcrest/go-API-template/lib/errors"
	"github.com/gilcrest/go-API-template/lib/usr"
	"github.com/gilcrest/httplog"
)

// handleUserCreate creates a user in the database
func (s *server) handleUserCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		// request is the expected service request fields
		type request struct {
			Username     string `json:"username"`
			Password     string `json:"password"`
			MobileID     string `json:"mobile_id"`
			Email        string `json:"email"`
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			UpdateUserID string `json:"udpate_user_id"`
		}

		// response is the expected service response fields
		type response struct {
			Username       string         `json:"username"`
			MobileID       string         `json:"mobile_id"`
			Email          string         `json:"email"`
			FirstName      string         `json:"first_name"`
			LastName       string         `json:"last_name"`
			UpdateUserID   string         `json:"update_user_id"`
			UpdateUnixTime int64          `json:"created"`
			Audit          *httplog.Audit `json:"audit"`
		}

		// Get logger instance
		log := s.logger

		// retrieve the context from the http.Request
		ctx := req.Context()

		var err error

		// Declare cur as an instance of createUserRequest
		// Decode JSON HTTP request body into a Decoder type
		//  and unmarshal that into cur
		rqst := new(request)
		err = json.NewDecoder(req.Body).Decode(&rqst)
		defer req.Body.Close()
		if err != nil {
			err = HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Invalid,
				Err:  err,
			}
			httpError(w, err)
			return
		}

		// declare a new instance of usr.User
		usr := new(usr.User)

		err = usr.SetUsername(rqst.Username)
		if err != nil {
			err = HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Validation,
				Err:  err,
			}
			httpError(w, err)
			return
		}
		err = usr.SetPassword(ctx, rqst.Password)
		if err != nil {
			err = HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Validation,
				Err:  err,
			}
			httpError(w, err)
			return
		}
		err = usr.SetMobileID(rqst.MobileID)
		if err != nil {
			err = HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Validation,
				Err:  err,
			}
			httpError(w, err)
			return
		}
		err = usr.SetEmail(rqst.Email)
		if err != nil {
			err = HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Validation,
				Err:  err,
			}
			httpError(w, err)
			return
		}
		err = usr.SetFirstName(rqst.FirstName)
		if err != nil {
			err = HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Validation,
				Err:  err,
			}
			httpError(w, err)
			return
		}
		err = usr.SetLastName(rqst.LastName)
		if err != nil {
			err = HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Validation,
				Err:  err,
			}
			httpError(w, err)
			return
		}
		err = usr.SetUpdateClientID("TBD")
		if err != nil {
			err = HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Validation,
				Err:  err,
			}
			httpError(w, err)
			return
		}
		err = usr.SetUpdateUserID(rqst.UpdateUserID)
		if err != nil {
			err = HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Validation,
				Err:  err,
			}
			httpError(w, err)
			return
		}

		// Call the create method of the User object to validate data and write to db
		err = usr.Create(ctx, log)
		if err != nil {
			err = HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Validation,
				Err:  err,
			}
			httpError(w, err)
			return
		}

		// get a new DB Tx
		tx, err := s.ds.BeginTx(ctx, nil, datastore.AppDB)
		if err != nil {
			err = HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Database,
				Err:  err,
			}
			httpError(w, err)
			return
		}

		// Call the create method of the User object to write
		// to the database
		err = usr.CreateDB(ctx, log, tx)
		if err != nil {
			err = HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Database,
				Err:  err,
			}
			httpError(w, err)
			return
		}

		if !usr.UpdateTimestamp().IsZero() {
			err := tx.Commit()
			if err != nil {
				err = HTTPErr{
					Code: http.StatusBadRequest,
					Kind: errors.Database,
					Err:  err,
				}
				httpError(w, err)
				return
			}
		} else {
			err = tx.Rollback()
			if err != nil {
				err = HTTPErr{
					Code: http.StatusBadRequest,
					Kind: errors.Database,
					Err:  err,
				}
				httpError(w, err)
				return
			}
		}

		aopt := new(httplog.AuditOpts)
		aopt.Host = true
		aopt.Port = true
		aopt.Path = true
		aopt.Query = true

		// If we successfully committed the db transaction, we can consider this
		// transaction successful and return a response with the response body
		aud, err := httplog.NewAudit(ctx, aopt)
		if err != nil {
			err = HTTPErr{
				Code: http.StatusInternalServerError,
				Kind: errors.Other,
				Err:  err,
			}
			httpError(w, err)
			return
		}

		resp := new(response)
		resp.Audit = aud
		resp.Username = usr.Username()
		resp.MobileID = usr.MobileID()
		resp.Email = usr.Email()
		resp.FirstName = usr.FirstName()
		resp.LastName = usr.LastName()
		resp.UpdateUserID = usr.UpdateUserID()
		resp.UpdateUnixTime = usr.UpdateTimestamp().Unix()

		// Encode usr struct to JSON for the response body
		json.NewEncoder(w).Encode(*resp)
		if err != nil {
			err = HTTPErr{
				Code: http.StatusInternalServerError,
				Kind: errors.Other,
				Err:  err,
			}
			httpError(w, err)
			return
		}
	}
}