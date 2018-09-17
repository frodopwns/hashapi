package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

//TestGetHashHandler tests general error situations around GETS to /hash
func TestGetHashHandler(t *testing.T) {

	var testCases = []struct {
		path   string
		status int
		msg    string
		desc   string
	}{
		{"/hash", http.StatusBadRequest, NoHashIdErr, "GET to /hash with no id"},
		{"/hash/", http.StatusBadRequest, NoHashIdErr, "GET to /hash/ with no id"},
		{"/hash/stuff", http.StatusBadRequest, BadHadhIdErr, "GET to /hash/stuff"},
		{"/hash/12/stuff", http.StatusBadRequest, BadHadhIdErr, "GET to /hash/23/stuff"},
		{"/hash/12", http.StatusNotFound, HashIdNotFound, "GET to non exisitng id"},
		{"/hash/12/", http.StatusNotFound, HashIdNotFound, "GET to non existing id with trailing slash"},
	}
	for _, tc := range testCases {

		req, err := http.NewRequest("GET", tc.path, nil)
		if err != nil {
			t.Fatal(err)
		}

		env := &Env{
			HashMap: NewHashMap(),
			Stats:   NewStats(),
		}
		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		handler := http.Handler(Handler{env, hashHandler})

		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		if status := rr.Code; status != tc.status {
			t.Errorf("%s\nhandler returned wrong status code: got %v want %v",
				tc.desc,
				status,
				tc.status,
			)
		}

		// Check the response body is what we expect.
		expected := tc.msg
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("%s\nhandler returned unexpected body: got %v want %v",
				tc.desc,
				rr.Body.String(),
				expected,
			)
		}
	}

}

// TestPostHashHandler tests POSTS to /hash
func TestPostHashHandler(t *testing.T) {

	var testCases = []struct {
		desc    string
		path    string
		payload string
		status  int
		msg     string
	}{
		{"post with no password payload", "/hash", "", http.StatusBadRequest, NoPayloadPresentErr},
		{"good post", "/hash", "password=test", http.StatusOK, "1"},
	}
	for i, tc := range testCases {
		log.Printf("Starting test %d: %s", i, tc.desc)

		env := &Env{
			HashMap: NewHashMap(),
			Stats:   NewStats(),
		}
		handler := http.Handler(Handler{env, hashHandler})
		ts := httptest.NewServer(handler)

		parts := strings.Split(tc.payload, "=")

		vals := url.Values{}
		if len(parts) > 1 {
			vals = url.Values{
				parts[0]: {parts[1]},
			}
		}
		rr, err := http.PostForm(ts.URL+tc.path, vals)
		if err != nil {
			t.Fatal(err)
		}

		// Check the status code is what we expect.
		if status := rr.StatusCode; status != tc.status {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, tc.status)
		}

		defer rr.Body.Close()
		body, _ := ioutil.ReadAll(rr.Body)

		// Check the response body is what we expect.
		expected := tc.msg
		if !strings.Contains(string(body), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v",
				string(body), expected)
		}
	}

}

// TestStatsHandler tests statistics gathering functionarlity
func TestStatsHandler(t *testing.T) {

	var testCases = []struct {
		desc     string
		expected string
	}{
		{"get to empty /stats", `{"total": 0, "average": NaN}`},
		{"get to /stats after 1 request", `"total": 1,`},
	}
	for i, tc := range testCases {
		log.Printf("Starting test %d: %s", i, tc.desc)
		req, err := http.NewRequest("GET", "/stats", nil)
		if err != nil {
			t.Fatal(err)
		}

		env := &Env{
			HashMap: NewHashMap(),
			Stats:   NewStats(),
		}

		if i > 0 {
			env.Stats.Update(54.0)
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		handler := http.Handler(Handler{env, statsHandler})

		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		if status := rr.Code; status != 200 {
			t.Errorf("%s\nhandler returned wrong status code: got %v want %v",
				"get empty stats",
				status,
				200,
			)
		}

		// Check the response body is what we expect.
		expected := tc.expected
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("%s\nhandler returned unexpected body: got %v want %v",
				"get empty stats",
				rr.Body.String(),
				expected,
			)
		}
	}
}

// TestHashMe tests to make sure a simple string will continue to be hashed the same way
func TestHashMe(t *testing.T) {
	tests := [][]string{
		{"test", "7iaw3Ur350mqGo7jwQrpkj9hiYB3Lkc_iBml1JQODbJ6wYX4oOHV-E-IvIh_1nsUNzLDBMxfqa2Ob1f1ACio_w=="},
	}

	for _, tset := range tests {
		if hashed := HashMe(tset[0]); hashed != tset[1] {
			t.Errorf("function returned unexpected response: got %v want %v",
				hashed,
				tset[1],
			)
		}
	}
}

// TestHashMap does some basic sanity tests around the HashMap
func TestHashMap(t *testing.T) {
	m := NewHashMap()

	expected := "test"
	key := m.Save(expected)
	if v, ok := m.data[key]; ok {
		if v != expected {
			t.Errorf("hashmap returned unexpected value: got %v want %v",
				v,
				expected,
			)
		}
	}

	expected = "test2"
	m.Update(key, expected)
	if v, ok := m.data[key]; ok {
		if v != expected {
			t.Errorf("hashmap returned unexpected value: got %v want %v",
				v,
				expected,
			)
		}
	}

	if v, ok := m.Get(key); ok {
		if v != expected {
			t.Errorf("hashmap returned unexpected value: got %v want %v",
				v,
				expected,
			)
		}
	}
}
