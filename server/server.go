package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"

	openapi "github.com/urunimi/grpc-open-api/proto"
)

// Backend implements the protobuf interface
type Backend struct {
	mu       *sync.RWMutex
	articles []*openapi.Article
}

// New initializes a new Backend struct.
func New() *Backend {
	return &Backend{
		mu: &sync.RWMutex{},
	}
}

// AddArticle adds an article to the in-memory store.
func (b *Backend) AddArticle(ctx context.Context, req *openapi.AddArticleRequest) (*openapi.Article, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	article := &openapi.Article{
		Id:          uuid.Must(uuid.NewV4()).String(),
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		CreatedAt:   ptypes.TimestampNow(),
	}
	b.articles = append(b.articles, article)

	return article, nil
}

// ListArticles lists all articles in the store.
func (b *Backend) ListArticles(_ *openapi.ListArticlesRequest, srv openapi.ArticleService_ListArticlesServer) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, article := range b.articles {
		err := srv.Send(article)
		if err != nil {
			return err
		}
	}

	return nil
}

// CustomErrorHandler defines the way we want errors to be shown to the articles.
func CustomErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, err error) {
	st := status.Convert(err)

	httpStatus := runtime.HTTPStatusFromCode(st.Code())
	w.WriteHeader(httpStatus)

	w.Write([]byte(st.Message()))
}
