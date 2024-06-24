package load

import "github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"

type LoadService struct {
	loadRepo *LoadRepository
}

func NewLoadService(loadRepo *LoadRepository, tumblebug *tumblebug.TumblebugClient) *LoadService {
	return &LoadService{
		loadRepo: loadRepo,
	}
}

func (l *LoadService) InstallAgent() {

}
