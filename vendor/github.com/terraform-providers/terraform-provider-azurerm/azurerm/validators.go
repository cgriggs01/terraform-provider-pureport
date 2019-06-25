package azurerm

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/satori/uuid"
)

func validateRFC3339Date(v interface{}, k string) (warnings []string, errors []error) {
	dateString := v.(string)

	if _, err := date.ParseTime(time.RFC3339, dateString); err != nil {
		errors = append(errors, fmt.Errorf("%q is an invalid RFC3339 date: %+v", k, err))
	}

	return warnings, errors
}

func validateUUID(v interface{}, k string) (warnings []string, errors []error) {
	if _, err := uuid.FromString(v.(string)); err != nil {
		errors = append(errors, fmt.Errorf("%q is an invalid UUUID: %s", k, err))
	}
	return warnings, errors
}

func evaluateSchemaValidateFunc(i interface{}, k string, validateFunc schema.SchemaValidateFunc) (bool, error) { // nolint: unparam
	_, errors := validateFunc(i, k)

	errorStrings := []string{}
	for _, e := range errors {
		errorStrings = append(errorStrings, e.Error())
	}

	if len(errors) > 0 {
		return false, fmt.Errorf(strings.Join(errorStrings, "\n"))
	}

	return true, nil
}

func validateIso8601Duration() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		matched, _ := regexp.MatchString(`^P([0-9]+Y)?([0-9]+M)?([0-9]+W)?([0-9]+D)?(T([0-9]+H)?([0-9]+M)?([0-9]+(\.?[0-9]+)?S)?)?$`, v)

		if !matched {
			errors = append(errors, fmt.Errorf("expected %s to be in ISO 8601 duration format, got %s", k, v))
		}
		return warnings, errors
	}
}

func validateAzureVirtualMachineTimeZone() schema.SchemaValidateFunc {
	// Candidates are listed here: http://jackstromberg.com/2017/01/list-of-time-zones-consumed-by-azure/
	candidates := []string{
		"",
		"Afghanistan Standard Time",
		"Alaskan Standard Time",
		"Arab Standard Time",
		"Arabian Standard Time",
		"Arabic Standard Time",
		"Argentina Standard Time",
		"Atlantic Standard Time",
		"AUS Central Standard Time",
		"AUS Eastern Standard Time",
		"Azerbaijan Standard Time",
		"Azores Standard Time",
		"Bahia Standard Time",
		"Bangladesh Standard Time",
		"Belarus Standard Time",
		"Canada Central Standard Time",
		"Cape Verde Standard Time",
		"Caucasus Standard Time",
		"Cen. Australia Standard Time",
		"Central America Standard Time",
		"Central Asia Standard Time",
		"Central Brazilian Standard Time",
		"Central Europe Standard Time",
		"Central European Standard Time",
		"Central Pacific Standard Time",
		"Central Standard Time (Mexico)",
		"Central Standard Time",
		"China Standard Time",
		"Dateline Standard Time",
		"E. Africa Standard Time",
		"E. Australia Standard Time",
		"E. Europe Standard Time",
		"E. South America Standard Time",
		"Eastern Standard Time (Mexico)",
		"Eastern Standard Time",
		"Egypt Standard Time",
		"Ekaterinburg Standard Time",
		"Fiji Standard Time",
		"FLE Standard Time",
		"Georgian Standard Time",
		"GMT Standard Time",
		"Greenland Standard Time",
		"Greenwich Standard Time",
		"GTB Standard Time",
		"Hawaiian Standard Time",
		"India Standard Time",
		"Iran Standard Time",
		"Israel Standard Time",
		"Jordan Standard Time",
		"Kaliningrad Standard Time",
		"Korea Standard Time",
		"Libya Standard Time",
		"Line Islands Standard Time",
		"Magadan Standard Time",
		"Mauritius Standard Time",
		"Middle East Standard Time",
		"Montevideo Standard Time",
		"Morocco Standard Time",
		"Mountain Standard Time (Mexico)",
		"Mountain Standard Time",
		"Myanmar Standard Time",
		"N. Central Asia Standard Time",
		"Namibia Standard Time",
		"Nepal Standard Time",
		"New Zealand Standard Time",
		"Newfoundland Standard Time",
		"North Asia East Standard Time",
		"North Asia Standard Time",
		"Pacific SA Standard Time",
		"Pacific Standard Time (Mexico)",
		"Pacific Standard Time",
		"Pakistan Standard Time",
		"Paraguay Standard Time",
		"Romance Standard Time",
		"Russia Time Zone 10",
		"Russia Time Zone 11",
		"Russia Time Zone 3",
		"Russian Standard Time",
		"SA Eastern Standard Time",
		"SA Pacific Standard Time",
		"SA Western Standard Time",
		"Samoa Standard Time",
		"SE Asia Standard Time",
		"Singapore Standard Time",
		"South Africa Standard Time",
		"Sri Lanka Standard Time",
		"Syria Standard Time",
		"Taipei Standard Time",
		"Tasmania Standard Time",
		"Tokyo Standard Time",
		"Tonga Standard Time",
		"Turkey Standard Time",
		"Ulaanbaatar Standard Time",
		"US Eastern Standard Time",
		"US Mountain Standard Time",
		"UTC",
		"UTC+12",
		"UTC-02",
		"UTC-11",
		"Venezuela Standard Time",
		"Vladivostok Standard Time",
		"W. Australia Standard Time",
		"W. Central Africa Standard Time",
		"W. Europe Standard Time",
		"West Asia Standard Time",
		"West Pacific Standard Time",
		"Yakutsk Standard Time",
	}
	return validation.StringInSlice(candidates, true)
}

func validateCollation() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		matched, _ := regexp.MatchString(`^[A-Za-z0-9_. ]+$`, v)

		if !matched {
			errors = append(errors, fmt.Errorf("%s contains invalid characters, only underscores are supported, got %s", k, v))
			return
		}

		return warnings, errors
	}
}

func validateFilePath() schema.SchemaValidateFunc {
	return func(v interface{}, k string) (warnings []string, errors []error) {
		val := v.(string)

		if !strings.HasPrefix(val, "/") {
			errors = append(errors, fmt.Errorf("%q must start with `/`", k))
		}

		return warnings, errors
	}
}
