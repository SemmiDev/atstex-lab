package validate

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	id_translations "github.com/go-playground/validator/v10/translations/id"
	"github.com/google/uuid"
)

var (
	instance   *validator.Validate
	translator ut.Translator
	once       sync.Once
)

type Config struct {
	UseIndonesianTranslations bool
}

func Init() {
	InitWithConfig(Config{UseIndonesianTranslations: true})
}

func InitWithConfig(cfg Config) {
	once.Do(func() {
		instance = validator.New()
		instance.RegisterTagNameFunc(jsonFieldName)

		if cfg.UseIndonesianTranslations {
			setupIndonesianTranslations()
		}

		registerBuiltinValidations()

		if cfg.UseIndonesianTranslations {
			registerBuiltinTranslations()
			overrideDefaultTranslations()
		}
	})
}

func Validator() *validator.Validate {
	Init()
	return instance
}

func Translator() ut.Translator {
	Init()
	return translator
}

func Struct(obj interface{}) map[string]string {
	v := Validator()
	t := Translator()

	if err := v.Struct(obj); err != nil {
		return FormatErrors(err, t)
	}
	return nil
}

func StructWithMessage(obj interface{}, customMessages map[string]string) map[string]string {
	errs := Struct(obj)
	if errs == nil || len(customMessages) == 0 {
		return errs
	}
	for field, msg := range customMessages {
		if _, exists := errs[field]; exists {
			errs[field] = msg
		}
	}
	return errs
}

func RegisterCustom(tag string, fn validator.Func, message string) {
	v := Validator()
	_ = v.RegisterValidation(tag, fn)
	registerTranslation(tag, message)
}

func registerBuiltinValidations() {
	_ = instance.RegisterValidation("uuid_any", validateUUIDAnyCase)
	_ = instance.RegisterValidation("base64_pdf", validateBase64PDF)
	_ = instance.RegisterValidation("safe_text", validateSafeText)
	_ = instance.RegisterValidation("base64", validateBase64)
	_ = instance.RegisterValidation("base64_image", validateBase64Image)
}

func validateUUIDAnyCase(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	if v == "" {
		return true
	}
	_, err := uuid.Parse(v)
	return err == nil
}

func validateBase64PDF(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return true
	}
	const prefix = "data:application/pdf;base64,"
	if !strings.HasPrefix(s, prefix) {
		return false
	}
	data := s[len(prefix):]
	if len(data) == 0 {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(data)
	return err == nil
}

var safeTextRegex = regexp.MustCompile(`^[a-zA-Z0-9 .,_\-\(\)]+$`)

func validateSafeText(fl validator.FieldLevel) bool {
	f := fl.Field()
	if f.Kind() != reflect.String {
		return false
	}
	v := f.String()
	if v == "" {
		return true
	}
	return safeTextRegex.MatchString(v)
}

func validateBase64(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return true
	}
	if !strings.HasPrefix(s, "data:") || !strings.Contains(s, ";base64,") {
		return false
	}
	parts := strings.SplitN(s, ";base64,", 2)
	if len(parts) != 2 || len(parts[1]) == 0 {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(parts[1])
	return err == nil
}

func validateBase64Image(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return true
	}
	if !strings.HasPrefix(s, "data:image/") || !strings.Contains(s, ";base64,") {
		return false
	}
	parts := strings.SplitN(s, ";base64,", 2)
	if len(parts) != 2 || len(parts[1]) == 0 {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(parts[1])
	return err == nil
}

func setupIndonesianTranslations() {
	loc := id.New()
	uni := ut.New(loc, loc)
	translator, _ = uni.GetTranslator("id")
	_ = id_translations.RegisterDefaultTranslations(instance, translator)
}

func registerBuiltinTranslations() {
	translations := map[string]string{
		"uuid_any":     "Format UUID tidak valid",
		"base64_pdf":   "Format file harus PDF (Base64)",
		"safe_text":    "Hanya boleh berisi huruf, angka, spasi, titik, koma, strip, dan underscore",
		"base64":       "Format Base64 tidak valid",
		"base64_image": "Format gambar tidak valid",
	}
	for tag, msg := range translations {
		registerTranslation(tag, msg)
	}
}

func overrideDefaultTranslations() {
	overrides := map[string]string{
		"required": "Wajib diisi",
		"email":    "Format email tidak valid",
		"numeric":  "Harus berupa angka",
		"min":      "Minimal {0} karakter/digit",
		"max":      "Maksimal {0} karakter/digit",
		"len":      "Harus tepat {0} karakter/digit",
		"gte":      "Harus lebih besar atau sama dengan {0}",
		"lte":      "Harus lebih kecil atau sama dengan {0}",
		"gt":       "Harus lebih besar dari {0}",
		"lt":       "Harus lebih kecil dari {0}",
		"alpha":    "Hanya boleh berisi huruf",
		"alphanum": "Hanya boleh berisi huruf dan angka",
		"url":      "Format URL tidak valid",
		"oneof":    "Harus salah satu dari {0}",
	}
	for tag, msg := range overrides {
		registerTranslation(tag, msg)
	}
}

func registerTranslation(tag, message string) {
	_ = instance.RegisterTranslation(tag, translator,
		func(ut ut.Translator) error {
			return ut.Add(tag, message, true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			msg := message
			if fe.Param() != "" {
				msg = strings.Replace(msg, "{0}", fe.Param(), -1)
			}
			return msg
		},
	)
}

func FormatErrors(err error, trans ut.Translator) map[string]string {
	result := make(map[string]string)
	valErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return result
	}

	for _, fe := range valErrs {
		fieldPath := cleanFieldPath(fe.Namespace())
		if _, exists := result[fieldPath]; exists {
			continue // keep the first error per field
		}
		result[fieldPath] = errorMessage(fe, trans)
	}

	return result
}

func errorMessage(fe validator.FieldError, trans ut.Translator) string {
	if trans != nil {
		if msg := fe.Translate(trans); msg != "" {
			return msg
		}
	}
	return fallbackMessage(fe)
}

func fallbackMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "Wajib diisi"
	case "email":
		return "Format email tidak valid"
	case "uuid_any":
		return "Format UUID tidak valid"
	case "numeric":
		return "Harus berupa angka"
	case "base64_pdf":
		return "Format file harus PDF (Base64)"
	case "min":
		return fmt.Sprintf("Minimal %s karakter/digit", fe.Param())
	case "max":
		return fmt.Sprintf("Maksimal %s karakter/digit", fe.Param())
	case "len":
		return fmt.Sprintf("Harus tepat %s karakter/digit", fe.Param())
	case "gte":
		return fmt.Sprintf("Harus lebih besar atau sama dengan %s", fe.Param())
	case "lte":
		return fmt.Sprintf("Harus lebih kecil atau sama dengan %s", fe.Param())
	case "gt":
		return fmt.Sprintf("Harus lebih besar dari %s", fe.Param())
	case "lt":
		return fmt.Sprintf("Harus lebih kecil dari %s", fe.Param())
	case "alpha":
		return "Hanya boleh berisi huruf"
	case "alphanum":
		return "Hanya boleh berisi huruf dan angka"
	case "url":
		return "Format URL tidak valid"
	case "safe_text":
		return "Hanya boleh berisi huruf, angka, spasi, titik, koma, strip, dan underscore"
	case "base64":
		return "Format Base64 tidak valid"
	case "base64_image":
		return "Format gambar tidak valid"
	default:
		return fmt.Sprintf("Validasi gagal untuk aturan '%s'", fe.Tag())
	}
}

func cleanFieldPath(namespace string) string {
	parts := strings.SplitN(namespace, ".", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return namespace
}

func jsonFieldName(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}
