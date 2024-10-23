package ratings

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ServiceWeaver/weaver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	userAddedRatings = make(map[int]map[string]int) // in-memory ratings
	unavailable      = false
	healthy          = true
	db               *sql.DB
	mongoClient      *mongo.Client
)

// Definindo uma estrutura específica para o retorno dos ratings
type RatingResponse struct {
	weaver.AutoMarshal
	ID      int            `json:"id"`
	Ratings map[string]int `json:"ratings"`
}

type Ratings interface {
	GetRatings(ctx context.Context, productId int) (RatingResponse, error)
	PostRatings(ctx context.Context, productIdStr string, requestBody []byte) (RatingResponse, error)
}

type ratings struct {
	weaver.Implements[Ratings]
}

// Função de inicialização para lidar com as variáveis de ambiente e configurar banco de dados
func (r *ratings) Init(ctx context.Context) error {
	if os.Getenv("SERVICE_VERSION") == "v-unavailable" {
		// make the service unavailable once in 60 seconds
		go func() {
			for {
				unavailable = !unavailable
				time.Sleep(60 * time.Second)
			}
		}()
	}

	if os.Getenv("SERVICE_VERSION") == "v-unhealthy" {
		// make the service unhealthy every 15 minutes
		go func() {
			for {
				healthy = !healthy
				unavailable = !unavailable
				time.Sleep(15 * time.Minute)
			}
		}()
	}

	// Configura conexão com o banco de dados MySQL
	if os.Getenv("SERVICE_VERSION") == "v2" {
		dbType := os.Getenv("DB_TYPE")
		if dbType == "mysql" {
			var err error
			host := os.Getenv("MYSQL_DB_HOST")
			port := os.Getenv("MYSQL_DB_PORT")
			user := os.Getenv("MYSQL_DB_USER")
			password := os.Getenv("MYSQL_DB_PASSWORD")
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/ratingsdb", user, password, host, port)

			db, err = sql.Open("mysql", dsn)
			if err != nil {
				log.Fatal("Could not connect to MySQL database:", err)
			}
		}
	}

	return nil
}

// Função para obter ratings
func (r *ratings) GetRatings(ctx context.Context, productId int) (RatingResponse, error) {
	if os.Getenv("SERVICE_VERSION") == "v-unavailable" || os.Getenv("SERVICE_VERSION") == "v-unhealthy" {
		if unavailable {
			return RatingResponse{}, fmt.Errorf("service unavailable")
		}
	}

	// Versão com banco de dados
	if os.Getenv("SERVICE_VERSION") == "v2" {
		var firstRating, secondRating int

		// Conectar ao banco MySQL
		if os.Getenv("DB_TYPE") == "mysql" {
			err := db.Ping()
			if err != nil {
				return RatingResponse{}, fmt.Errorf("could not connect to ratings database")
			}

			rows, err := db.Query("SELECT Rating FROM ratings LIMIT 2")
			if err != nil {
				return RatingResponse{}, fmt.Errorf("could not perform select")
			}
			defer rows.Close()

			count := 0
			for rows.Next() {
				if count == 0 {
					err = rows.Scan(&firstRating)
				} else if count == 1 {
					err = rows.Scan(&secondRating)
				}
				if err != nil {
					return RatingResponse{}, fmt.Errorf("could not retrieve ratings")
				}
				count++
			}

			if count == 0 {
				return RatingResponse{}, fmt.Errorf("ratings not found")
			}

		} else { // Conectar ao MongoDB
			var err error
			mongoURL := os.Getenv("MONGO_DB_URL")

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
			if err != nil {
				log.Fatal("Could not connect to MongoDB:", err)
			}

			collection := mongoClient.Database("test").Collection("ratings")
			cursor, err := collection.Find(ctx, bson.M{})
			if err != nil {
				return RatingResponse{}, fmt.Errorf("could not connect to ratings database")
			}
			defer cursor.Close(ctx)

			var ratingsData []bson.M
			if err = cursor.All(ctx, &ratingsData); err != nil {
				return RatingResponse{}, fmt.Errorf("could not parse ratings data")
			}

			firstRating, secondRating = 0, 0

			if len(ratingsData) > 0 {
				if val, ok := ratingsData[0]["rating"].(int32); ok {
					firstRating = int(val)
				}
			}

			if len(ratingsData) > 1 {
				if val, ok := ratingsData[1]["rating"].(int32); ok {
					secondRating = int(val)
				}
			}
		}

		// Montar a resposta usando a nova estrutura
		return RatingResponse{
			ID: productId,
			Ratings: map[string]int{
				"Reviewer1": firstRating,
				"Reviewer2": secondRating,
			},
		}, nil

	} else {
		// Se não for a versão v2, usar dados locais
		return getLocalReviews(productId), nil
	}
}

// Função para postar ratings
func (r *ratings) PostRatings(ctx context.Context, productIdStr string, requestBody []byte) (RatingResponse, error) {
	productId, err := strconv.Atoi(productIdStr)
	if err != nil {
		return RatingResponse{}, fmt.Errorf("please provide numeric product ID")
	}

	var ratings map[string]int
	if err := json.Unmarshal(requestBody, &ratings); err != nil {
		return RatingResponse{}, fmt.Errorf("please provide valid ratings JSON")
	}

	if os.Getenv("SERVICE_VERSION") == "v2" {
		return RatingResponse{}, fmt.Errorf("Post not implemented for database backed ratings")
	}

	// Processar avaliações localmente e retornar o resultado
	return putLocalReviews(productId, ratings), nil
}

func putLocalReviews(productId int, ratings map[string]int) RatingResponse {
	userAddedRatings[productId] = ratings
	return getLocalReviews(productId)
}

func getLocalReviews(productId int) RatingResponse {
	if val, ok := userAddedRatings[productId]; ok {
		return RatingResponse{ID: productId, Ratings: val}
	}

	return RatingResponse{
		ID: productId,
		Ratings: map[string]int{
			"Reviewer1": 5,
			"Reviewer2": 4,
		},
	}
}
