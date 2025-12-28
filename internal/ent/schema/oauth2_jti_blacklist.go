package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// OAuth2JTIBlacklist holds the schema definition for JWT Token ID blacklist.
// Used to prevent JWT replay attacks by tracking used JTIs.
type OAuth2JTIBlacklist struct {
	ent.Schema
}

// Annotations of the OAuth2JTIBlacklist.
func (OAuth2JTIBlacklist) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth2_jti_blacklist", Schema: "identify"},
	}
}

// Fields of the OAuth2JTIBlacklist.
func (OAuth2JTIBlacklist) Fields() []ent.Field {
	return []ent.Field{
		field.String("signature").
			NotEmpty().
			Unique().
			Comment("JWT ID (jti) claim - unique identifier for the token"),
		field.Time("expires_at").
			Comment("When this JTI entry expires and can be cleaned up"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Indexes of the OAuth2JTIBlacklist.
func (OAuth2JTIBlacklist) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("expires_at"),
	}
}

// Edges of the OAuth2JTIBlacklist.
func (OAuth2JTIBlacklist) Edges() []ent.Edge {
	return nil
}
