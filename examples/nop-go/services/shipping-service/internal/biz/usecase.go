// Package biz 物流服务业务逻辑层
package biz

import (
	"context"
	"time"

	"nop-go/services/shipping-service/internal/data"
	"nop-go/services/shipping-service/internal/models"
	shareErrors "nop-go/shared/errors"
)

type ShipmentUseCase struct {
	shipmentRepo data.ShipmentRepository
	itemRepo     data.ShipmentItemRepository
	methodRepo   data.ShippingMethodRepository
	trackingRepo data.ShipmentTrackingRepository
}

func NewShipmentUseCase(
	shipmentRepo data.ShipmentRepository,
	itemRepo data.ShipmentItemRepository,
	methodRepo data.ShippingMethodRepository,
	trackingRepo data.ShipmentTrackingRepository,
) *ShipmentUseCase {
	return &ShipmentUseCase{
		shipmentRepo: shipmentRepo,
		itemRepo:     itemRepo,
		methodRepo:   methodRepo,
		trackingRepo: trackingRepo,
	}
}

func (uc *ShipmentUseCase) CreateShipment(ctx context.Context, req *models.CreateShipmentRequest) (*models.Shipment, error) {
	shipment := &models.Shipment{
		OrderID:          req.OrderID,
		ShippingMethod:   req.ShippingMethod,
		ShippingProvider: req.ShippingProvider,
		Status:           "pending",
	}

	if err := uc.shipmentRepo.Create(ctx, shipment); err != nil {
		return nil, err
	}

	for _, item := range req.Items {
		shipmentItem := &models.ShipmentItem{
			ShipmentID:  shipment.ID,
			OrderItemID: item.OrderItemID,
			ProductID:   0,
			ProductName: "",
			Quantity:    item.Quantity,
		}
		uc.itemRepo.Create(ctx, shipmentItem)
	}

	return shipment, nil
}

func (uc *ShipmentUseCase) GetShipment(ctx context.Context, id uint64) (*models.Shipment, error) {
	shipment, err := uc.shipmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrShipmentNotFound
	}
	return shipment, nil
}

func (uc *ShipmentUseCase) GetShipmentByOrderID(ctx context.Context, orderID uint64) (*models.Shipment, error) {
	return uc.shipmentRepo.GetByOrderID(ctx, orderID)
}

func (uc *ShipmentUseCase) UpdateTracking(ctx context.Context, shipmentID uint64, req *models.UpdateTrackingRequest) error {
	shipment, err := uc.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	if req.TrackingNumber != "" {
		shipment.TrackingNumber = req.TrackingNumber
	}

	if req.Status != "" {
		shipment.Status = req.Status
		if req.Status == "shipped" {
			now := time.Now()
			shipment.ShippedAt = &now
		} else if req.Status == "delivered" {
			now := time.Now()
			shipment.DeliveredAt = &now
		}
	}

	if err := uc.shipmentRepo.Update(ctx, shipment); err != nil {
		return err
	}

	if req.Status != "" && req.Description != "" {
		tracking := &models.ShipmentTracking{
			ShipmentID:  shipmentID,
			Status:      req.Status,
			Location:    req.Location,
			Description: req.Description,
			OccurredAt:  time.Now(),
		}
		uc.trackingRepo.Create(ctx, tracking)
	}

	return nil
}

func (uc *ShipmentUseCase) ListShipments(ctx context.Context, page, pageSize int) ([]*models.Shipment, int64, error) {
	return uc.shipmentRepo.List(ctx, page, pageSize)
}

type ShippingMethodUseCase struct {
	methodRepo data.ShippingMethodRepository
}

func NewShippingMethodUseCase(methodRepo data.ShippingMethodRepository) *ShippingMethodUseCase {
	return &ShippingMethodUseCase{methodRepo: methodRepo}
}

func (uc *ShippingMethodUseCase) CreateMethod(ctx context.Context, m *models.ShippingMethod) error {
	return uc.methodRepo.Create(ctx, m)
}

func (uc *ShippingMethodUseCase) GetMethod(ctx context.Context, id uint64) (*models.ShippingMethod, error) {
	return uc.methodRepo.GetByID(ctx, id)
}

func (uc *ShippingMethodUseCase) ListMethods(ctx context.Context) ([]*models.ShippingMethod, error) {
	return uc.methodRepo.List(ctx)
}

func (uc *ShippingMethodUseCase) UpdateMethod(ctx context.Context, m *models.ShippingMethod) error {
	return uc.methodRepo.Update(ctx, m)
}

func (uc *ShippingMethodUseCase) DeleteMethod(ctx context.Context, id uint64) error {
	return uc.methodRepo.Delete(ctx, id)
}