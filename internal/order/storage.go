package order


import(
	"context"
)

type Repository interface {
	Save(ctx context.Context, ord Order) (string, error)
	FindAll(ctx context.Context) ([]Order,error)
	FindById(ctx context.Context,id string) (Order,error)
	
}
