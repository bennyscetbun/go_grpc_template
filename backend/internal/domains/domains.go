package domains

import (
	"net/http"
	"strings"
	"sync"
)

type Domains struct {
	DefaultHandler http.Handler
	handlers       map[string]http.Handler
	mut            *sync.Mutex
}

func NewDomain() *Domains {
	return &Domains{
		DefaultHandler: nil,
		handlers:       map[string]http.Handler{},
		mut:            &sync.Mutex{},
	}
}

func (s *Domains) GetOrCreateDomainsHandler(domain string, builder func() http.Handler) http.Handler {
	s.mut.Lock()
	defer s.mut.Unlock()
	handler := s.handlers[domain]
	if handler == nil {
		s.handlers[domain] = builder()
		handler = s.handlers[domain]
	}
	return handler
}

func (s Domains) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if mux := s.handlers[strings.ToLower(r.Host)]; mux != nil {
		mux.ServeHTTP(w, r)
	} else if s.DefaultHandler != nil {
		s.DefaultHandler.ServeHTTP(w, r)
	} else {
		http.Error(w, "Not found", 404)
	}
}

func (s Domains) GetDomains() []string {
	s.mut.Lock()
	defer s.mut.Unlock()

	keys := make([]string, 0, len(s.handlers))
	for k := range s.handlers {
		keys = append(keys, k)
	}
	return keys
}
