package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Order struct {
	ID    int
	Image string
	Title string
	Power int
	Coeff float32
}

type Trash struct {
	ID         int
	Components []Order
	Status     string
}

type CalcBid struct {
	ID          int
	ComponentID int
	BidUPS      int
}

func (r *Repository) GetTrash() ([]Trash, error) {
	order1, err := r.GetOrder(1)
	if err != nil {
		return []Trash{}, err
	}

	order2, err := r.GetOrder(2)
	if err != nil {
		return []Trash{}, err
	}

	order3, err := r.GetOrder(3)
	if err != nil {
		return []Trash{}, err
	}

	trash := []Trash{
		{
			ID:         1,
			Components: []Order{order1, order2},
			Status:     "Active",
		},
		{
			ID:         2,
			Components: []Order{order1, order3, order2},
			Status:     "Done",
		},
	}

	return trash, nil
}

func (r *Repository) GetOrders() ([]Order, error) {
	orders := []Order{
		{
			ID:    1,
			Image: "http://127.0.0.1:9000/test/Monitor.png",
			Title: "Монитор",
			Power: 70,
			Coeff: 0.7,
		},
		{
			ID:    2,
			Image: "http://127.0.0.1:9000/test/Router.png",
			Title: "Роутер",
			Power: 13,
			Coeff: 0.9,
		},
		{
			ID:    3,
			Image: "http://127.0.0.1:9000/test/Pc.png",
			Title: "Компьютер",
			Power: 700,
			Coeff: 0.9,
		},
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return orders, nil
}

func (r *Repository) GetOrder(id int) (Order, error) {
	orders, err := r.GetOrders()
	if err != nil {
		return Order{}, err
	}

	for _, order := range orders {
		if order.ID == id {
			return order, nil
		}
	}

	return Order{}, fmt.Errorf("Not found")
}

func (r *Repository) GetOrdersByTitle(title string) ([]Order, error) {
	orders, err := r.GetOrders()
	if err != nil {
		return []Order{}, err
	}

	var result []Order
	for _, order := range orders {
		if strings.Contains(strings.ToLower(order.Title), strings.ToLower(title)) {
			result = append(result, order)
		}
	}

	return result, nil
}
