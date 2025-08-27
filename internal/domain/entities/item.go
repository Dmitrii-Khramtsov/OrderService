// github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities/item.go
package entities

type Item struct {
	ChrtID      int    `json:"chrt_id" db:"chrt_id"`
	TrackNumber string `json:"track_number" db:"track_number"`
	Price       int    `json:"price" db:"price"`
	RID         string `json:"rid" db:"rid"`
	Name        string `json:"name" db:"name"`
	Sale        int    `json:"sale" db:"sale"`
	Size        string `json:"size" db:"size"`
	TotalPrice  int    `json:"total_price" db:"total_price"`
	NmID        int    `json:"nm_id" db:"nm_id"`
	Brand       string `json:"brand" db:"brand"`
	Status      int    `json:"status" db:"status"`
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
		     i.Status == other.Status
}
