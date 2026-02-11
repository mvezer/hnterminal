package repository

import (
	"encoding/json"
	api "hnterminal/internal/apiclient"
	"hnterminal/internal/utils"
	"log"
	"strconv"
	"sync"

	badger "github.com/dgraph-io/badger/v4"
)

const MAX_ITEM_GET_BATCH_SIZE = 5

type ItemIds []int
type Item struct {
	By            string  `json:"by"`
	Id            int     `json:"id"`
	CommentsCount int     `json:"descendants"`
	Kids          ItemIds `json:"kids"`
	Score         int     `json:"score"`
	Time          int     `json:"time"`
	Title         string  `json:"title"`
	Type          string  `json:"type"`
	Url           string  `json:"url"`
	Parent        int     `json:"parent"`
	Text          string  `json:"text"`
	IsDead        bool    `json:"dead"`
	Parts         []Item  `json:"parts"`
	IsDeleted     bool    `json:"deleted"`
}
type User struct {
	Id        int `json:"id"`
	CreadedAt int `json:"created_at"`
}
type Repository struct {
	db         *badger.DB
	apiClient  *api.ApiClient
	updatedIds map[int]bool
	wg         *sync.WaitGroup
}

func New(apiClient *api.ApiClient) *Repository {
	opts := badger.DefaultOptions("/tmp/badger")
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	client := apiClient
	if client == nil {
		client = api.New(nil)
	}
	return &Repository{db, client, make(map[int]bool, 0), &sync.WaitGroup{}}
}

func (r *Repository) SetUpdatedIds(ids []int) {
	for _, id := range ids {
		r.updatedIds[id] = true
	}
}

func (r *Repository) GetItem(id int) (*Item, error) {
	var item *Item
	if !r.updatedIds[id] {
		var err error
		item, err = r.LoadItemFromCache(id)
		if err != nil && err != badger.ErrKeyNotFound {
			log.Printf("error while getting item from the repository: %v", err)
			return nil, err
		}
	}
	if item == nil {
		apiBytes, apiError := r.apiClient.GetItem(id)
		if apiError != nil {
			log.Printf("error while getting item from the hacker-news API: %v", apiError)
			return nil, apiError
		}
		jsonError := json.Unmarshal(apiBytes, &item)
		if jsonError != nil {
			return nil, jsonError
		}

		r.SaveItemToCache(id, item)
	}
	return item, nil
}

func (r *Repository) GetItems(ids []int) ([]*Item, error) {
	items := make([]*Item, len(ids))
	processedCount := 0
	for processedCount < len(ids) {
		for i := processedCount; i < utils.Min(processedCount+MAX_ITEM_GET_BATCH_SIZE, len(ids)); i++ {
			r.wg.Go(func() {
				item, err := r.GetItem(ids[i])
				if err != nil {
					items[i] = nil
				} else {
					items[i] = item
				}
			})
		}
		r.wg.Wait()
		processedCount += (MAX_ITEM_GET_BATCH_SIZE - 1)
	}
	return items, nil
}

func (r *Repository) LoadItemFromCache(id int) (*Item, error) {
	var item Item
	cacheError := r.db.View(func(txn *badger.Txn) error {
		cachedBytes, err := txn.Get([]byte(strconv.Itoa(id)))
		if err != nil {
			return err
		}
		extractValueError := cachedBytes.Value(func(val []byte) error {
			jsonError := json.Unmarshal(val, &item)
			if jsonError != nil {
				return jsonError
			}
			return nil
		})
		if extractValueError != nil {
			return extractValueError
		}
		return nil
	})
	if cacheError != nil {
		return nil, cacheError
	}
	return &item, nil
}

func (r *Repository) SaveItemToCache(id int, item *Item) error {
	err := r.db.Update(func(txn *badger.Txn) error {
		bytes, err := json.Marshal(item)
		if err != nil {
			return err
		}
		err = txn.Set([]byte(strconv.Itoa(id)), bytes)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (r *Repository) Close() {
	r.db.Close()
}
