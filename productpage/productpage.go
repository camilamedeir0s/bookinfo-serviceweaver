package productpage

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

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

// productPageHandler serves the product page with product details and reviews.
func (s *Server) productPageHandler(w http.ResponseWriter, r *http.Request) {
	productID := 1 // Usando um ID de produto padrão

	// Chamando o serviço de detalhes para obter o BookDetails
	bookDetails, err := s.details.Get().GetBookDetails(context.Background(), productID, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get book details: %v", err), http.StatusInternalServerError)
		return
	}

	// Garantindo que bookDetails é do tipo concreto BookDetails
	// Aqui não precisamos de conversão, já que GetBookDetails retorna BookDetails diretamente

	// Chamando o serviço de avaliações
	reviewsResponse, err := s.reviews.Get().BookReviewsByID(context.Background(), fmt.Sprintf("%d", productID))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get book reviews: %v", err), http.StatusInternalServerError)
		return
	}

	// Obtendo o produto de exemplo
	products := getProducts()
	product := products[0]

	// Preparando os dados para passar ao template
	data := map[string]interface{}{
		"detailsStatus": http.StatusOK,
		"reviewsStatus": http.StatusOK,
		"product":       product,
		"details":       bookDetails, // Passando o BookDetails diretamente ao template
		"reviews":       reviewsResponse.Reviews,
		"user":          "", // Usuário não implementado
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
