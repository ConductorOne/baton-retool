// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: c1/connector/v2/config.proto

package v2

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on SchemaServiceGetSchemaRequest with the
// rules defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *SchemaServiceGetSchemaRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on SchemaServiceGetSchemaRequest with
// the rules defined in the proto definition for this message. If any rules
// are violated, the result is a list of violation errors wrapped in
// SchemaServiceGetSchemaRequestMultiError, or nil if none found.
func (m *SchemaServiceGetSchemaRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *SchemaServiceGetSchemaRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return SchemaServiceGetSchemaRequestMultiError(errors)
	}

	return nil
}

// SchemaServiceGetSchemaRequestMultiError is an error wrapping multiple
// validation errors returned by SchemaServiceGetSchemaRequest.ValidateAll()
// if the designated constraints aren't met.
type SchemaServiceGetSchemaRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m SchemaServiceGetSchemaRequestMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m SchemaServiceGetSchemaRequestMultiError) AllErrors() []error { return m }

// SchemaServiceGetSchemaRequestValidationError is the validation error
// returned by SchemaServiceGetSchemaRequest.Validate if the designated
// constraints aren't met.
type SchemaServiceGetSchemaRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SchemaServiceGetSchemaRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SchemaServiceGetSchemaRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SchemaServiceGetSchemaRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SchemaServiceGetSchemaRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SchemaServiceGetSchemaRequestValidationError) ErrorName() string {
	return "SchemaServiceGetSchemaRequestValidationError"
}

// Error satisfies the builtin error interface
func (e SchemaServiceGetSchemaRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSchemaServiceGetSchemaRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SchemaServiceGetSchemaRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SchemaServiceGetSchemaRequestValidationError{}

// Validate checks the field values on SchemaServiceGetSchemaResponse with the
// rules defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *SchemaServiceGetSchemaResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on SchemaServiceGetSchemaResponse with
// the rules defined in the proto definition for this message. If any rules
// are violated, the result is a list of violation errors wrapped in
// SchemaServiceGetSchemaResponseMultiError, or nil if none found.
func (m *SchemaServiceGetSchemaResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *SchemaServiceGetSchemaResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Version

	if all {
		switch v := interface{}(m.GetSchema()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, SchemaServiceGetSchemaResponseValidationError{
					field:  "Schema",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, SchemaServiceGetSchemaResponseValidationError{
					field:  "Schema",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetSchema()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return SchemaServiceGetSchemaResponseValidationError{
				field:  "Schema",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return SchemaServiceGetSchemaResponseMultiError(errors)
	}

	return nil
}

// SchemaServiceGetSchemaResponseMultiError is an error wrapping multiple
// validation errors returned by SchemaServiceGetSchemaResponse.ValidateAll()
// if the designated constraints aren't met.
type SchemaServiceGetSchemaResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m SchemaServiceGetSchemaResponseMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m SchemaServiceGetSchemaResponseMultiError) AllErrors() []error { return m }

// SchemaServiceGetSchemaResponseValidationError is the validation error
// returned by SchemaServiceGetSchemaResponse.Validate if the designated
// constraints aren't met.
type SchemaServiceGetSchemaResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SchemaServiceGetSchemaResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SchemaServiceGetSchemaResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SchemaServiceGetSchemaResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SchemaServiceGetSchemaResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SchemaServiceGetSchemaResponseValidationError) ErrorName() string {
	return "SchemaServiceGetSchemaResponseValidationError"
}

// Error satisfies the builtin error interface
func (e SchemaServiceGetSchemaResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSchemaServiceGetSchemaResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SchemaServiceGetSchemaResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SchemaServiceGetSchemaResponseValidationError{}

// Validate checks the field values on ConfigSchema with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ConfigSchema) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ConfigSchema with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ConfigSchemaMultiError, or
// nil if none found.
func (m *ConfigSchema) ValidateAll() error {
	return m.validate(true)
}

func (m *ConfigSchema) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	for idx, item := range m.GetFields() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ConfigSchemaValidationError{
						field:  fmt.Sprintf("Fields[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ConfigSchemaValidationError{
						field:  fmt.Sprintf("Fields[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ConfigSchemaValidationError{
					field:  fmt.Sprintf("Fields[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	// no validation rules for DisplayName

	// no validation rules for HelpUrl

	if all {
		switch v := interface{}(m.GetIcon()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ConfigSchemaValidationError{
					field:  "Icon",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ConfigSchemaValidationError{
					field:  "Icon",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetIcon()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ConfigSchemaValidationError{
				field:  "Icon",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if all {
		switch v := interface{}(m.GetLogo()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ConfigSchemaValidationError{
					field:  "Logo",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ConfigSchemaValidationError{
					field:  "Logo",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetLogo()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ConfigSchemaValidationError{
				field:  "Logo",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return ConfigSchemaMultiError(errors)
	}

	return nil
}

// ConfigSchemaMultiError is an error wrapping multiple validation errors
// returned by ConfigSchema.ValidateAll() if the designated constraints aren't met.
type ConfigSchemaMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ConfigSchemaMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ConfigSchemaMultiError) AllErrors() []error { return m }

// ConfigSchemaValidationError is the validation error returned by
// ConfigSchema.Validate if the designated constraints aren't met.
type ConfigSchemaValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ConfigSchemaValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ConfigSchemaValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ConfigSchemaValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ConfigSchemaValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ConfigSchemaValidationError) ErrorName() string { return "ConfigSchemaValidationError" }

// Error satisfies the builtin error interface
func (e ConfigSchemaValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sConfigSchema.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ConfigSchemaValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ConfigSchemaValidationError{}

// Validate checks the field values on Field with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Field) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Field with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in FieldMultiError, or nil if none found.
func (m *Field) ValidateAll() error {
	return m.validate(true)
}

func (m *Field) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Name

	// no validation rules for HelpUrl

	// no validation rules for DisplayName

	// no validation rules for Placeholder

	switch v := m.Field.(type) {
	case *Field_Str:
		if v == nil {
			err := FieldValidationError{
				field:  "Field",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if all {
			switch v := interface{}(m.GetStr()).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, FieldValidationError{
						field:  "Str",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, FieldValidationError{
						field:  "Str",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(m.GetStr()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return FieldValidationError{
					field:  "Str",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *Field_Select:
		if v == nil {
			err := FieldValidationError{
				field:  "Field",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if all {
			switch v := interface{}(m.GetSelect()).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, FieldValidationError{
						field:  "Select",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, FieldValidationError{
						field:  "Select",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(m.GetSelect()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return FieldValidationError{
					field:  "Select",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *Field_Random:
		if v == nil {
			err := FieldValidationError{
				field:  "Field",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if all {
			switch v := interface{}(m.GetRandom()).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, FieldValidationError{
						field:  "Random",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, FieldValidationError{
						field:  "Random",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(m.GetRandom()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return FieldValidationError{
					field:  "Random",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *Field_File:
		if v == nil {
			err := FieldValidationError{
				field:  "Field",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if all {
			switch v := interface{}(m.GetFile()).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, FieldValidationError{
						field:  "File",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, FieldValidationError{
						field:  "File",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(m.GetFile()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return FieldValidationError{
					field:  "File",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	default:
		_ = v // ensures v is used
	}

	if len(errors) > 0 {
		return FieldMultiError(errors)
	}

	return nil
}

// FieldMultiError is an error wrapping multiple validation errors returned by
// Field.ValidateAll() if the designated constraints aren't met.
type FieldMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m FieldMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m FieldMultiError) AllErrors() []error { return m }

// FieldValidationError is the validation error returned by Field.Validate if
// the designated constraints aren't met.
type FieldValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e FieldValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e FieldValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e FieldValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e FieldValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e FieldValidationError) ErrorName() string { return "FieldValidationError" }

// Error satisfies the builtin error interface
func (e FieldValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sField.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = FieldValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = FieldValidationError{}

// Validate checks the field values on StringField with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringField) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringField with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringFieldMultiError, or
// nil if none found.
func (m *StringField) ValidateAll() error {
	return m.validate(true)
}

func (m *StringField) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Secret

	if all {
		switch v := interface{}(m.GetValueValidator()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, StringFieldValidationError{
					field:  "ValueValidator",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, StringFieldValidationError{
					field:  "ValueValidator",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetValueValidator()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return StringFieldValidationError{
				field:  "ValueValidator",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return StringFieldMultiError(errors)
	}

	return nil
}

// StringFieldMultiError is an error wrapping multiple validation errors
// returned by StringField.ValidateAll() if the designated constraints aren't met.
type StringFieldMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m StringFieldMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m StringFieldMultiError) AllErrors() []error { return m }

// StringFieldValidationError is the validation error returned by
// StringField.Validate if the designated constraints aren't met.
type StringFieldValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e StringFieldValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e StringFieldValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e StringFieldValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e StringFieldValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e StringFieldValidationError) ErrorName() string { return "StringFieldValidationError" }

// Error satisfies the builtin error interface
func (e StringFieldValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sStringField.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = StringFieldValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = StringFieldValidationError{}

// Validate checks the field values on SelectField with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *SelectField) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on SelectField with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in SelectFieldMultiError, or
// nil if none found.
func (m *SelectField) ValidateAll() error {
	return m.validate(true)
}

func (m *SelectField) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	for idx, item := range m.GetItems() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, SelectFieldValidationError{
						field:  fmt.Sprintf("Items[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, SelectFieldValidationError{
						field:  fmt.Sprintf("Items[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return SelectFieldValidationError{
					field:  fmt.Sprintf("Items[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return SelectFieldMultiError(errors)
	}

	return nil
}

// SelectFieldMultiError is an error wrapping multiple validation errors
// returned by SelectField.ValidateAll() if the designated constraints aren't met.
type SelectFieldMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m SelectFieldMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m SelectFieldMultiError) AllErrors() []error { return m }

// SelectFieldValidationError is the validation error returned by
// SelectField.Validate if the designated constraints aren't met.
type SelectFieldValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SelectFieldValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SelectFieldValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SelectFieldValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SelectFieldValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SelectFieldValidationError) ErrorName() string { return "SelectFieldValidationError" }

// Error satisfies the builtin error interface
func (e SelectFieldValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSelectField.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SelectFieldValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SelectFieldValidationError{}

// Validate checks the field values on RandomStringField with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *RandomStringField) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on RandomStringField with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// RandomStringFieldMultiError, or nil if none found.
func (m *RandomStringField) ValidateAll() error {
	return m.validate(true)
}

func (m *RandomStringField) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Length

	if len(errors) > 0 {
		return RandomStringFieldMultiError(errors)
	}

	return nil
}

// RandomStringFieldMultiError is an error wrapping multiple validation errors
// returned by RandomStringField.ValidateAll() if the designated constraints
// aren't met.
type RandomStringFieldMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m RandomStringFieldMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m RandomStringFieldMultiError) AllErrors() []error { return m }

// RandomStringFieldValidationError is the validation error returned by
// RandomStringField.Validate if the designated constraints aren't met.
type RandomStringFieldValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e RandomStringFieldValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e RandomStringFieldValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e RandomStringFieldValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e RandomStringFieldValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e RandomStringFieldValidationError) ErrorName() string {
	return "RandomStringFieldValidationError"
}

// Error satisfies the builtin error interface
func (e RandomStringFieldValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sRandomStringField.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = RandomStringFieldValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = RandomStringFieldValidationError{}

// Validate checks the field values on FileField with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *FileField) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on FileField with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in FileFieldMultiError, or nil
// if none found.
func (m *FileField) ValidateAll() error {
	return m.validate(true)
}

func (m *FileField) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Secret

	if all {
		switch v := interface{}(m.GetValueValidator()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, FileFieldValidationError{
					field:  "ValueValidator",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, FileFieldValidationError{
					field:  "ValueValidator",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetValueValidator()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return FileFieldValidationError{
				field:  "ValueValidator",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return FileFieldMultiError(errors)
	}

	return nil
}

// FileFieldMultiError is an error wrapping multiple validation errors returned
// by FileField.ValidateAll() if the designated constraints aren't met.
type FileFieldMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m FileFieldMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m FileFieldMultiError) AllErrors() []error { return m }

// FileFieldValidationError is the validation error returned by
// FileField.Validate if the designated constraints aren't met.
type FileFieldValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e FileFieldValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e FileFieldValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e FileFieldValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e FileFieldValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e FileFieldValidationError) ErrorName() string { return "FileFieldValidationError" }

// Error satisfies the builtin error interface
func (e FileFieldValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sFileField.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = FileFieldValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = FileFieldValidationError{}

// Validate checks the field values on SelectField_Item with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *SelectField_Item) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on SelectField_Item with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// SelectField_ItemMultiError, or nil if none found.
func (m *SelectField_Item) ValidateAll() error {
	return m.validate(true)
}

func (m *SelectField_Item) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for DisplayName

	// no validation rules for Value

	if len(errors) > 0 {
		return SelectField_ItemMultiError(errors)
	}

	return nil
}

// SelectField_ItemMultiError is an error wrapping multiple validation errors
// returned by SelectField_Item.ValidateAll() if the designated constraints
// aren't met.
type SelectField_ItemMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m SelectField_ItemMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m SelectField_ItemMultiError) AllErrors() []error { return m }

// SelectField_ItemValidationError is the validation error returned by
// SelectField_Item.Validate if the designated constraints aren't met.
type SelectField_ItemValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SelectField_ItemValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SelectField_ItemValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SelectField_ItemValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SelectField_ItemValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SelectField_ItemValidationError) ErrorName() string { return "SelectField_ItemValidationError" }

// Error satisfies the builtin error interface
func (e SelectField_ItemValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSelectField_Item.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SelectField_ItemValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SelectField_ItemValidationError{}
