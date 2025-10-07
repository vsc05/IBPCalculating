package ds

type Component struct {
	ID       int     `json:"id"`
	Image    string  `json:"image",omitempty`
	Title    string  `json:"title"`
	Power    int     `json:"power"`
	Coeff    float32 `json:"coeff"`
	IsDelete bool    `json:"is_delete"`
}
