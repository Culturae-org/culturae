// backend/internal/usecase/dataset_reader.go

package usecase

import (
	"github.com/google/uuid"
	"github.com/Culturae-org/culturae/internal/model"
)

type DatasetReader interface {
	GetDatasetByID(datasetType string, id uuid.UUID) (model.DatasetInfo, error)

	GetDefaultDataset(datasetType string) (model.DatasetInfo, error)
}
