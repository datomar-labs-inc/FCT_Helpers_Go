package valid

import (
	"context"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/volatiletech/null/v8"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/go-playground/validator.v9/non-standard/validators"
	en2 "gopkg.in/go-playground/validator.v9/translations/en"
	"reflect"
)

var validate *validator.Validate
var UniversalTranslator ut.Translator

func init() {
	ent := en.New()

	var uni = ut.New(ent, ent)

	trans, _ := uni.GetTranslator("en")

	UniversalTranslator = trans

	validate = validator.New()

	// register all sql.Null* types to use the ValidateValuer CustomTypeFunc
	validate.RegisterCustomTypeFunc(ValidateNullString, null.String{}, &null.String{})
	_ = validate.RegisterValidation("notblank", validators.NotBlank)

	_ = en2.RegisterDefaultTranslations(validate, trans)
}

func ValidateStruct(ctx context.Context, s any) error {
	return validate.StructCtx(ctx, s)
}

func ValidateNullString(field reflect.Value) interface{} {
	var emptyStr *string

	if s, ok := field.Interface().(null.String); ok && s.Valid {
		return s.String
	}

	if s, ok := field.Interface().(*null.String); ok && s != nil && s.Valid {
		return s.String
	}

	return emptyStr
}