package postgres

import (
	"context"
	"database/sql"
	"fmt"
	

	"github.com/pressly/goose/v3"

	"github.com/srKazuya/ordersPET/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(cfg Config) (*Storage, error) {
	const op = "storage.postgres.NewStrorage"

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return nil, fmt.Errorf("%s migrate:%w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveOrder(ctx context.Context, order *storage.Order) error {
	const op = "storage.postgres.SaveOrder"

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("%s failed to begin transaction: %w", op, err)
	}

	defer func(){
		if p := recover(); p!=nil{
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return fmt.Errorf("%s insert into orders: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO deliveries (
			order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
		order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("%s insert into deliveries: %w",op, err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO payments (
			transaction, order_uid, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, order.Payment.Transaction, order.OrderUID, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("%s insert into payments: %w",op, err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO items (
			order_uid, chrt_id, track_number, price, rid,
			name, sale, size, total_price, nm_id, brand, status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`)
	if err != nil {
		return fmt.Errorf("%s prepare insert items: %w", op, err)
	}
	defer stmt.Close()

	for _, item := range order.Items {
		_, err = stmt.ExecContext(ctx,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID,
			item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("%s insert into items: %w", op, err)
		}
	}

	return nil
}


