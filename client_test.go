package consul

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/segmentio/objconv/json"
)

func TestClient(t *testing.T) {
	t.Run("Basic", testClientBasic)
	t.Run("Session", testSession)
	t.Run("checkSession", testCheckSession)
}

func testClientBasic(t *testing.T) {
	tests := []struct {
		method string
		path   string
		query  Query
		send   map[string]string
		recv   map[string]string
	}{
		{
			method: "GET",
			path:   "/",
			query:  Query{{"question", "universe"}},
			recv:   map[string]string{"answer": "42"},
			send:   map[string]string{},
		},
		{
			method: "POST",
			path:   "/hello/world",
			query:  nil,
			recv:   map[string]string{"answer": "42"},
			send:   map[string]string{},
		},
		{
			method: "PUT",
			path:   "/hello/world",
			query:  nil,
			recv:   map[string]string{},
			send:   map[string]string{},
		},
		{
			method: "DELETE",
			path:   "/hello/world",
			query:  nil,
			recv:   map[string]string{},
			send:   map[string]string{},
		},
	}

	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			server, client := newServerClient(func(res http.ResponseWriter, req *http.Request) {
				var send map[string]string

				switch {
				case req.Method != test.method:
					t.Error("invalid method:", req.Method)

				case !reflect.DeepEqual(req.URL.Query(), append(test.query, Param{"dc", "dc1"}).Values()):
					t.Error("invalid query string:", req.URL.RawQuery)
				}

				if err := json.NewDecoder(req.Body).Decode(&send); err != nil {
					t.Error(err)
				}

				if !reflect.DeepEqual(send, test.send) {
					t.Error(send)
				}

				json.NewEncoder(res).Encode(test.recv)
			})
			defer server.Close()

			var recv map[string]string
			if err := client.Do(context.Background(), test.method, test.path, test.query, test.send, &recv); err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(recv, test.recv) {
				t.Error(recv)
			}
		})
	}

}

func testCheckSession(t *testing.T) {
	if err := (&Client{}).checkSession("test"); err == nil {
		t.Error("no error detected by consul.(*Client).checkSession on a client that had no session")
	}
}

func newServerClient(handler func(http.ResponseWriter, *http.Request)) (server *httptest.Server, client *Client) {
	server = httptest.NewServer(http.HandlerFunc(handler))
	client = &Client{
		Address:    server.URL,
		UserAgent:  "test",
		Datacenter: "dc1",
	}
	return
}