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
	"github.com/camilamedeir0s/bookinfo-serviceweaver/ratings"
	"github.com/camilamedeir0s/bookinfo-serviceweaver/reviews"
)

//go:embed static/* templates/*
var embeddedFiles embed.FS

type Server struct {
	weaver.Implements[weaver.Main]
	handler     http.Handler
	productpage weaver.Listener
	details     weaver.Ref[details.Details]
	reviews     weaver.Ref[reviews.Reviews]
	ratings     weaver.Ref[ratings.Ratings]
	templates   *template.Template
}

// Product represents a product.
type Product struct {
	ID              int           `json:"id"`
	Title           string        `json:"title"`
	DescriptionHtml template.HTML `json:"descriptionHtml"`
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

	r.Handle("/", weaver.InstrumentHandler("index", http.HandlerFunc(s.indexHandler)))
	r.Handle("/health", weaver.InstrumentHandler("health", http.HandlerFunc(s.healthHandler)))
	r.Handle("/productpage", weaver.InstrumentHandler("productpage-reviews-details", http.HandlerFunc(s.productPageHandler)))
	r.Handle("/api/v1/products", weaver.InstrumentHandler("products", http.HandlerFunc(s.productsHandler)))
	r.Handle("/api/v1/products/{id}", weaver.InstrumentHandler("product", http.HandlerFunc(s.productHandler)))
	r.Handle("/api/v1/products/{id}/reviews", weaver.InstrumentHandler("product-reviews", http.HandlerFunc(s.productReviewsHandler)))
	r.Handle("/api/v1/products/{id}/ratings", weaver.InstrumentHandler("product-ratings", http.HandlerFunc(s.productRatingsHandler)))

	// Static content não precisa de tracing, pode manter normal:
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
	ctx := r.Context()

	// Obtendo os detalhes do livro
	bookDetails, err := s.details.Get().GetBookDetails(ctx, productID, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get book details: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println(bookDetails)

	// Obtendo as avaliações do livro
	reviewsResponse, err := s.reviews.Get().BookReviewsByID(ctx, fmt.Sprintf("%d", productID))
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
		{
			ID:              1,
			Title:           "The Comedy of Errors",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Comedy_of_Errors\">Wikipedia Summary</a>: The Comedy of Errors is one of <b>William Shakespeare's</b> early plays. It is his shortest and one of his most farcical comedies, with a major part of the humour coming from slapstick and mistaken identity, in addition to puns and word play.",
		},
		{
			ID:              2,
			Title:           "Hamlet",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Hamlet\">Wikipedia Summary</a>: <b>Hamlet</b> is a tragedy written by William Shakespeare sometime between 1599 and 1601. It is Shakespeare's longest play and is among the most powerful and influential tragedies in English literature.",
		},
		{
			ID:              3,
			Title:           "Romeo and Juliet",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Romeo_and_Juliet\">Wikipedia Summary</a>: <b>Romeo and Juliet</b> is a tragedy written early in the career of William Shakespeare about two young star-crossed lovers whose deaths ultimately reconcile their feuding families.",
		},
		{
			ID:              4,
			Title:           "Macbeth",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Macbeth\">Wikipedia Summary</a>: <b>Macbeth</b> is a tragedy by William Shakespeare. It is thought to have been first performed in 1606. It is one of Shakespeare's most famous and popular works.",
		},
		{
			ID:              5,
			Title:           "Othello",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Othello\">Wikipedia Summary</a>: <b>Othello</b> is a tragedy by William Shakespeare, believed to have been written in 1603. The play centers on the character Othello, a Moorish general, and his ensign, Iago.",
		},
		{
			ID:              6,
			Title:           "A Midsummer Night's Dream",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/A_Midsummer_Night%27s_Dream\">Wikipedia Summary</a>: <b>A Midsummer Night's Dream</b> is a comedy written by William Shakespeare in 1595/96. The play is one of Shakespeare's most popular works for the stage and is widely performed across the world.",
		},
		{
			ID:              7,
			Title:           "Julius Caesar",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Julius_Caesar_(play)\">Wikipedia Summary</a>: <b>Julius Caesar</b> is a tragedy by William Shakespeare, believed to have been written in 1599. It is one of several Roman plays that Shakespeare wrote, based on true events from Roman history.",
		},
		{
			ID:              8,
			Title:           "The Tempest",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Tempest\">Wikipedia Summary</a>: <b>The Tempest</b> is a play by William Shakespeare, probably written in 1610–1611. It is considered one of Shakespeare's late romances.",
		},
		{
			ID:              9,
			Title:           "King Lear",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/King_Lear\">Wikipedia Summary</a>: <b>King Lear</b> is a tragedy written by William Shakespeare. It depicts the gradual descent into madness of the title character, after he disposes of his kingdom giving bequests to two of his three daughters based on their flattery of him.",
		},
		{
			ID:              10,
			Title:           "Twelfth Night",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Twelfth_Night\">Wikipedia Summary</a>: <b>Twelfth Night</b> is a comedy by William Shakespeare, believed to have been written around 1601–1602. It centers on the twins Viola and Sebastian, who are separated in a shipwreck.",
		},
		{
			ID:              11,
			Title:           "The Merchant of Venice",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Merchant_of_Venice\">Wikipedia Summary</a>: <b>The Merchant of Venice</b> is a 16th-century play by William Shakespeare in which a merchant in Venice must default on a large loan provided by a Jewish moneylender, Shylock.",
		},
		{
			ID:              12,
			Title:           "Much Ado About Nothing",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Much_Ado_About_Nothing\">Wikipedia Summary</a>: <b>Much Ado About Nothing</b> is a comedy by William Shakespeare thought to have been written in 1598 and 1599. The play was included in the First Folio, published in 1623.",
		},
		{
			ID:              13,
			Title:           "Richard III",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Richard_III_(play)\">Wikipedia Summary</a>: <b>Richard III</b> is a historical play by William Shakespeare, believed to have been written around 1593. It depicts the Machiavellian rise to power and subsequent short reign of King Richard III of England.",
		},
		{
			ID:              14,
			Title:           "Antony and Cleopatra",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Antony_and_Cleopatra\">Wikipedia Summary</a>: <b>Antony and Cleopatra</b> is a tragedy by William Shakespeare. It was first performed around 1607 and depicts the relationship between Cleopatra and Mark Antony.",
		},
		{
			ID:              15,
			Title:           "Coriolanus",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Coriolanus\">Wikipedia Summary</a>: <b>Coriolanus</b> is a tragedy by William Shakespeare, believed to have been written between 1605 and 1608. The play is based on the life of the legendary Roman leader Caius Marcius Coriolanus.",
		},
		{
			ID:              16,
			Title:           "Henry V",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Henry_V_(play)\">Wikipedia Summary</a>: <b>Henry V</b> is a history play by William Shakespeare, believed to have been written near 1599. It focuses on King Henry V of England, and events before and after the Battle of Agincourt during the Hundred Years' War.",
		},
		{
			ID:              17,
			Title:           "As You Like It",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/As_You_Like_It\">Wikipedia Summary</a>: <b>As You Like It</b> is a pastoral comedy by William Shakespeare believed to have been written in 1599 and first published in the First Folio in 1623.",
		},
		{
			ID:              18,
			Title:           "Measure for Measure",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Measure_for_Measure\">Wikipedia Summary</a>: <b>Measure for Measure</b> is a play by William Shakespeare, believed to have been written in 1603 or 1604. It is often classified as a comedy, though its themes are more serious and introspective.",
		},
		{
			ID:              19,
			Title:           "The Taming of the Shrew",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Taming_of_the_Shrew\">Wikipedia Summary</a>: <b>The Taming of the Shrew</b> is a comedy by William Shakespeare, believed to have been written between 1590 and 1592.",
		},
		{
			ID:              20,
			Title:           "All's Well That Ends Well",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/All%27s_Well_That_Ends_Well\">Wikipedia Summary</a>: <b>All's Well That Ends Well</b> is a play by William Shakespeare, first published in the First Folio in 1623, although it was probably written earlier.",
		},
		{
			ID:              21,
			Title:           "Timon of Athens",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Timon_of_Athens\">Wikipedia Summary</a>: <b>Timon of Athens</b> is a play by William Shakespeare, probably written in collaboration with Thomas Middleton, about the fortunes and misfortunes of the title character, a wealthy Athenian.",
		},
		{
			ID:              22,
			Title:           "Titus Andronicus",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Titus_Andronicus\">Wikipedia Summary</a>: <b>Titus Andronicus</b> is a tragedy by William Shakespeare, believed to have been written between 1588 and 1593, probably in collaboration with George Peele.",
		},
		{
			ID:              23,
			Title:           "Love's Labour's Lost",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Love%27s_Labour%27s_Lost\">Wikipedia Summary</a>: <b>Love's Labour's Lost</b> is one of William Shakespeare's early comedies, believed to have been written in the mid-1590s.",
		},
		{
			ID:              24,
			Title:           "Pericles, Prince of Tyre",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Pericles,_Prince_of_Tyre\">Wikipedia Summary</a>: <b>Pericles, Prince of Tyre</b> is a play written at least in part by William Shakespeare and first performed in 1608.",
		},
		{
			ID:              25,
			Title:           "Troilus and Cressida",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Troilus_and_Cressida\">Wikipedia Summary</a>: <b>Troilus and Cressida</b> is a tragedy by William Shakespeare, believed to have been written in 1602.",
		},
		{
			ID:              26,
			Title:           "Cymbeline",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Cymbeline\">Wikipedia Summary</a>: <b>Cymbeline</b> is a play by William Shakespeare, based on legends concerning the early Celtic British King Cunobeline.",
		},
		{
			ID:              27,
			Title:           "The Two Gentlemen of Verona",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Two_Gentlemen_of_Verona\">Wikipedia Summary</a>: <b>The Two Gentlemen of Verona</b> is a comedy by William Shakespeare, and one of his earliest plays, believed to have been written in the early 1590s.",
		},
		{
			ID:              28,
			Title:           "Henry IV, Part 1",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Henry_IV,_Part_1\">Wikipedia Summary</a>: <b>Henry IV, Part 1</b> is a history play by William Shakespeare, believed to have been written no later than 1597.",
		},
		{
			ID:              29,
			Title:           "Henry IV, Part 2",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Henry_IV,_Part_2\">Wikipedia Summary</a>: <b>Henry IV, Part 2</b> is a history play by William Shakespeare, believed to have been written between 1596 and 1599.",
		},
		{
			ID:              30,
			Title:           "Henry VI, Part 1",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Henry_VI,_Part_1\">Wikipedia Summary</a>: <b>Henry VI, Part 1</b> is a history play by William Shakespeare, believed to have been written in 1591.",
		},
		{
			ID:              31,
			Title:           "Henry VI, Part 2",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Henry_VI,_Part_2\">Wikipedia Summary</a>: <b>Henry VI, Part 2</b> is a history play by William Shakespeare, believed to have been written in 1591.",
		},
		{
			ID:              32,
			Title:           "Henry VI, Part 3",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Henry_VI,_Part_3\">Wikipedia Summary</a>: <b>Henry VI, Part 3</b> is a history play by William Shakespeare, believed to have been written in 1591.",
		},
		{
			ID:              33,
			Title:           "Henry VIII",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Henry_VIII_(play)\">Wikipedia Summary</a>: <b>Henry VIII</b> is a collaborative history play, believed to have been written by William Shakespeare and John Fletcher in 1613.",
		},
		{
			ID:              34,
			Title:           "The Winter's Tale",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Winter%27s_Tale\">Wikipedia Summary</a>: <b>The Winter's Tale</b> is a play by William Shakespeare, originally published in the First Folio of 1623, and it is considered one of Shakespeare's late romances.",
		},
		{
			ID:              35,
			Title:           "The Two Noble Kinsmen",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Two_Noble_Kinsmen\">Wikipedia Summary</a>: <b>The Two Noble Kinsmen</b> is a play co-written by William Shakespeare and John Fletcher, published in 1634.",
		},
		{
			ID:              36,
			Title:           "Edward III",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Edward_III_(play)\">Wikipedia Summary</a>: <b>Edward III</b> is a play sometimes attributed to William Shakespeare, though its authorship is disputed.",
		},
		{
			ID:              37,
			Title:           "Thomas More",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Thomas_More_(play)\">Wikipedia Summary</a>: <b>Thomas More</b> is a play written by several playwrights, including William Shakespeare. It is based on the life of Sir Thomas More.",
		},
		{
			ID:              38,
			Title:           "Arden of Faversham",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Arden_of_Faversham\">Wikipedia Summary</a>: <b>Arden of Faversham</b> is an Elizabethan play sometimes attributed to William Shakespeare, revolving around the murder of Thomas Arden.",
		},
		{
			ID:              39,
			Title:           "Cardenio",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Cardenio\">Wikipedia Summary</a>: <b>Cardenio</b> is a lost play, attributed to William Shakespeare and John Fletcher, based on an episode in Miguel de Cervantes's Don Quixote.",
		},
		{
			ID:              40,
			Title:           "Sir Thomas More",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Sir_Thomas_More_(play)\">Wikipedia Summary</a>: <b>Sir Thomas More</b> is a play by several writers, including William Shakespeare, portraying the life of Thomas More.",
		},
		{
			ID:              41,
			Title:           "Fair Em",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Fair_Em\">Wikipedia Summary</a>: <b>Fair Em</b> is an Elizabethan play of uncertain authorship sometimes attributed to William Shakespeare.",
		},
		{
			ID:              42,
			Title:           "Mucedorus",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Mucedorus\">Wikipedia Summary</a>: <b>Mucedorus</b> is an Elizabethan play, sometimes attributed to William Shakespeare, that tells the tale of a shepherd prince and his love.",
		},
		{
			ID:              43,
			Title:           "The Birth of Merlin",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Birth_of_Merlin\">Wikipedia Summary</a>: <b>The Birth of Merlin</b> is a play sometimes attributed to William Shakespeare and William Rowley.",
		},
		{
			ID:              44,
			Title:           "The Merry Devil of Edmonton",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Merry_Devil_of_Edmonton\">Wikipedia Summary</a>: <b>The Merry Devil of Edmonton</b> is an anonymous Elizabethan comedy that has been occasionally attributed to William Shakespeare.",
		},
		{
			ID:              45,
			Title:           "The Arraignment of Paris",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Arraignment_of_Paris\">Wikipedia Summary</a>: <b>The Arraignment of Paris</b> is a play by George Peele, though once attributed to William Shakespeare.",
		},
		{
			ID:              46,
			Title:           "Locrine",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Locrine\">Wikipedia Summary</a>: <b>Locrine</b> is a play that was once attributed to William Shakespeare, though its authorship is disputed.",
		},
		{
			ID:              47,
			Title:           "The London Prodigal",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_London_Prodigal\">Wikipedia Summary</a>: <b>The London Prodigal</b> is a play included in the Shakespeare Apocrypha, attributed to Shakespeare, but authorship is uncertain.",
		},
		{
			ID:              48,
			Title:           "A Yorkshire Tragedy",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/A_Yorkshire_Tragedy\">Wikipedia Summary</a>: <b>A Yorkshire Tragedy</b> is a play included in the Shakespeare Apocrypha and was attributed to Shakespeare in 1608.",
		},
		{
			ID:              49,
			Title:           "The Puritan",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Puritan_(play)\">Wikipedia Summary</a>: <b>The Puritan</b> is a comedy sometimes attributed to William Shakespeare, though authorship is disputed.",
		},
		{
			ID:              50,
			Title:           "A Knack to Know a Knave",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/A_Knack_to_Know_a_Knave\">Wikipedia Summary</a>: <b>A Knack to Know a Knave</b> is an Elizabethan play sometimes included in the Shakespeare Apocrypha.",
		},
		{
			ID:              51,
			Title:           "The Troublesome Reign of King John",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Troublesome_Reign_of_King_John\">Wikipedia Summary</a>: <b>The Troublesome Reign of King John</b> is an anonymous Elizabethan play occasionally attributed to William Shakespeare.",
		},
		{
			ID:              52,
			Title:           "The Tragedy of Caesar and Pompey",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Tragedy_of_Caesar_and_Pompey\">Wikipedia Summary</a>: <b>The Tragedy of Caesar and Pompey</b> is a play occasionally attributed to Shakespeare but likely not his work.",
		},
		{
			ID:              53,
			Title:           "The Taming of A Shrew",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Taming_of_A_Shrew\">Wikipedia Summary</a>: <b>The Taming of A Shrew</b> is an anonymous play that parallels Shakespeare's The Taming of the Shrew.",
		},
		{
			ID:              54,
			Title:           "Edward IV",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Edward_IV_(play)\">Wikipedia Summary</a>: <b>Edward IV</b> is a play sometimes associated with Shakespeare's era, though its authorship is uncertain.",
		},
		{
			ID:              55,
			Title:           "Sir John Oldcastle",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Sir_John_Oldcastle_(play)\">Wikipedia Summary</a>: <b>Sir John Oldcastle</b> is a play sometimes attributed to William Shakespeare.",
		},
		{
			ID:              56,
			Title:           "The Famous Victories of Henry V",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Famous_Victories_of_Henry_V\">Wikipedia Summary</a>: <b>The Famous Victories of Henry V</b> is an anonymous play that portrays events later covered in Shakespeare's Henry IV and Henry V plays.",
		},
		{
			ID:              57,
			Title:           "The Second Maiden's Tragedy",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Second_Maiden%27s_Tragedy\">Wikipedia Summary</a>: <b>The Second Maiden's Tragedy</b> is a Jacobean play sometimes associated with Shakespeare.",
		},
		{
			ID:              58,
			Title:           "Philaster",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Philaster\">Wikipedia Summary</a>: <b>Philaster</b> is a play by Beaumont and Fletcher, sometimes associated with Shakespearean style.",
		},
		{
			ID:              59,
			Title:           "The Revenger's Tragedy",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Revenger%27s_Tragedy\">Wikipedia Summary</a>: <b>The Revenger's Tragedy</b> is a Jacobean play, sometimes attributed to Thomas Middleton but originally considered as possibly Shakespeare's work.",
		},
		{
			ID:              60,
			Title:           "The Spanish Tragedy",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Spanish_Tragedy\">Wikipedia Summary</a>: <b>The Spanish Tragedy</b> is an influential play by Thomas Kyd, sometimes linked to Shakespeare's work due to thematic similarities.",
		},
		{
			ID:              61,
			Title:           "The Maid's Tragedy",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Maid%27s_Tragedy\">Wikipedia Summary</a>: <b>The Maid's Tragedy</b> is a play by Beaumont and Fletcher, showcasing themes of betrayal and revenge that echo Shakespearean elements.",
		},
		{
			ID:              62,
			Title:           "The Changeling",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Changeling_(play)\">Wikipedia Summary</a>: <b>The Changeling</b> is a tragedy by Thomas Middleton and William Rowley, often compared to Shakespeare for its dark themes and psychological depth.",
		},
		{
			ID:              63,
			Title:           "The Roaring Girl",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Roaring_Girl\">Wikipedia Summary</a>: <b>The Roaring Girl</b> is a comedy by Thomas Middleton and Thomas Dekker, reflecting themes of gender roles and identity.",
		},
		{
			ID:              64,
			Title:           "The Duchess of Malfi",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Duchess_of_Malfi\">Wikipedia Summary</a>: <b>The Duchess of Malfi</b> is a tragedy by John Webster, celebrated for its poetic language and complex themes similar to Shakespearean tragedies.",
		},
		{
			ID:              65,
			Title:           "Doctor Faustus",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Doctor_Faustus_(play)\">Wikipedia Summary</a>: <b>Doctor Faustus</b> is a tragedy by Christopher Marlowe that explores themes of ambition and the supernatural, drawing comparisons to Shakespeare's work.",
		},
		{
			ID:              66,
			Title:           "The Alchemist",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Alchemist_(play)\">Wikipedia Summary</a>: <b>The Alchemist</b> is a comedy by Ben Jonson that satirizes greed and gullibility, sharing comedic traits with Shakespeare's works.",
		},
		{
			ID:              67,
			Title:           "Volpone",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Volpone\">Wikipedia Summary</a>: <b>Volpone</b> is a dark comedy by Ben Jonson, often discussed alongside Shakespearean works for its satirical tone.",
		},
		{
			ID:              68,
			Title:           "Bartholomew Fair",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Bartholomew_Fair_(play)\">Wikipedia Summary</a>: <b>Bartholomew Fair</b> is a comedy by Ben Jonson that explores human nature through humor, similar to Shakespeare’s comedies.",
		},
		{
			ID:              69,
			Title:           "The Jew of Malta",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Jew_of_Malta\">Wikipedia Summary</a>: <b>The Jew of Malta</b> is a play by Christopher Marlowe, known for its themes of revenge and ambition.",
		},
		{
			ID:              70,
			Title:           "Women Beware Women",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Women_Beware_Women\">Wikipedia Summary</a>: <b>Women Beware Women</b> is a tragedy by Thomas Middleton, often compared to Shakespeare for its tragic themes.",
		},
		{
			ID:              71,
			Title:           "The Revenger's Tragedy",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Revenger%27s_Tragedy\">Wikipedia Summary</a>: <b>The Revenger's Tragedy</b> is a dark play by Thomas Middleton, with themes reminiscent of Shakespearean revenge tragedies.",
		},
		{
			ID:              72,
			Title:           "A Chaste Maid in Cheapside",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/A_Chaste_Maid_in_Cheapside\">Wikipedia Summary</a>: <b>A Chaste Maid in Cheapside</b> is a satire by Thomas Middleton, noted for its humorous exploration of society similar to Shakespeare's comedies.",
		},
		{
			ID:              73,
			Title:           "The Honest Whore",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Honest_Whore\">Wikipedia Summary</a>: <b>The Honest Whore</b> is a two-part play by Thomas Dekker, exploring themes of morality and redemption.",
		},
		{
			ID:              74,
			Title:           "The Knight of the Burning Pestle",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Knight_of_the_Burning_Pestle\">Wikipedia Summary</a>: <b>The Knight of the Burning Pestle</b> is a satirical play by Francis Beaumont, noted for its humor and unique structure.",
		},
		{
			ID:              75,
			Title:           "The Spanish Tragedy",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Spanish_Tragedy\">Wikipedia Summary</a>: <b>The Spanish Tragedy</b> is a revenge play by Thomas Kyd that influenced Shakespeare's Hamlet.",
		},
		{
			ID:              76,
			Title:           "Arden of Faversham",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Arden_of_Faversham\">Wikipedia Summary</a>: <b>Arden of Faversham</b> is an Elizabethan play sometimes attributed to Shakespeare, known for its focus on domestic tragedy.",
		},
		{
			ID:              77,
			Title:           "Edward II",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Edward_II_(play)\">Wikipedia Summary</a>: <b>Edward II</b> is a historical tragedy by Christopher Marlowe, comparable to Shakespeare's history plays.",
		},
		{
			ID:              78,
			Title:           "The Shoemaker's Holiday",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Shoemaker%27s_Holiday\">Wikipedia Summary</a>: <b>The Shoemaker's Holiday</b> is a comedy by Thomas Dekker that offers a glimpse into the lives of common people.",
		},
		{
			ID:              79,
			Title:           "Dido, Queen of Carthage",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Dido,_Queen_of_Carthage_(play)\">Wikipedia Summary</a>: <b>Dido, Queen of Carthage</b> is a play by Christopher Marlowe that explores themes of love and betrayal.",
		},
		{
			ID:              80,
			Title:           "The White Devil",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_White_Devil\">Wikipedia Summary</a>: <b>The White Devil</b> is a tragedy by John Webster, noted for its dark themes and complex characters.",
		},
		{
			ID:              81,
			Title:           "The Witch",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Witch_(Middleton_play)\">Wikipedia Summary</a>: <b>The Witch</b> is a tragedy by Thomas Middleton, exploring themes of witchcraft and moral corruption.",
		},
		{
			ID:              82,
			Title:           "Philaster",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Philaster\">Wikipedia Summary</a>: <b>Philaster</b> is a romantic drama by Beaumont and Fletcher, known for its themes of love and identity.",
		},
		{
			ID:              83,
			Title:           "The Knight of Malta",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Knight_of_Malta\">Wikipedia Summary</a>: <b>The Knight of Malta</b> is a tragicomedy by John Fletcher and Philip Massinger.",
		},
		{
			ID:              84,
			Title:           "The Widow's Tears",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Widow%27s_Tears\">Wikipedia Summary</a>: <b>The Widow's Tears</b> is a dark comedy by George Chapman, known for its exploration of grief and resilience.",
		},
		{
			ID:              85,
			Title:           "The Atheist's Tragedy",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Atheist%27s_Tragedy\">Wikipedia Summary</a>: <b>The Atheist's Tragedy</b> is a play by Cyril Tourneur, exploring themes of morality and vengeance.",
		},
		{
			ID:              86,
			Title:           "Cupid's Revenge",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Cupid%27s_Revenge\">Wikipedia Summary</a>: <b>Cupid's Revenge</b> is a tragedy by Beaumont and Fletcher that explores the destructive nature of unrequited love.",
		},
		{
			ID:              87,
			Title:           "The Island Princess",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Island_Princess\">Wikipedia Summary</a>: <b>The Island Princess</b> is a tragicomedy by John Fletcher that explores themes of loyalty and honor.",
		},
		{
			ID:              88,
			Title:           "The Sea Voyage",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Sea_Voyage\">Wikipedia Summary</a>: <b>The Sea Voyage</b> is a tragicomedy by Fletcher and Massinger, known for its adventurous plot and exotic setting.",
		},
		{
			ID:              89,
			Title:           "Bonduca",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Bonduca\">Wikipedia Summary</a>: <b>Bonduca</b> is a historical tragedy by John Fletcher, focusing on the resistance of a British queen against Rome.",
		},
		{
			ID:              90,
			Title:           "The Scornful Lady",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Scornful_Lady\">Wikipedia Summary</a>: <b>The Scornful Lady</b> is a comedy by Beaumont and Fletcher that deals with themes of love and pride.",
		},
		{
			ID:              91,
			Title:           "The Faithful Shepherdess",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Faithful_Shepherdess\">Wikipedia Summary</a>: <b>The Faithful Shepherdess</b> is a pastoral tragicomedy by John Fletcher, exploring themes of love and innocence.",
		},
		{
			ID:              92,
			Title:           "The Parliament of Love",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Parliament_of_Love\">Wikipedia Summary</a>: <b>The Parliament of Love</b> is a comedy by Philip Massinger, exploring themes of love and virtue.",
		},
		{
			ID:              93,
			Title:           "The Emperor of the East",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Emperor_of_the_East\">Wikipedia Summary</a>: <b>The Emperor of the East</b> is a tragicomedy by Philip Massinger that examines themes of power and morality.",
		},
		{
			ID:              94,
			Title:           "The Fatal Dowry",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Fatal_Dowry\">Wikipedia Summary</a>: <b>The Fatal Dowry</b> is a tragedy by Philip Massinger and Nathan Field, noted for its examination of justice and honor.",
		},
		{
			ID:              95,
			Title:           "The Renegado",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Renegado\">Wikipedia Summary</a>: <b>The Renegado</b> is a tragicomedy by Philip Massinger, exploring themes of faith and redemption.",
		},
		{
			ID:              96,
			Title:           "A New Way to Pay Old Debts",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/A_New_Way_to_Pay_Old_Debts\">Wikipedia Summary</a>: <b>A New Way to Pay Old Debts</b> is a comedy by Philip Massinger, known for its satire of greed and corruption.",
		},
		{
			ID:              97,
			Title:           "The Roman Actor",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Roman_Actor\">Wikipedia Summary</a>: <b>The Roman Actor</b> is a tragedy by Philip Massinger that explores themes of tyranny and justice.",
		},
		{
			ID:              98,
			Title:           "The City Madam",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_City_Madam\">Wikipedia Summary</a>: <b>The City Madam</b> is a comedy by Philip Massinger that criticizes social ambition and greed.",
		},
		{
			ID:              99,
			Title:           "The Unnatural Combat",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/The_Unnatural_Combat\">Wikipedia Summary</a>: <b>The Unnatural Combat</b> is a tragedy by Philip Massinger that delves into themes of betrayal and family conflict.",
		},
		{
			ID:              100,
			Title:           "Believe as You List",
			DescriptionHtml: "<a href=\"https://en.wikipedia.org/wiki/Believe_as_You_List\">Wikipedia Summary</a>: <b>Believe as You List</b> is a play by Philip Massinger, exploring themes of loyalty and identity.",
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

func (s *Server) productRatingsHandler(w http.ResponseWriter, r *http.Request) {
	// Extrai o `id` do produto da URL usando `r.URL.Path`
	pathParts := strings.Split(r.URL.Path, "/")

	// Verifica se a URL tem pelo menos 5 partes para corresponder ao formato `/api/v1/products/{id}/ratings`
	if len(pathParts) < 5 {
		http.Error(w, "Invalid product URL", http.StatusBadRequest)
		return
	}

	// O `productID` está na quinta parte da URL (índice 4)
	productID := pathParts[4]

	// Converte o `productID` para um número inteiro
	id, err := strconv.Atoi(productID)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Chamada direta ao método `GetRatings` do componente `ratings`
	ratingsResponse, err := s.ratings.Get().GetRatings(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get product ratings: %v", err), http.StatusInternalServerError)
		return
	}

	// Serializa `ratingsResponse` para JSON
	response, err := json.Marshal(ratingsResponse)
	if err != nil {
		http.Error(w, "Failed to marshal product ratings", http.StatusInternalServerError)
		return
	}

	// Envia a resposta JSON ao cliente
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Product page is healthy")
}
