package phpcompat

import (
	"reflect"
	"testing"

	"github.com/wptide/pkg/tide"
)

var (
	testMessages = map[string]tide.PhpcsFilesMessage{
		"PHPCompatibility.PHP.RemovedConstants.intl_idna_variant_2003Deprecated": {
			Message: `The constant "INTL_IDNA_VARIANT_2003" is deprecated since PHP 7.2`,
			Source:  "PHPCompatibility.PHP.RemovedConstants.intl_idna_variant_2003Deprecated",
			Type:    "WARNING",
		},
		"PHPCompatibility.PHP.NewConstants.ill_illtrpFound": {
			Message: `The constant "ILL_ILLTRP" is not present in PHP version 5.2 or earlier`,
			Source:  "PHPCompatibility.PHP.NewConstants.ill_illtrpFound",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.NewFunctions.random_bytesFound": {
			Message: "The function random_bytes() is not present in PHP version 5.6 or earlier",
			Source:  "PHPCompatibility.PHP.NewFunctions.random_bytesFound",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.ForbiddenNames.constFound": {
			Message: "Function name, class name, namespace name or constant name can not be reserved keyword 'const' (since version all)",
			Source:  "PHPCompatibility.PHP.ForbiddenNames.constFound",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.DeprecatedFunctions.mysqli_send_long_dataDeprecatedRemoved": {
			Message: "Function mysqli_send_long_data() is deprecated since PHP 5.3 and removed since PHP 5.4; Use mysqli_stmt::send_long_data() instead",
			Source:  "PHPCompatibility.PHP.DeprecatedFunctions.mysqli_send_long_dataDeprecatedRemoved",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.DeprecatedFunctions.mcrypt_cfbDeprecatedRemoved": {
			Message: "Function mcrypt_cfb() is deprecated since PHP 5.5 and removed since PHP 7.0",
			Source:  "PHPCompatibility.PHP.DeprecatedFunctions.mcrypt_cfbDeprecatedRemoved",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.ForbiddenNames.cloneFound": {
			Message: "Function name, class name, namespace name or constant name can not be reserved keyword 'clone' (since version 5.0)",
			Source:  "PHPCompatibility.PHP.ForbiddenNames.cloneFound",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.DeprecatedFunctions.mcrypt_generic_deinitDeprecated": {
			Message: "Function mcrypt_generic_deinit() is deprecated since PHP 7.1; Use OpenSSL instead",
			Source:  "PHPCompatibility.PHP.DeprecatedFunctions.mcrypt_generic_deinitDeprecated",
			Type:    "WARNING",
		},
		"PHPCompatibility.PHP.DynamicAccessToStatic.Found": {
			Message: "Static class properties and methods, as well as class constants, could not be accessed using a dynamic (variable) classname in PHP 5.2 or earlier.",
			Source:  "PHPCompatibility.PHP.DynamicAccessToStatic.Found",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.DeprecatedPHP4StyleConstructors.Found": {
			Message: "Use of deprecated PHP4 style class constructor is not supported since PHP 7.",
			Source:  "PHPCompatibility.PHP.DeprecatedPHP4StyleConstructors.Found",
			Type:    "WARNING",
		},
		"PHPCompatibility.PHP.ForbiddenNamesAsDeclared.resourceFound": {
			Message: "'resource' is a soft reserved keyword as of PHP version 7.0 and should not be used to name a class, interface or trait or as part of a namespace (T_CLASS)",
			Source:  "PHPCompatibility.PHP.ForbiddenNamesAsDeclared.resourceFound",
			Type:    "WARNING",
		},
		"PHPCompatibility.PHP.ValidIntegers.HexNumericStringFound": {
			Message: "The behaviour of hexadecimal numeric strings was inconsistent prior to PHP 7 and support has been removed in PHP 7. Found: '0xaa78b5'",
			Source:  "PHPCompatibility.PHP.ValidIntegers.HexNumericStringFound",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.ValidIntegers.InvalidOctalIntegerFound": {
			Message: "Invalid octal integer detected. Prior to PHP 7 this would lead to a truncated number. From PHP 7 onwards this causes a parse error. Found: 038",
			Source:  "PHPCompatibility.PHP.ValidIntegers.InvalidOctalIntegerFound",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.EmptyNonVariable.Found": {
			Message: "Only variables can be passed to empty() prior to PHP 5.5.",
			Source:  "PHPCompatibility.PHP.EmptyNonVariable.Found",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.TernaryOperators.MiddleMissing": {
			Message: "Middle may not be omitted from ternary operators in PHP < 5.3",
			Source:  "PHPCompatibility.PHP.TernaryOperators.MiddleMissing",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.NonStaticMagicMethods.__getMethodVisibility": {
			Message: "Visibility for magic method __get must be public. Found: private",
			Source:  "PHPCompatibility.PHP.NonStaticMagicMethods.__getMethodVisibility",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.ShortArray.Found": {
			Message: "Short array syntax (open) is available since 5.4",
			Source:  "PHPCompatibility.PHP.ShortArray.Found",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.FakeAvailableSinceWarning": {
			Message: "Future state to test is available since 5.4",
			Source:  "PHPCompatibility.PHP.FakeAvailableSinceWarning",
			Type:    "WARNING",
		},
		"PHPCompatibility.PHP.ForbiddenSwitchWithMultipleDefaultBlocks.Found": {
			Message: "Switch statements can not have multiple default blocks since PHP 7.0",
			Source:  "PHPCompatibility.PHP.ForbiddenSwitchWithMultipleDefaultBlocks.Found",
			Type:    "ERROR",
		},
		"PHPCompatibility.PHP.FakeSinceWarning": {
			Message: "Future state to test since PHP 7.0",
			Source:  "PHPCompatibility.PHP.FakeSinceWarning",
			Type:    "WARNING",
		},
		"Unknown.Code": {
			Source:  "Unknown.Code",
			Message: "Unknown",
		},
	}
)

func TestParse(t *testing.T) {

	warningMap := map[string]bool{
		"Unknown.Code": true,
	}

	for _, message := range testMessages {
		t.Run(message.Source, func(t *testing.T) {
			compat, err := Parse(message)
			if reflect.TypeOf(compat).String() != "phpcompat.Compatibility" {
				t.Errorf("phpcompat.Parse(): %v, want %v", reflect.TypeOf(compat), "phpcompat.Compatibility")
			}

			expectWarning, ok := warningMap[message.Source]
			if err != nil && !ok && expectWarning == false {
				t.Errorf("phpcompat.Parse(): %v", err)
			}
		})
	}

}
