package db

import (
	"context"
	"fmt"
	"log/slog"
	"task1/internal/order"
	"task1/pkg/client"

	"github.com/go-playground/validator/v10"
)

type Repository struct {
	client client.CLient
	Logger *slog.Logger
}

func NewRepository(client client.CLient, logger *slog.Logger) order.Repository {
	return &Repository{
		client: client,
		Logger: logger,
	}
}
func (r *Repository) Save(ctx context.Context, ord order.Order) (string, error) {
	tx, err := r.client.Begin(ctx)
	if err != nil {
		r.Logger.Error("Ошибка при создании транзакции", "error", err)
		return "", err
	}
	defer tx.Rollback(ctx)

	if err := validateOrder(ord); err != nil {
		return "", err
	}

	var orderUID string
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (
			track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING order_uid`,
		ord.TrackNumber, ord.Entry, ord.Locale, ord.InternalSignature,
		ord.CustomerID, ord.DeliveryService, ord.ShardKey, ord.SmID,
		ord.DateCreated, ord.OofShard,
	).Scan(&orderUID)
	if err != nil {
		r.Logger.Error("Ошибка при вставке order", "error", err)
		return "", err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO delivery (
			 order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		orderUID, ord.Delivery.Name, ord.Delivery.Phone,
		ord.Delivery.Zip, ord.Delivery.City, ord.Delivery.Address,
		ord.Delivery.Region, ord.Delivery.Email,
	)
	if err != nil {
		r.Logger.Error("Ошибка при вставке delivery", "error", err)
		return "", err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO payment (
			 order_uid, transaction, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		orderUID, ord.Payment.Transaction, ord.Payment.RequestID,
		ord.Payment.Currency, ord.Payment.Provider, ord.Payment.Amount,
		ord.Payment.PaymentDT, ord.Payment.Bank, ord.Payment.DeliveryCost,
		ord.Payment.GoodsTotal, ord.Payment.CustomFee,
	)
	if err != nil {
		r.Logger.Error("Ошибка при вставке payment", "error", err)
		return "", err
	}

	for _, item := range ord.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO items (
				order_uid, chrt_id, track_number, price, rid,
				name, sale, size, total_price, nm_id, brand, status
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			orderUID, item.ChrtID, item.TrackNumber, item.Price,
			item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			r.Logger.Error("Ошибка при вставке item", "error", err)
			return "", err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		r.Logger.Error("Ошибка при коммите транзакции", "error", err)
		return "", err
	}

	return orderUID, nil
}

func (r *Repository) FindAll(ctx context.Context) ([]order.Order, error) {
	query := `SELECT o.order_uid,o.track_number,o.entry,o.locale,o.internal_signature,
		o.customer_id,o.delivery_service,o.shardkey,o.sm_id,o.date_created,o.oof_shard,
		d.delivery_id,d.name,d.phone,d.zip,d.city,d.address,d.region,d.email,
		p.payment_id,p.transaction,p.request_id,p.currency,p.provider,p.amount,
		p.payment_dt,p.bank,p.delivery_cost,p.goods_total,p.custom_fee,
		i.item_id,i.chrt_id,i.track_number,i.price,i.rid,i.name,i.sale,i.size,i.total_price,
		i.nm_id,i.brand,i.status
		FROM orders o LEFT JOIN delivery d ON o.order_uid=d.order_uid
		LEFT JOIN payment p ON o.order_uid=p.order_uid
		LEFT JOIN items i ON o.order_uid=i.order_uid`
	rows, err := r.client.Query(ctx, query)
	if err != nil {
		r.Logger.Error("Ошибка при чтении запросе FindAll", "error", err)
		return nil, err
	}
	ordersMap := make(map[string]*order.Order)
	defer rows.Close()
	for rows.Next() {
		var d order.Delivery
		var p order.Payment
		var i order.Item
		var o order.Order
		err = rows.Scan(
			&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature,
			&o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard,
			&d.DeliveryID, &d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email,
			&p.PaymentID, &p.Transaction, &p.RequestID, &p.Currency, &p.Provider, &p.Amount,
			&p.PaymentDT, &p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee,
			&i.ItemID, &i.ChrtID, &i.TrackNumber, &i.Price, &i.Rid, &i.Name, &i.Sale,
			&i.Size, &i.TotalPrice, &i.NmID, &i.Brand, &i.Status,
		)
		if err != nil {
			r.Logger.Error("Ошибка при чтении orders", "error", err)
			return nil, err
		}
		existingOrder, ok := ordersMap[o.OrderUID]
		if !ok {
			o.Delivery = &d
			o.Payment = &p
			o.Items = []*order.Item{}
			ordersMap[o.OrderUID] = &o
			existingOrder = &o
		}
		existingOrder.Items = append(existingOrder.Items, &i)
	}
	orders := make([]order.Order, 0, len(ordersMap))
	for _, o := range ordersMap {
		orders = append(orders, *o)
	}
	return orders, nil

}
func (r *Repository) FindById(ctx context.Context, id string) (order.Order, error) {
	var o order.Order
	query := `SELECT o.order_uid,o.track_number,o.entry,o.locale,o.internal_signature,
		o.customer_id,o.delivery_service,o.shardkey,o.sm_id,o.date_created,o.oof_shard,
		d.delivery_id,d.name,d.phone,d.zip,d.city,d.address,d.region,d.email,
		p.payment_id,p.transaction,p.request_id,p.currency,p.provider,p.amount,
		p.payment_dt,p.bank,p.delivery_cost,p.goods_total,p.custom_fee,
		i.item_id,i.chrt_id,i.track_number,i.price,i.rid,i.name,i.sale,i.size,i.total_price,
		i.nm_id,i.brand,i.status
		FROM orders o LEFT JOIN delivery d ON o.order_uid=d.order_uid
		LEFT JOIN payment p ON o.order_uid=p.order_uid
		LEFT JOIN items i ON o.order_uid=i.order_uid WHERE o.order_uid=$1`
	rows, err := r.client.Query(ctx, query, id)
	if err != nil {
		r.Logger.Error("Ошибка при выполнении запроса FindByID", "error", err)
		return o, err
	}
	defer rows.Close()
	o.Items = []*order.Item{}
	for rows.Next() {
		var d order.Delivery
		var p order.Payment
		var i order.Item
		err = rows.Scan(
			&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature,
			&o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard,
			&d.DeliveryID, &d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email,
			&p.PaymentID, &p.Transaction, &p.RequestID, &p.Currency, &p.Provider, &p.Amount,
			&p.PaymentDT, &p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee,
			&i.ItemID, &i.ChrtID, &i.TrackNumber, &i.Price, &i.Rid, &i.Name, &i.Sale,
			&i.Size, &i.TotalPrice, &i.NmID, &i.Brand, &i.Status,
		)
		if err != nil {
			r.Logger.Error("Ошибка при сканировании строки FindById", "error", err)
			return o, err
		}
		if o.Delivery == nil {
			o.Delivery = &d
		}
		if o.Payment == nil {
			o.Payment = &p
		}
		o.Items = append(o.Items, &i)
		if len(o.Items) == 0 && o.Delivery == nil && o.Payment == nil {
			return o, fmt.Errorf("заказ с order_uid=%s не найден", id)
		}

	}
	return o, nil
}

func validateOrder(ord order.Order) error {
	validate := validator.New()
	err := validate.Struct(ord)
	if err != nil {
		return err
	}

	return nil
}
