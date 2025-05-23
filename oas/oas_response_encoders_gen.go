// Code generated by ogen, DO NOT EDIT.

package oas

import (
	"net/http"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	ht "github.com/ogen-go/ogen/http"
)

func encodeAddTeamMemberResponse(response *TeamMemberResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	response.Encode(e)
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeChangeUserPasswordResponse(response *ChangeUserPasswordNoContent, w http.ResponseWriter) error {
	w.WriteHeader(204)

	return nil
}

func encodeCreatePersonalAccessTokenResponse(response *PersonalAccessTokenCreationResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	response.Encode(e)
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeCreateRepositoryResponse(response *RepositoryResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	response.Encode(e)
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeCreateTeamResponse(response *TeamResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	response.Encode(e)
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeCreateUserResponse(response *UserResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	response.Encode(e)
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeDeletePersonalAccessTokenResponse(response *DeletePersonalAccessTokenNoContent, w http.ResponseWriter) error {
	w.WriteHeader(204)

	return nil
}

func encodeDeleteRepositoryResponse(response *DeleteRepositoryNoContent, w http.ResponseWriter) error {
	w.WriteHeader(204)

	return nil
}

func encodeDeleteTeamResponse(response *DeleteTeamNoContent, w http.ResponseWriter) error {
	w.WriteHeader(204)

	return nil
}

func encodeDeleteUserResponse(response *DeleteUserNoContent, w http.ResponseWriter) error {
	w.WriteHeader(204)

	return nil
}

func encodeGetPersonalAccessTokenResponse(response *PersonalAccessTokenResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	response.Encode(e)
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeGetRepositoryResponse(response *RepositoryResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	response.Encode(e)
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeGetTeamResponse(response *TeamResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	response.Encode(e)
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeGetUserResponse(response *UserResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	response.Encode(e)
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeListPersonalAccessTokensResponse(response []PersonalAccessTokenResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	e.ArrStart()
	for _, elem := range response {
		elem.Encode(e)
	}
	e.ArrEnd()
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeListRepositoriesResponse(response []RepositoryResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	e.ArrStart()
	for _, elem := range response {
		elem.Encode(e)
	}
	e.ArrEnd()
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeListTeamMembersResponse(response []TeamMemberResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	e.ArrStart()
	for _, elem := range response {
		elem.Encode(e)
	}
	e.ArrEnd()
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeListTeamsResponse(response []TeamResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	e.ArrStart()
	for _, elem := range response {
		elem.Encode(e)
	}
	e.ArrEnd()
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeListUsersResponse(response []UserResponse, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	e := new(jx.Encoder)
	e.ArrStart()
	for _, elem := range response {
		elem.Encode(e)
	}
	e.ArrEnd()
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeRemoveTeamMemberResponse(response *RemoveTeamMemberNoContent, w http.ResponseWriter) error {
	w.WriteHeader(204)

	return nil
}

func encodeErrorResponse(response *ErrorStatusCode, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	code := response.StatusCode
	if code == 0 {
		// Set default status code.
		code = http.StatusOK
	}
	w.WriteHeader(code)

	e := new(jx.Encoder)
	response.Response.Encode(e)
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	if code >= http.StatusInternalServerError {
		return errors.Wrapf(ht.ErrInternalServerErrorResponse, "code: %d, message: %s", code, http.StatusText(code))
	}
	return nil

}
