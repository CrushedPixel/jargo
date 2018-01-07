package parser

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestParseFilterParameters(t *testing.T) {
	// test omitted operator
	query := make(url.Values)
	query[`filter[id]`] = []string{"1,2,3"}
	filters, err := ParseFilterParameters(query)
	assert.Nil(t, err)
	assert.Equal(t, filters["id"], map[string][]string{"EQ": {"1", "2", "3"}})

	// test operators
	// eq
	query = make(url.Values)
	query[`filter[id][eq]`] = []string{"1,2,3"}
	filters, err = ParseFilterParameters(query)
	assert.Nil(t, err)
	assert.Equal(t, filters["id"], map[string][]string{"EQ": {"1", "2", "3"}})

	// ne
	query = make(url.Values)
	query[`filter[id][ne]`] = []string{"1,2,3"}
	filters, err = ParseFilterParameters(query)
	assert.Nil(t, err)
	assert.Equal(t, filters["id"], map[string][]string{"NE": {"1", "2", "3"}})

	// like
	query = make(url.Values)
	query[`filter[id][like]`] = []string{"1,2,3"}
	filters, err = ParseFilterParameters(query)
	assert.Nil(t, err)
	assert.Equal(t, filters["id"], map[string][]string{"LIKE": {"1", "2", "3"}})

	// lt
	query = make(url.Values)
	query[`filter[id][lt]`] = []string{"1,2,3"}
	filters, err = ParseFilterParameters(query)
	assert.Nil(t, err)
	assert.Equal(t, filters["id"], map[string][]string{"LT": {"1", "2", "3"}})

	// lte
	query = make(url.Values)
	query[`filter[id][lte]`] = []string{"1,2,3"}
	filters, err = ParseFilterParameters(query)
	assert.Nil(t, err)
	assert.Equal(t, filters["id"], map[string][]string{"LTE": {"1", "2", "3"}})

	// gt
	query = make(url.Values)
	query[`filter[id][gt]`] = []string{"1,2,3"}
	filters, err = ParseFilterParameters(query)
	assert.Nil(t, err)
	assert.Equal(t, filters["id"], map[string][]string{"GT": {"1", "2", "3"}})

	// gte
	query = make(url.Values)
	query[`filter[id][gte]`] = []string{"1,2,3"}
	filters, err = ParseFilterParameters(query)
	assert.Nil(t, err)
	assert.Equal(t, filters["id"], map[string][]string{"GTE": {"1", "2", "3"}})

	// test case-insensitivity of operators
	query = make(url.Values)
	query[`filter[id][LiKe]`] = []string{"1,2,3"}
	filters, err = ParseFilterParameters(query)
	assert.Nil(t, err)
	assert.Equal(t, filters["id"], map[string][]string{"LIKE": {"1", "2", "3"}})

	// test invalid operators
	query = make(url.Values)
	query[`filter[id][custom]`] = []string{"1,2,3"}
	filters, err = ParseFilterParameters(query)
	assert.EqualError(t, err, errInvalidOperator("custom").Error())
}
