package internal

/*func TestValidateType(t *testing.T) {
	var tests = []struct {
		name   string
		input  *TypeStmt
		errors *ErrorCollection
	}{
		{
			name: "rule 2 success on struct or union if fields index are unique",
			input: &TypeStmt{
				Name: &IdentifierStmt{Lit: "A"},
				Fields: &[]*FieldStmt{
					{Index: &PrimitiveValueStmt{RawValue: int64(1), Primitive: Primitive_Int64}},
					{Index: &PrimitiveValueStmt{RawValue: int64(2), Primitive: Primitive_Int64}},
					{Index: &PrimitiveValueStmt{RawValue: int64(3), Primitive: Primitive_Int64}},
				},
			},
			errors: &ErrorCollection{},
		},
		{
			name: "rule 2 fails on struct or union if fields index are not unique",
			input: &TypeStmt{
				Name: &IdentifierStmt{Lit: "A"},
				Fields: &[]*FieldStmt{
					{Index: &PrimitiveValueStmt{RawValue: int64(1), Primitive: Primitive_Int64}},
					{Index: &PrimitiveValueStmt{RawValue: int64(2), Primitive: Primitive_Int64}},
					{Index: &PrimitiveValueStmt{RawValue: int64(1), Primitive: Primitive_Int64}},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> index 1 already defined for a field"),
			},
		},
		{
			name: "rule 2 fails on enum if fields index does not start with 0",
			input: &TypeStmt{
				Name:     &IdentifierStmt{Lit: "A"},
				Modifier: Token_Enum,
				Fields: &[]*FieldStmt{
					{Index: &PrimitiveValueStmt{RawValue: int64(1), Primitive: Primitive_Int64}},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> expected the first field in an enum to have the index 0, given index 1"),
			},
		},
		{
			name: "rule 2 fails on enum if fields index are not correlative",
			input: &TypeStmt{
				Name:     &IdentifierStmt{Lit: "A"},
				Modifier: Token_Enum,
				Fields: &[]*FieldStmt{
					{Index: &PrimitiveValueStmt{RawValue: int64(0), Primitive: Primitive_Int64}},
					{Index: &PrimitiveValueStmt{RawValue: int64(1), Primitive: Primitive_Int64}},
					{Index: &PrimitiveValueStmt{RawValue: int64(3), Primitive: Primitive_Int64}},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> field indexes in an enum must be correlative"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer(new(ScopeCollection))
			analyzer.skipFields = true
			analyzer.currentContext = &ResolvedContext{
				Owner: &Ast{File: &File{Pkg: "root"}},
			}
			analyzer.validateType(tt.input)
			require.Equal(t, tt.errors, analyzer.errors)
		})
	}
}

func TestValidateField(t *testing.T) {
	var tests = []struct {
		name           string
		input          *FieldStmt
		currentContext *ResolvedContext
		errors         *ErrorCollection
	}{
		{
			name: "rule 1.a success with exactly one type argument",
			input: &FieldStmt{
				ValueType: &ValueTypeStmt{
					Ident: &IdentifierStmt{Lit: "list"},
					TypeArguments: &[]*ValueTypeStmt{
						{Ident: &IdentifierStmt{Lit: "string"}},
					},
				},
			},
			errors: &ErrorCollection{},
		},
		{
			name: "rule 1.a fails if zero type argument is specified",
			input: &FieldStmt{
				ValueType: &ValueTypeStmt{
					Ident:         &IdentifierStmt{Lit: "list"},
					TypeArguments: &[]*ValueTypeStmt{},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> list expects exactly one type argument"),
			},
		},
		{
			name: "rule 1.a fails with more than one type argument",
			input: &FieldStmt{
				ValueType: &ValueTypeStmt{
					Ident: &IdentifierStmt{Lit: "list"},
					TypeArguments: &[]*ValueTypeStmt{
						{Ident: &IdentifierStmt{Lit: "string"}},
						{Ident: &IdentifierStmt{Lit: "int64"}},
					},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> list expects exactly one type argument"),
			},
		},
		{
			name: "rule 1.b success with exactly two type arguments",
			input: &FieldStmt{
				ValueType: &ValueTypeStmt{
					Ident: &IdentifierStmt{Lit: "map"},
					TypeArguments: &[]*ValueTypeStmt{
						{Ident: &IdentifierStmt{Lit: "string"}},
						{Ident: &IdentifierStmt{Lit: "int64"}},
					},
				},
			},
			errors: &ErrorCollection{},
		},
		{
			name: "rule 1.b fails with more than two type arguments",
			input: &FieldStmt{
				ValueType: &ValueTypeStmt{
					Ident: &IdentifierStmt{Lit: "map"},
					TypeArguments: &[]*ValueTypeStmt{
						{Ident: &IdentifierStmt{Lit: "string"}},
						{Ident: &IdentifierStmt{Lit: "int64"}},
						{Ident: &IdentifierStmt{Lit: "bool"}},
					},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> map expects exactly two type arguments"),
			},
		},
		{
			name: "rule 1.b fails with zero type arguments",
			input: &FieldStmt{
				ValueType: &ValueTypeStmt{
					Ident:         &IdentifierStmt{Lit: "map"},
					TypeArguments: &[]*ValueTypeStmt{},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> map expects exactly two type arguments"),
			},
		},
		{
			name: "rule 2 fails if default value does not match value type",
			input: &FieldStmt{
				ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "string"}},
				DefaultValue: &PrimitiveValueStmt{
					RawValue:  int64(25),
					Primitive: Primitive_Int64,
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> field's default value is not of type string, it is int64"),
			},
		},
		{
			name: "rule 2 success if default value matches value type",
			input: &FieldStmt{
				ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "int64"}},
				DefaultValue: &PrimitiveValueStmt{
					RawValue:  int64(25),
					Primitive: Primitive_Int64,
				},
			},
			errors: &ErrorCollection{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer(new(ScopeCollection))
			analyzer.validateField(tt.input, false, false)

			if tt.currentContext != nil {
				analyzer.currentContext = tt.currentContext
			}

			require.Equal(t, tt.errors, analyzer.errors)
		})
	}
}

func TestValidateMetadata(t *testing.T) {
	var tests = []struct {
		name   string
		input  *MapValueStmt
		errors *ErrorCollection
	}{
		{
			name: "rule 1 success",
			input: &MapValueStmt{
				{
					Key:   &PrimitiveValueStmt{RawValue: "string", Primitive: Primitive_String},
					Value: &PrimitiveValueStmt{RawValue: "string", Primitive: Primitive_String},
				},
			},
			errors: &ErrorCollection{},
		},
		{
			name: "rule 1 fails if not string",
			input: &MapValueStmt{
				{
					Key: &PrimitiveValueStmt{Primitive: Primitive_Int64},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> metadata map keys must be of type string"),
			},
		},
		{
			name: "rule 2 success if value is string, bool, int64 or float64",
			input: &MapValueStmt{
				{
					Key:   &PrimitiveValueStmt{RawValue: "1", Primitive: Primitive_String},
					Value: &PrimitiveValueStmt{Primitive: Primitive_String},
				},
				{
					Key:   &PrimitiveValueStmt{RawValue: "2", Primitive: Primitive_String},
					Value: &PrimitiveValueStmt{Primitive: Primitive_Bool},
				},
				{
					Key:   &PrimitiveValueStmt{RawValue: "3", Primitive: Primitive_String},
					Value: &PrimitiveValueStmt{Primitive: Primitive_Int64},
				},
				{
					Key:   &PrimitiveValueStmt{RawValue: "4", Primitive: Primitive_String},
					Value: &PrimitiveValueStmt{Primitive: Primitive_Float64},
				},
			},
			errors: &ErrorCollection{},
		},
		{
			name: "rule 2 fails if value is not string, bool, int64 or float64",
			input: &MapValueStmt{
				{
					Key:   &PrimitiveValueStmt{RawValue: "1", Primitive: Primitive_String},
					Value: &PrimitiveValueStmt{Primitive: Primitive_List},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> metadata map values must be one of the following types: string|bool|float64|int64, given: list"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer(new(ScopeCollection))
			analyzer.validateMetadata(tt.input)
			require.Equal(t, tt.errors, analyzer.errors)
		})
	}
}

func TestValidateMap(t *testing.T) {
	var tests = []struct {
		name   string
		input  *MapValueStmt
		errors *ErrorCollection
	}{
		{
			name: "rule 1 success",
			input: &MapValueStmt{
				{
					Key:   &PrimitiveValueStmt{RawValue: "string", Primitive: Primitive_String},
					Value: &PrimitiveValueStmt{RawValue: "string", Primitive: Primitive_String},
				},
			},
			errors: &ErrorCollection{},
		},
		{
			name: "rule 1 fails with list key",
			input: &MapValueStmt{
				{
					Key: &ListValueStmt{
						&PrimitiveValueStmt{Primitive: Primitive_String},
					},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> map's keys cannot be of type list, map, null or a custom type"),
			},
		},
		{
			name: "rule 1 fails with map key",
			input: &MapValueStmt{
				{
					Key: &MapValueStmt{
						{
							Key: &PrimitiveValueStmt{Primitive: Primitive_String},
						},
					},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> map's keys cannot be of type list, map, null or a custom type"),
			},
		},
		{
			name: "rule 1 fails with null key",
			input: &MapValueStmt{
				{
					Key: &PrimitiveValueStmt{Primitive: Primitive_Null},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> map's keys cannot be of type list, map, null or a custom type"),
			},
		},
		{
			name: "rule 2 with string key",
			input: &MapValueStmt{
				{
					Key: &PrimitiveValueStmt{RawValue: "a", Primitive: Primitive_String},
				},
				{
					Key: &PrimitiveValueStmt{RawValue: "a", Primitive: Primitive_String},
				},
			},
			errors: &ErrorCollection{
				errors.New(`[analyzer] 0:0 -> key "a" already exists in map`),
			},
		},
		{
			name: "rule 2 with int key",
			input: &MapValueStmt{
				{
					Key: &PrimitiveValueStmt{RawValue: int64(2), Primitive: Primitive_Int64},
				},
				{
					Key: &PrimitiveValueStmt{RawValue: int64(2), Primitive: Primitive_Int64},
				},
			},
			errors: &ErrorCollection{
				errors.New(`[analyzer] 0:0 -> key "2" already exists in map`),
			},
		},
		{
			name: "rule 2 with float key",
			input: &MapValueStmt{
				{
					Key: &PrimitiveValueStmt{RawValue: float64(2.5), Primitive: Primitive_Float64},
				},
				{
					Key: &PrimitiveValueStmt{RawValue: float64(2.5), Primitive: Primitive_Float64},
				},
			},
			errors: &ErrorCollection{
				errors.New(`[analyzer] 0:0 -> key "2.5" already exists in map`),
			},
		},
		{
			name: "rule 2 with bool key",
			input: &MapValueStmt{
				{
					Key: &PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool},
				},
				{
					Key: &PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool},
				},
			},
			errors: &ErrorCollection{
				errors.New(`[analyzer] 0:0 -> key "true" already exists in map`),
			},
		},
		{
			name: "rule 2 with enum key",
			input: &MapValueStmt{
				{
					Key: &TypeValueStmt{
						TypeName: &IdentifierStmt{Lit: "MyEnum"},
						RawValue: &IdentifierStmt{Lit: "red"},
					},
				},
				{
					Key: &TypeValueStmt{
						TypeName: &IdentifierStmt{Lit: "MyEnum"},
						RawValue: &IdentifierStmt{Lit: "red"},
					},
				},
			},
			errors: &ErrorCollection{
				errors.New(`[analyzer] 0:0 -> key "MyEnum.red" already exists in map`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer(new(ScopeCollection))
			analyzer.validateMap(tt.input)
			require.Equal(t, tt.errors, analyzer.errors)
		})
	}
}*/
