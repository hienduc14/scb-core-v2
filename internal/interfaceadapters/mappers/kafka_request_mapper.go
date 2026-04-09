package mappers

import (
	"encoding/json"
	"fmt"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"
)

type KafkaScanRequestMessage struct {
	RequestUUID   string            `json:"requestUUID"`
	ProjectURI    string            `json:"projectUri"`
	TargetURL     string            `json:"targetUrl"`
	Scanner       string            `json:"scanner"`
	ProjectType   string            `json:"projectType"`
	AuthToken     string            `json:"authToken"`
	Branch        string            `json:"branch"`
	CommitID      string            `json:"commitId"`
	RefID         interface{}       `json:"refId"`
	RequestTime   string            `json:"requestTime"`
	Tag           string            `json:"tag"`
	NucleiConfig  map[string]string `json:"nucleiConfig"`
	RequestParams map[string]string `json:"requestParams"`
}

func DecodeScanRequest(payload []byte) (entities.ScanRequest, error) {
	var message KafkaScanRequestMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		return entities.ScanRequest{}, err
	}
	
	var refID string
	if message.RefID != nil {
		switch v := message.RefID.(type) {
		case string:
			refID = v
		case float64:
			refID = fmt.Sprintf("%.0f", v)
		default:
			refID = fmt.Sprintf("%v", v)
		}
	}

	return entities.ScanRequest{
		RequestUUID:  message.RequestUUID,
		ScannerAlias: message.Scanner,
		ProjectType:  valueobjects.ProjectType(message.ProjectType),
		Target: entities.ScanTarget{
			RepositoryURL: message.ProjectURI,
			ImageRef:      message.ProjectURI,
			TargetURL:     message.TargetURL,
		},
		AuthToken:     message.AuthToken,
		Branch:        message.Branch,
		BaseCommit:    message.CommitID,
		RefID:         refID,
		RequestTime:   message.RequestTime,
		PlatformTag:   message.Tag,
		NucleiConfig:  message.NucleiConfig,
		RequestParams: message.RequestParams,
	}, nil
}
