package cart

import (
	"fmt"
	"net/http"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) HandleCartRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, Cart!")
}
