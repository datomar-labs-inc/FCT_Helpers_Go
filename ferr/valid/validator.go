package valid

import (
	"context"
	"github.com/datomar-labs-inc/FCT_Helpers_Go/maybe"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
	en2 "github.com/go-playground/validator/v10/translations/en"
	"github.com/volatiletech/null/v8"
	"reflect"
	"regexp"
)

var validate *validator.Validate
var UniversalTranslator ut.Translator

var simpleTextRegex *regexp.Regexp

func init() {
	ent := en.New()

	var uni = ut.New(ent, ent)

	trans, _ := uni.GetTranslator("en")

	UniversalTranslator = trans

	validate = validator.New()

	err := validate.RegisterValidation("simpletext", validateSimpleText, false)
	if err != nil {
		panic(err)
	}

	err = validate.RegisterValidation("isempty", validateStringIsEmpty, false)
	if err != nil {
		panic(err)
	}

	simpleTextRegex = regexp.MustCompile(`^[\d\sa-zA-Z\-._]+$`)

	// register all sql.Null* types to use the ValidateValuer CustomTypeFunc
	validate.RegisterCustomTypeFunc(ValidateNullString, null.String{}, &null.String{})

	// TODO register all maybe.Maybe types to use the ValidateValuer CustomTypeFunc
	validate.RegisterCustomTypeFunc(maybe.ValidateValuer, maybe.Maybe[string]{}, &maybe.Maybe[int]{})

	_ = validate.RegisterValidation("notblank", validators.NotBlank)

	_ = en2.RegisterDefaultTranslations(validate, trans)
}

func ValidateStruct(ctx context.Context, s any) error {
	return validate.StructCtx(ctx, s)
}

func ValidateNullString(field reflect.Value) any {
	var emptyStr *string

	if s, ok := field.Interface().(null.String); ok && s.Valid {
		return s.String
	}

	if s, ok := field.Interface().(*null.String); ok && s != nil && s.Valid {
		return s.String
	}

	return emptyStr
}

func validateSimpleText(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	return simpleTextRegex.MatchString(str)
}

func validateSimpleTextOrEmpty(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if str == "" {
		return true
	}
	return simpleTextRegex.MatchString(str)
}

func validateStringIsEmpty(fl validator.FieldLevel) bool {
	return fl.Field().String() == ""
}
