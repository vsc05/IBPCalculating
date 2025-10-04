package repository

import (
	"Lab1/internal/app/ds"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (r *Repository) GetComponents() ([]ds.Component, error) {
	var Components []ds.Component
	err := r.db.Find(&Components).Error
	if err != nil {
		return nil, err
	}
	if len(Components) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return Components, nil
}

func (r *Repository) GetComponent(id int) (ds.Component, error) {
	component := ds.Component{}
	err := r.db.Where("id = ?", id).First(&component).Error
	if err != nil {
		return ds.Component{}, err
	}
	return component, nil
}

func (r *Repository) GetComponentsByTitle(title string) ([]ds.Component, error) {
	var components []ds.Component
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&components).Error
	if err != nil {
		return nil, err
	}
	return components, nil
}

func (r *Repository) AddComponentToBid(bidID, componentID int) error {
	var existing ds.CalcUPS

	err := r.db.Where("bid_id = ? AND component_id = ?", bidID, componentID).
		First(&existing).Error

	if err == nil {
		// запись есть → восстанавливаем
		return r.db.Model(&ds.Component{}).
			Where("id = ?", componentID).
			Update("is_delete", false).Error
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// записи реально нет → создаём новую
		newComp := ds.CalcUPS{
			BidID:       uint(bidID),
			ComponentID: uint(componentID),
		}
		return r.db.Create(&newComp).Error
	}

	return err
}

func (r *Repository) UpdateCalcUPS(bidID, componentID int, incomingCurrent, calculatedPower, batteryLife int) error {
	return r.db.Model(&ds.CalcUPS{}).
		Where("bid_id = ? AND component_id = ?", bidID, componentID).
		UpdateColumns(map[string]interface{}{
			"incoming_current": incomingCurrent,
			"calculated_power": calculatedPower,
			"battery_life":     batteryLife,
		}).Error
}

func (r *Repository) UpdateBidStatus(bidID int, status string) error {
	return r.db.Model(&ds.BidUPS{}).
		Where("id = ?", bidID).
		UpdateColumn("status", status).Error
}

func (r *Repository) GetCalcUPSs(bidID int) ([]ds.Component, error) {
	var components []ds.Component

	err := r.db.Table("calc_ups bc").
		Select("c.id, c.image, c.title, c.power, c.coeff, bc.incoming_current, bc.battery_life, bc.calculated_power, c.is_delete").
		Joins("JOIN components c ON bc.component_id = c.id").
		Where("bc.bid_id = ? and c.is_delete = ?", bidID, false).
		Scan(&components).Error

	if err != nil {
		return nil, err
	}
	return components, nil
}

func (r *Repository) GetBinCount(creatorID int) int64 {
	var bidID uint
	var count int64

	err := r.db.Model(&ds.BidUPS{}).
		Where("creator_id = ? AND status = ?", creatorID, "черновик").
		Select("id").
		First(&bidID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.CalcUPS{}).
		Joins("JOIN components ON calc_ups.component_id = components.id").
		Where("calc_ups.bid_id = ? AND components.is_delete = ?", bidID, false).
		Count(&count).Error
	if err != nil {
		logrus.Println("Error counting calc_ups:", err)
	}

	return count
}

func (r *Repository) DeleteComponent(bidID, componentID uint) error {
	err := r.db.Model(&ds.Component{}).
		Where("id = ?", componentID).
		UpdateColumn("is_delete", true).Error

	if err != nil {
		return fmt.Errorf("ошибка при удалении компонента (bid_id=%d, component_id=%d): %w", bidID, componentID, err)
	}
	return nil
}

func (r *Repository) GetCalcPower(bidID int) float64 {
	var sumPower float64

	err := r.db.Model(&ds.CalcUPS{}).
		Where("bid_id = ?", bidID).
		Select("COALESCE(SUM(calculated_power), 0)").
		Scan(&sumPower).Error

	if err != nil {
		logrus.Printf("Ошибка подсчета мощности для заявки %d: %v", bidID, err)
		return 0
	}

	return sumPower
}
