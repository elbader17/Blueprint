package domain

// Config represents the top-level structure of the blueprint JSON
type Config struct {
	ProjectName        string           `json:"project_name"`
	Database           Database         `json:"database"`
	FirestoreProjectID string           `json:"firestore_project_id,omitempty"` // Deprecated: use Database.ProjectID
	Auth               *Auth            `json:"auth,omitempty"`
	Payments           *Payments        `json:"payments,omitempty"`
	Pagination         *Pagination      `json:"pagination,omitempty"`
	RateLimit          *RateLimitConfig `json:"rate_limit,omitempty"`
	Models             []Model          `json:"models"`
}

// Pagination configures the default pagination settings
type Pagination struct {
	DefaultLimit int `json:"default_limit"`
}

// Database configures the database driver
type Database struct {
	Type      string `json:"type"`                 // "firestore", "postgresql", "mongodb"
	ProjectID string `json:"project_id,omitempty"` // For Firestore
	URL       string `json:"url,omitempty"`        // For Postgres/Mongo
}

// Auth configures the authentication module
type Auth struct {
	Enabled        bool   `json:"enabled"`
	Provider       string `json:"provider"` // "firebase" (default) or "jwt"
	UserCollection string `json:"user_collection"`
}

// Payments configures the payment module
type Payments struct {
	Enabled          bool   `json:"enabled"`
	Provider         string `json:"provider"` // e.g., "mercadopago"
	TransactionsColl string `json:"transactions_collection"`
}

// Validation represents field validation rules
type Validation struct {
	Required    bool     `json:"required,omitempty"`
	Min         *float64 `json:"min,omitempty"`
	Max         *float64 `json:"max,omitempty"`
	Regex       string   `json:"regex,omitempty"`
	EnumValues  []string `json:"enum_values,omitempty"`
}

// Model represents a data model definition
type Model struct {
	Name       string            `json:"name"`
	Protected  bool              `json:"protected"`
	SoftDelete bool             `json:"soft_delete,omitempty"`
	Fields     map[string]string `json:"fields"`
	Relations  map[string]string `json:"relations"`
	Validations map[string]Validation `json:"validations,omitempty"`
}

// RateLimitConfig configures rate limiting
type RateLimitConfig struct {
	Enabled           bool `json:"enabled"`
	RequestsPerMinute int  `json:"requests_per_minute"`
}
