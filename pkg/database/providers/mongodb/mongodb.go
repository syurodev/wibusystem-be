package mongodb

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"wibusystem/pkg/database/config"
	"wibusystem/pkg/database/interfaces"
)

// MongoProvider implements the DocumentDatabase interface
type MongoProvider struct {
	client   *mongo.Client
	database *mongo.Database
	config   *config.DocumentConfig
}

// NewMongoProvider creates a new MongoDB database provider
func NewMongoProvider(config *config.DocumentConfig) (*MongoProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("mongodb config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid mongodb config: %w", err)
	}

	return &MongoProvider{
		config: config,
	}, nil
}

// Connect establishes connection to MongoDB
func (m *MongoProvider) Connect(ctx context.Context) error {
	clientOptions := options.Client().ApplyURI(m.config.URI)

	// Set connection pool options
	if m.config.MaxPoolSize != nil {
		clientOptions.SetMaxPoolSize(*m.config.MaxPoolSize)
	}
	if m.config.MinPoolSize != nil {
		clientOptions.SetMinPoolSize(*m.config.MinPoolSize)
	}
	if m.config.MaxConnIdleTime != nil {
		clientOptions.SetMaxConnIdleTime(*m.config.MaxConnIdleTime)
	}

	// Set timeout options
	clientOptions.SetConnectTimeout(m.config.ConnectTimeout)
	clientOptions.SetServerSelectionTimeout(m.config.ServerSelectionTimeout)
	clientOptions.SetSocketTimeout(m.config.SocketTimeout)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	m.database = client.Database(m.config.Database)

	log.Printf("Successfully connected to MongoDB database: %s", m.config.Database)
	return nil
}

// Close closes the MongoDB connection
func (m *MongoProvider) Close() error {
	if m.client != nil {
		ctx := context.Background()
		if err := m.client.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
		}
		log.Printf("MongoDB connection closed for database: %s", m.config.Database)
	}
	return nil
}

// Health checks if the MongoDB connection is healthy
func (m *MongoProvider) Health(ctx context.Context) error {
	if m.client == nil {
		return fmt.Errorf("mongodb provider not connected")
	}
	return m.client.Ping(ctx, nil)
}

// GetType returns the database type
func (m *MongoProvider) GetType() interfaces.DatabaseType {
	return interfaces.MongoDB
}

// BeginTx starts a new session (MongoDB transactions require sessions)
func (m *MongoProvider) BeginTx(ctx context.Context) (interfaces.Transaction, error) {
	if m.client == nil {
		return nil, fmt.Errorf("mongodb provider not connected")
	}

	session, err := m.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start MongoDB session: %w", err)
	}

	// Start transaction
	if err := session.StartTransaction(); err != nil {
		session.EndSession(ctx)
		return nil, fmt.Errorf("failed to start MongoDB transaction: %w", err)
	}

	return &MongoTransaction{session: session}, nil
}

// CreateCollection creates a new collection
func (m *MongoProvider) CreateCollection(ctx context.Context, name string) error {
	if m.database == nil {
		return fmt.Errorf("mongodb provider not connected")
	}

	return m.database.CreateCollection(ctx, name)
}

// DropCollection drops a collection
func (m *MongoProvider) DropCollection(ctx context.Context, name string) error {
	if m.database == nil {
		return fmt.Errorf("mongodb provider not connected")
	}

	collection := m.database.Collection(name)
	return collection.Drop(ctx)
}

// ListCollections lists all collections in the database
func (m *MongoProvider) ListCollections(ctx context.Context) ([]string, error) {
	if m.database == nil {
		return nil, fmt.Errorf("mongodb provider not connected")
	}

	cursor, err := m.database.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	return cursor, nil
}

// CreateIndex creates an index on a collection
func (m *MongoProvider) CreateIndex(ctx context.Context, collectionName string, index interfaces.IndexDefinition) error {
	if m.database == nil {
		return fmt.Errorf("mongodb provider not connected")
	}

	collection := m.database.Collection(collectionName)

	// Convert IndexDefinition to MongoDB index model
	keys := bson.D{}
	for field, order := range index.Fields {
		keys = append(keys, bson.E{Key: field, Value: order})
	}

	indexModel := mongo.IndexModel{
		Keys: keys,
	}

	// Set index options
	indexOptions := options.Index()
	if index.Name != "" {
		indexOptions.SetName(index.Name)
	}
	if index.Unique {
		indexOptions.SetUnique(true)
	}
	if index.Sparse {
		indexOptions.SetSparse(true)
	}
	if index.TTL != nil {
		indexOptions.SetExpireAfterSeconds(int32(index.TTL.Seconds()))
	}

	indexModel.Options = indexOptions

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// DropIndex drops an index from a collection
func (m *MongoProvider) DropIndex(ctx context.Context, collectionName, indexName string) error {
	if m.database == nil {
		return fmt.Errorf("mongodb provider not connected")
	}

	collection := m.database.Collection(collectionName)
	_, err := collection.Indexes().DropOne(ctx, indexName)
	if err != nil {
		return fmt.Errorf("failed to drop index: %w", err)
	}

	return nil
}

// GetDatabase returns the MongoDB database instance for advanced operations
func (m *MongoProvider) GetDatabase() *mongo.Database {
	return m.database
}

// GetClient returns the MongoDB client instance for advanced operations
func (m *MongoProvider) GetClient() *mongo.Client {
	return m.client
}

// GetCollection returns a collection for direct operations
func (m *MongoProvider) GetCollection(name string) *mongo.Collection {
	if m.database == nil {
		return nil
	}
	return m.database.Collection(name)
}
