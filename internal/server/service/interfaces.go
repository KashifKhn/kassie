package service

import (
	"github.com/KashifKhn/kassie/internal/server/db"
	"github.com/KashifKhn/kassie/internal/server/state"
	"github.com/KashifKhn/kassie/internal/shared/config"
	"github.com/gocql/gocql"
)

type SessionStore interface {
	Create(id string, profile *config.Profile, conn *db.Session) *state.Session
	Get(id string) (*state.Session, error)
	Delete(id string)
	CloseAll()
	Close()
}

type ConnectionPool interface {
	GetOrCreate(profileName string, cfg *db.ConnectionConfig) (*gocql.Session, error)
}

type ProfileProvider interface {
	GetProfile(name string) (*config.Profile, error)
	GetProfiles() []config.Profile
}
