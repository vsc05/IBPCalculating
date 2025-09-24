package repository

import (
	"Lab1/internal/app/ds"
	"errors"
	"fmt"

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
	var existing ds.BidComponent

	err := r.db.Where("bid_id = ? AND component_id = ?", bidID, componentID).
		First(&existing).Error

	if err == nil {
		// запись есть → восстанавливаем
		return r.db.Model(&ds.BidComponent{}).
			Where("bid_id = ? AND component_id = ?", bidID, componentID).
			Update("is_delete", false).Error
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// записи реально нет → создаём новую
		newComp := ds.BidComponent{
			BidID:       uint(bidID),
			ComponentID: uint(componentID),
			IsDelete:    false,
		}
		return r.db.Create(&newComp).Error
	}

	return err
}
