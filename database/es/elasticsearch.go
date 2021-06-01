package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/quan-xie/tuba/log"
)

// TODO more config to add .
type Config struct {
	Addresses []string
}

type ElasticClient struct {
	conf   *Config
	client *elasticsearch.Client
}

func NewElasticsearch(c *Config) *ElasticClient {
	var (
		err error
		es  *elasticsearch.Client
	)
	cfg := elasticsearch.Config{
		Addresses: c.Addresses,
	}
	es, err = elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	return &ElasticClient{
		conf:   c,
		client: es,
	}
}

func (e *ElasticClient) CreateIndex(index, body string) (res *esapi.Response, err error) {
	// IndicesCreate
	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  strings.NewReader(body),
	}
	res, err = req.Do(context.Background(), e.client)
	if err != nil {
		log.Errorf("CreateIndex error %v", err)
		return
	}
	defer res.Body.Close()
	return
}

func (e *ElasticClient) DeleteIndex(index []string) (res *esapi.Response, err error) {
	req := esapi.IndicesDeleteRequest{Index: index}
	res, err = req.Do(context.Background(), e.client)
	if err != nil {
		log.Errorf("DeleteIndex error %v", err)
		return
	}
	defer res.Body.Close()
	fmt.Println(res.String())
	return
}

func (e *ElasticClient) CreateDocument(index, docID, body string) (err error) {
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: docID,
		Body:       strings.NewReader(body),
		Refresh:    "true",
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), e.client)
	if err != nil {
		log.Errorf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Errorf("[%s] Error indexing document ID=%s", res.Status(), docID)
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Infof("Error parsing the response body: %s", err)
		} else {
			// Print the response status and indexed document version.
			log.Infof("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
		}
	}
	return
}

func (e *ElasticClient) UpdateDocumentByID(index, docID, body string) (err error) {
	req := esapi.UpdateRequest{Index: index, DocumentID: docID, Body: strings.NewReader(body)}
	res, err := req.Do(context.Background(), e.client)
	if err != nil {
		log.Errorf("UpdateDocumentByID error %v", err)
		return
	}
	defer res.Body.Close()
	log.Infof("UpdateDocumentByID response %v ", res.String())
	return
}

func (e *ElasticClient) UpdateDocumentByQurey(body string) (err error) {
	req := esapi.DeleteByQueryRequest{Index: []string{"test_index"}, Body: strings.NewReader(body)}
	res, err := req.Do(context.Background(), e.client)
	if err != nil {
		log.Errorf("UpdateDocumentByQurey error %v", err)
		return
	}
	defer res.Body.Close()
	log.Infof("UpdateDocumentByQurey response %v ", res.String())
	return
}

// Qurey query from es .
func (e *ElasticClient) Qurey(index string, query map[string]interface{}, res interface{}) (err error) {
	var (
		buf  bytes.Buffer
		resp *esapi.Response
	)
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		log.Errorf("Error encoding query: %v", err)
		return
	}

	// Perform the search request.
	resp, err = e.client.Search(
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(index),
		e.client.Search.WithBody(&buf),
		e.client.Search.WithTrackTotalHits(true),
		e.client.Search.WithPretty(),
	)
	if err != nil {
		log.Errorf("Error getting response: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.IsError() {
		var tmpMap map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&tmpMap); err != nil {
			log.Errorf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Errorf("[%s] %s: %s",
				resp.Status(),
				tmpMap["error"].(map[string]interface{})["type"],
				tmpMap["error"].(map[string]interface{})["reason"],
			)
		}
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Errorf("Error parsing the response body: %s", err)
	}
	return
}
