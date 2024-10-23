package productpage

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/ServiceWeaver/weaver"
	"github.com/camilamedeir0s/bookinfo-serviceweaver/details"
	"github.com/camilamedeir0s/bookinfo-serviceweaver/reviews"
)

var (
	//go:embed static/*
	staticFS embed.FS
)

type Server struct {
	weaver.Implements[weaver.Main]
	handler     http.Handler
	productpage weaver.Listener
	details     weaver.Ref[details.Details] // Adiciona o campo details aqui
	reviews     weaver.Ref[reviews.Reviews]
}

// Serve initializes the product page service.
func Serve(ctx context.Context, s *Server) error {
	// Set up static file serving.
	staticHTML, err := fs.Sub(fs.FS(staticFS), "static")
	if err != nil {
		return err
	}

	// Call the details service to get the book details for a sample book ID (e.g., 1).
	bookID := 1
	bookDetails, err := s.details.Get().GetBookDetails(ctx, bookID, nil)
	if err != nil {
		return fmt.Errorf("failed to get book details: %w", err)
	}

	// Call the reviews service to get the reviews for the same book ID.
	bookReviews, err := s.reviews.Get().BookReviewsByID(ctx, fmt.Sprintf("%d", bookID))
	if err != nil {
		return fmt.Errorf("failed to get book reviews: %w", err)
	}

	// Print the details and reviews for the book.
	fmt.Println("Book Details:", bookDetails)
	fmt.Println("Book Reviews:", bookReviews)

	// Print the book details to the log.
	s.Logger(ctx).Info("Fetched book details", "bookDetails", bookDetails)

	// Set up basic routing.
	r := http.NewServeMux()
	r.Handle("/", http.FileServer(http.FS(staticHTML)))

	// Set handler and log initialization.
	s.handler = r
	s.Logger(ctx).Debug("ProductPage service is up", "address", s.productpage)

	// Serve requests on the Service Weaver listener.
	return http.Serve(s.productpage, s.handler)
}
