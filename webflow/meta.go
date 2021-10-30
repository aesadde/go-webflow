package webflow

import (
	"context"
)

// MetaService handles communication with the meta related
// methods of the Webflow API.
//
// Webflow API docs: https://developers.webflow.com/#meta
type MetaService service

func (s *MetaService) GetCurrentAuthorizationInfo(ctx context.Context) (interface{}, error) {

	u := "/info"
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var body interface{}
	_, err = s.client.Do(ctx, req, &body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *MetaService) GetCurrentAuthorizedUser(ctx context.Context) (interface{}, error) {

	u := "/user"
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var body interface{}
	_, err = s.client.Do(ctx, req, &body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
