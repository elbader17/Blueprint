package generator

const PostgresBaseTemplate = `package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	List(ctx context.Context, collection string) ([]map[string]interface{}, error)
	Get(ctx context.Context, collection, id string) (map[string]interface{}, error)
	Create(ctx context.Context, collection string, data map[string]interface{}) (string, error)
	Update(ctx context.Context, collection, id string, data map[string]interface{}) error
	Delete(ctx context.Context, collection, id string) error
	Close()
}

type PostgresRepository struct {
	Pool *pgxpool.Pool
}

func NewPostgresRepository(url string) (Repository, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return &PostgresRepository{Pool: pool}, nil
}

func (r *PostgresRepository) Close() {
	if r.Pool != nil {
		r.Pool.Close()
	}
}

// Helper methods for generic operations (simplified for this template)
func (r *PostgresRepository) List(ctx context.Context, table string) ([]map[string]interface{}, error) {
	// Implementation would use dynamic SQL
	return nil, fmt.Errorf("generic List not implemented for Postgres adapter")
}

func (r *PostgresRepository) Get(ctx context.Context, table, id string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("generic Get not implemented for Postgres adapter")
}

func (r *PostgresRepository) Create(ctx context.Context, table string, data map[string]interface{}) (string, error) {
	return "", fmt.Errorf("generic Create not implemented for Postgres adapter")
}

func (r *PostgresRepository) Update(ctx context.Context, table, id string, data map[string]interface{}) error {
	return fmt.Errorf("generic Update not implemented for Postgres adapter")
}

func (r *PostgresRepository) Delete(ctx context.Context, table, id string) error {
	return fmt.Errorf("generic Delete not implemented for Postgres adapter")
}
`

const PostgresRepoTemplate = `package db

import (
	"context"
	"fmt"
	"{{.ProjectName}}/internal/domain"
)

type {{.Model.Name | title}}Repository struct {
	repo *PostgresRepository
}

func New{{.Model.Name | title}}Repository(repo *PostgresRepository) *{{.Model.Name | title}}Repository {
	_, err := repo.Pool.Exec(context.Background(), "{{.CreateTableSQL}}")
	if err != nil {
		fmt.Printf("Error creating table {{.Model.Name}}: %v\n", err)
	}
	return &{{.Model.Name | title}}Repository{repo: repo}
}

func (r *{{.Model.Name | title}}Repository) List(ctx context.Context) ([]*domain.{{.Model.Name | title}}, error) {
	rows, err := r.repo.Pool.Query(ctx, "SELECT {{.SelectColumns}} FROM {{.Model.Name}}")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.{{.Model.Name | title}}
	for rows.Next() {
		var m domain.{{.Model.Name | title}}
		fields := []interface{}{&m.ID}
		{{range $f := .Fields}}
		fields = append(fields, &m.{{$f | pascal}})
		{{end}}
		
		if err := rows.Scan(fields...); err != nil {
			return nil, err
		}
		results = append(results, &m)
	}
	return results, nil
}

func (r *{{.Model.Name | title}}Repository) Get(ctx context.Context, id string) (*domain.{{.Model.Name | title}}, error) {
	var m domain.{{.Model.Name | title}}
	fields := []interface{}{&m.ID}
	{{range $f := .Fields}}
	fields = append(fields, &m.{{$f | pascal}})
	{{end}}

	query := "SELECT {{.SelectColumns}} FROM {{.Model.Name}} WHERE id = $1"
	err := r.repo.Pool.QueryRow(ctx, query, id).Scan(fields...)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *{{.Model.Name | title}}Repository) Create(ctx context.Context, m *domain.{{.Model.Name | title}}) (string, error) {
	query := "INSERT INTO {{.Model.Name}} ({{.InsertColumns}}) VALUES ({{.InsertPlaceholders}}) RETURNING id"
	
	values := []interface{}{
		{{range $f := .Fields}}m.{{$f | pascal}},
		{{end}}
	}

	var id string
	err := r.repo.Pool.QueryRow(ctx, query, values...).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *{{.Model.Name | title}}Repository) Update(ctx context.Context, id string, m *domain.{{.Model.Name | title}}) error {
	query := "UPDATE {{.Model.Name}} SET {{.UpdateSet}} WHERE id = ${{add .TotalFields 1}}"
	
	values := []interface{}{
		{{range $f := .Fields}}m.{{$f | pascal}},
		{{end}}
		id,
	}

	_, err := r.repo.Pool.Exec(ctx, query, values...)
	return err
}

func (r *{{.Model.Name | title}}Repository) Delete(ctx context.Context, id string) error {
	_, err := r.repo.Pool.Exec(ctx, "DELETE FROM {{.Model.Name}} WHERE id = $1", id)
	return err
}
`

const MongoBaseTemplate = `package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository interface {
	List(ctx context.Context, collection string) ([]map[string]interface{}, error)
	Get(ctx context.Context, collection, id string) (map[string]interface{}, error)
	Create(ctx context.Context, collection string, data map[string]interface{}) (string, error)
	Update(ctx context.Context, collection, id string, data map[string]interface{}) error
	Delete(ctx context.Context, collection, id string) error
	Close()
}

type MongoRepository struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewMongoRepository(url, dbName string) (Repository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}

	return &MongoRepository{
		Client: client,
		DB:     client.Database(dbName),
	}, nil
}

func (r *MongoRepository) Close() {
	if r.Client != nil {
		r.Client.Disconnect(context.Background())
	}
}

func (r *MongoRepository) List(ctx context.Context, collection string) ([]map[string]interface{}, error) {
	cursor, err := r.DB.Collection(collection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *MongoRepository) Get(ctx context.Context, collection, id string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := r.DB.Collection(collection).FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	return result, err
}

func (r *MongoRepository) Create(ctx context.Context, collection string, data map[string]interface{}) (string, error) {
	res, err := r.DB.Collection(collection).InsertOne(ctx, data)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", res.InsertedID), nil
}

func (r *MongoRepository) Update(ctx context.Context, collection, id string, data map[string]interface{}) error {
	_, err := r.DB.Collection(collection).UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": data})
	return err
}

func (r *MongoRepository) Delete(ctx context.Context, collection, id string) error {
	_, err := r.DB.Collection(collection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}
`

const MongoRepoTemplate = `package db

import (
	"context"
	"{{.ProjectName}}/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type {{.Model.Name | title}}Repository struct {
	repo *MongoRepository
}

func New{{.Model.Name | title}}Repository(repo *MongoRepository) *{{.Model.Name | title}}Repository {
	return &{{.Model.Name | title}}Repository{repo: repo}
}

func (r *{{.Model.Name | title}}Repository) List(ctx context.Context) ([]*domain.{{.Model.Name | title}}, error) {
	cursor, err := r.repo.DB.Collection("{{.Model.Name}}").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var results []*domain.{{.Model.Name | title}}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *{{.Model.Name | title}}Repository) Get(ctx context.Context, id string) (*domain.{{.Model.Name | title}}, error) {
	objID, _ := primitive.ObjectIDFromHex(id)
	var m domain.{{.Model.Name | title}}
	err := r.repo.DB.Collection("{{.Model.Name}}").FindOne(ctx, bson.M{"_id": objID}).Decode(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *{{.Model.Name | title}}Repository) Create(ctx context.Context, m *domain.{{.Model.Name | title}}) (string, error) {
	res, err := r.repo.DB.Collection("{{.Model.Name}}").InsertOne(ctx, m)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *{{.Model.Name | title}}Repository) Update(ctx context.Context, id string, m *domain.{{.Model.Name | title}}) error {
	objID, _ := primitive.ObjectIDFromHex(id)
	_, err := r.repo.DB.Collection("{{.Model.Name}}").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": m})
	return err
}

func (r *{{.Model.Name | title}}Repository) Delete(ctx context.Context, id string) error {
	objID, _ := primitive.ObjectIDFromHex(id)
	_, err := r.repo.DB.Collection("{{.Model.Name}}").DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
`
