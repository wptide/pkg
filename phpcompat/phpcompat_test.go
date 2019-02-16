package phpcompat

import (
	"reflect"
	"testing"

	"github.com/wptide/pkg/tide"
)

func TestBreaksVersions(t *testing.T) {
	type args struct {
		message tide.PhpcsFilesMessage
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"PHPCompatibility.PHP.NewConstants.ill_illtrpFound",
			args{
				testMessages["PHPCompatibility.PHP.NewConstants.ill_illtrpFound"],
			},
			[]string{"5.2"},
		},
		{
			"PHPCompatibility.PHP.NewFunctions.random_bytesFound",
			args{
				testMessages["PHPCompatibility.PHP.NewFunctions.random_bytesFound"],
			},
			[]string{"5.2", "5.3", "5.4", "5.5", "5.6"},
		},
		{
			"PHPCompatibility.PHP.ForbiddenNames.constFound",
			args{
				testMessages["PHPCompatibility.PHP.ForbiddenNames.constFound"],
			},
			[]string{"5.2", "5.3", "5.4", "5.5", "5.6", "7.0", "7.1", "7.2"},
		},
		{
			"PHPCompatibility.PHP.DeprecatedFunctions.mysqli_send_long_dataDeprecatedRemoved",
			args{
				testMessages["PHPCompatibility.PHP.DeprecatedFunctions.mysqli_send_long_dataDeprecatedRemoved"],
			},
			[]string{"5.4", "5.5", "5.6", "7.0", "7.1", "7.2"},
		},
		{
			"PHPCompatibility.PHP.DeprecatedFunctions.mcrypt_cfbDeprecatedRemoved",
			args{
				testMessages["PHPCompatibility.PHP.DeprecatedFunctions.mcrypt_cfbDeprecatedRemoved"],
			},
			[]string{"7.0", "7.1", "7.2"},
		},
		{
			"PHPCompatibility.PHP.ForbiddenNames.cloneFound",
			args{
				testMessages["PHPCompatibility.PHP.ForbiddenNames.cloneFound"],
			},
			[]string{"5.2", "5.3", "5.4", "5.5", "5.6", "7.0", "7.1", "7.2"},
		},
		{
			"PHPCompatibility.PHP.DynamicAccessToStatic.Found",
			args{
				testMessages["PHPCompatibility.PHP.DynamicAccessToStatic.Found"],
			},
			[]string{"5.2"},
		},
		{
			"PHPCompatibility.PHP.ValidIntegers.HexNumericStringFound",
			args{
				testMessages["PHPCompatibility.PHP.ValidIntegers.HexNumericStringFound"],
			},
			[]string{"7.0", "7.1", "7.2"},
		},
		{
			"PHPCompatibility.PHP.ValidIntegers.InvalidOctalIntegerFound",
			args{
				testMessages["PHPCompatibility.PHP.ValidIntegers.InvalidOctalIntegerFound"],
			},
			[]string{"7.0", "7.1", "7.2"},
		},
		{
			"PHPCompatibility.PHP.LanguageConstructs.NewEmptyNonVariableFound",
			args{
				testMessages["PHPCompatibility.PHP.LanguageConstructs.NewEmptyNonVariableFound"],
			},
			[]string{"5.2", "5.3", "5.4"},
		},
		{
			"PHPCompatibility.PHP.EmptyNonVariable.Found",
			args{
				testMessages["PHPCompatibility.PHP.EmptyNonVariable.Found"],
			},
			[]string{"5.2", "5.3", "5.4"},
		},
		{
			"PHPCompatibility.PHP.TernaryOperators.MiddleMissing",
			args{
				testMessages["PHPCompatibility.PHP.TernaryOperators.MiddleMissing"],
			},
			[]string{"5.2"},
		},
		{
			"PHPCompatibility.PHP.NonStaticMagicMethods.__getMethodVisibility",
			args{
				testMessages["PHPCompatibility.PHP.NonStaticMagicMethods.__getMethodVisibility"],
			},
			[]string{"5.2", "5.3", "5.4", "5.5", "5.6", "7.0", "7.1", "7.2"},
		},
		{
			"PHPCompatibility.PHP.ShortArray.Found",
			args{
				testMessages["PHPCompatibility.PHP.ShortArray.Found"],
			},
			[]string{"5.2", "5.3"},
		},
		{
			"PHPCompatibility.PHP.ForbiddenSwitchWithMultipleDefaultBlocks.Found",
			args{
				testMessages["PHPCompatibility.PHP.ForbiddenSwitchWithMultipleDefaultBlocks.Found"],
			},
			[]string{"7.0", "7.1", "7.2"},
		},
		{
			// Warnings only, no breaks.
			"PHPCompatibility.PHP.DeprecatedFunctions.mcrypt_generic_deinitDeprecated",
			args{
				testMessages["PHPCompatibility.PHP.DeprecatedFunctions.mcrypt_generic_deinitDeprecated"],
			},
			nil,
		},
		{
			// Another Warning test.
			"PHPCompatibility.PHP.DeprecatedPHP4StyleConstructors.Found",
			args{
				testMessages["PHPCompatibility.PHP.DeprecatedPHP4StyleConstructors.Found"],
			},
			nil,
		},
		{
			"Unknown.Code",
			args{
				testMessages["Unknown.Code"],
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BreaksVersions(tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BreaksVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNonBreakingVersions(t *testing.T) {
	type args struct {
		message tide.PhpcsFilesMessage
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"PHPCompatibility.PHP.RemovedConstants.intl_idna_variant_2003Deprecated",
			args{
				testMessages["PHPCompatibility.PHP.RemovedConstants.intl_idna_variant_2003Deprecated"],
			},
			[]string{"7.2"},
		},
		{
			"PHPCompatibility.PHP.DeprecatedFunctions.mcrypt_generic_deinitDeprecated",
			args{
				testMessages["PHPCompatibility.PHP.DeprecatedFunctions.mcrypt_generic_deinitDeprecated"],
			},
			[]string{"7.1", "7.2"},
		},
		{
			"PHPCompatibility.PHP.DeprecatedPHP4StyleConstructors.Found",
			args{
				testMessages["PHPCompatibility.PHP.DeprecatedPHP4StyleConstructors.Found"],
			},
			[]string{"7.0", "7.1", "7.2"},
		},
		{
			"PHPCompatibility.PHP.ForbiddenNamesAsDeclared.resourceFound",
			args{
				testMessages["PHPCompatibility.PHP.ForbiddenNamesAsDeclared.resourceFound"],
			},
			[]string{"7.0", "7.1", "7.2"},
		},
		{
			"PHPCompatibility.PHP.FakeAvailableSinceWarning",
			args{
				testMessages["PHPCompatibility.PHP.FakeAvailableSinceWarning"],
			},
			[]string{"5.2", "5.3"},
		},
		{
			"PHPCompatibility.PHP.FakeSinceWarning",
			args{
				testMessages["PHPCompatibility.PHP.FakeSinceWarning"],
			},
			[]string{"7.0", "7.1", "7.2"},
		},
		{
			"PHPCompatibility.PHP.NewFunctions.random_bytesFound",
			args{
				testMessages["PHPCompatibility.PHP.NewFunctions.random_bytesFound"],
			},
			nil,
		},
		{
			"Unknown.Code",
			args{
				testMessages["Unknown.Code"],
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NonBreakingVersions(tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NonBreakingVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPreviousVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"5.5.2 -> 5.5.1",
			args{
				"5.5.2",
			},
			"5.5.1",
		},
		{
			"5.5.0 -> 5.4.45",
			args{
				"5.5.0",
			},
			"5.4.45",
		},
		{
			"5.5 -> 5.4.45",
			args{
				"5.5",
			},
			"5.4.45",
		},
		{
			"7.2.0 -> 7.1.13",
			args{
				"7.2.0",
			},
			"7.1.13",
		},
		{
			"7.1.0 -> 7.0.26",
			args{
				"7.1",
			},
			"7.0.26",
		},
		{
			"7.0.0 -> 5.6.32",
			args{
				"7.0.0",
			},
			"5.6.32",
		},
		{
			"5.6 -> 5.5.37",
			args{
				"5.6",
			},
			"5.5.37",
		},
		{
			"5.5 -> 5.4.45",
			args{
				"5.5",
			},
			"5.4.45",
		},
		{
			"5.4 -> 5.3.29",
			args{
				"5.4",
			},
			"5.3.29",
		},
		{
			"5.3 -> 5.2.17",
			args{
				"5.3",
			},
			"5.2.17",
		},
		{
			"5.2.0 -> 5.2.0",
			args{
				"5.2.0",
			},
			"5.2.0",
		},
		{
			"all -> all",
			args{
				"all",
			},
			"all",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PreviousVersion(tt.args.version); got != tt.want {
				t.Errorf("PreviousVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionParts(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
		want2 int
	}{
		{
			"all -> 0 0 0",
			args{
				"all",
			},
			0,
			0,
			0,
		},
		{
			"7.2.1 -> 7 2 1",
			args{
				"7.2.1",
			},
			7,
			2,
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := VersionParts(tt.args.version)
			if got != tt.want {
				t.Errorf("VersionParts() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("VersionParts() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("VersionParts() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestGetVersionParts(t *testing.T) {
	type args struct {
		version string
		lowIn   string
	}
	tests := []struct {
		name           string
		args           args
		wantLow        string
		wantHigh       string
		wantMajorMinor string
		wantReported   string
	}{
		{
			"all",
			args{
				"all",
				"",
			},
			"all",
			"all",
			"all",
			"all",
		},
		{
			"5.6.4",
			args{
				"5.6.4",
				"",
			},
			"5.6.0",
			"5.6.4",
			"5.6",
			"5.6.4",
		},
		{
			"5.6.4, 5.5.1",
			args{
				"5.6.4",
				"5.5.1",
			},
			"5.5.1",
			"5.6.4",
			"5.6",
			"5.6.4",
		},
		{
			"7.2",
			args{
				"7.2",
				"",
			},
			"7.2.0",
			"7.2.1",
			"7.2",
			"7.2",
		},
		{
			"5.4",
			args{
				"5.4",
				"5.2",
			},
			"5.2.0",
			"5.4.45",
			"5.4",
			"5.4",
		},
		{
			"5.0",
			args{
				"5.0",
				"",
			},
			"5.2.0",
			"5.2.0",
			"5.2",
			"5.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLow, gotHigh, gotMajorMinor, gotReported := GetVersionParts(tt.args.version, tt.args.lowIn)
			if gotLow != tt.wantLow {
				t.Errorf("GetVersionParts() gotLow = %v, want %v", gotLow, tt.wantLow)
			}
			if gotHigh != tt.wantHigh {
				t.Errorf("GetVersionParts() gotHigh = %v, want %v", gotHigh, tt.wantHigh)
			}
			if gotMajorMinor != tt.wantMajorMinor {
				t.Errorf("GetVersionParts() gotMajorMinor = %v, want %v", gotMajorMinor, tt.wantMajorMinor)
			}
			if gotReported != tt.wantReported {
				t.Errorf("GetVersionParts() gotReported = %v, want %v", gotReported, tt.wantReported)
			}
		})
	}
}

func TestPhpMajorVersions(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			"Get all major.minor versions",
			[]string{"5.2", "5.3", "5.4", "5.5", "5.6", "7.0", "7.1", "7.2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PhpMajorVersions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PhpMajorVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeVersions(t *testing.T) {
	type args struct {
		n [][]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"Merge [5.2,5.3] with [5.3,5.4,5.5]",
			args{
				[][]string{
					{"5.2", "5.3"},
					{"5.3", "5.4", "5.5"},
				},
			},
			[]string{"5.2", "5.3", "5.4", "5.5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeVersions(tt.args.n...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExcludeVersions(t *testing.T) {
	type args struct {
		versions []string
		exclude  []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"5.6",
			args{
				[]string{"5.2", "5.3", "5.4", "5.5", "5.6", "7.0", "7.1", "7.2"},
				[]string{"5.6"},
			},
			[]string{"5.2", "5.3", "5.4", "5.5", "7.0", "7.1", "7.2"},
		},
		{
			"5.2, 5.3, 7.1, 7.2",
			args{
				[]string{"5.2", "5.3", "5.4", "5.5", "5.6", "7.0", "7.1", "7.2"},
				[]string{"5.2", "5.3", "7.1", "7.2"},
			},
			[]string{"5.4", "5.5", "5.6", "7.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExcludeVersions(tt.args.versions, tt.args.exclude); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExcludeVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}
