package auth

import (
	"errors"
	"fmt"
	"github.com/RangelReale/osin"
)

var (
	CLIENT_NOT_FOUND    = errors.New("Client not found")
	AUTHORIZE_NOT_FOUND = errors.New("Authorize not found")
	ACCESS_NOT_FOUND    = errors.New("Access not found")
	REFRESH_NOT_FOUND   = errors.New("Refresh not found")
)

type MyStorage struct {
	clients   map[string]osin.Client
	authorize map[string]*osin.AuthorizeData
	access    map[string]*osin.AccessData
	refresh   map[string]string
}

func NewMyStorage() *MyStorage {
	r := &MyStorage{
		clients:   make(map[string]osin.Client),
		authorize: make(map[string]*osin.AuthorizeData),
		access:    make(map[string]*osin.AccessData),
		refresh:   make(map[string]string),
	}

	r.clients["test"] = &osin.DefaultClient{
		Id:          "test",
		Secret:      "mysecret",
		RedirectUri: "https://myapp.appspot.com/oauth2/callback",
	}

	return r
}

func (s *MyStorage) Clone() osin.Storage {
	return s
}

func (s *MyStorage) Close() {
}

func (s *MyStorage) GetClient(id string) (osin.Client, error) {
	fmt.Printf("GetClient: %s\n", id)
	if c, ok := s.clients[id]; ok {
		return c, nil
	}
	return nil, CLIENT_NOT_FOUND
}

func (s *MyStorage) SetClient(id string, client osin.Client) error {
	fmt.Printf("SetClient: %s\n", id)
	s.clients[id] = client
	return nil
}

func (s *MyStorage) SaveAuthorize(data *osin.AuthorizeData) error {
	fmt.Printf("SaveAuthorize: %s\n", data.Code)
	s.authorize[data.Code] = data
	return nil
}

func (s *MyStorage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	fmt.Printf("LoadAuthorize: %s\n", code)
	if d, ok := s.authorize[code]; ok {
		return d, nil
	}
	return nil, AUTHORIZE_NOT_FOUND
}

func (s *MyStorage) RemoveAuthorize(code string) error {
	fmt.Printf("RemoveAuthorize: %s\n", code)
	delete(s.authorize, code)
	return nil
}

func (s *MyStorage) SaveAccess(data *osin.AccessData) error {
	fmt.Printf("SaveAccess: %s\n", data.AccessToken)
	s.access[data.AccessToken] = data
	if data.RefreshToken != "" {
		s.refresh[data.RefreshToken] = data.AccessToken
	}
	return nil
}

func (s *MyStorage) LoadAccess(code string) (*osin.AccessData, error) {
	fmt.Printf("LoadAccess: %s\n", code)
	if d, ok := s.access[code]; ok {
		return d, nil
	}
	return nil, ACCESS_NOT_FOUND
}

func (s *MyStorage) RemoveAccess(code string) error {
	fmt.Printf("RemoveAccess: %s\n", code)
	delete(s.access, code)
	return nil
}

func (s *MyStorage) LoadRefresh(code string) (*osin.AccessData, error) {
	fmt.Printf("LoadRefresh: %s\n", code)
	if d, ok := s.refresh[code]; ok {
		return s.LoadAccess(d)
	}
	return nil, REFRESH_NOT_FOUND
}

func (s *MyStorage) RemoveRefresh(code string) error {
	fmt.Printf("RemoveRefresh: %s\n", code)
	delete(s.refresh, code)
	return nil
}
