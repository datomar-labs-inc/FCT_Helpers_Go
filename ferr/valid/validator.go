package valid

import (
	"context"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	en2 "gopkg.in/go-playground/validator.v9/translations/en"
)

var validate *validator.Validate
var UniversalTranslator ut.Translator

func init() {
	ent := en.New()

	var uni = ut.New(ent, ent)

	trans, _ := uni.GetTranslator("en")

	UniversalTranslator = trans

	validate = validator.New()
	_ = en2.RegisterDefaultTranslations(validate, trans)
}

func ValidateStruct(ctx context.Context, s any) error {
	return validate.StructCtx(ctx, s)
}
