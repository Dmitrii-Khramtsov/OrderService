// github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities/item.go
package entities

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func (i *Item) Equal(other Item) bool {
	return i.ChrtID == other.ChrtID &&
		     i.TrackNumber == other.TrackNumber &&
		     i.Price == other.Price &&
		     i.RID == other.RID &&
		     i.Name == other.Name &&
		     i.Sale == other.Sale &&
		     i.Size == other.Size &&
		     i.TotalPrice == other.TotalPrice &&
		     i.NmID == other.NmID &&
		     i.Brand == other.Brand &&
		     i.Price == other.Status &&
		     true
}
