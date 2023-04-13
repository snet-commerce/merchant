package merchant

import (
	"context"

	"github.com/snet-commerce/merchant/internal/ent"
	"github.com/snet-commerce/validation"
)

type CreateMerchantProps struct {
	Name            string
	Description     *string
	PrimaryEmail    string
	SecondaryEmail  *string
	PrimaryNumber   string
	SecondaryNumber *string
	Active          bool
}

type CreateInteractor struct {
	client *ent.MerchantClient
}

func NewCreateInteractor(client *ent.MerchantClient) *CreateInteractor {
	return &CreateInteractor{client: client}
}

func (i *CreateInteractor) Process(ctx context.Context, props CreateMerchantProps) (*ent.Merchant, error) {
	res := validation.NewResult()

}
