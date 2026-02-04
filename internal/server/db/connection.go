package db

import (
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/KashifKhn/kassie/internal/shared/config"
	"github.com/gocql/gocql"
)

var (
	ErrNoHosts          = fmt.Errorf("no hosts configured")
	ErrInvalidPort      = fmt.Errorf("invalid port number")
	ErrConnectionFailed = fmt.Errorf("failed to connect to cluster")
	ErrPoolClosed       = fmt.Errorf("connection pool is closed")
)

type ConnectionConfig struct {
	Hosts         []string
	Port          int
	Keyspace      string
	Username      string
	Password      string
	Consistency   gocql.Consistency
	Timeout       time.Duration
	PoolSize      int
	SSLEnabled    bool
	SSLCertPath   string
	SSLKeyPath    string
	SSLCAPath     string
	SSLSkipVerify bool
}

type Pool struct {
	mu          sync.RWMutex
	connections map[string]*gocql.Session
	configs     map[string]*ConnectionConfig
	closed      bool
}

func NewPool() *Pool {
	return &Pool{
		connections: make(map[string]*gocql.Session),
		configs:     make(map[string]*ConnectionConfig),
		closed:      false,
	}
}

func (p *Pool) GetOrCreate(profileName string, cfg *ConnectionConfig) (*gocql.Session, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, ErrPoolClosed
	}

	if session, exists := p.connections[profileName]; exists {
		if !session.Closed() {
			return session, nil
		}
		delete(p.connections, profileName)
	}

	session, err := createSession(cfg)
	if err != nil {
		return nil, err
	}

	p.connections[profileName] = session
	p.configs[profileName] = cfg

	return session, nil
}

func (p *Pool) Get(profileName string) (*gocql.Session, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, ErrPoolClosed
	}

	session, exists := p.connections[profileName]
	if !exists {
		return nil, fmt.Errorf("connection not found for profile: %s", profileName)
	}

	if session.Closed() {
		return nil, fmt.Errorf("connection closed for profile: %s", profileName)
	}

	return session, nil
}

func (p *Pool) Close(profileName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, exists := p.connections[profileName]
	if !exists {
		return nil
	}

	session.Close()
	delete(p.connections, profileName)
	delete(p.configs, profileName)

	return nil
}

func (p *Pool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, session := range p.connections {
		session.Close()
	}

	p.connections = make(map[string]*gocql.Session)
	p.configs = make(map[string]*ConnectionConfig)
	p.closed = true
}

func (p *Pool) ListProfiles() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	profiles := make([]string, 0, len(p.connections))
	for name := range p.connections {
		profiles = append(profiles, name)
	}

	return profiles
}

func createSession(cfg *ConnectionConfig) (*gocql.Session, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Port = cfg.Port
	cluster.Keyspace = cfg.Keyspace
	cluster.Consistency = cfg.Consistency
	cluster.Timeout = cfg.Timeout
	cluster.NumConns = cfg.PoolSize

	if cfg.Username != "" && cfg.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	if cfg.SSLEnabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: cfg.SSLSkipVerify,
		}
		cluster.SslOpts = &gocql.SslOptions{
			Config: tlsConfig,
		}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	return session, nil
}

func validateConfig(cfg *ConnectionConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if len(cfg.Hosts) == 0 {
		return ErrNoHosts
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return ErrInvalidPort
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}

	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 5
	}

	if cfg.Consistency == 0 {
		cfg.Consistency = gocql.Quorum
	}

	return nil
}

func ProfileToConnectionConfig(profile *config.Profile) *ConnectionConfig {
	cfg := &ConnectionConfig{
		Hosts:       profile.Hosts,
		Port:        profile.Port,
		Keyspace:    profile.Keyspace,
		Consistency: gocql.Quorum,
		Timeout:     10 * time.Second,
		PoolSize:    5,
	}

	if profile.Auth != nil {
		cfg.Username = profile.Auth.Username
		cfg.Password = profile.Auth.Password
	}

	if profile.SSL != nil {
		cfg.SSLEnabled = profile.SSL.Enabled
		cfg.SSLCertPath = profile.SSL.CertPath
		cfg.SSLKeyPath = profile.SSL.KeyPath
		cfg.SSLCAPath = profile.SSL.CAPath
		cfg.SSLSkipVerify = profile.SSL.InsecureSkipVerify
	}

	return cfg
}
