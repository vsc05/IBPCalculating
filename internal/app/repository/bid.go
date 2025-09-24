package repository

import (
	"Lab1/internal/app/ds"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (r *Repository) UpdateBidComponent(bidID, componentID int, incomingCurrent, calculatedPower, batteryLife int) error {
	return r.db.Model(&ds.BidComponent{}).
		Where("bid_id = ? AND component_id = ?", bidID, componentID).
		UpdateColumns(map[string]interface{}{
			"incoming_current": incomingCurrent,
			"calculated_power": calculatedPower,
			"battery_life":     batteryLife,
		}).Error
}

func (r *Repository) UpdateBidStatus(bidID int, status string) error {
	return r.db.Model(&ds.Bid{}).
		Where("id = ?", bidID).
		UpdateColumn("status", status).Error
}

func (r *Repository) GetBidComponents(bidID int) ([]ds.Component, error) {
	var components []ds.Component

	err := r.db.Table("bid_components bc").
		Select("c.id, c.image, c.title, c.power, c.coeff, bc.incoming_current, bc.battery_life, bc.calculated_power, bc.is_delete").
		Joins("JOIN components c ON bc.component_id = c.id").
		Where("bc.bid_id = ? and bc.is_delete = ?", bidID, false).
		Scan(&components).Error

	if err != nil {
		return nil, err
	}
	return components, nil
}

func (r *Repository) GetBinCount(creatorID int) int64 {
	var bidID uint
	var count int64

	err := r.db.Model(&ds.Bid{}).
		Where("creator_id = ? AND status = ?", creatorID, "черновик").
		Select("id").
		First(&bidID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.BidComponent{}).
		Where("bid_id = ? and is_delete = false", bidID).
		Count(&count).Error
	if err != nil {
		logrus.Println("Error counting bid_components:", err)
	}

	return count
}

func (r *Repository) DeleteComponent(bidID, componentID uint) error {
	err := r.db.Model(&ds.BidComponent{}).
		Where("bid_id = ? AND component_id = ?", bidID, componentID).
		UpdateColumn("is_delete", true).Error

	if err != nil {
		return fmt.Errorf("ошибка при удалении компонента (bid_id=%d, component_id=%d): %w", bidID, componentID, err)
	}
	return nil
}

func (r *Repository) GetCalcPower(bidID int) float64 {
	var sumPower float64

	err := r.db.Model(&ds.BidComponent{}).
		Where("bid_id = ?", bidID).
		Select("COALESCE(SUM(calculated_power), 0)").
		Scan(&sumPower).Error

	if err != nil {
		logrus.Printf("Ошибка подсчета мощности для заявки %d: %v", bidID, err)
		return 0
	}

	return sumPower
}
