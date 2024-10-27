package productpage

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/ServiceWeaver/weaver"
	"github.com/camilamedeir0s/bookinfo-serviceweaver/details"
	"github.com/camilamedeir0s/bookinfo-serviceweaver/reviews"
)

//go:embed templates/*
var embeddedFiles embed.FS

type Server struct {
	weaver.Implements[weaver.Main]
	handler     http.Handler
	productpage weaver.Listener
	details     weaver.Ref[details.Details]
	reviews     weaver.Ref[reviews.Reviews]
	templates   *template.Template
}

// Product represents a product.
type Product struct {
	ID              int    `json:"id"`
	Title           string `json:"title"`
	DescriptionHtml string `json:"descriptionHtml"`
}

// Service represents a service with children.
type Service struct {
	Name     string    `json:"name"`
	Children []Service `json:"children"`
}

// Serve initializes the product page service.
func Serve(ctx context.Context, s *Server) error {
	// Set up static file serving.
	staticHTML, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		return err
	}

	// Load templates
	s.templates, err = template.ParseFS(embeddedFiles, "templates/*.html")
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// Set up routing
	r := http.NewServeMux()
	r.Handle("/", http.HandlerFunc(s.indexHandler))
	r.HandleFunc("/productpage", s.productPageHandler)
	r.HandleFunc("/api/v1/products", s.productsHandler)
	r.HandleFunc("/api/v1/products/{id}", s.productHandler)
	r.HandleFunc("/api/v1/products/{id}/reviews", s.productReviewsHandler)
	r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticHTML))))

	// Set handler and log initialization.
	s.handler = r
	s.Logger(ctx).Debug("ProductPage service is up", "address", s.productpage)

	// Serve requests on the Service Weaver listener.
	return http.Serve(s.productpage, s.handler)
}

// indexHandler serves the index page with the service table.
func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	// Create the services structure for rendering.
	services := Service{
		Name: "ProductPage",
		Children: []Service{
			{Name: "Details", Children: nil},
			{Name: "Reviews", Children: []Service{
				{Name: "Ratings", Children: nil},
			}},
		},
	}

	// Convert the services structure to an HTML table.
	table := jsonToHTMLTable(services)

	// Set the content type to HTML.
	w.Header().Set("Content-Type", "text/html")

	// Execute the template, passing in the service table.
	if err := s.templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"serviceTable": template.HTML(table),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) productPageHandler(w http.ResponseWriter, r *http.Request) {
	productID := 1 // ID de produto padrão

	// Obtendo os detalhes do livro
	bookDetails, err := s.details.Get().GetBookDetails(context.Background(), productID, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get book details: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println(bookDetails)

	// Obtendo as avaliações do livro
	reviewsResponse, err := s.reviews.Get().BookReviewsByID(context.Background(), fmt.Sprintf("%d", productID))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get book reviews: %v", err), http.StatusInternalServerError)
		return
	}

	// Processa as avaliações e cria StarsSlice e EmptyStarsSlice
	processedReviews := processReviewsWithStarsSlice(reviewsResponse.Reviews)

	// Obtenção de produto exemplo
	products := getProducts()
	product := products[0]

	// Preparando os dados para passar ao template
	data := map[string]interface{}{
		"detailsStatus": http.StatusOK,
		"reviewsStatus": http.StatusOK,
		"product":       product,
		"details":       bookDetails,
		"reviews":       processedReviews, // Reviews processados
		"user":          "",               // Usuário não implementado
		"rating": map[string]interface{}{
			"stars": makeSeq(4), // Exemplo de valor de rating
			"color": "yellow",
		},
	}

	w.Header().Set("Content-Type", "text/html")

	// Renderizando o template productpage.html com os dados
	if err := s.templates.ExecuteTemplate(w, "productpage.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// makeSeq gera uma sequência de números de 0 até n-1
func makeSeq(n int) []int {
	seq := make([]int, n)
	for i := 0; i < n; i++ {
		seq[i] = i
	}
	return seq
}

// jsonToHTMLTable converts a Service structure to an HTML table.
func jsonToHTMLTable(service Service) string {
	html := "<table class='table table-condensed table-bordered table-hover'>"
	html += "<thead><tr><th>Name</th><th>Endpoint</th><th>Children</th></tr></thead>"
	html += "<tbody>"
	html += buildHTMLTableRow(service)
	html += "</tbody></table>"
	return html
}

// buildHTMLTableRow builds HTML rows for a service and its children.
func buildHTMLTableRow(service Service) string {
	row := "<tr>"
	row += fmt.Sprintf("<td>%s</td>", service.Name)
	row += "<td>Interno</td>" // Use "Interno" instead of the endpoint

	if len(service.Children) > 0 {
		row += "<td><table>"
		for _, child := range service.Children {
			row += buildHTMLTableRow(child)
		}
		row += "</table></td>"
	} else {
		row += "<td>None</td>"
	}
	row += "</tr>"
	return row
}

// getProducts returns a sample list of products.
func getProducts() []Product {
	return []Product{
		{
			ID:              0,
			Title:           "The Comedy of Errors",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Comedy_of_Errors\">Wikipedia Summary</a>: The Comedy of Errors is one of <b>William Shakespeare's</b> early plays. It is his shortest and one of his most farcical comedies, with a major part of the humour coming from slapstick and mistaken identity, in addition to puns and word play.",
		},
	}
}

func processReviewsWithStarsSlice(reviews []reviews.Review) []map[string]interface{} {
	var processedReviews []map[string]interface{}

	for _, review := range reviews {
		stars := review.Rating.Stars

		// Criando slices de estrelas cheias e vazias
		starsSlice := make([]int, stars)
		emptyStarsSlice := make([]int, 5-stars) // Assume-se que a escala é de 5 estrelas

		// Cria um mapa para passar ao template
		reviewMap := map[string]interface{}{
			"Reviewer":        review.Reviewer,
			"Text":            review.Text,
			"Rating":          review.Rating,
			"StarsSlice":      starsSlice,
			"EmptyStarsSlice": emptyStarsSlice,
			"Color":           review.Rating.Color,
		}

		processedReviews = append(processedReviews, reviewMap)
	}

	return processedReviews
}

func (s *Server) productsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	products := getProducts()

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)

	err := enc.Encode(products)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	w.Write(buf.Bytes())
}

func (s *Server) productHandler(w http.ResponseWriter, r *http.Request) {
	// Extrai o `id` do produto da URL usando `r.URL.Path`
	pathParts := strings.Split(r.URL.Path, "/")

	// Verifica se a URL tem pelo menos 5 partes para corresponder ao formato `/api/v1/products/{id}`
	if len(pathParts) < 5 {
		http.Error(w, "Invalid product URL", http.StatusBadRequest)
		return
	}

	// O `productID` está na quinta parte da URL (índice 4)
	productID := pathParts[4]

	// Converte o productID para um número inteiro
	id, err := strconv.Atoi(productID)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Chamada direta ao método `GetBookDetails` do componente `details`
	details, err := s.details.Get().GetBookDetails(r.Context(), id, nil)
	if err != nil {
		http.Error(w, "Failed to fetch product details", http.StatusInternalServerError)
		return
	}

	// Serializa `details` para JSON
	response, err := json.Marshal(details)
	if err != nil {
		http.Error(w, "Failed to marshal product details", http.StatusInternalServerError)
		return
	}

	// Envia a resposta JSON ao cliente
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (s *Server) productReviewsHandler(w http.ResponseWriter, r *http.Request) {
	// Extrai o `id` do produto da URL usando `r.URL.Path`
	pathParts := strings.Split(r.URL.Path, "/")

	// Verifica se a URL tem pelo menos 5 partes para corresponder ao formato `/api/v1/products/{id}/reviews`
	if len(pathParts) < 5 {
		http.Error(w, "Invalid product URL", http.StatusBadRequest)
		return
	}

	// O `productID` está na quinta parte da URL (índice 4)
	productID := pathParts[4]

	// Chamada direta ao método `BookReviewsByID` do componente `reviews`
	reviewsResponse, err := s.reviews.Get().BookReviewsByID(r.Context(), productID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get book reviews: %v", err), http.StatusInternalServerError)
		return
	}

	// Serializa `reviewsResponse` para JSON
	response, err := json.Marshal(reviewsResponse)
	if err != nil {
		http.Error(w, "Failed to marshal product reviews", http.StatusInternalServerError)
		return
	}

	// Envia a resposta JSON ao cliente
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
