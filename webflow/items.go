package webflow

import (
	"context"
	"fmt"
)

// ItemsService handles communication with the items related
// methods of the Webflow API.
//
// ref: https://developers.webflow.com/#items
type ItemsService service

//ListItemsOptions defines the available list options
type ListItemsOptions struct {
	ListOptions
}

// Item is a dynamic object. Its fields depend on the fields a user creates for a collection
// ref: https://developers.webflow.com/#item-model
type Item map[string]interface{}

// ItemResponse models a response returned by the Items Webflow API
type ItemResponse struct {
	Items []Item
	PageInfo
}

//ListItems lists all the items of the given collection
// ref: https://developers.webflow.com/#get-all-items-for-a-collection
func (s *ItemsService) ListItems(ctx context.Context, collectionId string, opts *ListItemsOptions) (*ItemResponse, error) {
	u := fmt.Sprintf("/collections/%s/items", collectionId)

	u, err := addOptions(u, opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var items *ItemResponse
	_, err = s.client.Do(ctx, req, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

//GetItem retrieves the specified item
// ref: https://developers.webflow.com/#get-single-item
func (s *ItemsService) GetItem(ctx context.Context, collectionId string, itemId string) (*ItemResponse, error) {

	u := fmt.Sprintf("/collections/%s/items/%s", collectionId, itemId)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var item *ItemResponse
	_, err = s.client.Do(ctx, req, &item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *ItemsService) createItem(ctx context.Context, collectionId string, item Item, live bool) (Item, error) {

	u := fmt.Sprintf("/collections/%s/items", collectionId)

	// create the item and publish it. ref: https://developers.webflow.com/#create-new-live-collection-item
	if live {
		u += "?live=true"
	}

	req, err := s.client.NewRequest("POST", u, item)
	if err != nil {
		return nil, err
	}

	var created Item
	_, err = s.client.Do(ctx, req, &created)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// CreateItem creates a new item inside the given collection
// ref: https://developers.webflow.com/#create-new-collection-item
func (s *ItemsService) CreateItem(ctx context.Context, collectionId string, item Item) (Item, error) {
	return s.createItem(ctx, collectionId, item, false)
}

// CreateItemLive creates a new item inside the given collection and publish it
// ref: https://developers.webflow.com/#create-new-live-collection-item
func (s *ItemsService) CreateItemLive(ctx context.Context, collectionId string, item Item) (Item, error) {
	return s.createItem(ctx, collectionId, item, true)
}

func (s *ItemsService) updateItem(ctx context.Context, collectionId string, itemId string, item Item, live bool) (Item, error) {

	u := fmt.Sprintf("/collections/%s/items/%s", collectionId, itemId)

	// create the item and publish it. ref: https://developers.webflow.com/#update-live-collection-item
	if live {
		u += "?live=true"
	}

	req, err := s.client.NewRequest("PUT", u, item)
	if err != nil {
		return nil, err
	}

	var updated Item
	_, err = s.client.Do(ctx, req, &updated)
	if err != nil {
		return nil, err
	}

	return updated, nil

}

// UpdateItem updates (replaces) the collection item with the given fields
// ref: https://developers.webflow.com/#update-collection-item
func (s *ItemsService) UpdateItem(ctx context.Context, collectionId string, itemId string, item Item) (Item, error) {
	return s.updateItem(ctx, collectionId, itemId, item, false)

}

// UpdateItemLive updates (replaces) the collection item with the given fields and publishes the changes
// ref: https://developers.webflow.com/#update-live-collection-item
func (s *ItemsService) UpdateItemLive(ctx context.Context, collectionId string, itemId string, item Item) (Item, error) {
	return s.updateItem(ctx, collectionId, itemId, item, true)
}

func (s *ItemsService) patchItem(ctx context.Context, collectionId string, itemId string, item Item, live bool) (Item, error) {

	u := fmt.Sprintf("/collections/%s/items/%s", collectionId, itemId)

	// create the item and publish it. ref: https://developers.webflow.com/#update-live-collection-item
	if live {
		u += "?live=true"
	}

	req, err := s.client.NewRequest("PATCH", u, item)
	if err != nil {
		return nil, err
	}

	var updated Item
	_, err = s.client.Do(ctx, req, &updated)
	if err != nil {
		return nil, err
	}

	return updated, nil

}

// PatchItem updates the selected fields of the given item
// ref: https://developers.webflow.com/#patch-collection-item
func (s *ItemsService) PatchItem(ctx context.Context, collectionId string, itemId string, item Item) (Item, error) {
	return s.patchItem(ctx, collectionId, itemId, item, false)

}

// PatchItemLive updates the selected fields of the given item and publishes the changes
// ref: https://developers.webflow.com/#patch-live-collection-item
func (s *ItemsService) PatchItemLive(ctx context.Context, collectionId string, itemId string, item Item) (Item, error) {
	return s.patchItem(ctx, collectionId, itemId, item, true)
}

type DeleteItemResponse struct {
	Deleted int `json:"deleted"`
}

func (s *ItemsService) DeleteItem(ctx context.Context, collectionId string, itemId string) (*DeleteItemResponse, error) {
	u := fmt.Sprintf("/collections/%s/items/%s", collectionId, itemId)

	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	var deleted *DeleteItemResponse
	_, err = s.client.Do(ctx, req, &deleted)
	if err != nil {
		return nil, err
	}

	return deleted, nil
}
