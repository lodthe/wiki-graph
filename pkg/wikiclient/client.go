package wikiclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/ratelimit"
)

const EnglishWikipediaURL = `https://en.wikipedia.org/w/api.php`
const MaxRPS = 50

type Client struct {
	apiURL  string
	limiter ratelimit.Limiter

	httpCli *http.Client
}

func New(apiURL string, maxRPS int, httpCli ...*http.Client) *Client {
	client := &Client{
		httpCli: http.DefaultClient,
		apiURL:  apiURL,
		limiter: ratelimit.New(maxRPS),
	}

	if len(httpCli) == 1 {
		client.httpCli = httpCli[0]
	}

	return client
}

func (c *Client) GetMentionedPages(pageTitle string) (titles []string, err error) {
	var cursor *string
	for {
		newBatch, nextCursor, err := c.getLinks(pageTitle, cursor)
		if err != nil {
			return nil, err
		}

		titles = append(titles, newBatch...)

		cursor = nextCursor
		if cursor == nil {
			break
		}
	}

	return titles, nil
}

func (c *Client) getLinks(title string, cursor *string) (titles []string, nextCursor *string, err error) {
	params := url.Values{}
	params.Add("action", "query")
	params.Add("prop", "links")
	params.Add("pllimit", "max")
	params.Add("format", "json")
	params.Add("titles", title)
	if cursor != nil {
		params.Add("plcontinue", *cursor)
	}

	c.limiter.Take()

	resp, err := c.httpCli.Get(fmt.Sprintf("%s?%s", c.apiURL, params.Encode()))
	if err != nil {
		time.Sleep(time.Second)
		return nil, nil, err
	}
	defer resp.Body.Close()

	type Response struct {
		Continue *struct {
			Plcontinue string `json:"plcontinue"`
			Continue   string `json:"continue"`
		} `json:"continue"`
		Query struct {
			Normalized *[]struct {
				From string `json:"from"`
				To   string `json:"to"`
			} `json:"normalized"`
			Pages map[string]struct {
				Pageid int    `json:"pageid"`
				Ns     int    `json:"ns"`
				Title  string `json:"title"`
				Links  []struct {
					Ns    int    `json:"ns"`
					Title string `json:"title"`
				} `json:"links"`
			} `json:"pages"`
		} `json:"query"`
	}

	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, nil, errors.Wrap(err, "decode failed")
	}

	if len(response.Query.Pages) == 0 {
		return nil, nil, nil
	}

	var pageID string
	for id := range response.Query.Pages {
		pageID = id
	}

	titles = make([]string, 0, len(response.Query.Pages[pageID].Links))
	for _, link := range response.Query.Pages[pageID].Links {
		titles = append(titles, link.Title)
	}

	if response.Continue != nil {
		nextCursor = &response.Continue.Plcontinue
	}

	return titles, nextCursor, nil
}
