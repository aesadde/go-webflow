package webflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-querystring/query"
	"golang.org/x/oauth2"
)

const (
	defaultBaseURL = "https://api.webflow.com/"
	apiVersion     = "1.0.0"
	userAgent      = "go-webflow"

	headerRateLimit     = "X-RateLimit-Limit"
	headerRateRemaining = "X-RateLimit-Remaining"

	headerApiVersion = "Accept-Version"
)

var errNonNilContext = errors.New("context must be non-nil")

// A Client manages communication with the Webflow API.
type Client struct {
	clientMu sync.Mutex   // clientMu protects the client during calls that modify the CheckRedirect func.
	client   *http.Client // HTTP client used to communicate with the API.

	// Base URL for API requests. Defaults to the public Webflow API, but can be
	// set to a domain endpoint to use with Webflow Enterprise. BaseURL should
	// always be specified with a trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the Webflow API.
	UserAgent string

	rateMu    sync.Mutex
	rateLimit Rate // Rate limits for the client as determined by the most recent API calls.

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services used for talking to different parts of the Webflow API.
	//TODO: Add services
	Meta        *MetaService
	Sites       *SitesService
	Domains     *DomainsService
	Collections *CollectionsService
	Items       *ItemsService
	Ecommerce   *EcommerceService
	Webhooks    *WebhooksService
}

type service struct {
	client *Client
}

// Client returns the http.Client used by this Webflow client.
func (c *Client) Client() *http.Client {
	c.clientMu.Lock()
	defer c.clientMu.Unlock()
	clientCopy := *c.client
	return &clientCopy
}

func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return newClient(tc)
}

func newClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{client: httpClient, BaseURL: baseURL, UserAgent: userAgent}
	c.common.client = c

	c.Meta = (*MetaService)(&c.common)
	c.Sites = (*SitesService)(&c.common)
	c.Domains = (*DomainsService)(&c.common)
	c.Collections = (*CollectionsService)(&c.common)
	c.Items = (*ItemsService)(&c.common)
	c.Ecommerce = (*EcommerceService)(&c.common)
	c.Webhooks = (*WebhooksService)(&c.common)

	return c
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set(headerApiVersion, apiVersion)

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil
}

// Response is a Webflow API response. This wraps the standard http.Response
// returned from Webflow and provides convenient access to things like
// pagination links.
type Response struct {
	*http.Response

	// These fields provide the page values for paginating through a set of
	// results. Any or all of these may be set to the zero value for
	// responses that are not part of a paginated set, or for which there
	// are no additional pages.
	//
	// These fields support what is called "offset pagination" and should
	// be used with the ListOptions struct.
	NextPage  int
	PrevPage  int
	FirstPage int
	LastPage  int

	// Explicitly specify the Rate type so Rate's String() receiver doesn't
	// propagate to Response.
	Rate Rate

	// token's expiration date
	TokenExpiration time.Time
}

// newResponse creates a new Response for the provided http.Response.
// r must not be nil.
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	//response.populatePageValues()
	response.Rate = parseRate(r)
	//response.TokenExpiration = parseTokenExpiration(r)
	return response
}

// Rate represents the rate limit for the current client.
type Rate struct {
	// The number of requests per hour the client is currently limited to.
	Limit int `json:"limit"`

	// The number of remaining requests the client can make this hour.
	Remaining int `json:"remaining"`

	// The time at which the current rate limit will reset.
	Reset time.Time `json:"reset"`
}

// parseRate parses the rate related headers.
func parseRate(r *http.Response) Rate {
	var rate Rate
	if limit := r.Header.Get(headerRateLimit); limit != "" {
		rate.Limit, _ = strconv.Atoi(limit)
	}
	if remaining := r.Header.Get(headerRateRemaining); remaining != "" {
		rate.Remaining, _ = strconv.Atoi(remaining)
	}

	return rate
}

// RateLimitError occurs when Webflow returns 403 Forbidden response with a rate limit
// remaining value of 0.
type RateLimitError struct {
	Rate     Rate           // Rate specifies last known rate limit for the client
	Response *http.Response // HTTP response that caused this error
	Message  string         `json:"message"` // error message
}

func (r *RateLimitError) Error() string {
	return fmt.Sprintf("%v %v: %d %v %v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Message, time.Until(r.Rate.Reset))
}

// Is returns whether the provided error equals this error.
func (r *RateLimitError) Is(target error) bool {
	v, ok := target.(*RateLimitError)
	if !ok {
		return false
	}

	return r.Rate == v.Rate &&
		r.Message == v.Message &&
		compareHTTPResponse(r.Response, v.Response)
}

// compareHTTPResponse returns whether two http.Response objects are equal or not.
// Currently, only StatusCode is checked. This function is used when implementing the
// Is(error) bool interface for the custom error types in this package.
func compareHTTPResponse(r1, r2 *http.Response) bool {
	if r1 == nil && r2 == nil {
		return true
	}

	if r1 != nil && r2 != nil {
		return r1.StatusCode == r2.StatusCode
	}
	return false
}

// checkRateLimitBeforeDo does not make any network calls, but uses existing knowledge from
// current client state in order to quickly check if *RateLimitError can be immediately returned
// from Client.Do, and if so, returns it so that Client.Do can skip making a network API call unnecessarily.
// Otherwise it returns nil, and Client.Do should proceed normally.
func (c *Client) checkRateLimitBeforeDo(req *http.Request) *RateLimitError {
	c.rateMu.Lock()
	rate := c.rateLimit
	c.rateMu.Unlock()
	if !rate.Reset.IsZero() && rate.Remaining == 0 && time.Now().Before(rate.Reset) {
		// Create a fake response.
		resp := &http.Response{
			Status:     http.StatusText(http.StatusForbidden),
			StatusCode: http.StatusForbidden,
			Request:    req,
			Header:     make(http.Header),
			Body:       ioutil.NopCloser(strings.NewReader("")),
		}
		return &RateLimitError{
			Rate:     rate,
			Response: resp,
			Message: fmt.Sprintf("API rate limit of %v still exceeded until %v, not making remote request.",
				rate.Limit, rate.Reset),
		}
	}

	return nil
}

// BareDo sends an API request and lets you handle the api response. If an error
// or API Error occurs, the error will contain more information. Otherwise you
// are supposed to read and close the response's Body. If rate limit is exceeded
// and reset time is in the future, BareDo returns *RateLimitError immediately
// without making a network API call.
//
// The provided ctx must be non-nil, if it is nil an error is returned. If it is
// canceled or times out, ctx.Err() will be returned.
func (c *Client) BareDo(ctx context.Context, req *http.Request) (*Response, error) {
	if ctx == nil {
		return nil, errNonNilContext
	}

	//if bypass := ctx.Value(bypassRateLimitCheck); bypass == nil {
	// If we've hit rate limit, don't make further requests before Reset time.
	if err := c.checkRateLimitBeforeDo(req); err != nil {
		return &Response{
			Response: err.Response,
			Rate:     err.Rate,
		}, err
	}
	//}

	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return nil, err
	}

	response := newResponse(resp)

	c.rateMu.Lock()
	c.rateLimit = response.Rate
	c.rateMu.Unlock()

	err = CheckResponse(resp)
	return response, err
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer interface,
// the raw response body will be written to v, without attempting to first
// decode it. If v is nil, and no error happens, the response is returned as is.
// If rate limit is exceeded and reset time is in the future, Do returns
// *RateLimitError immediately without making a network API call.
//
// The provided ctx must be non-nil, if it is nil an error is returned. If it
// is canceled or times out, ctx.Err() will be returned.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	resp, err := c.BareDo(ctx, req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	switch v := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(v, resp.Body)
	default:
		decErr := json.NewDecoder(resp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}
	return resp, err
}

// An ErrorResponse reports one or more errors caused by an API request.
// Webflow API docs: https://developers.webflow.com/?shell#errors
type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Msg      string         `json:"msg"`
	Code     int            `json:"code"`
	Name     string         `json:"name"`
	Path     string         `json:"path"`
	Err      string         `json:"err"`
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v %+v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Msg, r.Err)
}

func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}

	switch {
	case r.StatusCode == http.StatusTooManyRequests && r.Header.Get(headerRateRemaining) == "0":
		return &RateLimitError{
			Rate:     parseRate(r),
			Response: errorResponse.Response,
			Message:  errorResponse.Msg,
		}
	default:
		return errorResponse
	}
}

// ListOptions specifies the optional parameters to various List methods that
// support offset pagination.
type ListOptions struct {
	// Maximum number of Items to be returned (max limit: 100)
	Limit int `url:"limit,omitempty"`

	// 	Offset used for pagination if collection has more than limit items
	Offset int `url:"offset,omitempty"`
}

// addOptions adds the parameters in opts as URL query parameters to s. opts
// must be a struct whose fields may contain "url" tags.
func addOptions(s string, opts interface{}) (string, error) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opts)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

//PageInfo holds the pagination information for a Webflow API request
type PageInfo struct {
	Count  int `json:"count,omitempty"`
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
	Total  int `json:"total,omitempty"`
}
