// backend/internal/usecase/dataset_reader_adapter.go

package usecase

import (
	"github.com/Culturae-org/culturae/internal/model"
	adminUsecase "github.com/Culturae-org/culturae/internal/usecase/admin"

	"github.com/google/uuid"
)

type DatasetReaderAdapter struct {
	adminDatasets *adminUsecase.AdminDatasetsUsecase
}

func NewDatasetReaderAdapter(
	adminDatasets *adminUsecase.AdminDatasetsUsecase,
	) DatasetReader {
	return &DatasetReaderAdapter{
		adminDatasets: adminDatasets,
	}
}

// -----------------------------------------------
// Dataset Reader Adapter Methods
//
// - GetDatasetByID
// - GetDefaultDataset
//
// -----------------------------------------------

func (a *DatasetReaderAdapter) GetDatasetByID(datasetType string, id uuid.UUID) (model.DatasetInfo, error) {
	unified, err := a.adminDatasets.GetDataset(datasetType, id)
	if err != nil {
		return model.DatasetInfo{}, err
	}

	return model.DatasetInfo{
		ID:        unified.ID,
		Type:      unified.Type,
		Slug:      unified.Slug,
		Name:      unified.Name,
		IsActive:  unified.IsActive,
		IsDefault: unified.IsDefault,
	}, nil
}

func (a *DatasetReaderAdapter) GetDefaultDataset(datasetType string) (model.DatasetInfo, error) {
	unified, err := a.adminDatasets.GetDefaultDataset(datasetType)
	if err != nil {
		return model.DatasetInfo{}, err
	}

	return model.DatasetInfo{
		ID:        unified.ID,
		Type:      unified.Type,
		Slug:      unified.Slug,
		Name:      unified.Name,
		IsActive:  unified.IsActive,
		IsDefault: unified.IsDefault,
	}, nil
}
