package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// readyzResponse is the response shape for GET /ant/readyz. It follows pattern
// B of c-mig-common STANDARD-READYZ (cmig-workflow
// c-mig-common/design/07-DESIGN/STANDARD-READYZ.md): a structured body that
// callers can inspect for per-dependency reachable/authenticated state in
// addition to the HTTP status code. cm-ant itself has no separate
// initialization step, so Initialized is always true when Ready is true.
type readyzResponse struct {
	Ready        bool       `json:"ready"`
	Initialized  bool       `json:"initialized"`
	Message      string     `json:"message"`
	Dependencies *DepResult `json:"dependencies"`
}

// @Id AntServerReadiness
// @Summary Check CM-Ant API server readiness
// @Description Returns CM-Ant server readiness including DB and outbound
// @Description dependency (cb-spider, cb-tumblebug) reachability and
// @Description authentication. Per STANDARD-READYZ pattern B, the response
// @Description body always carries per-dependency status; the HTTP status is
// @Description 200 when every dependency is reachable and authenticated, and
// @Description 503 otherwise. Results are cached briefly to limit outbound
// @Description call rate; see STANDARD-READYZ for details.
// @Tags [Server Health]
// @Accept json
// @Produce json
// @Success 200 {object} readyzResponse "CM-Ant is ready (all dependencies healthy)"
// @Failure 503 {object} readyzResponse "CM-Ant is not ready (see dependencies for the failing component)"
// @Router /readyz [get]
func (s *AntServer) readyz(c echo.Context) error {
	res := s.getDependencyStatus(c.Request().Context())

	body := readyzResponse{
		Ready:        res.Ready,
		Initialized:  res.Ready,
		Dependencies: res,
	}

	if !res.Ready {
		body.Message = "CM-Ant is not ready - see dependencies for the failing component"
		return c.JSON(http.StatusServiceUnavailable, body)
	}
	body.Message = "CM-Ant is ready"
	return c.JSON(http.StatusOK, body)
}
