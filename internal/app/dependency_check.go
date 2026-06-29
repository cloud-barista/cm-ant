package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloud-barista/cm-ant/internal/infra/outbound/spider"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"
	"gorm.io/gorm"
)

// depCacheTTL controls how long the /ant/readyz handler reuses a previous
// dependency check result before re-checking. Per STANDARD-READYZ §8, this
// limits outbound dependency call rate to at most one full check per TTL.
// Startup uses CheckStartupDependencies which bypasses the cache.
const depCacheTTL = 30 * time.Second

// DepStatus captures the result of probing a single dependency.
type DepStatus struct {
	Reachable     bool      `json:"reachable"`
	Authenticated bool      `json:"authenticated"`
	LastCheck     time.Time `json:"lastCheck"`
	Error         string    `json:"error,omitempty"`
}

// DepResult captures the result of probing all dependencies cm-ant needs to
// service requests. Per STANDARD-READYZ §3, a dependency is considered fully
// healthy only when both Reachable and Authenticated are true (cb-tumblebug
// also requires Ready==true && Initialized==true; that check happens inside
// tumblebug_client.ReadyzWithContext).
type DepResult struct {
	Ready     bool      `json:"ready"`
	DB        DepStatus `json:"db"`
	Spider    DepStatus `json:"spider"`
	Tumblebug DepStatus `json:"tumblebug"`
}

type depCache struct {
	mu      sync.Mutex
	last    *DepResult
	lastSet time.Time
}

// checkDependencies runs the full STANDARD-READYZ §3 check sequence:
//  1. DB connectivity (gorm handle + Ping)
//  2. cb-spider /readyz reachability
//  3. cb-spider /cloudos with Basic Auth (auth enforcement)
//  4. cb-tumblebug /readyz body inspection (Ready && Initialized)
//  5. cb-tumblebug /cloudInfo with Basic Auth (auth enforcement)
//
// It always probes live; callers wanting the cached result should use
// getDependencyStatus.
func (s *AntServer) checkDependencies(ctx context.Context) *DepResult {
	now := time.Now()
	res := &DepResult{}

	// 1. DB
	res.DB = probeDB(ctx, s.db, now)

	// 2/3. cb-spider
	res.Spider = probeSpider(ctx, s.spiderClient, now)

	// 4/5. cb-tumblebug
	res.Tumblebug = probeTumblebug(ctx, s.tumblebugClient, now)

	res.Ready = res.DB.Reachable && res.DB.Authenticated &&
		res.Spider.Reachable && res.Spider.Authenticated &&
		res.Tumblebug.Reachable && res.Tumblebug.Authenticated
	return res
}

func probeDB(ctx context.Context, db *gorm.DB, now time.Time) DepStatus {
	if db == nil {
		return DepStatus{Reachable: false, LastCheck: now,
			Error: "cm-ant DB connection is nil. Check ANT_DB_HOST/USER/PASSWORD and DB connectivity"}
	}
	sqlDB, err := db.DB()
	if err != nil {
		return DepStatus{Reachable: false, LastCheck: now,
			Error: fmt.Sprintf("failed to acquire cm-ant DB handle: %v. Check DB driver/configuration", err)}
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return DepStatus{Reachable: false, LastCheck: now,
			Error: fmt.Sprintf("cm-ant DB ping failed: %v. Check ANT_DB_HOST/USER/PASSWORD and network", err)}
	}
	// DB connectivity == authentication for DB (creds are part of the conn).
	return DepStatus{Reachable: true, Authenticated: true, LastCheck: now}
}

func probeSpider(ctx context.Context, c *spider.SpiderClient, now time.Time) DepStatus {
	if c == nil {
		return DepStatus{Reachable: false, LastCheck: now,
			Error: "cb-spider client is nil"}
	}
	endpoint := c.Endpoint()
	if err := c.ReadyzWithContext(ctx); err != nil {
		return DepStatus{Reachable: false, LastCheck: now,
			Error: fmt.Sprintf("cb-spider unreachable (host=%s): %v. Check cb-spider container state and network", endpoint, err)}
	}
	if err := c.AuthCheckWithContext(ctx); err != nil {
		if errors.Is(err, spider.ErrUnauthorized) {
			return DepStatus{Reachable: true, Authenticated: false, LastCheck: now,
				Error: fmt.Sprintf("cb-spider authentication failed (host=%s, HTTP 401). Verify ANT_SPIDER_USERNAME and ANT_SPIDER_PASSWORD env vars (or config Spider.Username/Password) on the cm-ant side", endpoint)}
		}
		return DepStatus{Reachable: true, Authenticated: false, LastCheck: now,
			Error: fmt.Sprintf("cb-spider auth check call error (host=%s, GET /spider/cloudos): %v", endpoint, err)}
	}
	return DepStatus{Reachable: true, Authenticated: true, LastCheck: now}
}

func probeTumblebug(ctx context.Context, c *tumblebug.TumblebugClient, now time.Time) DepStatus {
	if c == nil {
		return DepStatus{Reachable: false, LastCheck: now,
			Error: "cb-tumblebug client is nil"}
	}
	endpoint := c.Endpoint()
	if err := c.ReadyzWithContext(ctx); err != nil {
		switch {
		case errors.Is(err, tumblebug.ErrNotReady):
			return DepStatus{Reachable: false, LastCheck: now,
				Error: fmt.Sprintf("cb-tumblebug unreachable (host=%s): %v. Check cb-tumblebug container state and network", endpoint, err)}
		case errors.Is(err, tumblebug.ErrNotInitialized):
			return DepStatus{Reachable: true, Authenticated: false, LastCheck: now,
				Error: fmt.Sprintf("cb-tumblebug not initialized (host=%s): %v. Complete cb-tumblebug initialization procedure", endpoint, err)}
		default:
			return DepStatus{Reachable: false, LastCheck: now,
				Error: fmt.Sprintf("cb-tumblebug unreachable (host=%s): %v. Check cb-tumblebug container state and network", endpoint, err)}
		}
	}
	if err := c.AuthCheckWithContext(ctx); err != nil {
		if errors.Is(err, tumblebug.ErrUnauthorized) {
			return DepStatus{Reachable: true, Authenticated: false, LastCheck: now,
				Error: fmt.Sprintf("cb-tumblebug authentication failed (host=%s, HTTP 401). Verify ANT_TUMBLEBUG_USERNAME and ANT_TUMBLEBUG_PASSWORD env vars (or config Tumblebug.Username/Password) on the cm-ant side", endpoint)}
		}
		return DepStatus{Reachable: true, Authenticated: false, LastCheck: now,
			Error: fmt.Sprintf("cb-tumblebug auth check call error (host=%s, GET /tumblebug/cloudInfo): %v", endpoint, err)}
	}
	return DepStatus{Reachable: true, Authenticated: true, LastCheck: now}
}

// getDependencyStatus returns the most recent dependency check result, running
// a fresh probe only when the cache TTL has expired. Used by /ant/readyz.
func (s *AntServer) getDependencyStatus(ctx context.Context) *DepResult {
	s.depCache.mu.Lock()
	defer s.depCache.mu.Unlock()

	if s.depCache.last != nil && time.Since(s.depCache.lastSet) < depCacheTTL {
		return s.depCache.last
	}

	s.depCache.last = s.checkDependencies(ctx)
	s.depCache.lastSet = time.Now()
	return s.depCache.last
}

// CheckStartupDependencies runs a fresh dependency probe (no cache) and returns
// an aggregated error if any dependency is not fully healthy. Called from main
// right after NewAntServer so the container exits with a clear message when a
// dependency is missing or misconfigured, per STANDARD-READYZ §6.2.
func (s *AntServer) CheckStartupDependencies(ctx context.Context) error {
	res := s.checkDependencies(ctx)
	if res.Ready {
		// Seed the cache so the first /ant/readyz call after startup does not
		// double-probe the dependencies that we just verified.
		s.depCache.mu.Lock()
		s.depCache.last = res
		s.depCache.lastSet = time.Now()
		s.depCache.mu.Unlock()
		return nil
	}

	var msgs []string
	if !res.DB.Reachable || !res.DB.Authenticated {
		msgs = append(msgs, res.DB.Error)
	}
	if !res.Spider.Reachable || !res.Spider.Authenticated {
		msgs = append(msgs, res.Spider.Error)
	}
	if !res.Tumblebug.Reachable || !res.Tumblebug.Authenticated {
		msgs = append(msgs, res.Tumblebug.Error)
	}
	return errors.New(strings.Join(msgs, " | "))
}
