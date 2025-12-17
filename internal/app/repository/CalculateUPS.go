package repository

import (
	"DIA3Course/internal/app/ds"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const djangoCalculationURL = "http://localhost:8000/calculate_bid/"

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

	// 1. Поиск существующей записи по BidID и ComponentID
	// Мы ожидаем, что CalcUPS имеет уникальный индекс по этой паре.
	err := r.db.Where("bid_id = ? AND component_id = ?", bidID, componentID).
		First(&existing).Error

	if err == nil {
		// Условие 1: Запись НАЙДЕНА

		// 1.1. Определяем новое количество (Увеличиваем текущее на 1)
		// Если Count в структуре ds.CalcUPS имеет тип int, то GORM преобразует NULL в 0.
		newCount := existing.Count + 1

		// 1.2. Обновляем существующую запись в CalcUPS, увеличивая Count
		updateErr := r.db.Model(&existing).
			Updates(map[string]interface{}{
				"count": newCount,
				// Также можно добавить обновление других полей, например, снятие флага удаления, если он есть
				// "is_delete": false,
			}).Error

		if updateErr != nil {
			return updateErr
		}

		// ВАЖНО: Ваша старая логика обновляла ds.Component. Убираем ее,
		// так как AddComponentToBid должно работать только с CalcUPS.
		// Если нужно обновить Component, это должно быть отдельным шагом,
		// но, вероятно, это было лишним действием.
		// return r.db.Model(&ds.Component{}).
		//    Where("id = ?", componentID).
		//    Update("is_delete", false).Error

		return nil // Успешно обновили количество
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Условие 2: Запись НЕ НАЙДЕНА

		// Создаем новую запись, устанавливая Count = 1
		newComp := ds.CalcUPS{
			BidID:       uint(bidID),
			ComponentID: uint(componentID),
			Count:       1, // Устанавливаем начальное количество
		}
		return r.db.Create(&newComp).Error
	}

	// Условие 3: Другая ошибка GORM
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

func (r *Repository) GetUserCart(userLogin string) (int64, error) {
	// Находим ID пользователя по логину
	var user ds.User
	err := r.db.Where("login = ?", userLogin).First(&user).Error
	if err != nil {
		return 0, fmt.Errorf("user not found: %w", err)
	}

	var count int64
	err = r.db.Model(&ds.CalcUPS{}).
		Joins("JOIN bid_ups ON bid_ups.id = calc_ups.bid_id").
		Where("bid_ups.creator_id = ? AND bid_ups.status = ?", user.ID, "черновик").
		Count(&count).Error

	return count, err
}

func (r *Repository) GetUserCartDetails(userLogin string) (uint, int, error) {
	// Находим ID пользователя по логину
	var user ds.User
	err := r.db.Where("login = ?", userLogin).First(&user).Error
	if err != nil {
		return 0, 0, fmt.Errorf("user not found: %w", err)
	}

	// Ищем черновик заявки пользователя
	var bid ds.BidUPS
	err = r.db.Where("creator_id = ? AND status = ?", user.ID, "черновик").First(&bid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Если черновика нет, возвращаем нулевые значения
			return 0, 0, nil
		}
		return 0, 0, err
	}

	// Считаем количество услуг в заявке (записей в таблице CalcUPS для этой заявки)
	var itemsCount int64
	err = r.db.Model(&ds.CalcUPS{}).Where("bid_id = ?", bid.ID).Count(&itemsCount).Error
	if err != nil {
		return 0, 0, err
	}

	return bid.ID, int(itemsCount), nil
}

func (r *Repository) GetAllBidUPS(filters map[string]interface{}) ([]ds.BidUPS, error) {
	var applications []ds.BidUPS

	// Базовый запрос с исключением удаленных
	db := r.db.Preload("Creator").
		Preload("Moderator").
		Preload("Components").
		Where("status NOT IN ?", []string{"удалён"})

	// Применяем фильтры если они переданы
	if filters != nil {
		// Фильтрация по диапазону дат
		if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
			if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
				db = db.Where("date_update BETWEEN ? AND ?", startDate, endDate)
			} else {
				db = db.Where("date_update >= ?", startDate)
			}
		} else if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
			db = db.Where("date_update <= ?", endDate)
		}

		// Фильтрация по статусу
		if status, ok := filters["status"].(string); ok && status != "" {
			db = db.Where("status = ?", status)
		}
	}

	err := db.Find(&applications).Error
	if err != nil {
		return nil, err
	}

	// Добавляем вычисляемое поле для каждой заявки
	for i := range applications {
		calculatedCount := 0
		for _, component := range applications[i].Components {
			if component.CalculatedPower != 0 { // Предполагаем, что 0 означает "пустое"
				calculatedCount++
			}
		}
		applications[i].CalculatedPowerCount = calculatedCount
	}

	return applications, nil
}

func (r *Repository) GetAllComponents(nameFilter string) ([]ds.Component, error) {
	var components []ds.Component
	db := r.db

	if nameFilter != "" {
		db = db.Where("LOWER(title) LIKE LOWER(?)", "%"+nameFilter+"%")
	}

	err := db.Order("title").Find(&components).Error
	if err != nil {
		return nil, err
	}
	return components, nil
}

func (r *Repository) ProcessBidUPS(applicationID int, moderatorID int, status string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var currentBid ds.BidUPS
	// Важно: Загружаем компонент, так как его данные нужны для Django
	if err := tx.Preload("Components.Component").First(&currentBid, applicationID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("заявка с ID %d не найдена: %w", applicationID, err)
	}

	// --- 1. Проверки статуса --- (остаются синхронными)
	if status == "завершена" && currentBid.Status != "сформирован" {
		tx.Rollback()
		return fmt.Errorf("для завершения заявка должна быть в статусе 'сформирован', текущий статус: '%s'", currentBid.Status)
	}
	if status == "отклонена" && currentBid.Status == "завершена" {
		tx.Rollback()
		return fmt.Errorf("нельзя отклонить уже завершенную заявку")
	}
	if status != "отклонена" && status != "завершена" {
		tx.Rollback()
		return fmt.Errorf("Допустимые статусы отклонена или завершена")
	}

	// --- 2. Логика отклонения --- (остается синхронной)
	if status == "отклонена" {
		err := tx.Model(&ds.BidUPS{}).Where("id = ?", applicationID).Updates(map[string]interface{}{
			"status":       status,
			"date_finish":  time.Now(),
			"moderator_id": moderatorID,
		}).Error
		if err != nil {
			tx.Rollback()
			return err
		}
		return tx.Commit().Error
	}

	// --- 3. Логика ЗАВЕРШЕНИЯ (АСИНХРОННЫЙ ЗАПУСК) ---
	if currentBid.Status == "сформирован" && status == "завершена" {
		// 3.1 Обновляем статус на "в_процессе_расчета"
		if err := tx.Model(&ds.BidUPS{}).Where("id = ?", applicationID).Update("status", "рассчитан").Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit().Error; err != nil {
			return err
		}

		// 3.2 Запускаем асинхронную отправку запроса в Django
		go r.sendCalculationRequest(applicationID, currentBid, moderatorID)

		// Быстрый ответ клиенту
		return nil
	}

	return tx.Commit().Error
}

// Вспомогательная функция для отправки запроса в Django
func (r *Repository) sendCalculationRequest(applicationID int, bid ds.BidUPS, moderatorID int) {
	// Структура данных, необходимая Django-сервису
	requestData := map[string]interface{}{
		"application_id":     int(applicationID),
		"incoming_current":   bid.IncomingCurrent,
		"moderator_id_final": int(moderatorID),
		"components":         bid.Components, // Содержит вложенные CalcUPS и Component
	}

	// Маршалируем данные в JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		fmt.Printf("Ошибка маршалинга данных заявки %d: %v\n", applicationID, err)
		return
	}

	// Отправка POST-запроса
	resp, err := http.Post(djangoCalculationURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка отправки POST-запроса на Django-сервис для заявки %d: %v\n", applicationID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted { // Ожидаем 202 Accepted
		fmt.Printf("Django-сервис вернул неожиданный статус %d для заявки %d\n", resp.StatusCode, applicationID)
		// Здесь можно добавить логику, чтобы откатить статус заявки или уведомить администратора
	}
}

// Новая функция: принимает результаты расчетов от Django и обновляет БД
func (r *Repository) UpdateCalculatedPower(bidID int, results []ds.BidCalcResult, moderatorID int) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Обновляем CalculatedPower для каждой записи CalcUPS
	for _, result := range results {
		if err := tx.Model(&ds.CalcUPS{}).Where("id = ?", result.CalcID).Update("calculated_power", result.CalculatedPower).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 2. Финальное обновление статуса заявки на "завершена"
	err := tx.Model(&ds.BidUPS{}).Where("id = ?", bidID).Updates(map[string]interface{}{
		"status":       "завершена",
		"date_finish":  time.Now(),
		"moderator_id": moderatorID,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *Repository) DeleteBidUPS(applicationID int) error {
	return r.db.Model(&ds.BidUPS{}).Where("id = ?", applicationID).Updates(map[string]interface{}{
		"status":      "удален",
		"date_finish": time.Now(),
	}).Error
}

func (r *Repository) SetBidUPS(applicationID, incomingCurrent int) error {
	return r.db.Model(&ds.BidUPS{}).
		Where("id = ?", applicationID).
		Updates(map[string]interface{}{
			"incoming_current": incomingCurrent,
		}).Error
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

// ⭐️ Обновленная сигнатура: Принимаем BidID и ComponentID ⭐️
func (r *Repository) SetCalcUPS(bidID int, componentID int, batteryLife int, count int) error {

	// ⭐️ КЛЮЧЕВОЕ ИЗМЕНЕНИЕ: Ищем по BidID И ComponentID ⭐️
	// Обновляем запись CalcUPS, которая уникально определяется парой BidID-ComponentID
	result := r.db.Model(&ds.CalcUPS{}).
		Where("bid_id = ? AND component_id = ?", bidID, componentID).
		Updates(map[string]interface{}{
			"battery_life": batteryLife,
			"count":        count,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// Запись CalcUPS не найдена, потому что BidID-ComponentID пара отсутствует
		return fmt.Errorf("ресурс CalcUPS не найден (неверная пара BidID-ComponentID)")
	}

	return nil
}

// GetDraftBid возвращает черновую заявку текущего пользователя
func (r *Repository) GetDraftBid(userLogin string) (*ds.BidUPS, error) {
	var bid ds.BidUPS

	// Ищем черновик по статусу и логину пользователя через JOIN с таблицей users
	err := r.db.
		Joins("JOIN users ON bid_ups.creator_id = users.id").
		Where("bid_ups.status = ? AND users.login = ?", "черновик", userLogin).
		First(&bid).Error

	if err != nil {
		return nil, err
	}
	return &bid, nil
}

func (r *Repository) GetOrCreateDraftBid(userLogin string) (*ds.BidUPS, error) {
	// Сначала пытаемся найти существующий черновик
	bid, err := r.GetDraftBid(userLogin)
	if err == nil {
		return bid, nil
	}

	// Если черновик не найден, создаем нового пользователя (если не существует)
	var user ds.User
	if err := r.db.Where("login = ?", userLogin).First(&user).Error; err != nil {
		// Пользователь не найден - возможно, нужно создать
		// В зависимости от логики приложения
		return nil, fmt.Errorf("пользователь не найден: %v", err)
	}

	// Создаем новую черновую заявку
	newBid := ds.BidUPS{
		Status:          "черновик",
		DateUpdate:      time.Now(),
		CreatorID:       user.ID,
		IncomingCurrent: 220, // значение по умолчанию
	}

	if err := r.db.Create(&newBid).Error; err != nil {
		return nil, err
	}

	return &newBid, nil
}

func (r *Repository) Register(user *ds.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetBidUPSByID(id uint) (*ds.BidUPS, error) {
	var bidUPS ds.BidUPS

	err := r.db.
		Preload("Creator").
		Preload("Moderator").
		Preload("Components.Component").
		Where("id = ?", id).
		First(&bidUPS).Error

	if err != nil {
		return nil, err
	}

	return &bidUPS, nil
}
