package webflow

import (
	"context"
	"fmt"
	"time"
)

// SitesService handles communication with the meta related
// methods of the Webflow API.
//
// ref: https://developers.webflow.com/#sites
type SitesService service

// Site models a Webflow API site
// ref: https://developers.webflow.com/#model
type Site struct {
	Id            string    `json:"_id"`
	CreatedOn     time.Time `json:"createdOn"`
	Name          string    `json:"name"`
	ShortName     string    `json:"shortName"`
	LastPublished time.Time `json:"lastPublished"`
	PreviewUrl    string    `json:"previewUrl"`
	Timezone      string    `json:"timezone"`
	Database      string    `json:"database"`
}

type ListSitesOptions struct {
	ListOptions
}

// ListSites lists of all sites the provided access token is able to access.
// ref: https://developers.webflow.com/#list-sites
func (s *SitesService) ListSites(ctx context.Context, opts *ListSitesOptions) ([]*Site, error) {

	u := "/sites"

	u, err := addOptions(u, opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var sites []*Site
	_, err = s.client.Do(ctx, req, &sites)
	if err != nil {
		return nil, err
	}

	return sites, nil
}

//GetSite retrieves the specified site
// ref: https://developers.webflow.com/#get-specific-site
func (s *SitesService) GetSite(ctx context.Context, siteId string) (*Site, error) {

	u := fmt.Sprintf("/sites/%s", siteId)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var site *Site
	_, err = s.client.Do(ctx, req, &site)
	if err != nil {
		return nil, err
	}

	return site, nil
}

type PublishSiteResponse struct {
	Queued bool `json:"queued"`
}

// PublishSite publishes a new site
// ref: https://developers.webflow.com/#publish-site
func (s *SitesService) PublishSite(ctx context.Context, siteId string, domains []string) (*PublishSiteResponse, error) {

	u := fmt.Sprintf("/sites/%s/publish", siteId)

	req, err := s.client.NewRequest("POST", u, map[string][]string{"domains": domains})
	if err != nil {
		return nil, err
	}

	var published *PublishSiteResponse
	_, err = s.client.Do(ctx, req, &published)
	if err != nil {
		return nil, err
	}

	return published, nil
}
