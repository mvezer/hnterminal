package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	HN_BASE_URL     = "https://hacker-news.firebaseio.com/v0/"
	DEFAULT_TIMEOUT = time.Second * 10
)

type ApiClient struct {
	client *http.Client
}

func New(httpClient *http.Client) *ApiClient {
	if httpClient == nil {
		httpClient = CreateHttpClient(DEFAULT_TIMEOUT)
	}
	return &ApiClient{httpClient}
}

/*
GetTopStoryIds returns the newest item ids from the Hacker News API
*/
func (api *ApiClient) GetTopStoryIds() ([]int, error) {
	response, err := api.client.Get(fmt.Sprintf("%s/topstories.json", HN_BASE_URL))
	if err != nil {
		return nil, fmt.Errorf("error while getting the newest items: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code while getting the newest items: %d", response.StatusCode)
	}

	var stories []int
	err = json.NewDecoder(response.Body).Decode(&stories)
	if err != nil {
		return nil, fmt.Errorf("error while decoding the top stories response: %w", err)
	}

	return stories, nil
}

/*
GetItem returns the item by id as raw bytes from the Hacker News API
*/
func (api *ApiClient) GetItem(itemId int) ([]byte, error) {
	response, err := api.client.Get(fmt.Sprintf("%s/item/%d.json", HN_BASE_URL, itemId))
	if err != nil {
		return nil, fmt.Errorf("error while getting the item from the hacker-news API: %w", err)
	}
	item, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading the item response: %w", err)
	}
	return item, nil
}

func CreateHttpClient(timeout time.Duration) *http.Client {
	if timeout == 0 {
		timeout = DEFAULT_TIMEOUT
	}
	return &http.Client{
		Timeout: timeout,
	}
}
