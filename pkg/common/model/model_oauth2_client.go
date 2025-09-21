package model

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "time"
)

// OAuth2Client represents an OAuth2 client application
type OAuth2Client struct {
    ID                        string    `json:"id" db:"id"`
    ClientSecretHash          string    `json:"-" db:"client_secret_hash"` // Never expose in JSON
    RedirectURIs              []string  `json:"redirect_uris" db:"redirect_uris"`
    GrantTypes                []string  `json:"grant_types" db:"grant_types"`
    ResponseTypes             []string  `json:"response_types" db:"response_types"`
    Scopes                    []string  `json:"scopes" db:"scopes"`
    Audience                  []string  `json:"audience" db:"audience"`
    Public                    bool      `json:"public" db:"public"`
    ClientName                *string   `json:"client_name,omitempty" db:"client_name"`
    ClientURI                 *string   `json:"client_uri,omitempty" db:"client_uri"`
    LogoURI                   *string   `json:"logo_uri,omitempty" db:"logo_uri"`
    Contacts                  []string  `json:"contacts" db:"contacts"`
    TOSURI                    *string   `json:"tos_uri,omitempty" db:"tos_uri"`
    PolicyURI                 *string   `json:"policy_uri,omitempty" db:"policy_uri"`
    JWKSURI                   *string   `json:"jwks_uri,omitempty" db:"jwks_uri"`
    JWKS                      *string   `json:"jwks,omitempty" db:"jwks"`
    SectorIdentifierURI       *string   `json:"sector_identifier_uri,omitempty" db:"sector_identifier_uri"`
    SubjectType               *string   `json:"subject_type,omitempty" db:"subject_type"`
    TokenEndpointAuthMethod   *string   `json:"token_endpoint_auth_method,omitempty" db:"token_endpoint_auth_method"`
    UserinfoSignedResponseAlg *string   `json:"userinfo_signed_response_alg,omitempty" db:"userinfo_signed_response_alg"`
    CreatedAt                 time.Time `json:"created_at" db:"created_at"`
    UpdatedAt                 time.Time `json:"updated_at" db:"updated_at"`
}

// StringArray handles PostgreSQL array types in Go
type StringArray []string

// Value implements driver.Valuer interface
func (a *StringArray) Value() (driver.Value, error) {
    if len(*a) == 0 {
        return "{}", nil
    }
    jsonData, err := json.Marshal(*a)
    if err != nil {
        return nil, err
    }
    return fmt.Sprintf("{%s}", string(jsonData)), nil
}

// Scan implements sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
    if value == nil {
        *a = StringArray{}
        return nil
    }

    switch v := value.(type) {
    case []byte:
        return json.Unmarshal(v, a)
    case string:
        return json.Unmarshal([]byte(v), a)
    default:
        return fmt.Errorf("cannot scan %T into StringArray", value)
    }
}
