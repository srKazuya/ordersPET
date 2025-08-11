package get

import (
	"context"

	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"


	"github.com/srKazuya/ordersPET/internal/service/getter"
	"github.com/srKazuya/ordersPET/internal/storage"

	"github.com/srKazuya/ordersPET/internal/lib/logger/sl"

	resp "github.com/srKazuya/ordersPET/internal/lib/validators"
)


type DeliveryRequest struct {
	Name    string `json:"name" validate:"required"`
	Phone   string `json:"phone" validate:"required,e164"`
	Zip     string `json:"zip" validate:"required"`
	City    string `json:"city" validate:"required"`
	Address string `json:"address" validate:"required"`
	Region  string `json:"region" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
}

type PaymentRequest struct {
	Transaction  string `json:"transaction" validate:"required"`
	RequestID    string `json:"request_id" validate:"required"`
	Currency     string `json:"currency" validate:"required,len=3"`
	Provider     string `json:"provider" validate:"required"`
	Amount       int    `json:"amount" validate:"required,gt=0"`
	PaymentDT    int64  `json:"payment_dt" validate:"required"`
	Bank         string `json:"bank" validate:"required"`
	DeliveryCost int    `json:"delivery_cost" validate:"required"`
	GoodsTotal   int    `json:"goods_total" validate:"required"`
	CustomFee    int    `json:"custom_fee" validate:"required"`
}

type ItemRequest struct {
	ChrtID      int    `json:"chrt_id" validate:"required"`
	TrackNumber string `json:"track_number" validate:"required"`
	Price       int    `json:"price" validate:"required,gt=0"`
	RID         string `json:"rid" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Sale        int    `json:"sale" validate:"gte=0"`
	Size        string `json:"size" validate:"required"`
	TotalPrice  int    `json:"total_price" validate:"required,gt=0"`
	NmID        int    `json:"nm_id" validate:"required"`
	Brand       string `json:"brand" validate:"required"`
	Status      int    `json:"status" validate:"required"`
}
type Order struct {
	OrderUID          string          `json:"order_uid" validate:"required,alphanum"`
	TrackNumber       string          `json:"track_number" validate:"required"`
	Entry             string          `json:"entry" validate:"required"`
	Delivery          DeliveryRequest `json:"delivery" validate:"required,dive"`
	Payment           PaymentRequest  `json:"payment" validate:"required,dive"`
	Items             []ItemRequest   `json:"items" validate:"required,min=1,dive"`
	Locale            string          `json:"locale" validate:"required,alpha"`
	InternalSignature string          `json:"internal_signature" validate:"required"`
	CustomerID        string          `json:"customer_id" validate:"required"`
	DeliveryService   string          `json:"delivery_service" validate:"required"`
	ShardKey          string          `json:"shardkey" validate:"required"`
	SmID              int             `json:"sm_id" validate:"required"`
	DateCreated       time.Time       `json:"date_created" validate:"required"`
	OofShard          string          `json:"oof_shard" validate:"required"`
}

type Response struct {
	resp.ValidationResponse
	Order Order
}

func New(log *slog.Logger, getter orderGetter.OrderGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.order.Get"

		log = log.With(
			slog.String("op", op),
		)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		orderUID := chi.URLParam(r, "order_uid")
		
		order, err := getter.GetOrderByUID(ctx, orderUID)
		if err != nil {
			log.Error("failed to get order", sl.Err(err))
			render.JSON(w, r, resp.Error("failde to get order"))
			return
		}

		log.Info("order getted", slog.String("uid: ", order.OrderUID))
		responseOK(w, r, order)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, order storage.Order) {
	render.JSON(w, r, Response{
		ValidationResponse: resp.OK(),
		Order: Order{
			OrderUID:    order.OrderUID,
			TrackNumber: order.TrackNumber,
			Entry:       order.Entry,
			Delivery: DeliveryRequest{
				Name:    order.Delivery.Name,
				Phone:   order.Delivery.Phone,
				Zip:     order.Delivery.Zip,
				City:    order.Delivery.City,
				Address: order.Delivery.Address,
				Region:  order.Delivery.Region,
				Email:   order.Delivery.Email,
			},
			Payment: PaymentRequest{
				Transaction:  order.Payment.Transaction,
				RequestID:    order.Payment.RequestID,
				Currency:     order.Payment.Currency,
				Provider:     order.Payment.Provider,
				Amount:       order.Payment.Amount,
				PaymentDT:    order.Payment.PaymentDT,
				Bank:         order.Payment.Bank,
				DeliveryCost: order.Payment.DeliveryCost,
				GoodsTotal:   order.Payment.GoodsTotal,
				CustomFee:    order.Payment.CustomFee,
			},
			Items: func(items []storage.Item) []ItemRequest {
				result := make([]ItemRequest, len(items))
				for i, item := range items {
					result[i] = ItemRequest{
						ChrtID:      item.ChrtID,
						TrackNumber: item.TrackNumber,
						Price:       item.Price,
						RID:         item.RID,
						Name:        item.Name,
						Sale:        item.Sale,
						Size:        item.Size,
						TotalPrice:  item.TotalPrice,
						NmID:        item.NmID,
						Brand:       item.Brand,
						Status:      item.Status,
					}
				}
				return result
			}(order.Items),
			Locale:            order.Locale,
			InternalSignature: order.InternalSignature,
			CustomerID:        order.CustomerID,
			DeliveryService:   order.DeliveryService,
			ShardKey:          order.ShardKey,
			SmID:              order.SmID,
			DateCreated:       order.DateCreated,
			OofShard:          order.OofShard,
		},
	})
}
