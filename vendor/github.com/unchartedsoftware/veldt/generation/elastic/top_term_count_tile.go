package elastic

import (
	"encoding/json"

	"github.com/unchartedsoftware/veldt"
	"github.com/unchartedsoftware/veldt/binning"
)

// TopTermCountTile represents an elasticsearch implementation of the
// top term count tile.
type TopTermCountTile struct {
	Bivariate
	TopTerms
	Tile
}

// NewTopTermCountTile instantiates and returns a new tile struct.
func NewTopTermCountTile(host, port string) veldt.TileCtor {
	return func() (veldt.Tile, error) {
		t := &TopTermCountTile{}
		t.Host = host
		t.Port = port
		return t, nil
	}
}

// Parse parses the provided JSON object and populates the tiles attributes.
func (t *TopTermCountTile) Parse(params map[string]interface{}) error {
	err := t.Bivariate.Parse(params)
	if err != nil {
		return err
	}
	return t.TopTerms.Parse(params)
}

// Create generates a tile from the provided URI, tile coordinate and query
// parameters.
func (t *TopTermCountTile) Create(uri string, coord *binning.TileCoord, query veldt.Query) ([]byte, error) {
	// get client
	client, err := NewClient(t.Host, t.Port)
	if err != nil {
		return nil, err
	}
	// create search service
	search := client.Search().
		Index(uri).
		Size(0)

	// create root query
	q, err := t.CreateQuery(query)
	if err != nil {
		return nil, err
	}
	// add tiling query
	q.Must(t.Bivariate.GetQuery(coord))
	// set the query
	search.Query(q)
	// get agg
	aggs := t.TopTerms.GetAggs()
	// set the aggregation
	search.Aggregation("top-terms", aggs["top-terms"])
	// send query
	res, err := search.Do()
	if err != nil {
		return nil, err
	}
	// get terms
	terms, err := t.TopTerms.GetTerms(&res.Aggregations)
	if err != nil {
		return nil, err
	}
	// encode
	counts := make(map[string]uint32)
	for term, bucket := range terms {
		counts[term] = uint32(bucket.DocCount)
	}
	// marshal results
	return json.Marshal(counts)
}
