package webflow

import (
	"context"
	"fmt"
	"time"
)

// CollectionsService handles communication with the collections related
// methods of the Webflow API.
//
// ref: https://developers.webflow.com/#collections
type CollectionsService service

//Collection models a Webflow API collection
// ref: https://developers.webflow.com/#model16
type Collection struct {
	Id           string        `json:"_id"`
	LastUpdated  time.Time     `json:"lastUpdated"`
	CreatedOn    time.Time     `json:"createdOn"`
	Name         string        `json:"name"`
	Slug         string        `json:"slug"`
	SingularName string        `json:"singularName"`
	Fields       []interface{} `json:"fields"`
}

type ListCollectionsOptions struct {
	ListOptions
}

// ListCollections lists all the collections for the given site
// ref: https://developers.webflow.com/#list-collections
func (s *CollectionsService) ListCollections(ctx context.Context, site string, opts *ListCollectionsOptions) ([]*Collection, error) {

	u := fmt.Sprintf("/sites/%s/collections", site)

	u, err := addOptions(u, opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var collections []*Collection
	_, err = s.client.Do(ctx, req, &collections)
	if err != nil {
		return nil, err
	}

	return collections, nil
}

//GetCollection retrieves the specified collection
// ref: https://developers.webflow.com/#get-collection-with-full-schema
func (s *CollectionsService) GetCollection(ctx context.Context, collectionId string) (*Collection, error) {

	u := fmt.Sprintf("/collections/%s", collectionId)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var collection *Collection
	_, err = s.client.Do(ctx, req, &collection)
	if err != nil {
		return nil, err
	}

	return collection, nil
}
