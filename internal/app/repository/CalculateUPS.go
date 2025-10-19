package repository

import (
	"Lab1/internal/app/ds"
	"errors"
	"fmt"
	"time"

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
		return r.db.Model(&ds.Component{}).
			Where("id = ?", componentID).
			Update("is_delete", false).Error
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
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

func (r *Repository) CreateUser(user *ds.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByID(id int) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByUsername(username string) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("login = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUser(id int, updates map[string]interface{}) error {
	result := r.db.Model(&ds.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("пользователь не найден")
	}
	return nil
}

func (r *Repository) CreateComponent(resource *ds.Component) error {
	resource.IsDelete = false
	return r.db.Create(resource).Error
}

func (r *Repository) UpdateComponent(id int, updates map[string]interface{}) error {
	result := r.db.Model(&ds.Component{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ресурс не найден")
	}
	return nil
}

func (r *Repository) DeleteComponentPostman(id int) error {
	result := r.db.Model(&ds.Component{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_delete": true,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ресурс не найден")
	}
	return nil
}

func (r *Repository) SetComponentImage(id int, imageURL string) error {
	return r.db.Model(&ds.Component{}).Where("id = ?", id).Update("image", imageURL).Error
}

func (r *Repository) GetUserCart() (int64, error) {
	userID, err := r.GetActiveUser()
	if err != nil {
		return 0, err
	}

	var count int64
	err = r.db.Model(&ds.CalcUPS{}).
		Joins("JOIN bid_ups ON bid_ups.id = calc_ups.bid_id").
		Where("bid_ups.creator_id = ? AND bid_ups.status = ?", userID, "черновик").
		Count(&count).Error

	return count, err
}

func (r *Repository) GetAllBidUPS() ([]ds.BidUPS, error) {
	var applications []ds.BidUPS
	err := r.db.Preload("Components.Component").
		Where("status NOT IN ?", []string{"черновик", "удалён"}).
		Find(&applications).Error
	return applications, err
}

func (r *Repository) GetAllComponents() ([]ds.Component, error) {
	var components []ds.Component
	err := r.db.Find(&components).Error
	if err != nil {
		return nil, err
	}
	return components, nil
}

func (r *Repository) DeclineBidUPS(applicationID int, moderatorID int, status string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var currentBid ds.BidUPS
	if err := tx.First(&currentBid, applicationID).Error; err != nil {
		tx.Rollback()
		return err
	}

	finalStatus := status
	if status == "завершена" && currentBid.Status != "сформирован" {
		tx.Rollback()
		return fmt.Errorf("для завершения заявка должна быть в статусе 'сформирован', текущий статус: '%s'", currentBid.Status)
	}

	err := tx.Model(&ds.BidUPS{}).Where("id = ?", applicationID).Updates(map[string]interface{}{
		"status":       finalStatus,
		"date_finish":  time.Now(),
		"moderator_id": moderatorID,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if currentBid.Status == "сформирован" && status == "завершена" {
		var calcUPSRecords []ds.CalcUPS
		if err := tx.Where("bid_id = ?", applicationID).Find(&calcUPSRecords).Error; err != nil {
			tx.Rollback()
			return err
		}

		for _, calcRecord := range calcUPSRecords {
			var component ds.Component
			if err := tx.First(&component, calcRecord.ComponentID).Error; err != nil {
				tx.Rollback()
				return err
			}

			calculatedPower := calcRecord.BatteryLife * (calcRecord.IncomingCurrent + component.Power) * int(component.Coeff)

			if err := tx.Model(&ds.CalcUPS{}).Where("id = ?", calcRecord.ID).Update("calculated_power", calculatedPower).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

func (r *Repository) DeleteBidUPS(applicationID int, moderatorID int) error {
	return r.db.Model(&ds.BidUPS{}).Where("id = ?", applicationID).Updates(map[string]interface{}{
		"status":       "удален",
		"date_finish":  time.Now(),
		"moderator_id": moderatorID,
	}).Error
}

func (r *Repository) SetBidUPS(id, weight, productivity int) error {
	var application ds.BidUPS

	result := r.db.First(&application, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("заявка не найдена")
		}
		return result.Error
	}

	result = r.db.Model(&application).Updates(map[string]interface{}{
		"creator_id":   weight,
		"moderator_id": productivity,
	})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *Repository) FormBidUPS(id int) error {
	var application ds.BidUPS

	result := r.db.First(&application, "id = ? AND status = ?", id, "черновик")
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("заявка не найдена")
		}
		return result.Error
	}

	result = r.db.Model(&application).Update("status", "сформирован")
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *Repository) DeleteCalcUPS(applicationID, resourceID int) error {
	result := r.db.Where("bid_id = ? AND component_id = ?", applicationID, resourceID).
		Delete(&ds.CalcUPS{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ресурс не найден в заявке")
	}
	return nil
}

func (r *Repository) SetCalcUPS(applicationID, resourceID int, component_id int, coefficient int, power int) error {

	result := r.db.Model(&ds.CalcUPS{}).
		Where("id = ? AND bid_id = ? AND component_id = ?", resourceID, applicationID, component_id).
		Update("battery_life", coefficient).
		Update("incoming_current", power)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ресурс не найден в заявке")
	}
	return nil
}

func (r *Repository) GetDraftBid() (*ds.BidUPS, error) {
	var bid ds.BidUPS
	err := r.db.Where("status = ?", "черновик").First(&bid).Error
	if err != nil {
		return nil, err
	}
	return &bid, nil
}

func (r *Repository) GetActiveUser() (int, error) {
	var bid ds.BidUPS
	err := r.db.Where("status = ?", "черновик").First(&bid).Error
	if err != nil {
		return 0, err
	}
	return int(bid.CreatorID), nil
}
