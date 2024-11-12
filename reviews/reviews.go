package reviews

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/ServiceWeaver/weaver"
	"github.com/camilamedeir0s/bookinfo-serviceweaver/ratings"
)

type Review struct {
	weaver.AutoMarshal
	Reviewer string `json:"reviewer"`
	Text     string `json:"text"`
	Rating   Rating `json:"rating,omitempty"`
}

type Rating struct {
	weaver.AutoMarshal
	Stars int    `json:"stars"`
	Color string `json:"color"`
}

type Response struct {
	weaver.AutoMarshal
	ID          string   `json:"id"`
	PodName     string   `json:"podname"`
	ClusterName string   `json:"clustername"`
	Reviews     []Review `json:"reviews"`
}

var (
	starColor      = getEnv("STAR_COLOR", "black")
	podHostname    = getEnv("HOSTNAME", "unknown")
	clusterName    = getEnv("CLUSTER_NAME", "unknown")
	ratingsEnabled = getEnvAsBool("ENABLE_RATINGS", true)
)

// Definição do componente Reviews
type Reviews interface {
	BookReviewsByID(ctx context.Context, productId string) (Response, error)
}

type reviews struct {
	weaver.Implements[Reviews]
	ratingsComponent weaver.Ref[ratings.Ratings] // Referência ao componente Ratings
}

// Função principal para buscar reviews por ID de produto
func (r *reviews) BookReviewsByID(ctx context.Context, productId string) (Response, error) {
	starsReviewer1 := -1
	starsReviewer2 := -1

	// Verifica se as avaliações estão habilitadas
	if ratingsEnabled {
		ratingsResponse, err := r.getRatings(ctx, productId)
		if err == nil {
			// Extrai as estrelas para cada reviewer do ratingsResponse
			if reviewer1, exists := ratingsResponse.Ratings["Reviewer1"]; exists {
				starsReviewer1 = reviewer1
			}
			if reviewer2, exists := ratingsResponse.Ratings["Reviewer2"]; exists {
				starsReviewer2 = reviewer2
			}
		} else {
			// Log de erro ou tratamento de erro pode ser adicionado aqui
			fmt.Printf("Error getting ratings: %v\n", err)
		}
	}

	// Gera a resposta final com as reviews e os ratings
	response := r.getJsonResponse(productId, starsReviewer1, starsReviewer2)
	return response, nil
}

func (r *reviews) getRatings(ctx context.Context, productId string) (ratings.RatingResponse, error) {
	// Converte productId de string para int
	productIdInt, err := strconv.Atoi(productId)
	if err != nil {
		return ratings.RatingResponse{}, fmt.Errorf("invalid product ID: %w", err)
	}

	// Chama o componente Ratings diretamente com o productId convertido
	ratingsComponent := r.ratingsComponent.Get()
	ratingResponse, err := ratingsComponent.GetRatings(ctx, productIdInt)
	if err != nil {
		return ratings.RatingResponse{}, fmt.Errorf("error getting ratings: %w", err)
	}
	return ratingResponse, nil
}

// Função para gerar a resposta JSON
func (r *reviews) getJsonResponse(productId string, starsReviewer1, starsReviewer2 int) Response {
	reviews := []Review{
		{
			Reviewer: "Reviewer1",
			Text:     "An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!",
		},
		{
			Reviewer: "Reviewer2",
			Text:     "Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare.",
		},
		{
			Reviewer: "Reviewer3",
			Text:     "A thought-provoking play with complex characters and captivating dialogues.",
		},
		{
			Reviewer: "Reviewer4",
			Text:     "Engaging storyline but falls short on character development.",
		},
		{
			Reviewer: "Reviewer5",
			Text:     "An interesting mix of humor and tragedy, with moments of brilliance.",
		},
		{
			Reviewer: "Reviewer6",
			Text:     "The plot is intriguing, though some parts feel rushed.",
		},
		{
			Reviewer: "Reviewer7",
			Text:     "A fine piece of literature with deep thematic elements.",
		},
		{
			Reviewer: "Reviewer8",
			Text:     "One of the less popular works, but it deserves more recognition.",
		},
		{
			Reviewer: "Reviewer9",
			Text:     "The humor is well-placed, though it may not be for everyone.",
		},
		{
			Reviewer: "Reviewer10",
			Text:     "A beautifully written play that captures the essence of human emotions.",
		},
	}

	// Verifica se as avaliações estão habilitadas e define os ratings
	if ratingsEnabled {
		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				if starsReviewer1 != -1 {
					reviews[i].Rating = Rating{Stars: starsReviewer1, Color: starColor}
				} else {
					reviews[i].Rating = Rating{Stars: -1, Color: "Ratings service is unavailable"}
				}
			} else {
				if starsReviewer2 != -1 {
					reviews[i].Rating = Rating{Stars: starsReviewer2, Color: starColor}
				} else {
					reviews[i].Rating = Rating{Stars: -1, Color: "Ratings service is unavailable"}
				}
			}
		}
	}

	// Retorna a estrutura de resposta
	return Response{
		ID:          productId,
		PodName:     podHostname,
		ClusterName: clusterName,
		Reviews:     reviews,
	}
}

// Funções utilitárias para variáveis de ambiente
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return value == "true"
	}
	return fallback
}
