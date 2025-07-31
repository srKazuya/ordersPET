package save

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/srKazuya/ordersPET/internal/kafka"

	"github.com/srKazuya/ordersPET/internal/lib/logger/sl"

	resp "github.com/srKazuya/ordersPET/internal/lib/validators"
)

type Request struct {
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

type Response struct {
	resp.ValidationResponse
	TrackNumber string
}

func New(log *slog.Logger, prod *kafka.Producer, topic string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.order.Save"

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request BODY is empty")

			render.JSON(w, r, resp.Error("empty request"))
			return
		}

		if err != nil {
			log.Error("failed todecode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request body"))
			return
		}

		log.Info("request body decoded")

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invaild request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		msgBytes, err := json.Marshal(req)
		if err != nil {
			log.Error("failed to marshal request", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to marshal request"))
			return
		}

		err = prod.Produce(string(msgBytes), topic)
		if err != nil {
			log.Error("failed to produse order", sl.Err(err))
			render.JSON(w, r, resp.Error("failde to save transaction"))
			return
		}

		log.Info("order added", slog.String("trackNumber: ", req.TrackNumber))
		responseOK(w, r, req.TrackNumber)

	}
}

func responseOK(w http.ResponseWriter, r *http.Request, trackNumber string) {
	render.JSON(w, r, Response{
		ValidationResponse: resp.OK(),
		TrackNumber:        trackNumber,
	})
}
