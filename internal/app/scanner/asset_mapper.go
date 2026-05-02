// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"context"
	"fmt"

	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/entity"
)

type AssetMapperDB interface {
	GetComponentInstance(id int64) (*entity.ComponentInstance, error)
	CreateScannerAssetMapping(mapping *entity.ScannerAssetMapping) (*entity.ScannerAssetMapping, error)
	GetScannerAssetMappingByUri(scannerName, artifactUri string) (*entity.ScannerAssetMapping, error)
}

type AssetMapper interface {
	ResolveAsset(ctx context.Context, artifactUri, serviceId string) (*entity.ComponentInstance, error)
	RegisterAssetMapping(ctx context.Context, mapping *entity.ScannerAssetMapping) error
	GetAssetMapping(ctx context.Context, scannerName, artifactUri string) (*entity.ScannerAssetMapping, error)
}

type assetMapper struct {
	db AssetMapperDB
}

func NewAssetMapper(db AssetMapperDB) AssetMapper {
	return &assetMapper{db: db}
}

func (am *assetMapper) ResolveAsset(ctx context.Context, artifactUri, serviceId string) (*entity.ComponentInstance, error) {
	op := appErrors.Op("AssetMapper.ResolveAsset")

	if artifactUri == "" {
		return nil, appErrors.E(op, "artifact URI is required")
	}

	if serviceId == "" {
		return nil, appErrors.E(op, "service ID is required")
	}

	// POC: Look up asset mapping by artifact URI and scanner name
	// In production, this would implement more sophisticated resolution logic
	return nil, appErrors.E(op, "Asset resolution requires pre-configured ScannerAssetMapping (use RegisterAssetMapping to configure)")
}

func (am *assetMapper) RegisterAssetMapping(ctx context.Context, mapping *entity.ScannerAssetMapping) error {
	op := appErrors.Op("AssetMapper.RegisterAssetMapping")

	if mapping.ArtifactUri == "" {
		return appErrors.E(op, "artifact URI is required")
	}

	if mapping.ComponentInstanceId == 0 {
		return appErrors.E(op, "component instance ID is required")
	}

	if mapping.ServiceId == 0 {
		return appErrors.E(op, "service ID is required")
	}

	compInst, err := am.db.GetComponentInstance(mapping.ComponentInstanceId)
	if err != nil || compInst == nil {
		return appErrors.E(op, fmt.Sprintf("Component instance not found: %d", mapping.ComponentInstanceId))
	}

	// TODO: Verify component instance belongs to service

	_, err = am.db.CreateScannerAssetMapping(mapping)
	if err != nil {
		return appErrors.E(op, "Failed to create asset mapping", err)
	}

	return nil
}

func (am *assetMapper) GetAssetMapping(ctx context.Context, scannerName, artifactUri string) (*entity.ScannerAssetMapping, error) {
	op := appErrors.Op("AssetMapper.GetAssetMapping")

	mapping, err := am.db.GetScannerAssetMappingByUri(scannerName, artifactUri)
	if err != nil {
		return nil, appErrors.E(op, "Failed to retrieve asset mapping", err)
	}

	return mapping, nil
}
